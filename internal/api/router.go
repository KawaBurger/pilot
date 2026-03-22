package api

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kawaburger/pilot/internal/auth"
	"github.com/kawaburger/pilot/internal/claude"
	"github.com/kawaburger/pilot/internal/config"
)

// Server holds all dependencies needed by the API handlers.
type Server struct {
	Config    *config.Config
	Commands  *config.CommandsConfig
	Store     *auth.Store
	JWT       *auth.JWTManager
	Watchers  *claude.WatcherManager
	TmuxMap   sync.Map // sessionID -> tmuxName
	StartedAt time.Time
}

// NewRouter creates a gin.Engine with all routes wired up.
func NewRouter(s *Server) *gin.Engine {
	r := gin.Default()

	// CORS middleware.
	r.Use(cors.New(cors.Config{
		AllowOrigins:     s.Config.Server.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Public routes.
	r.GET("/api/health", s.HealthHandler)
	r.POST("/api/auth/login", s.LoginHandler)

	// Protected routes.
	protected := r.Group("/api", auth.Middleware(s.JWT))
	{
		protected.GET("/sessions", s.ListSessions)
		protected.GET("/sessions/:id", s.GetSession)
		protected.GET("/sessions/:id/messages", s.GetMessages)
		protected.POST("/sessions", s.NewSession)
		protected.POST("/sessions/:id/resume", s.ResumeSession)
		protected.POST("/sessions/:id/message", s.SendMessage)
		protected.POST("/sessions/:id/interrupt", s.InterruptSession)
		protected.GET("/sessions/:id/stream", s.StreamSession)
		protected.GET("/sessions/:id/terminal", s.TerminalContent)
		protected.POST("/sessions/:id/keys", s.SendKeysHandler)
		protected.GET("/commands", s.ListCommands)
	}

	// Serve frontend static files if web/dist exists.
	if webDir := findWebDist(); webDir != "" {
		r.NoRoute(serveFrontend(webDir))
	}

	return r
}

// findWebDist locates the web/dist directory relative to the executable or cwd.
func findWebDist() string {
	// Try relative to executable.
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Join(filepath.Dir(exe), "web", "dist")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	// Try relative to cwd (for development).
	if info, err := os.Stat("web/dist"); err == nil && info.IsDir() {
		abs, _ := filepath.Abs("web/dist")
		return abs
	}
	return ""
}

// serveFrontend returns a handler that serves static files, falling back to index.html for SPA routing.
func serveFrontend(webDir string) gin.HandlerFunc {
	fileServer := http.FileServer(http.Dir(webDir))
	return func(c *gin.Context) {
		// If the file exists, serve it directly.
		path := filepath.Join(webDir, c.Request.URL.Path)
		if _, err := fs.Stat(os.DirFS(webDir), c.Request.URL.Path[1:]); err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		// Otherwise serve index.html (SPA fallback).
		_ = path
		c.File(filepath.Join(webDir, "index.html"))
	}
}
