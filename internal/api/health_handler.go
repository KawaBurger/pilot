package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler returns the server version and uptime.
func (s *Server) HealthHandler(c *gin.Context) {
	uptime := time.Since(s.StartedAt).Round(time.Second)
	c.JSON(http.StatusOK, gin.H{
		"version": "0.1.0",
		"uptime":  uptime.String(),
	})
}
