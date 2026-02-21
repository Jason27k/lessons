# Lesson 7: API Gateway Concepts

## The Problem

You have multiple backend services — a user service, a product service, a billing service. Clients shouldn't need to know each service's address. You want:

- **One entry point** for all clients
- **Cross-cutting concerns** (auth, rate limiting, logging) in one place
- **Backend isolation** — swap, scale, or kill services without clients noticing

That's what an API gateway does. And Go's stdlib has most of what you need built in.

---

## 1. Reverse Proxy — `httputil.ReverseProxy`

A reverse proxy takes an incoming request, forwards it to a backend, and streams the response back. Go gives you `httputil.ReverseProxy`:

```go
package main

import (
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
)

func main() {
    // The backend we're proxying to
    target, _ := url.Parse("http://localhost:9001")

    proxy := httputil.NewSingleHostReverseProxy(target)

    http.Handle("/", proxy)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

That's a fully working reverse proxy. Every request to `:8080` gets forwarded to `:9001`.

### What `NewSingleHostReverseProxy` does under the hood:

1. Copies the incoming request
2. Rewrites the URL scheme, host, and path to point at the target
3. Forwards hop-by-hop headers appropriately
4. Streams the backend's response back to the client

### Customizing the proxy

The `Director` function lets you modify requests before they're forwarded:

```go
proxy := httputil.NewSingleHostReverseProxy(target)

// The default Director already sets scheme, host, path.
// Wrap it to add your own logic.
originalDirector := proxy.Director
proxy.Director = func(req *http.Request) {
    originalDirector(req)
    req.Header.Set("X-Forwarded-By", "my-gateway")
    req.Host = target.Host // preserve target's Host header
}
```

`ModifyResponse` lets you modify responses before they go back to the client:

```go
proxy.ModifyResponse = func(resp *http.Response) error {
    resp.Header.Set("X-Gateway", "true")
    // Return an error here to trigger ErrorHandler instead
    return nil
}
```

`ErrorHandler` handles backend failures:

```go
proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
    log.Printf("proxy error: %v", err)
    http.Error(w, "service unavailable", http.StatusBadGateway)
}
```

---

## 2. Routing to Multiple Backends

A gateway routes different paths to different services:

```go
func main() {
    mux := http.NewServeMux()

    usersSvc, _ := url.Parse("http://localhost:9001")
    productsSvc, _ := url.Parse("http://localhost:9002")

    mux.Handle("/users/", httputil.NewSingleHostReverseProxy(usersSvc))
    mux.Handle("/products/", httputil.NewSingleHostReverseProxy(productsSvc))

    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

Request to `/users/123` → forwarded to `http://localhost:9001/users/123`
Request to `/products/456` → forwarded to `http://localhost:9002/products/456`

---

## 3. Path Stripping

Sometimes you want the gateway path to differ from the backend path. For example, `/api/v1/users` on the gateway should hit `/users` on the backend:

```go
func stripPrefix(prefix string, proxy http.Handler) http.Handler {
    return http.StripPrefix(prefix, proxy)
}

mux.Handle("/api/v1/users/", stripPrefix("/api/v1", usersProxy))
```

`http.StripPrefix` is stdlib — it removes the prefix before the request reaches the proxy.

---

## 4. Rate Limiting

Rate limiting protects backends from getting overwhelmed. The simplest approach uses `golang.org/x/time/rate` (a token bucket):

```go
package main

import (
    "net/http"
    "golang.org/x/time/rate"
)

// Global rate limiter: 10 requests/sec, burst of 20
var limiter = rate.NewLimiter(10, 20)

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

This is a **global** limiter — all clients share it. For per-client limiting, you'd use a map of limiters keyed by IP (we'll build this in the next lesson).

### Token bucket in 30 seconds

- Bucket holds up to `burst` tokens
- Tokens are added at `rate` per second
- Each request takes one token
- No tokens left → rejected (429)

---

## 5. Load Balancing

When a backend has multiple instances, the gateway picks which one to send traffic to. Here's a simple round-robin:

```go
type RoundRobinProxy struct {
    targets []*url.URL
    next    uint64
}

func (rr *RoundRobinProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Atomically get next index
    idx := atomic.AddUint64(&rr.next, 1)
    target := rr.targets[idx%uint64(len(rr.targets))]

    proxy := httputil.NewSingleHostReverseProxy(target)
    proxy.ServeHTTP(w, r)
}
```

Common strategies:

| Strategy | How it works | Best for |
|----------|-------------|----------|
| Round-robin | Cycle through backends | Equal-capacity servers |
| Random | Pick a random backend | Simple, decent distribution |
| Least connections | Track active connections, pick lowest | Varying request durations |
| Weighted | Assign weights to backends | Mixed-capacity servers |

---

## 6. Auth at the Gateway

Instead of each backend validating tokens, do it once at the gateway:

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }

        // Validate token (check JWT, call auth service, etc.)
        if !isValid(token) {
            http.Error(w, "forbidden", http.StatusForbidden)
            return
        }

        // Optionally pass user info to backends via headers
        r.Header.Set("X-User-ID", extractUserID(token))

        next.ServeHTTP(w, r)
    })
}
```

Backends trust the gateway. They read `X-User-ID` from the header instead of validating tokens themselves.

---

## Key Rules

1. **`httputil.ReverseProxy` is the foundation** — don't write raw HTTP forwarding yourself
2. **Wrap the Director, don't replace it** — the default Director handles URL rewriting; add to it
3. **Always set `ErrorHandler`** — the default panics or writes a bare 502
4. **Host header matters** — backends behind load balancers often need `req.Host = target.Host`
5. **Rate limiting goes in the gateway, not backends** — centralized control, one place to tune

---

## Common Mistakes

| Mistake | Why it's wrong |
|---------|---------------|
| Creating a new proxy per request | Wastes resources; create proxies once at startup |
| Ignoring hop-by-hop headers | `Connection`, `Keep-Alive` etc. shouldn't be forwarded — `ReverseProxy` handles this for you |
| Rate limiting per-server instead of per-client | One client can still overwhelm you; per-IP limiting is usually what you want |
| No timeout on proxy requests | A hung backend blocks the gateway; set `Transport` timeouts |
| Trusting `X-Forwarded-For` blindly | Clients can spoof it; only trust it from known proxies |

---

## Your Turn

### Exercise 1: Two-Backend Gateway

Build a gateway that routes to two backends. You'll run three processes:

**Backend 1** (`backend1/main.go` — conceptually, just use main.go):
```go
// Pretend this runs on :9001
mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello from backend 1\n"))
})
```

**Backend 2** (`run separately on :9002`):
```go
// Pretend this runs on :9002
mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello from backend 2\n"))
})
```

**Your gateway** (`:8080`):
- `/svc1/` routes to `:9001`
- `/svc2/` routes to `:9002`
- Add a custom `ErrorHandler` that returns a JSON error when a backend is down

Test: start only one backend and hit both routes. One should work, one should return your JSON error.

### Exercise 2: Add Rate Limiting

Add a global rate limiter middleware (5 req/sec, burst 10) to your gateway. Test by hammering it:

```bash
for i in $(seq 1 20); do curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/svc1/; done
```

You should see 200s followed by 429s.

### Exercise 3: Custom Headers

Modify the proxy's `Director` to add an `X-Request-ID` header (use a counter or `uuid` — a simple counter is fine). Then add a `ModifyResponse` that logs the backend's response status.

---

<details>
<summary>Exercise 1 Answer</summary>

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
)

