# Lesson 9: Testing HTTP Servers

## The Problem

You've built handlers, middleware, and a full gateway. But how do you know they actually work without manually curling every endpoint? You need automated tests — and Go's `net/http/httptest` package makes this surprisingly easy. No need to start a real server or bind to a port.

---

## 1. `httptest.NewRecorder` — Testing Handlers Directly

The core idea: create a fake `ResponseWriter`, pass it to your handler, then inspect what was written.

```go
func TestHealthHandler(t *testing.T) {
    // Create a request
    req := httptest.NewRequest("GET", "/health", nil)

    // Create a recorder (implements http.ResponseWriter)
    rec := httptest.NewRecorder()

    // Call the handler directly — no server needed
    healthHandler(rec, req)

    // Check the response
    if rec.Code != http.StatusOK {
        t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
    }

    if rec.Body.String() != `{"status":"ok"}` {
        t.Errorf("got body %q, want %q", rec.Body.String(), `{"status":"ok"}`)
    }
}
```

`httptest.NewRecorder()` returns a `*httptest.ResponseRecorder` which captures:
- `Code` — the status code
- `Body` — a `*bytes.Buffer` with the response body
- `Header()` — the response headers

This is the bread and butter of handler testing. No ports, no goroutines, no cleanup.

---

## 2. `httptest.NewRequest` vs `http.NewRequest`

Both create requests, but there's a subtle difference:

```go
// httptest.NewRequest — panics on error, no need for error check
req := httptest.NewRequest("GET", "/users/42", nil)

// http.NewRequest — returns an error (use in production code)
req, err := http.NewRequest("GET", "/users/42", nil)
```

Use `httptest.NewRequest` in tests — it's cleaner because tests should panic on malformed test data anyway.

### Adding headers, query params, context:

```go
// Headers
req := httptest.NewRequest("GET", "/api/data", nil)
req.Header.Set("Authorization", "Bearer abc123")

// Query params (just put them in the URL)
req := httptest.NewRequest("GET", "/search?q=golang&page=2", nil)

// JSON body
body := strings.NewReader(`{"name":"Alice","age":30}`)
req := httptest.NewRequest("POST", "/users", body)
req.Header.Set("Content-Type", "application/json")
```

---

## 3. Testing with a Mux (Route Parameters)

If your handler relies on path parameters from Go 1.22+ routing, you need to register it on a mux first:

```go
func userHandler(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    fmt.Fprintf(w, "user: %s", id)
}

func TestUserHandler(t *testing.T) {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /users/{id}", userHandler)

    req := httptest.NewRequest("GET", "/users/42", nil)
    rec := httptest.NewRecorder()

    mux.ServeHTTP(rec, req)

    if rec.Body.String() != "user: 42" {
        t.Errorf("got %q, want %q", rec.Body.String(), "user: 42")
    }
}
```

**Key point:** if you call `userHandler(rec, req)` directly, `r.PathValue("id")` returns `""` because no mux parsed the path. Route the request through the mux in your test.

---

## 4. `httptest.NewServer` — Integration Tests

Sometimes you want a real HTTP server — to test the full stack including middleware, or to test code that makes HTTP calls (like your gateway proxying to a backend).

```go
func TestFullStack(t *testing.T) {
    // Start a real test server
    mux := http.NewServeMux()
    mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("world"))
    })

    ts := httptest.NewServer(mux)
    defer ts.Close() // shuts down the server when the test ends

    // Make a real HTTP request
    resp, err := http.Get(ts.URL + "/hello")
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    if string(body) != "world" {
        t.Errorf("got %q, want %q", string(body), "world")
    }
}
```

`httptest.NewServer`:
- Starts a real server on a random available port
- `ts.URL` gives you the base URL (e.g., `http://127.0.0.1:54321`)
- `ts.Close()` shuts it down — always defer this
- No port conflicts between parallel tests

---

## 5. Testing Middleware

Middleware wraps a handler, so test it by wrapping a known handler and checking what changes:

```go
func TestLoggingMiddleware(t *testing.T) {
    // Inner handler that returns 201
    inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusCreated)
        w.Write([]byte("created"))
    })

    // Wrap with middleware
    handler := loggingMiddleware(inner)

    req := httptest.NewRequest("POST", "/things", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    // Verify the middleware didn't interfere with the response
    if rec.Code != http.StatusCreated {
        t.Errorf("got status %d, want %d", rec.Code, http.StatusCreated)
    }
    if rec.Body.String() != "created" {
        t.Errorf("got body %q, want %q", rec.Body.String(), "created")
    }
}
```

For auth middleware, test both the happy path and the rejection:

```go
func TestAuthMiddleware(t *testing.T) {
    inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("secret"))
    })
    handler := authMiddleware(inner)

    t.Run("no token", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/secret", nil)
        rec := httptest.NewRecorder()
        handler.ServeHTTP(rec, req)

        if rec.Code != http.StatusUnauthorized {
            t.Errorf("got %d, want 401", rec.Code)
        }
    })

    t.Run("valid token", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/secret", nil)
        req.Header.Set("Authorization", "Bearer valid-token")
        rec := httptest.NewRecorder()
        handler.ServeHTTP(rec, req)

        if rec.Code != http.StatusOK {
            t.Errorf("got %d, want 200", rec.Code)
        }
    })
}
```

