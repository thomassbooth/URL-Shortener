package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
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
	limiter  *Limiter     // Add Limiter field
	srv      *http.Server // Add HTTP server field
}

// NewServer initializes a new server with the given configuration, worker pool, and database.
func NewServer(cfg Config, db *storage.Storage) *Server {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultListenAddr
	}
	//setup worker pool and handlers
	wp := internal.NewWorkerPool(5, db)
	handlers := NewHandlers(wp, db)
	// Initialize rate limiter with a limit of 10 requests per second and a burst of 20
	limiter := NewLimiter(rate.Limit(10), 20)
	srv := &http.Server{
		Addr: cfg.ListenAddr,
	}
	return &Server{
		Config:   cfg,
		Wp:       wp,
		handlers: handlers,
		db:       db,
		limiter:  limiter,
		srv:      srv, // Store the HTTP server
	}
}

// Start starts the server and listens for incoming requests.
func (s *Server) Start() error {
	// Wrap routes with rate limiting middleware
	http.Handle("/shorten", s.limiter.LimitMiddleware(http.HandlerFunc(s.handlers.HandleShortenURL)))
	http.Handle("/", s.limiter.LimitMiddleware(http.HandlerFunc(s.handlers.HandleRedirect)))
	s.srv.Handler = http.DefaultServeMux
	// Start the server in a separate goroutine with recovery handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic: %v", r)
				// Optionally restart the server if desired
				if err := s.Start(); err != nil {
					log.Fatalf("Server failed to restart: %v", err)
				}
			}
		}()

		// Start the server and handle potential errors
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed to start: %v", err)
		}
	}()
	fmt.Println("Server initialized and started, access routes /shorten and /{shortURL}")
	return nil
}

// Stop gracefully stops the server and other resources.
func (s *Server) Stop() {
	// Create a context with a timeout for the shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Attempt to gracefully shutdown the HTTP server
	if err := s.srv.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	} else {
		fmt.Println("HTTP server shut down gracefully")
	}

	// Stop the worker pool, close threads and close db connection
	s.Wp.Stop()
	if err := s.db.Close(); err != nil {
		fmt.Printf("Error closing the database: %v\n", err)
	} else {
		fmt.Println("Database connection closed")
	}

	fmt.Println("Server stopped gracefully")
}
