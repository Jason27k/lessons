# Lesson 8: Building a Simple API Gateway

## The Problem

Last lesson introduced the pieces — reverse proxies, rate limiting, Director hooks. This lesson puts them together into a real, production-shaped gateway with features you skipped:

- **Per-client rate limiting** (not just global)
- **Health checks** (don't send traffic to dead backends)
- **Circuit breaker** (stop hammering a failing backend)
- **Request logging with timing**

You'll build the whole thing in one file.

---

## 1. Per-Client Rate Limiting

Last lesson's global limiter has a problem: 10 well-behaved clients sharing a 10 req/sec limit means each gets ~1 req/sec. One aggressive client steals everyone else's quota.

Fix: a limiter **per IP address**.

```go
type IPRateLimiter struct {
    mu       sync.Mutex
    limiters map[string]*rate.Limiter
    rate     rate.Limit
    burst    int
}

func NewIPRateLimiter(r rate.Limit, burst int) *IPRateLimiter {
    return &IPRateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     r,
        burst:    burst,
    }
}

func (rl *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    limiter, exists := rl.limiters[ip]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.limiters[ip] = limiter
    }
    return limiter
}
```

Use it in middleware:

```go
var ipLimiter = NewIPRateLimiter(5, 10) // 5 req/sec per IP, burst 10

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip, _, _ := net.SplitHostPort(r.RemoteAddr)
        if !ipLimiter.GetLimiter(ip).Allow() {
            http.Error(w, `{"error":"rate_limit_exceeded"}`, http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

**Gotcha:** The `limiters` map grows forever. In production, you'd add a cleanup goroutine that removes entries not seen in the last N minutes. For this lesson, we'll keep it simple.

---

## 2. Health Checks

Before routing traffic, check if a backend is alive. The simplest approach: periodically hit a health endpoint.

```go
type Backend struct {
    URL     *url.URL
    Proxy   *httputil.ReverseProxy
    alive   bool
    mu      sync.RWMutex
}

func (b *Backend) IsAlive() bool {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return b.alive
}

func (b *Backend) SetAlive(alive bool) {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.alive = alive
}

func healthCheck(b *Backend, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for range ticker.C {
        resp, err := http.Get(b.URL.String() + "/health")
        if err != nil || resp.StatusCode != http.StatusOK {
            if b.IsAlive() {
                log.Printf("backend %s is DOWN", b.URL)
            }
            b.SetAlive(false)
            continue
        }
        resp.Body.Close()
        if !b.IsAlive() {
            log.Printf("backend %s is UP", b.URL)
        }
        b.SetAlive(true)
    }
}
```

Start a health check goroutine per backend:

```go
go healthCheck(backend, 10*time.Second)
```

Then in your handler, only route to alive backends:

```go
if !backend.IsAlive() {
    http.Error(w, `{"error":"service_unavailable"}`, http.StatusServiceUnavailable)
    return
}
backend.Proxy.ServeHTTP(w, r)
```

---

## 3. Circuit Breaker

Health checks run on a timer. But what if a backend dies *between* checks? A circuit breaker tracks failures in real-time:

```go
type CircuitBreaker struct {
    mu           sync.Mutex
    failures     int
    threshold    int
    openUntil    time.Time
    cooldown     time.Duration
}

func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        threshold: threshold,
        cooldown:  cooldown,
    }
}

// IsOpen returns true if the circuit is tripped (too many failures).
func (cb *CircuitBreaker) IsOpen() bool {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    if cb.failures >= cb.threshold {
        // Still in cooldown?
        if time.Now().Before(cb.openUntil) {
            return true
        }
        // Cooldown expired — allow one request through (half-open)
        cb.failures = cb.threshold - 1
        return false
    }
    return false
}

// RecordFailure increments the failure count and starts cooldown if threshold hit.
func (cb *CircuitBreaker) RecordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.failures++
    if cb.failures >= cb.threshold {
        cb.openUntil = time.Now().Add(cb.cooldown)
        log.Printf("circuit OPEN — cooling down for %v", cb.cooldown)
    }
}

// RecordSuccess resets the failure count.
func (cb *CircuitBreaker) RecordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.failures = 0
}
```

### The three states:

```
     success
  ┌───────────┐
  │           ▼
