package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/time/rate"
)

// Build a full API gateway with:
//
// 1. Per-IP rate limiting (5 req/sec, burst 10)
//    - Map of *rate.Limiter keyed by IP
//    - Use net.SplitHostPort(r.RemoteAddr) to get the IP
//
// 2. Circuit breaker (threshold: 3 failures, cooldown: 15s)
//    - Track failure count, open/close states
//    - Record failures in ErrorHandler, successes in ModifyResponse
//    - Check IsOpen() before proxying
//
// 3. Logging middleware (outermost)
//    - Wrap ResponseWriter to capture status code
//    - Log: method, path, status, duration
//
// 4. Two backend proxies with StripPrefix
//    - /svc1/ → :9001
//    - /svc2/ → :9002
//
// 5. Graceful shutdown
//
// Middleware order: logging(rateLimit(mux))

type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    burst,
	}
}

func (r *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	limiter, exists := r.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(r.rate, r.burst)
		r.limiters[ip] = limiter
	}
	return limiter
}

var rateLimiter = NewRateLimiter(5, 10)

func rateLimitMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !rateLimiter.GetLimiter(ip).Allow() {
			http.Error(w, `{"error":"rate_limit_exceeded"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type CircuitBreaker struct {
	mu        sync.Mutex
	failures  int
	threshold int
	openUntil time.Time
	cooldown  time.Duration
}

func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		cooldown:  cooldown,
	}
}

func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.failures >= cb.threshold {
		if time.Now().Before(cb.openUntil) {
			return true
		}
		cb.failures = cb.threshold - 1
		return false
	}

	return false
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	if cb.failures >= cb.threshold {
		cb.openUntil = time.Now().Add(cb.cooldown)
		log.Printf("circuit is open and cooling down for %v",
			cb.cooldown)
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
}

func getProxy(urlString string, circuitBreaker *CircuitBreaker) *httputil.ReverseProxy {
	url, err := url.Parse(urlString)

	if err != nil {
		log.Fatalf("Could not parse URL for %s", urlString)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		circuitBreaker.RecordFailure()
		http.Error(w, `{"error": "bad_gateway"}`, http.StatusBadGateway)
	}

	proxy.ModifyResponse = func(r *http.Response) error {
		if r.StatusCode < 500 {
			circuitBreaker.RecordSuccess()
		} else {
			circuitBreaker.RecordFailure()
		}
		return nil
	}

	return proxy
}

type StatusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *StatusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &StatusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		log.Printf("%s %s → %d (%v)",
			r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

func gatewayHandler(proxy *httputil.ReverseProxy, cb *CircuitBreaker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cb.IsOpen() {
			http.Error(w, `{"error":"service_unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		proxy.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	svc1CB := NewCircuitBreaker(3, 15*time.Second)
	svc2CB := NewCircuitBreaker(3, 15*time.Second)

	svc1Proxy := getProxy("http://localhost:9001", svc1CB)
	svc2Proxy := getProxy("http://localhost:9002", svc2CB)

	handler1 := http.StripPrefix("/svc1", gatewayHandler(svc1Proxy, svc1CB))
	handler2 := http.StripPrefix("/svc2", gatewayHandler(svc2Proxy, svc2CB))

	mux.Handle("/svc1/", handler1)
	mux.Handle("/svc2/", handler2)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      loggingMiddleware(rateLimitMiddleWare(mux)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("gateway on :8080")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server Error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	log.Println("server stopped cleanly")
}
