# Lesson 1: net/http Fundamentals

## The Big Picture

Go's `net/http` package is powerful enough to run production servers — no framework needed. Most Go web services use the standard library directly, maybe with a lightweight router on top. That's it.

Three things to understand:
1. **Handlers** — functions that process requests
2. **ServeMux** — a router that maps URLs to handlers
3. **Server** — listens for connections and dispatches to handlers

---

## Your First Server

```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, World!")
    })

    fmt.Println("Server running on :8080")
    http.ListenAndServe(":8080", nil)
}
```

Run it, then visit `http://localhost:8080/hello` in your browser (or `curl localhost:8080/hello`).

**What's happening:**
- `http.HandleFunc` registers a function for a URL pattern on the **default ServeMux**
- `http.ListenAndServe(":8080", nil)` starts listening — `nil` means "use the default ServeMux"
- Every request to `/hello` calls your function

---

## The Handler Interface

At the core of everything is one interface:

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

That's it. Anything with a `ServeHTTP` method is a handler. `http.HandleFunc` is just a convenience that wraps a plain function to satisfy this interface.

You can implement it yourself:

```go
type greeting struct {
    name string
}

func (g greeting) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %s!", g.name)
}

func main() {
    http.Handle("/greet", greeting{name: "Gopher"})
    http.ListenAndServe(":8080", nil)
}
```

**Why does this matter?** Middleware, routers, entire web frameworks — they're all just `Handler` implementations. Understanding this interface unlocks everything that comes later.

---

## Request and ResponseWriter

Your handler gets two things:

### `*http.Request` — everything about the incoming request

```go
func handler(w http.ResponseWriter, r *http.Request) {
    r.Method           // "GET", "POST", etc.
    r.URL.Path         // "/hello"
    r.URL.Query()      // query params as url.Values
    r.Header.Get("X-Custom")  // header value
    r.Body             // io.ReadCloser for POST/PUT body
}
```

### `http.ResponseWriter` — how you send a response back

```go
func handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")  // Set headers FIRST
    w.WriteHeader(http.StatusCreated)                     // Then status code
    w.Write([]byte(`{"status": "created"}`))             // Then body
}
```

**Order matters:**
1. Set headers
2. Write status code (optional — defaults to 200)
3. Write body

Once you call `Write()` or `WriteHeader()`, headers are sent and can't be changed.

---

## Creating Your Own ServeMux

Using the default global ServeMux works, but creating your own is better practice — it avoids global state:

```go
func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello!")
    })

    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })

    fmt.Println("Server running on :8080")
    http.ListenAndServe(":8080", mux)  // Pass your mux instead of nil
}
```

---

## Handling Errors from ListenAndServe

`ListenAndServe` returns an error (e.g., port already in use). Always handle it:

```go
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/hello", helloHandler)

    fmt.Println("Server running on :8080")
    if err := http.ListenAndServe(":8080", mux); err != nil {
        fmt.Println("Server error:", err)
    }
}
```

You'll often see `log.Fatal(http.ListenAndServe(...))` — this logs the error and exits.

---

## Your Turn

Try running the code in `main.go`. It sets up a server with three routes.

1. Run it with `go run main.go`
2. In another terminal, test each endpoint:
   - `curl localhost:8080/`
   - `curl localhost:8080/health`
   - `curl -X POST localhost:8080/echo -d '{"message": "hi"}'`
3. Try `curl localhost:8080/notaroute` — what happens?

**Questions to think about:**
1. What HTTP status code does `/health` return? (You didn't set one explicitly)
2. What happens if you move `w.WriteHeader(201)` *after* `w.Write(...)` in the echo handler?
3. Why do we use `log.Fatal` instead of just ignoring the error from `ListenAndServe`?

### Answers

1. **200 OK** — if you never call `WriteHeader()`, the first call to `Write()` automatically sends a 200.

2. **The 201 is ignored** — you'll see a warning in the server logs: `http: superfluous response.WriteHeader call`. The status was already sent as 200 when `Write()` was called. Header order matters!

3. **Because the server is the whole point of the program.** If it can't start (port taken, permissions issue), there's nothing useful the program can do. `log.Fatal` logs the error and exits with code 1. Silently ignoring it would leave you with a program that appears to run but isn't serving anything.

---

## Summary

| Concept | What It Does |
|---------|-------------|
| `http.HandleFunc(pattern, fn)` | Register a function as a handler on the default mux |
| `http.Handle(pattern, handler)` | Register a `Handler` implementation |
| `http.ListenAndServe(addr, handler)` | Start the server (`nil` = default mux) |
| `http.NewServeMux()` | Create your own router (preferred over default) |
| `http.ResponseWriter` | Write headers, status, and body to the response |
| `*http.Request` | Access method, URL, headers, body, query params |
| `Handler` interface | Just `ServeHTTP(ResponseWriter, *Request)` — everything builds on this |
