package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/edgetainer/edgetainer/internal/server/db"
	"github.com/edgetainer/edgetainer/internal/server/ssh"
	"github.com/edgetainer/edgetainer/internal/shared/logging"
)

// Server represents the API server
type Server struct {
	host       string
	port       int
	httpServer *http.Server
	database   *db.DB
	sshServer  *ssh.Server
	logger     *logging.Logger
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewServer creates a new API server
func NewServer(ctx context.Context, host string, port int, database *db.DB, sshServer *ssh.Server) (*Server, error) {
	serverCtx, cancel := context.WithCancel(ctx)

	logger := logging.WithComponent("api-server")

	return &Server{
		host:       host,
		port:       port,
		database:   database,
		sshServer:  sshServer,
		logger:     logger,
		ctx:        serverCtx,
		cancelFunc: cancel,
	}, nil
}

// Start starts the API server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	// Setup router
	router := http.NewServeMux()

	// Register API routes
	router.HandleFunc("/api/health", s.handleHealth)

	// Auth routes
	router.HandleFunc("/api/auth/login", s.handleLogin)
	router.HandleFunc("/api/auth/logout", s.handleLogout)
	router.HandleFunc("/api/auth/me", s.authMiddleware(s.handleGetCurrentUser))

	// Fleet routes
	router.HandleFunc("/api/fleets", s.authMiddleware(s.handleFleets))
	router.HandleFunc("/api/fleets/", s.authMiddleware(s.handleFleetByID)) // Handles /api/fleets/{id}

	// Device routes
	router.HandleFunc("/api/devices", s.authMiddleware(s.handleDevices))
	router.HandleFunc("/api/devices/", s.authMiddleware(s.handleDeviceByID)) // Handles /api/devices/{id}

	// Software routes
	router.HandleFunc("/api/software", s.authMiddleware(s.handleSoftware))
	router.HandleFunc("/api/software/", s.authMiddleware(s.handleSoftwareByID)) // Handles /api/software/{id}

	// Agent routes
	router.HandleFunc("/api/agent/heartbeat", s.handleAgentHeartbeat)
	router.HandleFunc("/api/agent/status", s.handleAgentStatus)

	// Provision routes
	router.HandleFunc("/api/provision/device", s.handleDeviceProvisioning) // Create new device provisioning config

	// Setup static file serving for web UI with SPA support
	var webDir string
	webDirs := []string{"./web", "/app/web"}
	for _, dir := range webDirs {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			webDir = dir
			s.logger.Info(fmt.Sprintf("Found web UI directory at %s", webDir))
			break
		}
	}

	if webDir != "" {
		// Create a SPA file server handler that serves index.html for unmatched routes
		router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// First, check if the requested file exists in the web directory
			path := webDir + r.URL.Path
			_, err := os.Stat(path)

			// If file doesn't exist or is a directory, serve index.html
			if os.IsNotExist(err) || (err == nil && r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/api/")) {
				// Check if the path contains a dot (likely a file extension)
				// If yes, let it 404 as normal, don't serve index.html for missing assets
				if !strings.Contains(r.URL.Path, ".") {
					s.logger.Info(fmt.Sprintf("Serving index.html for SPA route: %s", r.URL.Path))
					http.ServeFile(w, r, webDir+"/index.html")
					return
				}
			}

			// Otherwise serve the file directly
			http.FileServer(http.Dir(webDir)).ServeHTTP(w, r)
		})
	} else {
		s.logger.Warn("No web UI directory found")
	}

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.loggingMiddleware(router),
	}

	s.logger.Info(fmt.Sprintf("API server listening on %s", addr))

	// Start HTTP server
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error(fmt.Sprintf("HTTP server error: %v", err), err)
		}
	}()

	return nil
}

// Shutdown stops the API server
func (s *Server) Shutdown() {
	s.logger.Info("Shutting down API server")

	// Create a context with timeout for the graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("HTTP server shutdown error", err)
		}
	}

	// Signal the server context to cancel
	s.cancelFunc()

	s.logger.Info("API server shutdown complete")
}

// handleHealth handles the health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}

	jsonResponse(w, response, http.StatusOK)
}
