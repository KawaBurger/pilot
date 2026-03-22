package claude

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

// Session represents a Claude Code conversation session.
type Session struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Project   string `json:"project"`
	CWD       string `json:"cwd"`
	GitBranch string `json:"gitBranch"`
	Status    string `json:"status"` // "active" or "ended"
	JSONLPath       string `json:"jsonlPath"`
	UpdatedAt       int64  `json:"updatedAt"`       // JSONL file mod time (unix seconds)
	LastUserMessage string `json:"lastUserMessage"` // last user message preview
}

// activeSession is the in-memory representation of a sessions/*.json file.
type activeSession struct {
	SessionID string `json:"sessionId"`
	CWD       string `json:"cwd"`
	StartedAt int64  `json:"startedAt"`
	PID       int    `json:"pid"`
}

// DiscoverSessions scans claudeHome for all JSONL conversation files and returns
// Session metadata for each. Active sessions are listed first, then sorted by ID
// descending (newest first).
func DiscoverSessions(claudeHome string) ([]Session, error) {
	activeMap, err := loadActiveSessions(claudeHome)
	if err != nil {
		// Non-fatal: we just won't know which sessions are active.
		activeMap = make(map[string]bool)
	}

	pattern := filepath.Join(claudeHome, "projects", "*", "*.jsonl")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob sessions: %w", err)
	}

	var sessions []Session
	for _, path := range matches {
		id := strings.TrimSuffix(filepath.Base(path), ".jsonl")
		project := filepath.Base(filepath.Dir(path))

		cwd, gitBranch := extractMetadata(path)
		title := loadTitle(claudeHome, id)
		if title == "" {
			title = extractFirstUserMessage(path)
		}

		status := "ended"
		if activeMap[id] {
			status = "active"
		}

		var modTime int64
		if info, err := os.Stat(path); err == nil {
			modTime = info.ModTime().Unix()
		}

		sessions = append(sessions, Session{
			ID:              id,
			Title:           title,
			Project:         project,
			CWD:             cwd,
			GitBranch:       gitBranch,
			Status:          status,
			JSONLPath:       path,
			LastUserMessage: extractLastUserMessage(path),
			UpdatedAt: modTime,
		})
	}

	sort.Slice(sessions, func(i, j int) bool {
		// Active sessions first.
		if sessions[i].Status != sessions[j].Status {
			return sessions[i].Status == "active"
		}
		// Then by file modification time descending (most recently updated first).
		return sessions[i].UpdatedAt > sessions[j].UpdatedAt
	})

	return sessions, nil
}

// FindSession locates a single session by its ID under claudeHome.
func FindSession(claudeHome, sessionID string) (*Session, error) {
	pattern := filepath.Join(claudeHome, "projects", "*", sessionID+".jsonl")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob session: %w", err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	path := matches[0]
	project := filepath.Base(filepath.Dir(path))
	cwd, gitBranch := extractMetadata(path)
	title := loadTitle(claudeHome, sessionID)

	activeMap, _ := loadActiveSessions(claudeHome)
	status := "ended"
	if activeMap[sessionID] {
		status = "active"
	}

	return &Session{
		ID:        sessionID,
		Title:     title,
		Project:   project,
		CWD:       cwd,
		GitBranch: gitBranch,
		Status:    status,
		JSONLPath: path,
	}, nil
}

// loadActiveSessions reads sessions/*.json and returns a set of session IDs
// whose processes are still alive (verified via kill -0).
// ListActiveSessions returns the session IDs of all currently active Claude sessions
// by checking sessions/*.json and verifying the process is alive.
func ListActiveSessions(claudeHome string) ([]string, error) {
	active, err := loadActiveSessions(claudeHome)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(active))
	for id := range active {
		ids = append(ids, id)
	}
	return ids, nil
}

func loadActiveSessions(claudeHome string) (map[string]bool, error) {
	active := make(map[string]bool)

	pattern := filepath.Join(claudeHome, "sessions", "*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		log.Printf("[loadActiveSessions] glob error: %v (pattern=%s)", err, pattern)
		return active, err
	}
	log.Printf("[loadActiveSessions] claudeHome=%s, found %d session files", claudeHome, len(matches))

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("[loadActiveSessions] read %s error: %v", path, err)
			continue
		}
		var s activeSession
		if err := json.Unmarshal(data, &s); err != nil {
			log.Printf("[loadActiveSessions] unmarshal %s error: %v", path, err)
			continue
		}
		alive := isProcessAlive(s.PID)
		log.Printf("[loadActiveSessions] file=%s sid=%s pid=%d alive=%v", filepath.Base(path), s.SessionID, s.PID, alive)
		if s.PID > 0 && alive {
			active[s.SessionID] = true
		}
	}

	return active, nil
}

// isProcessAlive checks if a process with the given PID is still running.
func isProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

// loadTitle reads the session title from session-titles/<id>.
func loadTitle(claudeHome, sessionID string) string {
	path := filepath.Join(claudeHome, "session-titles", sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// extractLastUserMessage reads the JSONL and returns the last user message text, truncated to 80 chars.
func extractLastUserMessage(jsonlPath string) string {
	f, err := os.Open(jsonlPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	var last string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		msg, err := ParseLine(scanner.Bytes())
		if err != nil || msg.Type != "user" || msg.Message == nil {
			continue
		}
		for _, block := range msg.Message.Content {
			if block.Type == "text" && block.Text != "" {
				last = block.Text
			}
		}
	}
	if len(last) > 80 {
		last = last[:80] + "..."
	}
	return last
}

// extractFirstUserMessage reads the JSONL and returns the first user message text, truncated to 50 chars.
func extractFirstUserMessage(jsonlPath string) string {
	f, err := os.Open(jsonlPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		msg, err := ParseLine(scanner.Bytes())
		if err != nil || msg.Type != "user" || msg.Message == nil {
			continue
		}
		for _, block := range msg.Message.Content {
			if block.Type == "text" && block.Text != "" {
				text := block.Text
				if len(text) > 50 {
					text = text[:50] + "..."
				}
				return text
			}
		}
	}
	return ""
}

// extractMetadata reads the first line of a JSONL file to extract cwd and gitBranch.
func extractMetadata(jsonlPath string) (cwd, gitBranch string) {
	f, err := os.Open(jsonlPath)
	if err != nil {
		return "", ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
	if scanner.Scan() {
		msg, err := ParseLine(scanner.Bytes())
		if err != nil {
			return "", ""
		}
		return msg.CWD, msg.GitBranch
	}
	return "", ""
}