### `t.Run` — Subtests

`t.Run("name", func(t *testing.T) { ... })` creates a named subtest. Benefits:
- Output shows which case failed: `--- FAIL: TestAuthMiddleware/no_token`
- Run a single subtest: `go test -run TestAuthMiddleware/valid_token`
- Each subtest gets its own `t`, so `t.Fatalf` only stops that subtest

---

## 6. Table-Driven Tests

The Go testing pattern you'll see everywhere. Instead of writing separate test functions, define a table of inputs and expected outputs:

```go
func TestStatusHandler(t *testing.T) {
    tests := []struct {
        name       string
        method     string
        path       string
        wantCode   int
        wantBody   string
    }{
        {"health check", "GET", "/health", 200, `{"status":"ok"}`},
        {"not found", "GET", "/nope", 404, "404 page not found\n"},
        {"method not allowed", "POST", "/health", 405, "Method Not Allowed\n"},
    }

    mux := setupRoutes() // your function that returns a configured mux

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(tt.method, tt.path, nil)
            rec := httptest.NewRecorder()

            mux.ServeHTTP(rec, req)

            if rec.Code != tt.wantCode {
                t.Errorf("got status %d, want %d", rec.Code, tt.wantCode)
            }
            if rec.Body.String() != tt.wantBody {
                t.Errorf("got body %q, want %q", rec.Body.String(), tt.wantBody)
            }
        })
    }
}
```

Why table-driven:
- Adding a new test case = adding one struct — no new function
- Consistent assertion logic — if your check is wrong, fix it once
- Output is clean: `--- FAIL: TestStatusHandler/not_found`

---

## 7. Faking a Backend with `httptest.NewServer`

This is where it gets powerful. Remember your gateway that proxies to backends? Test it by creating fake backends:

```go
func TestGatewayProxy(t *testing.T) {
    // Fake backend
    backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"from":"backend","path":"` + r.URL.Path + `"}`))
    }))
    defer backend.Close()

    // Point gateway at fake backend instead of real one
    backendURL, _ := url.Parse(backend.URL)
    proxy := httputil.NewSingleHostReverseProxy(backendURL)

    mux := http.NewServeMux()
    mux.Handle("/svc1/", http.StripPrefix("/svc1", proxy))

    // Request through gateway
    req := httptest.NewRequest("GET", "/svc1/data", nil)
    rec := httptest.NewRecorder()
    mux.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("got status %d", rec.Code)
    }

    // Verify path was stripped correctly
    var resp map[string]string
    json.NewDecoder(rec.Body).Decode(&resp)
    if resp["path"] != "/data" {
        t.Errorf("backend saw path %q, want /data", resp["path"])
    }
}
```

This tests that `StripPrefix` works correctly — the backend should see `/data`, not `/svc1/data`.

---

## 8. Testing JSON Request/Response

A helper pattern for JSON handlers:

```go
func TestCreateUser(t *testing.T) {
    // Build JSON request
    reqBody := map[string]any{"name": "Alice", "age": 30}
    bodyBytes, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/users", bytes.NewReader(bodyBytes))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    createUserHandler(rec, req)

    // Check status
    if rec.Code != http.StatusCreated {
        t.Fatalf("got status %d, want 201", rec.Code)
    }

    // Parse JSON response
    var resp map[string]any
    if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
        t.Fatalf("invalid JSON response: %v", err)
    }

    if resp["name"] != "Alice" {
        t.Errorf("got name %v, want Alice", resp["name"])
    }
}
```

---

## Key Rules

1. **`NewRecorder` for unit tests** — fast, no ports, test handlers in isolation
2. **`NewServer` for integration tests** — real HTTP, test the full stack
3. **Route through the mux** if your handler uses `PathValue` — calling the handler directly won't parse path params
4. **Table-driven tests** — the Go way; one struct per case, loop with `t.Run`
5. **Fake backends with `NewServer`** — point your proxy at a test server instead of a real one
6. **Always `defer ts.Close()`** — test servers leak goroutines if not closed

---

## Common Mistakes

| Mistake | Why it's wrong |
|---------|---------------|
| Calling handler directly when it uses `PathValue` | Path params aren't parsed without a mux — you get empty strings |
| Forgetting `defer ts.Close()` | Test server goroutines leak, eventually hitting resource limits |
| Checking `rec.Code` without handler calling `WriteHeader` | Default is 200, which may mask bugs — be explicit in handlers |
| Not setting `Content-Type` on POST requests | Handler's JSON decoder may work but it's not testing the real flow |
| Using `http.NewRequest` instead of `httptest.NewRequest` in tests | Adds unnecessary error handling boilerplate |

---

## Your Turn

### Exercise: Test Your Gateway Components

Write tests in a file called `main_test.go` with these test cases:

1. **`TestRateLimiter`** — Create a `RateLimiter` with rate 1, burst 2. Call `GetLimiter("1.2.3.4").Allow()` — first two calls should return `true` (burst), third should return `false`.

2. **`TestCircuitBreaker`** — Create a `CircuitBreaker` with threshold 2 and a short cooldown (100ms). Record 2 failures → `IsOpen()` should be `true`. Wait for cooldown → `IsOpen()` should be `false` (half-open). Record a success → `IsOpen()` should be `false` (closed).

3. **`TestLoggingMiddleware`** — Wrap a handler that returns 201 with `loggingMiddleware`. Verify the response code and body pass through unchanged.

4. **`TestCircuitBreakerIntegration`** — Use `httptest.NewServer` to create a backend that returns 502. Set up a proxy with a circuit breaker (threshold 2). Make requests through `gatewayHandler` until the circuit opens. Verify the response changes from "bad_gateway" to "service_unavailable".

Use table-driven tests where it makes sense.

---

<details>
<summary>Full Answer</summary>

```go
package main

