package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListCommands returns the configured quick commands.
func (s *Server) ListCommands(c *gin.Context) {
	c.JSON(http.StatusOK, s.Commands.Commands)
}