func jsonError(w http.ResponseWriter, r *http.Request, err error) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadGateway)
    json.NewEncoder(w).Encode(map[string]string{
        "error":   "bad_gateway",
        "message": "backend unavailable",
    })
    log.Printf("proxy error for %s: %v", r.URL.Path, err)
}

func newProxy(target string) *httputil.ReverseProxy {
    u, _ := url.Parse(target)
    proxy := httputil.NewSingleHostReverseProxy(u)
    proxy.ErrorHandler = jsonError
    return proxy
}

func main() {
    mux := http.NewServeMux()

    mux.Handle("/svc1/", http.StripPrefix("/svc1", newProxy("http://localhost:9001")))
    mux.Handle("/svc2/", http.StripPrefix("/svc2", newProxy("http://localhost:9002")))

    log.Println("gateway on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

</details>

<details>
<summary>Exercise 2 Answer</summary>

```go
import "golang.org/x/time/rate"

var limiter = rate.NewLimiter(5, 10)

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusTooManyRequests)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "rate_limit_exceeded",
            })
            return
        }
        next.ServeHTTP(w, r)
    })
}

// In main():
log.Fatal(http.ListenAndServe(":8080", rateLimitMiddleware(mux)))
```

</details>

<details>
<summary>Exercise 3 Answer</summary>

```go
import (
    "fmt"
    "sync/atomic"
)

var requestID uint64

func newProxy(target string) *httputil.ReverseProxy {
    u, _ := url.Parse(target)
    proxy := httputil.NewSingleHostReverseProxy(u)

    originalDirector := proxy.Director
    proxy.Director = func(req *http.Request) {
        originalDirector(req)
        id := atomic.AddUint64(&requestID, 1)
        req.Header.Set("X-Request-ID", fmt.Sprintf("gw-%d", id))
    }

    proxy.ModifyResponse = func(resp *http.Response) error {
        log.Printf("backend responded: %d for %s", resp.StatusCode, resp.Request.URL.Path)
        return nil
    }

    proxy.ErrorHandler = jsonError
    return proxy
}
```

</details>

---

## Summary

| Concept | Key Point |
|---------|-----------|
| `httputil.ReverseProxy` | Stdlib reverse proxy — handles forwarding, headers, streaming |
| `Director` | Modify requests before forwarding (add headers, rewrite paths) |
| `ModifyResponse` | Modify responses before sending to client |
| `ErrorHandler` | Handle backend failures (default is a bare 502) |
| `http.StripPrefix` | Remove gateway path prefixes before forwarding |
| Rate limiting | `x/time/rate` token bucket — global or per-client |
| Load balancing | Round-robin, random, least connections, weighted |
| Gateway auth | Validate once at the gateway, pass user info via headers |