import (
    "net/http"
    "net/http/httptest"
    "net/http/httputil"
    "net/url"
    "testing"
    "time"
)

// ---- Test 1: Rate Limiter ----

func TestRateLimiter(t *testing.T) {
    rl := NewRateLimiter(1, 2) // 1 req/sec, burst of 2

    lim := rl.GetLimiter("1.2.3.4")

    if !lim.Allow() {
        t.Error("first request should be allowed (burst)")
    }
    if !lim.Allow() {
        t.Error("second request should be allowed (burst)")
    }
    if lim.Allow() {
        t.Error("third request should be rejected (burst exhausted)")
    }

    // Same IP returns same limiter
    lim2 := rl.GetLimiter("1.2.3.4")
    if lim != lim2 {
        t.Error("same IP should return the same limiter")
    }

    // Different IP gets a fresh limiter
    lim3 := rl.GetLimiter("5.6.7.8")
    if !lim3.Allow() {
        t.Error("new IP should be allowed")
    }
}

// ---- Test 2: Circuit Breaker ----

func TestCircuitBreaker(t *testing.T) {
    cb := NewCircuitBreaker(2, 100*time.Millisecond)

    // Closed initially
    if cb.IsOpen() {
        t.Fatal("circuit should start closed")
    }

    // Record failures to open it
    cb.RecordFailure()
    cb.RecordFailure()

    if !cb.IsOpen() {
        t.Fatal("circuit should be open after 2 failures")
    }

    // Wait for cooldown → half-open
    time.Sleep(150 * time.Millisecond)
    if cb.IsOpen() {
        t.Fatal("circuit should be half-open after cooldown")
    }

    // Record success → closed
    cb.RecordSuccess()
    if cb.IsOpen() {
        t.Fatal("circuit should be closed after success")
    }
}

// ---- Test 3: Logging Middleware ----

func TestLoggingMiddleware(t *testing.T) {
    inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusCreated)
        w.Write([]byte("created"))
    })

    handler := loggingMiddleware(inner)

    req := httptest.NewRequest("POST", "/things", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusCreated {
        t.Errorf("got status %d, want %d", rec.Code, http.StatusCreated)
    }
    if rec.Body.String() != "created" {
        t.Errorf("got body %q, want %q", rec.Body.String(), "created")
    }
}

// ---- Test 4: Circuit Breaker Integration ----

func TestCircuitBreakerIntegration(t *testing.T) {
    // Backend that always returns 502
    backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadGateway)
    }))
    defer backend.Close()

    backendURL, _ := url.Parse(backend.URL)
    cb := NewCircuitBreaker(2, 5*time.Second)
    proxy := httputil.NewSingleHostReverseProxy(backendURL)

    proxy.ModifyResponse = func(resp *http.Response) error {
        if resp.StatusCode < 500 {
            cb.RecordSuccess()
        } else {
            cb.RecordFailure()
        }
        return nil
    }

    handler := gatewayHandler(proxy, cb)

    // First two requests: backend returns 502, circuit breaker records failures
    for i := 0; i < 2; i++ {
        req := httptest.NewRequest("GET", "/", nil)
        rec := httptest.NewRecorder()
        handler.ServeHTTP(rec, req)

        if rec.Code != http.StatusBadGateway {
            t.Errorf("request %d: got status %d, want 502", i+1, rec.Code)
        }
    }

    // Third request: circuit is open, should get 503 without hitting backend
    req := httptest.NewRequest("GET", "/", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusServiceUnavailable {
        t.Errorf("got status %d, want 503 (circuit open)", rec.Code)
    }
}
```

</details>

---

## Summary

| Concept | Key Point |
|---------|-----------|
| `httptest.NewRecorder` | Fake `ResponseWriter` — captures code, body, headers without a real server |
| `httptest.NewRequest` | Like `http.NewRequest` but panics on error — cleaner for tests |
| `httptest.NewServer` | Real test server on random port — use for integration tests and fake backends |
| Table-driven tests | Slice of structs + `t.Run` loop — the standard Go testing pattern |
| `t.Run` subtests | Named sub-cases, individually runnable, clean failure output |
| Testing middleware | Wrap a known handler, verify the response passes through correctly |
| Fake backends | `httptest.NewServer` as a stand-in for real backends when testing proxies |