CLOSED ──failures──► OPEN ──cooldown──► HALF-OPEN
  ▲                                        │
  └────────── success ─────────────────────┘
              failure → back to OPEN
```

- **Closed** — normal operation, requests pass through
- **Open** — too many failures, reject immediately (fast-fail)
- **Half-open** — cooldown expired, allow one request to test if backend recovered

Wire it into the proxy's ErrorHandler:

```go
cb := NewCircuitBreaker(3, 30*time.Second) // open after 3 failures, wait 30s

proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
    cb.RecordFailure()
    http.Error(w, `{"error":"bad_gateway"}`, http.StatusBadGateway)
}

// In your handler, check circuit before proxying:
if cb.IsOpen() {
    http.Error(w, `{"error":"circuit_open"}`, http.StatusServiceUnavailable)
    return
}
proxy.ServeHTTP(w, r)
```

And record successes via `ModifyResponse`:

```go
proxy.ModifyResponse = func(resp *http.Response) error {
    if resp.StatusCode < 500 {
        cb.RecordSuccess()
    } else {
        cb.RecordFailure()
    }
    return nil
}
```

---

## 4. Request Logging with Timing

Wrap everything with a logging middleware that captures duration:

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Wrap ResponseWriter to capture status code
        wrapped := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
        next.ServeHTTP(wrapped, r)

        log.Printf("%s %s → %d (%v)",
            r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
    })
}

type statusRecorder struct {
    http.ResponseWriter
    statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
    sr.statusCode = code
    sr.ResponseWriter.WriteHeader(code)
}
```

This is a pattern you'll see everywhere — embedding `http.ResponseWriter` and overriding `WriteHeader` to capture the status code.

---

## 5. Putting It All Together

Here's the architecture of what you're building:

```
Client Request
    │
    ▼
┌──────────────────────┐
│   Logging Middleware  │  ← captures timing + status
├──────────────────────┤
│ Rate Limit Middleware │  ← per-IP limiting
├──────────────────────┤
│     ServeMux          │  ← routes /svc1/ and /svc2/
│  ┌────────┐ ┌───────┐│
│  │ proxy1 │ │proxy2 ││  ← each has circuit breaker
│  └────┬───┘ └───┬───┘│
└───────┼─────────┼────┘
        ▼         ▼
   Backend 1  Backend 2
```

Middleware stacks from outside in: `logging(rateLimit(mux))`

---

## Key Rules

1. **Per-client > global** rate limiting — global punishes everyone for one client's behavior
2. **Circuit breakers complement health checks** — health checks are periodic, circuit breakers are reactive
3. **Half-open state matters** — without it, the circuit never closes and the backend stays blacklisted forever
4. **`statusRecorder` pattern** — embed `ResponseWriter`, override `WriteHeader` to capture the status code
5. **Middleware order matters** — logging should be outermost (captures everything), rate limiting next

---

## Common Mistakes

| Mistake | Why it's wrong |
|---------|---------------|
| Per-IP limiter map with no cleanup | Memory leak — grows forever with unique IPs |
| Circuit breaker with no half-open state | Backend recovers but circuit never closes |
| Health check with no timeout | `http.Get` has no timeout by default — use a client with timeout |
| Logging middleware inside rate limiting | Rate-limited requests don't get logged |
| `statusRecorder` not setting default 200 | If handler never calls `WriteHeader`, you record 0 |

---

## Your Turn

### Exercise: Build the Full Gateway

Build a gateway in `main.go` with:

1. **Two backend proxies** (`/svc1/` → `:9001`, `/svc2/` → `:9002`)
2. **Per-IP rate limiting** (5 req/sec, burst 10)
3. **A circuit breaker** on each proxy (threshold: 3 failures, cooldown: 15 seconds)
4. **Logging middleware** that records method, path, status, and duration
5. **Graceful shutdown** (from Lesson 6)

**Test plan:**

1. Start a backend on `:9001` (just return "hello")
2. Start your gateway on `:8080`
3. `curl http://localhost:8080/svc1/` → should work
4. `curl http://localhost:8080/svc2/` → should fail (no backend), circuit breaker records failure
5. Hit `/svc2/` three more times → circuit opens, response changes to "circuit open"
6. Spam `/svc1/` rapidly → should see 429s after burst exhausted
7. Check your logs — every request should show method, path, status, duration

