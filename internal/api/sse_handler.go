package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kawaburger/pilot/internal/claude"
)

// StreamSession streams new messages from a Claude session via SSE.
func (s *Server) StreamSession(c *gin.Context) {
	sessionID := c.Param("id")

	session, err := claude.FindSession(s.claudeHome(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	watcher, err := s.Watchers.Get(sessionID, session.JSONLPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("start watcher: %v", err)})
		return
	}

	sub := watcher.Subscribe()
	defer watcher.Unsubscribe(sub)

	// Set SSE headers.
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	clientGone := c.Request.Context().Done()

	for {
		select {
		case <-clientGone:
			return
		case <-sub.Done:
			return
		case msg := <-sub.Ch:
			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			_, err = fmt.Fprintf(c.Writer, "event: message\ndata: %s\n\n", data)
			if err != nil {
				return
			}
			c.Writer.(http.Flusher).Flush()
		}
	}
}
