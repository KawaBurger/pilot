package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kawaburger/pilot/internal/claude"
	"github.com/kawaburger/pilot/internal/tmux"
)

// claudeHome expands ~ in the configured Claude home path.
func (s *Server) claudeHome() string {
	home := s.Config.Claude.Home
	if strings.HasPrefix(home, "~/") {
		userHome, _ := os.UserHomeDir()
		home = filepath.Join(userHome, home[2:])
	}
	return home
}

// ListSessions returns all discovered Claude sessions.
func (s *Server) ListSessions(c *gin.Context) {
	sessions, err := claude.DiscoverSessions(s.claudeHome())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sessions)
}

// GetSession returns a single session by ID.
func (s *Server) GetSession(c *gin.Context) {
	session, err := claude.FindSession(s.claudeHome(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, session)
}

// GetMessages returns the conversation history for a session.
func (s *Server) GetMessages(c *gin.Context) {
	session, err := claude.FindSession(s.claudeHome(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	afterUUID := c.Query("after")
	w := claude.NewWatcher()
	msgs, err := w.ReadHistory(session.JSONLPath, afterUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, msgs)
}

type newSessionRequest struct {
	CWD    string `json:"cwd"`
	Prompt string `json:"prompt"`
}

// NewSession creates a tmux session, starts Claude, and waits for the session
// to appear in the Claude home directory.
func (s *Server) NewSession(c *gin.Context) {
	var req newSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	cwd := req.CWD
	if cwd == "" {
		cwd = s.Config.Defaults.CWD
	}
	// Expand ~ to home directory.
	if cwd == "~" {
		cwd, _ = os.UserHomeDir()
	} else if strings.HasPrefix(cwd, "~/") {
		userHome, _ := os.UserHomeDir()
		cwd = filepath.Join(userHome, cwd[2:])
	}
	// Ensure cwd exists.
	if _, err := os.Stat(cwd); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("directory does not exist: %s", cwd)})
		return
	}

	// Generate a short tmux session name.
	short := uuid.New().String()[:8]
	tmuxName := fmt.Sprintf("pilot-%s", short)

	// Snapshot existing active session IDs BEFORE starting claude.
	existingIDs := make(map[string]bool)
	if existing, err := claude.ListActiveSessions(s.claudeHome()); err == nil {
		for _, id := range existing {
			existingIDs[id] = true
		}
	}

	if err := tmux.NewSession(tmuxName, cwd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("create tmux session: %v", err)})
		return
	}

	if err := tmux.StartClaude(tmuxName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("start claude: %v", err)})
		return
	}

	// Poll for a new active session to appear (500ms interval, 30s timeout).
	// We check sessions/*.json directly because JSONL files are only created
	// after the first message, not when claude starts.
	log.Printf("[NewSession] snapshot has %d existing IDs, tmux=%s, cwd=%s", len(existingIDs), tmuxName, cwd)
	var sessionID string
	deadline := time.Now().Add(30 * time.Second)
	pollCount := 0
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		pollCount++

		// If tmux session died, stop waiting.
		if !tmux.HasSession(tmuxName) {
			log.Printf("[NewSession] tmux session %s died", tmuxName)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "claude process exited unexpectedly"})
			return
		}

		active, err := claude.ListActiveSessions(s.claudeHome())
		if err != nil {
			log.Printf("[NewSession] poll %d: ListActiveSessions error: %v", pollCount, err)
			continue
		}
		if pollCount <= 3 || pollCount%10 == 0 {
			log.Printf("[NewSession] poll %d: found %d active sessions, existing=%d", pollCount, len(active), len(existingIDs))
		}
		for _, id := range active {
			if !existingIDs[id] {
				sessionID = id
				break
			}
		}
		if sessionID != "" {
			break
		}
	}

	if sessionID == "" {
		// Clean up the tmux session we created since claude never started.
		_ = tmux.KillSession(tmuxName)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "timed out waiting for claude session"})
		return
	}

	// Send the initial prompt if provided.
	if req.Prompt != "" {
		if err := tmux.SendMessage(tmuxName, req.Prompt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("send prompt: %v", err)})
			return
		}
	}

	s.TmuxMap.Store(sessionID, tmuxName)

	c.JSON(http.StatusOK, gin.H{
		"sessionId":   sessionID,
		"tmuxSession": tmuxName,
	})
}

// ResumeSession resumes an existing Claude session in a new tmux session.
func (s *Server) ResumeSession(c *gin.Context) {
	sessionID := c.Param("id")
	session, err := claude.FindSession(s.claudeHome(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	short := uuid.New().String()[:8]
	tmuxName := fmt.Sprintf("pilot-%s", short)

	cwd := session.CWD
	if cwd == "" {
		cwd = s.Config.Defaults.CWD
	}

	if err := tmux.NewSession(tmuxName, cwd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("create tmux session: %v", err)})
		return
	}

	if err := tmux.ResumeClaude(tmuxName, sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("resume claude: %v", err)})
		return
	}

	s.TmuxMap.Store(sessionID, tmuxName)

	c.JSON(http.StatusOK, gin.H{
		"sessionId":   sessionID,
		"tmuxSession": tmuxName,
	})
}

// SendMessage sends a message to a running Claude session via tmux.
func (s *Server) SendMessage(c *gin.Context) {
	sessionID := c.Param("id")

	var req struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	tmuxName, err := s.lookupTmux(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := tmux.SendMessage(tmuxName, req.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("send message: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// InterruptSession sends an interrupt (Escape) to a running Claude session.
func (s *Server) InterruptSession(c *gin.Context) {
	sessionID := c.Param("id")

	tmuxName, err := s.lookupTmux(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := tmux.Interrupt(tmuxName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("interrupt: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// TerminalContent returns the current tmux pane content for a session.
func (s *Server) TerminalContent(c *gin.Context) {
	sessionID := c.Param("id")

	tmuxName, err := s.lookupTmux(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	content, err := tmux.CapturePaneDynamic(tmuxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("capture pane: %v", err)})
		return
	}

	// Detect if there's a prompt waiting for input.
	hasPrompt := strings.Contains(content, "Do you want to") ||
		strings.Contains(content, "Yes, I trust this folder") ||
		strings.Contains(content, "Esc to cancel")

	c.JSON(http.StatusOK, gin.H{
		"content":   content,
		"hasPrompt": hasPrompt,
	})
}

// SendKeys sends raw key sequences to a session's tmux (for confirming prompts).
func (s *Server) SendKeysHandler(c *gin.Context) {
	sessionID := c.Param("id")

	var req struct {
		Keys []string `json:"keys" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	tmuxName, err := s.lookupTmux(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := tmux.SendKeys(tmuxName, req.Keys...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("send keys: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// lookupTmux retrieves the tmux session name for a given Claude session ID.
func (s *Server) lookupTmux(sessionID string) (string, error) {
	val, ok := s.TmuxMap.Load(sessionID)
	if !ok {
		return "", fmt.Errorf("no tmux session for %s", sessionID)
	}
	return val.(string), nil
}
