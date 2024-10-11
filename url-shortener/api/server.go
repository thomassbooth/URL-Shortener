package api

import (
	"fmt"
	"net/http"
	"url-shortener/url-shortener/internal"
	"url-shortener/url-shortener/storage"

	"golang.org/x/time/rate"
)

// Limiter struct (from your rate limiter code) omitted for brevity

const defaultListenAddr = ":8005"

// Config holds the configuration for the server.
type Config struct {
	ListenAddr string
	DSN        string // Add Data Source Name for PostgreSQL connection
}

// Server struct holds the server's configuration, worker pool, handlers, and rate limiter.
type Server struct {
	Config
	Wp       *internal.WorkerPool
	handlers *Handlers
	db       *storage.Storage
	limiter  *Limiter // Add Limiter field
}

// NewServer initializes a new server with the given configuration, worker pool, and database.
func NewServer(cfg Config, db *storage.Storage) *Server {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultListenAddr
	}

	wp := internal.NewWorkerPool(5, db)
	handlers := NewHandlers(wp, db)

	// Initialize rate limiter with a limit of 10 requests per second and a burst of 20
	limiter := NewLimiter(rate.Limit(10), 20)

	return &Server{
		Config:   cfg,
		Wp:       wp,
		handlers: handlers,
		db:       db,
		limiter:  limiter,
	}
}

// Start starts the server and listens for incoming requests and signals.
func (s *Server) Start() error {
	// Setup HTTP server and routes

	// Wrap routes with rate limiting middleware
	http.Handle("/shorten", s.limiter.LimitMiddleware(http.HandlerFunc(s.handlers.HandleShortenURL)))
	http.Handle("/", s.limiter.LimitMiddleware(http.HandlerFunc(s.handlers.HandleRedirect)))

	srv := &http.Server{Addr: s.ListenAddr}

	fmt.Printf("Starting server on %s\n", s.ListenAddr)
	// If the server fails to start, return the error
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %v", err)
	}

	return nil
}

// Stop gracefully stops the worker pool and closes the database connection.
func (s *Server) Stop() {
	// Close the worker pool and close threads
	s.Wp.Stop()

	// Close the database connection
	if err := s.db.Close(); err != nil {
		fmt.Printf("Error closing the database: %v", err)
	}

	fmt.Println("Server stopped gracefully")
}
