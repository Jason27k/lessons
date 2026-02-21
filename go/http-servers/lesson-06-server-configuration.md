# Lesson 6: Server Configuration

## The Problem

You've been starting servers with `http.ListenAndServe(":8080", mux)` — a one-liner that hides a lot of defaults. In production, those defaults will burn you:

- A slow client can hold a connection open **forever** (no read/write timeouts)
- Deploying a new version **drops in-flight requests** (no graceful shutdown)
- Traffic is **unencrypted** (no TLS)

This lesson is about taking control of the `http.Server` struct.

---

## 1. The `http.Server` Struct

Instead of `http.ListenAndServe`, create a server explicitly:

```go
srv := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}

log.Fatal(srv.ListenAndServe())
```

That's it — same behavior, but now you control the knobs.

---

## 2. Timeouts — What Each One Does

```
Client connects
    │
    ├─── ReadTimeout ────────────┐
    │    (reading request        │
    │     headers + body)        │
    │                            │
    ├─── ReadHeaderTimeout ──┐   │
    │    (just headers)      │   │
    │                        │   │
    ├─── WriteTimeout ───────────────┐
    │    (from end of request        │
    │     header read to end         │
    │     of response write)         │
    │                                │
    └─── IdleTimeout ────────────────────┐
         (between keep-alive requests)   │
```

| Timeout | Default | What happens if exceeded |
|---------|---------|------------------------|
| `ReadTimeout` | none (∞) | Connection closed — protects against slowloris |
| `ReadHeaderTimeout` | none (∞) | Connection closed — lighter alternative to ReadTimeout |
| `WriteTimeout` | none (∞) | Connection closed — protects against slow readers |
| `IdleTimeout` | none (∞) | Keep-alive connection closed — frees resources |

**Rule of thumb:** Always set at least `ReadTimeout` and `WriteTimeout`. Without them, a single malicious client can exhaust your server's file descriptors.

---

## 3. Graceful Shutdown

The problem: `ListenAndServe` stops immediately when you kill the process. Any request mid-flight gets a broken connection.

