# Goroutines Deep Dive

## Lesson 1: Synchronization with WaitGroups

### The Problem

When you launch goroutines, the main function doesn't wait for them to complete:

```go
func main() {
    go doWork()  // This starts...
    // ...but main exits immediately, killing the goroutine!
}
```

You might have used `time.Sleep()` as a workaround, but that's unreliable. How do you know how long to wait?

### The Solution: sync.WaitGroup

A `WaitGroup` is a counter that blocks until it reaches zero.

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var wg sync.WaitGroup

    for i := 1; i <= 3; i++ {
        wg.Add(1)  // Increment counter BEFORE starting goroutine

        go func(id int) {
            defer wg.Done()  // Decrement counter when done
            fmt.Printf("Worker %d finished\n", id)
        }(i)
    }

    wg.Wait()  // Block until counter is 0
    fmt.Println("All workers complete!")
}
```

### Key Rules

1. **Call `Add()` before `go`** - Not inside the goroutine! Otherwise there's a race.
2. **Use `defer wg.Done()`** - Ensures it runs even if the function panics.
3. **Pass WaitGroup by pointer** - If passing to functions, use `*sync.WaitGroup`.

### Common Mistake

```go
// WRONG - Add() inside goroutine creates a race condition
go func() {
    wg.Add(1)  // Main might call Wait() before this runs!
    defer wg.Done()
    // ...
}()
```

---

## Your Turn

Before we continue, try this:

1. What happens if you call `wg.Done()` more times than `wg.Add()`?
2. What happens if you forget to call `wg.Done()`?

### Answers

1. **Too many `Done()` calls** → `panic: sync: negative WaitGroup counter`
   - The WaitGroup doesn't allow the counter to go negative
   - This is a runtime panic, not a compile-time error

2. **Missing `Done()` calls** → `fatal error: all goroutines are asleep - deadlock!`
   - `Wait()` blocks forever because the counter never reaches 0
   - Go's runtime detects that no goroutine can make progress

---

## Summary

| Method | Purpose |
|--------|---------|
| `wg.Add(n)` | Increment counter by n (call before `go`) |
| `wg.Done()` | Decrement counter by 1 (use with `defer`) |
| `wg.Wait()` | Block until counter reaches 0 |

WaitGroups are perfect when you don't need data back from goroutines - you just need to know they finished.
