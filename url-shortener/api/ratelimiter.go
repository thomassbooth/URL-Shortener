package api

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter struct holds the limiter and a timestamp when it was last accessed.
type RateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Limiter struct stores rate limiters for each client IP.
type Limiter struct {
	clients map[string]*RateLimiter
	mu      sync.Mutex
	limit   rate.Limit
	burst   int
}

// NewLimiter initializes the rate limiter with a limit and burst rate.
func NewLimiter(r rate.Limit, b int) *Limiter {
	return &Limiter{
		clients: make(map[string]*RateLimiter),
		limit:   r,
		burst:   b,
	}
}

// GetLimiter returns the rate limiter for the given IP or creates a new one.
func (l *Limiter) GetLimiter(ip string) *rate.Limiter {
	// prevent race conditions
	l.mu.Lock()
	defer l.mu.Unlock()

	// Clean up old clients, we dont want to keep them around
	for ip, rl := range l.clients {
		if time.Since(rl.lastSeen) > 5*time.Minute {
			delete(l.clients, ip)
		}
	}

	// If the IP already has a limiter, update the lastSeen timestamp
	if _, exists := l.clients[ip]; !exists {
		// Create a new limiter for new clients
		l.clients[ip] = &RateLimiter{
			limiter:  rate.NewLimiter(l.limit, l.burst),
			lastSeen: time.Now(),
		}
	} else {
		l.clients[ip].lastSeen = time.Now()
	}

	return l.clients[ip].limiter
}

// LimitMiddleware enforces the rate limiting for each request. Decorator
func (l *Limiter) LimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Check if the IP exceeds the rate limit
		limiter := l.GetLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		// Call the next handler if rate limit is not exceeded
		next.ServeHTTP(w, r)
	})
}
