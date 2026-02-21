# Lesson 7: Context for Cancellation

## The Problem

Remember the done channel pattern?

```go
done := make(chan struct{})
go worker(done)
// ...
close(done)  // Signal cancellation
```

This works, but it gets messy:
- What if you need timeouts too?
- What if you have nested function calls that all need to stop?
- What if you need to pass request-scoped data?

---

## Enter context.Context

`context.Context` is the standard way to handle:
- **Cancellation** - stop work when no longer needed
- **Deadlines/Timeouts** - stop after a duration
- **Request-scoped values** - pass data through call chain

```go
import "context"
```

---

## Creating Contexts

### context.Background()

The root context - never canceled, no deadline, no values. Start here.

```go
ctx := context.Background()
```

### context.WithCancel()

Returns a context and a cancel function:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()  // Always call cancel to release resources

go worker(ctx)
// ...
cancel()  // Signals ctx.Done() channel
```

### context.WithTimeout()

Cancels automatically after a duration:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// ctx.Done() fires after 5 seconds OR when cancel() is called
```

### context.WithDeadline()

Cancels at a specific time:

```go
deadline := time.Now().Add(10 * time.Second)
ctx, cancel := context.WithDeadline(context.Background(), deadline)
defer cancel()
```

---

## Using Context in Goroutines

```go
func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            fmt.Println("worker: canceled:", ctx.Err())
            return
        default:
            // Do work
            fmt.Println("working...")
            time.Sleep(500 * time.Millisecond)
        }
    }
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    go worker(ctx)

    time.Sleep(3 * time.Second)  // Worker stops after 2s
}
```

---

## ctx.Done() and ctx.Err()

| Method | Returns |
|--------|---------|
| `ctx.Done()` | Channel that closes when context is canceled |
| `ctx.Err()` | `nil` if not canceled, `context.Canceled` or `context.DeadlineExceeded` |

```go
select {
case <-ctx.Done():
    if ctx.Err() == context.Canceled {
        fmt.Println("was canceled")
    } else if ctx.Err() == context.DeadlineExceeded {
        fmt.Println("timed out")
    }
}
```

---

## Context Propagation

Contexts form a tree - canceling a parent cancels all children:

```go
func main() {
    parent, cancel := context.WithCancel(context.Background())

    child1, _ := context.WithTimeout(parent, 5*time.Second)
    child2, _ := context.WithCancel(parent)

    cancel()  // Cancels parent, child1, AND child2
}
```

Pass context as the **first parameter** to functions:

```go
func fetchData(ctx context.Context, url string) ([]byte, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    return http.DefaultClient.Do(req)
}
```

---

## Common Pattern: HTTP Handler

```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()  // Already has timeout from server

    result, err := doWork(ctx)
    if err == context.Canceled {
        return  // Client disconnected, stop work
    }
    // ...
}
```

---

## Rules

1. **Always call cancel()** - use `defer cancel()` immediately
2. **Context is first parameter** - `func foo(ctx context.Context, ...)`
3. **Don't store in structs** - pass explicitly through function calls
4. **Don't pass nil** - use `context.Background()` or `context.TODO()`

---

## Your Turn

1. What's the difference between `context.WithTimeout` and `context.WithDeadline`?

2. What does this print?

```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()

select {
case <-ctx.Done():
    fmt.Println(ctx.Err())
case <-time.After(2 * time.Second):
    fmt.Println("completed")
}
```

3. Why should you always `defer cancel()` even if the context will timeout anyway?

### Answers

1. **Relative vs Absolute time:**
   - `WithTimeout(ctx, 5*time.Second)` - cancels 5 seconds from now (relative)
   - `WithDeadline(ctx, time.Now().Add(5*time.Second))` - cancels at a specific time (absolute)

   Use timeout for "give this operation N seconds." Use deadline for "this must complete by 3:00 PM."

2. **`context deadline exceeded`** - The 1-second timeout fires before the 2-second `time.After`. `ctx.Err()` returns `context.DeadlineExceeded` (not `context.Canceled`) because it was a timeout, not manual cancellation.

3. **Resource cleanup** - Creating a context with timeout/deadline spawns internal goroutines and timers. Calling `cancel()`:
   - Releases these resources immediately
   - Prevents leaks if the function returns early (error, early success)
   - Is safe to call multiple times (idempotent)

   Without `defer cancel()`, resources linger until the timeout expires.

---

## Summary

| Function | Cancels When |
|----------|--------------|
| `WithCancel` | `cancel()` is called |
| `WithTimeout` | Duration elapses OR `cancel()` |
| `WithDeadline` | Time reached OR `cancel()` |

| Error | Meaning |
|-------|---------|
| `context.Canceled` | `cancel()` was called |
| `context.DeadlineExceeded` | Timeout/deadline passed |

**Key pattern:** Always `defer cancel()` immediately after creating a cancelable context.
