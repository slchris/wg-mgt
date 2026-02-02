package router

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/slchris/wg-mgt/internal/handler"
	"github.com/slchris/wg-mgt/internal/middleware"
	"github.com/slchris/wg-mgt/internal/service"
)

// Config holds router configuration.
type Config struct {
	LocalOnly      bool
	ServerPort     int
	ServerHost     string
	AuthService    *service.AuthService
	NodeHandler    *handler.NodeHandler
	PeerHandler    *handler.PeerHandler
	NetworkHandler *handler.NetworkHandler
	AuthHandler    *handler.AuthHandler
	WebFS          fs.FS
}

// New creates a new router.
func New(cfg *Config) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)

	if cfg.LocalOnly {
		r.Use(middleware.LocalOnly)
	}

	// Health check - outside of all auth middleware
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			r.Get("/setup/check", cfg.AuthHandler.CheckSetup)
			r.Post("/setup", cfg.AuthHandler.Setup)
			r.Post("/auth/login", cfg.AuthHandler.Login)

			// Config endpoint - returns server config for frontend
			r.Get("/config", func(w http.ResponseWriter, r *http.Request) {
				config := map[string]interface{}{
					"api_port": cfg.ServerPort,
					"api_host": cfg.ServerHost,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"data":    config,
				})
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.AuthService))

			// User routes
			r.Get("/auth/me", cfg.AuthHandler.Me)
			r.Post("/auth/password", cfg.AuthHandler.ChangePassword)

			// Node routes
			r.Route("/nodes", func(r chi.Router) {
				r.Get("/", cfg.NodeHandler.GetAll)
				r.Post("/", cfg.NodeHandler.Create)
				r.Get("/{id}", cfg.NodeHandler.GetByID)
				r.Put("/{id}", cfg.NodeHandler.Update)
				r.Delete("/{id}", cfg.NodeHandler.Delete)
				r.Post("/{id}/check", cfg.NodeHandler.CheckStatus)
				r.Get("/{id}/wg-status", cfg.NodeHandler.GetWireGuardStatus)
				r.Get("/{id}/system-info", cfg.NodeHandler.GetSystemInfo)
				r.Post("/{id}/initialize", cfg.NodeHandler.InitializeWireGuard)
				r.Post("/{id}/save-config", cfg.NodeHandler.SaveWireGuardConfig)
				r.Post("/{id}/restart", cfg.NodeHandler.RestartWireGuard)
				r.Get("/{nodeId}/peers", cfg.PeerHandler.GetByNodeID)
			})

			// Peer routes
			r.Route("/peers", func(r chi.Router) {
				r.Get("/", cfg.PeerHandler.GetAll)
				r.Post("/", cfg.PeerHandler.Create)
				r.Get("/next-ip/{nodeId}", cfg.PeerHandler.GetNextIP)
				r.Post("/sync/{nodeId}", cfg.PeerHandler.SyncToServer)
				r.Get("/{id}", cfg.PeerHandler.GetByID)
				r.Put("/{id}", cfg.PeerHandler.Update)
				r.Delete("/{id}", cfg.PeerHandler.Delete)
				r.Get("/{id}/config", cfg.PeerHandler.GetConfig)
				r.Get("/{id}/qrcode", cfg.PeerHandler.GetQRCode)
			})

			// Network routes
			r.Route("/networks", func(r chi.Router) {
				r.Get("/", cfg.NetworkHandler.GetAll)
				r.Post("/", cfg.NetworkHandler.Create)
				r.Get("/{id}", cfg.NetworkHandler.GetByID)
				r.Put("/{id}", cfg.NetworkHandler.Update)
				r.Delete("/{id}", cfg.NetworkHandler.Delete)
			})
		})
	})

	// Serve static files for SPA
	if cfg.WebFS != nil {
		fileServer := http.FileServer(http.FS(cfg.WebFS))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// Try to serve the file
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}

			// Check if file exists
			f, err := cfg.WebFS.Open(path)
			if err != nil {
				// File not found, serve index.html for SPA routing
				r.URL.Path = "/"
				fileServer.ServeHTTP(w, r)
				return
			}
			f.Close()

			fileServer.ServeHTTP(w, r)
		})
	}

	return r
}
