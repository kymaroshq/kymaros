/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RouteRegistrar is a function that registers additional routes on the mux.
// Used by kymaros-pro to add premium API endpoints.
type RouteRegistrar func(mux *http.ServeMux, q *Queries)

// Server is the Kymaros API and static-file HTTP server.
type Server struct {
	client      client.Client
	restConfig  *rest.Config
	port        int
	staticDir   string // path to frontend/dist
	logger      *slog.Logger
	httpServer  *http.Server
	extraRoutes []RouteRegistrar
}

// NewServer creates a Server that serves the API and, optionally, the React frontend.
func NewServer(c client.Client, port int, staticDir string, rc *rest.Config) *Server {
	return &Server{
		client:     c,
		restConfig: rc,
		port:       port,
		staticDir:  staticDir,
		logger:     slog.Default(),
	}
}

// RegisterRoutes adds a RouteRegistrar that will be called during Start
// to register additional API routes (e.g. from kymaros-pro).
func (s *Server) RegisterRoutes(r RouteRegistrar) {
	s.extraRoutes = append(s.extraRoutes, r)
}

// Start starts the HTTP server (blocking). Call in a goroutine.
func (s *Server) Start(ctx context.Context) error {
	q := NewQueries(s.client, s.restConfig)
	mux := http.NewServeMux()

	// --- API routes ---
	mux.HandleFunc("GET /api/v1/health", HandleHealth(q))
	mux.HandleFunc("GET /api/v1/summary", HandleSummary(q))
	mux.HandleFunc("GET /api/v1/summary/daily", HandleDailyScores(q))
	mux.HandleFunc("GET /api/v1/tests", HandleTests(q))
	mux.HandleFunc("GET /api/v1/reports", HandleReports(q))
	mux.HandleFunc("GET /api/v1/reports/{name}/logs", HandleReportLogs(q))
	mux.HandleFunc("GET /api/v1/alerts", HandleAlerts(q))
	mux.HandleFunc("GET /api/v1/upcoming", HandleUpcoming(q))
	mux.HandleFunc("GET /api/v1/schedules", HandleSchedules(q))
	mux.HandleFunc("POST /api/v1/tests", HandleCreateTest(q))
	mux.HandleFunc("PUT /api/v1/tests/{name}", HandleUpdateTest(q))
	mux.HandleFunc("DELETE /api/v1/tests/{name}", HandleDeleteTest(q))
	mux.HandleFunc("POST /api/v1/tests/{name}/trigger", HandleTriggerTest(q))

	// --- Config routes ---
	mux.HandleFunc("GET /api/v1/config/provider", HandleGetProviderConfig(q))
	mux.HandleFunc("GET /api/v1/config/notifications", HandleGetNotificationConfig(q))
	mux.HandleFunc("POST /api/v1/config/notifications/test", HandleTestNotification(q))
	mux.HandleFunc("GET /api/v1/config/sandbox", HandleGetSandboxConfig(q))
	mux.HandleFunc("POST /api/v1/config/sandbox/cleanup", HandleCleanupOrphanSandboxes(q))

	// --- Pro routes (registered by kymaros-pro if present) ---
	for _, register := range s.extraRoutes {
		register(mux, q)
	}

	// --- Static files (SPA fallback) ---
	s.registerStaticHandler(mux)

	var handler http.Handler = mux
	handler = AuthMiddleware(handler)
	handler = CORSMiddleware(handler)

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	s.logger.Info("kymaros api server starting", "port", s.port, "static", s.staticDir)

	// Respect context cancellation.
	go func() {
		<-ctx.Done()
		_ = s.Shutdown(context.Background())
	}()

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("api server listen: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	s.logger.Info("kymaros api server shutting down")
	return s.httpServer.Shutdown(ctx)
}

// registerStaticHandler serves the React SPA from staticDir (if it exists).
// API routes are registered first on the mux, so they always take priority.
// Any request that does not match an API route and does not resolve to a real
// file is served index.html (SPA client-side routing).
func (s *Server) registerStaticHandler(mux *http.ServeMux) {
	if s.staticDir == "" {
		return
	}

	abs, err := filepath.Abs(s.staticDir)
	if err != nil {
		s.logger.Warn("static dir resolve failed, skipping", "dir", s.staticDir, "error", err)
		return
	}

	indexPath := filepath.Join(abs, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		s.logger.Info("no index.html in static dir, serving API only", "dir", abs)
		return
	}

	staticFS := os.DirFS(abs)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Never intercept API paths (safety net — mux already matched them).
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to serve the requested file.
		cleanPath := strings.TrimPrefix(r.URL.Path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		if _, err := fs.Stat(staticFS, cleanPath); err == nil {
			http.ServeFileFS(w, r, staticFS, cleanPath)
			return
		}

		// SPA fallback: serve index.html for any unmatched path.
		http.ServeFileFS(w, r, staticFS, "index.html")
	})

	s.logger.Info("serving static files", "dir", abs)
}
