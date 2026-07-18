package agentd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/agent"
)

// Server is the HTTP server for the Agent daemon.
type Server struct {
	addr          string
	maxConcurrent int
	pool          *Pool
	runs          sync.Map // run_id -> *agent.RunState
	startTime     time.Time
	root          string
	httpServer    *http.Server
}

// NewServer creates a new Agent daemon server.
func NewServer(root, addr string, maxConcurrent int) *Server {
	s := &Server{
		addr:          addr,
		maxConcurrent: maxConcurrent,
		startTime:     time.Now(),
		root:          root,
	}
	s.pool = NewPool(root, maxConcurrent)
	s.setupRoutes()
	return s
}

// setupRoutes registers all HTTP handlers.
func (s *Server) setupRoutes() {
	mux := http.NewServeMux()

	// REST API endpoints
	mux.HandleFunc("/api/v1/health", s.healthHandler)
	mux.HandleFunc("/api/v1/runs", s.listRunsHandler)
	mux.HandleFunc("/api/v1/runs/", s.routeByMethod)
	mux.HandleFunc("/api/v1/incidents", s.createIncidentHandler)
	mux.HandleFunc("/api/v1/dashboard", s.dashboardHandler)

	s.httpServer = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}
}

// routeByMethod routes /api/v1/runs/{id} and /api/v1/runs/{id}/confirm
func (s *Server) routeByMethod(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// /api/v1/runs/{id}/confirm
	if len(path) > len("/api/v1/runs/") {
		rest := path[len("/api/v1/runs/"):]
		for i, c := range rest {
			if c == '/' {
				// Check if this is the /confirm endpoint
				segment := rest[i+1:]
				if segment == "confirm" {
					s.confirmRunHandler(w, r)
					return
				}
				break
			}
		}
	}

	// /api/v1/runs/{id}
	s.getRunHandler(w, r)
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	logInfo("server", "starting on %s", s.addr)

	// Start signal handler
	go s.handleSignals(ctx)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server listen: %w", err)
	}
	return nil
}

// Stop gracefully stops the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	logInfo("server", "shutting down...")

	// Drain the pool (wait for running tasks)
	if err := s.pool.Drain(ctx); err != nil {
		logError("server", "pool drain: %v", err)
	}

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	logInfo("server", "shutdown complete")
	return nil
}

// handleSignals listens for OS signals.
func (s *Server) handleSignals(parentCtx context.Context) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-sigCh:
		logInfo("server", "received signal: %v", sig)
		ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
		defer cancel()
		if err := s.Stop(ctx); err != nil {
			logError("server", "stop failed: %v", err)
		}
		os.Exit(0)
	case <-parentCtx.Done():
		// Parent context cancelled
	}
}

// getRunState returns the RunState for a given run ID.
func (s *Server) getRunState(runID string) *agent.RunState {
	if v, ok := s.runs.Load(runID); ok {
		return v.(*agent.RunState)
	}
	return nil
}

// setRunState stores the RunState for a given run ID.
func (s *Server) setRunState(runID string, state *agent.RunState) {
	s.runs.Store(runID, state)
}

func logInfo(component, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[INFO] agentd.%s | "+format+"\n", append([]interface{}{component}, args...)...)
}

func logError(component, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] agentd.%s | "+format+"\n", append([]interface{}{component}, args...)...)
}
