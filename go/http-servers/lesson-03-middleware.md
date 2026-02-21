# Lesson 3: Middleware

## The Problem

Your API has 10 handlers. You want every one of them to:
- Log the request method and path
- Measure how long the handler took
- Recover from panics so one bad handler doesn't crash the whole server
- Check for a valid API key

Do you copy-paste that logic into all 10 handlers? Obviously not. You need a way to wrap behavior **around** handlers without touching the handlers themselves.

That's middleware.

---

## The Pattern

Middleware in Go is just a function that takes a handler and returns a new handler:

```go
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // do something BEFORE the handler
        next.ServeHTTP(w, r)
        // do something AFTER the handler
    })
}
```

That's it. The whole pattern. Everything else is just applying it.

Let's break down what's happening:
1. You receive `next` — the handler you're wrapping
2. You return a **new** handler (via `http.HandlerFunc` adapter)
3. Inside, you can run code before and/or after calling `next.ServeHTTP`

---

## A Real Example: Logging

```go
func logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}
```

Usage:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /health", healthHandler)

// Wrap the entire mux
server := logging(mux)
http.ListenAndServe(":8080", server)
```

Every request through `mux` now gets logged. The handlers don't know or care.

---

## Chaining Middleware

You'll usually want multiple middleware. Just nest them:

```go
server := logging(requireAuth(mux))
```

Request flow: **logging → requireAuth → your handler → requireAuth → logging**

It's like layers of an onion. The outermost middleware runs first on the way in and last on the way out.

For cleaner syntax, you can write a helper:

```go
func chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
    // Apply in reverse so the first middleware in the list runs first
    for i := len(middlewares) - 1; i >= 0; i-- {
        handler = middlewares[i](handler)
    }
    return handler
}

server := chain(mux, logging, requireAuth, recovery)
```

---

## Common Middleware Patterns

### Recovery (Panic Catcher)

A panicking handler shouldn't kill your server:

```go
func recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("PANIC: %v", err)
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### Auth Check

Short-circuit the chain by NOT calling `next`:

```go
func requireAPIKey(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        key := r.Header.Get("X-API-Key")
        if key != "secret123" {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return // Don't call next — request stops here
        }
        next.ServeHTTP(w, r)
    })
}
```

The key insight: **if you don't call `next.ServeHTTP`, the request never reaches the handler.** This is how auth, rate limiting, and validation middleware work.

### CORS

```go
func cors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return // Preflight request — respond and stop
        }

        next.ServeHTTP(w, r)
    })
}
```

---

## Per-Route vs Global Middleware

You can wrap **individual routes** instead of the whole mux:

```go
mux.Handle("GET /admin", requireAPIKey(http.HandlerFunc(adminHandler)))
mux.HandleFunc("GET /public", publicHandler)  // No auth needed
```

This is how you mix protected and public routes without a framework.

---

## Key Rules

1. **Middleware signature** is always `func(http.Handler) http.Handler` — this is the convention the entire ecosystem follows.
2. **Order matters** — `logging(auth(mux))` means logging runs first. If auth rejects, logging still sees the request.
3. **Don't call `next.ServeHTTP` twice** — that runs the handler twice.
4. **Headers must be set before `next.ServeHTTP`** if the handler writes the response (remember lesson 1: once you `Write()`, headers are locked).
5. **Recovery middleware should be outermost** — if it's inside another middleware that panics, it can't catch it.

---

## Your Turn

Run `main.go`. It has a server with three middleware (logging, recovery, auth) and a few routes.

1. `go run main.go`
2. Try these:
   - `curl localhost:8080/health` — does this need an API key?
   - `curl localhost:8080/secret` — what happens without a key?
   - `curl -H "X-API-Key: secret123" localhost:8080/secret` — now?
   - `curl localhost:8080/panic` — does the server crash?

**Exercises:**

1. Add a middleware called `requestID` that sets a `X-Request-ID` header on every response (use any string — even a counter is fine). Add it to the global chain.

2. The `/panic` endpoint triggers a panic. Check your terminal — does the recovery middleware log it? What status code does the client get?

3. Move `requireAPIKey` so it only protects `/secret` and `/admin`, but NOT `/health`. (Hint: per-route wrapping.)

### Answers

1. A simple `requestID` middleware:
```go
var counter atomic.Int64

func requestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := counter.Add(1)
        w.Header().Set("X-Request-ID", fmt.Sprintf("req-%d", id))
        next.ServeHTTP(w, r)
    })
}

// Add to chain:
server := chain(mux, logging, recovery, requestID)
```

2. Yes, the recovery middleware logs `PANIC: something went wrong` and the client gets a `500 Internal Server Error`. The server keeps running — only that single request is affected.

3. Per-route auth instead of global:
```go
// Remove requireAPIKey from the chain
server := chain(mux, logging, recovery)

// Wrap only the protected routes
mux.Handle("GET /secret", requireAPIKey(http.HandlerFunc(secretHandler)))
mux.Handle("GET /admin", requireAPIKey(http.HandlerFunc(adminHandler)))
mux.HandleFunc("GET /health", healthHandler)  // No auth
```

---

## Summary

| Concept | What It Means |
|---------|--------------|
| Middleware | A function: `func(http.Handler) http.Handler` |
| Chaining | Nest calls: `a(b(c(handler)))` — `a` runs first |
| Short-circuit | Don't call `next.ServeHTTP` to stop the chain |
| Global | Wrap the mux: `logging(mux)` |
| Per-route | Wrap one handler: `mux.Handle("/x", auth(handler))` |

| Common Middleware | Purpose |
|-------------------|---------|
| Logging | Log method, path, duration |
| Recovery | Catch panics, return 500 |
| Auth | Check credentials, reject if invalid |
| CORS | Set cross-origin headers, handle preflight |
| Request ID | Tag each request for tracing |