---

<details>
<summary>Full Answer</summary>

```go
package main

import (
    "context"
    "encoding/json"
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

// ---- Per-IP Rate Limiter ----

type IPRateLimiter struct {
    mu       sync.Mutex
    limiters map[string]*rate.Limiter
    rate     rate.Limit
    burst    int
}

func NewIPRateLimiter(r rate.Limit, burst int) *IPRateLimiter {
    return &IPRateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     r,
        burst:    burst,
    }
}

func (rl *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    lim, exists := rl.limiters[ip]
    if !exists {
        lim = rate.NewLimiter(rl.rate, rl.burst)
        rl.limiters[ip] = lim
    }
    return lim
}

// ---- Circuit Breaker ----

type CircuitBreaker struct {
    mu        sync.Mutex
    failures  int
    threshold int
    openUntil time.Time
    cooldown  time.Duration
}

func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
    return &CircuitBreaker{threshold: threshold, cooldown: cooldown}
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
    }
}

func (cb *CircuitBreaker) RecordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.failures = 0
}

// ---- Status Recorder ----

type statusRecorder struct {
    http.ResponseWriter
    statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
    sr.statusCode = code
    sr.ResponseWriter.WriteHeader(code)
}

// ---- Middleware ----

var ipLimiter = NewIPRateLimiter(5, 10)

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip, _, _ := net.SplitHostPort(r.RemoteAddr)
        if !ipLimiter.GetLimiter(ip).Allow() {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusTooManyRequests)
            json.NewEncoder(w).Encode(map[string]string{"error": "rate_limit_exceeded"})
            return
        }
        next.ServeHTTP(w, r)
    })
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        wrapped := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
        next.ServeHTTP(wrapped, r)
        log.Printf("%s %s → %d (%v)", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
    })
}

// ---- Proxy Setup ----

func newBackendHandler(target string, cb *CircuitBreaker) http.Handler {
    u, _ := url.Parse(target)
    proxy := httputil.NewSingleHostReverseProxy(u)

    proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
        cb.RecordFailure()
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadGateway)
        json.NewEncoder(w).Encode(map[string]string{"error": "bad_gateway"})
    }

    proxy.ModifyResponse = func(resp *http.Response) error {
        if resp.StatusCode < 500 {
            cb.RecordSuccess()
        } else {
            cb.RecordFailure()
        }
        return nil
    }

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if cb.IsOpen() {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusServiceUnavailable)
            json.NewEncoder(w).Encode(map[string]string{"error": "circuit_open"})
            return
        }
        proxy.ServeHTTP(w, r)
    })
}

// ---- Main ----

func main() {
    mux := http.NewServeMux()

    cb1 := NewCircuitBreaker(3, 15*time.Second)
    cb2 := NewCircuitBreaker(3, 15*time.Second)

    mux.Handle("/svc1/", http.StripPrefix("/svc1", newBackendHandler("http://localhost:9001", cb1)))
    mux.Handle("/svc2/", http.StripPrefix("/svc2", newBackendHandler("http://localhost:9002", cb2)))

    srv := &http.Server{
        Addr:         ":8080",
        Handler:      loggingMiddleware(rateLimitMiddleware(mux)),
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    go func() {
        log.Println("gateway on :8080")
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("server error: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("shutting down...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("forced shutdown: %v", err)
    }
    log.Println("server stopped cleanly")
}
```

</details>

---

## Summary

| Concept | Key Point |
|---------|-----------|
| Per-IP rate limiting | Map of `*rate.Limiter` keyed by IP; needs cleanup in production |
| Health checks | Periodic goroutine hitting `/health`; marks backends alive/dead |
| Circuit breaker | Tracks failures; opens after threshold; half-open after cooldown |
| `statusRecorder` | Embed `ResponseWriter`, override `WriteHeader` to capture status |
| Middleware order | Logging (outer) → rate limit → mux → circuit breaker → proxy |
| Full gateway | Combines all patterns from lessons 3, 6, and 7 into one service |
