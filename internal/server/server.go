package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/DiegoSantos90/chargeback-api/internal/api/http/handler"
	"github.com/DiegoSantos90/chargeback-api/internal/usecase"
)

// CreateChargebackUseCase interface defines the contract for creating chargebacks
type CreateChargebackUseCase interface {
	Execute(ctx context.Context, req usecase.CreateChargebackRequest) (*usecase.CreateChargebackResponse, error)
}

// Server represents the HTTP server
type Server struct {
	config            ServerConfig
	mux               *http.ServeMux
	chargebackHandler *handler.ChargebackHandler
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string `json:"port"`
}

// Validate validates the server configuration
func (c ServerConfig) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("port is required")
	}

	// Validate port is numeric
	if _, err := strconv.Atoi(c.Port); err != nil {
		return fmt.Errorf("port must be a valid number")
	}

	return nil
}

// NewServer creates a new HTTP server
func NewServer(config ServerConfig, createChargebackUC CreateChargebackUseCase) *Server {
	server := &Server{
		config:            config,
		mux:               http.NewServeMux(),
		chargebackHandler: handler.NewChargebackHandler(createChargebackUC),
	}

	server.setupRoutes()
	server.setupMiddleware()

	return server
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.mux.HandleFunc("/health", s.handleHealth)

	// Chargeback endpoints
	s.mux.HandleFunc("/chargebacks", s.chargebackHandler.CreateChargeback)
}

// setupMiddleware applies middleware to the server
func (s *Server) setupMiddleware() {
	// Middleware is applied through ServeHTTP method
}

// ServeHTTP implements http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply CORS middleware
	s.corsMiddleware(w, r)

	// Return early for OPTIONS requests (handled by CORS middleware)
	if r.Method == http.MethodOptions {
		return
	}

	// Create a response writer wrapper to capture status code
	wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	start := time.Now()

	// Check if route exists
	if !s.routeExists(r.URL.Path) {
		wrapped.Header().Set("Content-Type", "application/json")
		wrapped.WriteHeader(http.StatusNotFound)
		json.NewEncoder(wrapped).Encode(map[string]string{"error": "Not found"})
	} else {
		s.mux.ServeHTTP(wrapped, r)
	}

	// Log the request
	duration := time.Since(start)
	log.Printf("HTTP %s %s - %d - %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
}

// routeExists checks if a route exists
func (s *Server) routeExists(path string) bool {
	validRoutes := []string{"/health", "/chargebacks"}
	for _, route := range validRoutes {
		if path == route {
			return true
		}
	}
	return false
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"service":   "chargeback-api",
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// corsMiddleware handles CORS (Cross-Origin Resource Sharing)
func (s *Server) corsMiddleware(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	if err := s.config.Validate(); err != nil {
		return fmt.Errorf("invalid server configuration: %w", err)
	}

	addr := ":" + s.config.Port
	log.Printf("Starting server on %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      s,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server.ListenAndServe()
}