The fix: `Shutdown(ctx)` — stops accepting new connections, waits for in-flight requests to finish, then returns.

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
        log.Println("slow request started")
        time.Sleep(5 * time.Second) // simulate work
        w.Write([]byte("done\n"))
        log.Println("slow request finished")
    })

    srv := &http.Server{
        Addr:         ":8080",
        Handler:      mux,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    // Start server in a goroutine
    go func() {
        log.Println("server starting on :8080")
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("server error: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("shutting down...")

    // Give in-flight requests up to 30 seconds to finish
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("forced shutdown: %v", err)
    }

    log.Println("server stopped cleanly")
}
```

### The pattern step-by-step:

1. **Run `ListenAndServe` in a goroutine** — so the main goroutine is free to wait for signals
2. **Check for `http.ErrServerClosed`** — that's the normal return from `Shutdown`, not a real error
3. **Wait for SIGINT/SIGTERM** — ctrl+C or `kill <pid>`
4. **Call `Shutdown(ctx)` with a deadline** — gives in-flight requests a grace period
5. **If the deadline passes**, `Shutdown` returns an error and you can force-exit

---

## 4. TLS/HTTPS

For local dev and testing, generate a self-signed cert:

```bash
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem \
  -days 365 -nodes -subj '/CN=localhost'
```

Then swap `ListenAndServe` for `ListenAndServeTLS`:

```go
srv := &http.Server{
    Addr:         ":8443",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
}

log.Fatal(srv.ListenAndServeTLS("cert.pem", "key.pem"))
```

That's all it takes. In production, you'd typically use Let's Encrypt via `golang.org/x/crypto/acme/autocert`:

```go
m := &autocert.Manager{
    Cache:      autocert.DirCache("cert-cache"),
    Prompt:     autocert.AcceptTOS,
    HostPolicy: autocert.HostWhitelist("example.com"),
}

srv := &http.Server{
    Addr:      ":443",
    Handler:   mux,
    TLSConfig: m.TLSConfig(),
}

log.Fatal(srv.ListenAndServeTLS("", ""))
```

---

## 5. `MaxHeaderBytes`

One more useful knob — limits the size of request headers:

```go
srv := &http.Server{
    Addr:           ":8080",
    Handler:        mux,
    MaxHeaderBytes: 1 << 20, // 1 MB (default is also 1 MB)
}
```

The default (1 MB) is fine for most cases. Lower it if you want to reject requests with absurdly large headers early.

---

## Key Rules

1. **Never use `http.ListenAndServe` in production** — always create an `http.Server` with timeouts
2. **`Shutdown` vs `Close`** — `Shutdown` is graceful (waits), `Close` is immediate (drops connections). Use `Shutdown`.
3. **`ErrServerClosed` is not an error** — it's the expected return when `Shutdown` is called
4. **WriteTimeout must exceed your longest handler** — if a handler takes 30s, a 10s WriteTimeout will kill it
5. **Signal buffer size 1** — `make(chan os.Signal, 1)` so the signal isn't lost if you're not yet listening

---

## Common Mistakes

| Mistake | Why it's wrong |
|---------|---------------|
| `http.ListenAndServe` with no timeouts | Slowloris attacks can DoS your server |
| Checking `err != nil` instead of `err != http.ErrServerClosed` | You'll log a "fatal" on every clean shutdown |
| Calling `Shutdown` with `context.Background()` (no deadline) | Shutdown waits forever if a request never finishes |
| Setting WriteTimeout shorter than handler duration | Your handler's response gets cut off |
| Forgetting to run `ListenAndServe` in a goroutine | Main goroutine blocks, never reaches shutdown logic |

---

## Your Turn

### Exercise 1: Add Timeouts

Take this bare server and add proper timeouts:

```go
package main

import (
    "net/http"
    "time"
)

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(2 * time.Second) // simulate work
        w.Write([]byte("hello\n"))
    })

    // TODO: Create an http.Server with:
    //   - ReadTimeout: 5s
    //   - WriteTimeout: 10s
    //   - IdleTimeout: 120s
    // Then start it.
}
```

### Exercise 2: Graceful Shutdown

Extend your server with graceful shutdown. Test it:
1. Start the server
2. In another terminal, run: `curl http://localhost:8080/`
3. While the request is in-flight (during the 2s sleep), press ctrl+C
4. Verify the curl request completes successfully and the server logs "server stopped cleanly"

### Exercise 3: Timeout Conflict

What happens if you set `WriteTimeout: 1 * time.Second` but your handler sleeps for 2 seconds? Try it and observe the behavior. What does the client see?

---

<details>
<summary>Exercise 1 Answer</summary>

```go
srv := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}

log.Fatal(srv.ListenAndServe())
```

</details>

<details>
<summary>Exercise 2 Answer</summary>

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(2 * time.Second)
        w.Write([]byte("hello\n"))
    })

    srv := &http.Server{
        Addr:         ":8080",
        Handler:      mux,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    go func() {
        log.Println("listening on :8080")
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

<details>
<summary>Exercise 3 Answer</summary>

The client gets an **empty response** (or connection reset). The server closes the connection after 1 second because `WriteTimeout` expired before the handler called `w.Write()`. The handler's `Write` call silently fails. The lesson: `WriteTimeout` must be longer than your slowest handler.

</details>

---

## Summary

| Concept | Key Point |
|---------|-----------|
| `http.Server` | Always create one explicitly — don't rely on `ListenAndServe` defaults |
| `ReadTimeout` | Protects against slow/malicious request senders |
| `WriteTimeout` | Protects against slow readers; must exceed handler duration |
| `IdleTimeout` | Cleans up idle keep-alive connections |
| Graceful shutdown | `Shutdown(ctx)` waits for in-flight requests; use signal.Notify |
| `ErrServerClosed` | Normal return from `Shutdown` — not a real error |
| TLS | `ListenAndServeTLS` for self-signed; `autocert` for production |
