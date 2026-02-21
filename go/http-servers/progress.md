# Learning Progress - HTTP Servers & API Gateway

## Session: 2026-02-06

### Completed

- [x] net/http Fundamentals (lesson-01-net-http-fundamentals.md)
- [x] Routing (lesson-02-routing.md)
- [x] Middleware (lesson-03-middleware.md)
- [x] Request Handling
- [x] Response Patterns
- [x] Server Configuration
- [x] API Gateway Concepts
- [x] Building a Simple API Gateway
- [x] Testing HTTP Servers

### Notes

- Learner experimented with WriteHeader ordering (saw the superfluous WriteHeader warning)
- Noticed `/` catch-all behavior, which set up the `{$}` exact-match lesson nicely
- Asked about gorilla/mux status and Go router ecosystem — discussed chi, gin, echo, fiber tradeoffs
- Sticking with stdlib for learning; chi recommended as first step if stdlib isn't enough

---

### Key Takeaways

1. **Handler interface** — `ServeHTTP(ResponseWriter, *Request)` is the foundation; everything else builds on it
2. **ResponseWriter ordering** — set headers, then status code, then body. Once you `Write()`, headers are locked
3. **Own your mux** — `http.NewServeMux()` over the global default to avoid shared state
4. **Go 1.22 routing** — method prefixes (`"GET /users"`), path params (`{id}`), exact match (`{$}`), catch-all (`{path...}`)
5. **Precedence** — literal segments beat wildcards; conflicts panic at registration, not runtime
6. **Always use `http.Server`** — explicit struct with timeouts; never bare `ListenAndServe` in production
7. **Graceful shutdown pattern** — `ListenAndServe` in goroutine, signal.Notify on main, `Shutdown(ctx)` with deadline
8. **WriteTimeout must exceed handler duration** — otherwise the response gets silently dropped
9. **`httputil.ReverseProxy`** — stdlib reverse proxy; customize via Director, ModifyResponse, ErrorHandler
10. **StripPrefix for gateway routing** — gateway paths differ from backend paths; strip the prefix before forwarding
11. **Token bucket rate limiting** — `x/time/rate` for global limiting; per-client needs a map of limiters
12. **Per-IP rate limiting** — map of `*rate.Limiter` keyed by IP; needs cleanup goroutine in production
13. **Circuit breaker** — three states (closed/open/half-open); complements periodic health checks with real-time failure tracking
14. **`statusRecorder` pattern** — embed `ResponseWriter`, override `WriteHeader` to capture status codes for logging
15. **`httptest.NewRecorder`** — fake ResponseWriter for unit testing handlers without a real server
16. **`httptest.NewServer`** — real test server on random port for integration tests and faking backends
17. **Table-driven tests with `t.Run`** — idiomatic Go pattern; one struct per case, clean failure output
