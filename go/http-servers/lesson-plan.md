# HTTP Servers & API Gateway - Lesson Plan

## Prerequisites
- Go basics
- Goroutines and concurrency (completed)

## Topics

### 1. net/http Fundamentals
- `http.HandleFunc` and `http.Handle`
- The `Handler` interface
- `http.ListenAndServe`
- Request and ResponseWriter

### 2. Routing
- Default ServeMux limitations
- Path parameters and patterns (Go 1.22+)
- Popular routers: chi, gorilla/mux, httprouter

### 3. Middleware
- What middleware is and why it matters
- Writing middleware (function wrapping)
- Chaining middleware
- Common middleware: logging, auth, CORS, recovery

### 4. Request Handling
- Parsing JSON bodies
- Query parameters and path variables
- Headers and cookies
- Request validation

### 5. Response Patterns
- JSON responses
- Status codes
- Error handling patterns
- Streaming responses

### 6. Server Configuration
- Timeouts (read, write, idle)
- Graceful shutdown
- TLS/HTTPS

### 7. API Gateway Concepts
- What an API gateway does
- Reverse proxy with `httputil.ReverseProxy`
- Load balancing strategies
- Rate limiting
- Authentication/Authorization middleware
- Request/Response transformation

### 8. Building a Simple API Gateway
- Routing to multiple backends
- Health checks
- Circuit breaker pattern
- Logging and metrics

### 9. Testing HTTP Servers
- `httptest` package
- Testing handlers
- Integration tests

## Project Ideas
- REST API for a simple resource (CRUD)
- API gateway that proxies to multiple services
- Middleware library (auth, rate limit, logging)
