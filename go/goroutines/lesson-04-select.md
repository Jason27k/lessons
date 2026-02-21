# Lesson 4: The Select Statement

## The Problem

What if you need to receive from multiple channels?

```go
val1 := <-ch1  // Blocks here...
val2 := <-ch2  // ...can't get to this until ch1 sends
```

You're stuck waiting on `ch1` even if `ch2` has data ready.

---

## Select to the Rescue

`select` lets you wait on multiple channel operations simultaneously:

```go
select {
case val := <-ch1:
    fmt.Println("received from ch1:", val)
case val := <-ch2:
    fmt.Println("received from ch2:", val)
}
```

**How it works:**
- Blocks until ONE case is ready
- If multiple are ready, picks one **randomly**
- Executes that case, then continues after the select

---

## Basic Example

```go
func main() {
    ch1 := make(chan string)
    ch2 := make(chan string)

    go func() {
        time.Sleep(100 * time.Millisecond)
        ch1 <- "one"
    }()

    go func() {
        time.Sleep(200 * time.Millisecond)
        ch2 <- "two"
    }()

    // Wait for both, handling whichever comes first
    for i := 0; i < 2; i++ {
        select {
        case msg := <-ch1:
            fmt.Println("received:", msg)
        case msg := <-ch2:
            fmt.Println("received:", msg)
        }
    }
}
// Output:
// received: one
// received: two
```

---

## The Default Case: Non-Blocking Operations

Without `default`, select blocks. With `default`, it doesn't:

```go
select {
case val := <-ch:
    fmt.Println("received:", val)
default:
    fmt.Println("no value ready, moving on")
}
```

### Non-blocking send:

```go
select {
case ch <- value:
    fmt.Println("sent!")
default:
    fmt.Println("channel full or no receiver, skipping")
}
```

---

## Timeouts with time.After

`time.After(d)` returns a channel that sends after duration `d`:

```go
select {
case val := <-ch:
    fmt.Println("received:", val)
case <-time.After(1 * time.Second):
    fmt.Println("timed out waiting for value")
}
```

This is how you prevent waiting forever.

---

## Common Pattern: Done Channel

Gracefully stop a goroutine:

```go
func worker(done <-chan struct{}, jobs <-chan int) {
    for {
        select {
        case <-done:
            fmt.Println("worker: stopping")
            return
        case job := <-jobs:
            fmt.Println("worker: processing", job)
        }
    }
}

func main() {
    done := make(chan struct{})
    jobs := make(chan int)

    go worker(done, jobs)

    jobs <- 1
    jobs <- 2
    jobs <- 3

    close(done)  // Signal worker to stop
    time.Sleep(100 * time.Millisecond)
}
```

**Why `struct{}`?** It's zero bytes - we only care about the signal, not the value.

---

## Empty Select: Block Forever

```go
select {}  // Blocks forever (useful to keep main alive)
```

---

## Your Turn

Try this in main.go:

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch := make(chan int)

    go func() {
        time.Sleep(2 * time.Second)
        ch <- 42
    }()

    select {
    case val := <-ch:
        fmt.Println("received:", val)
    case <-time.After(1 * time.Second):
        fmt.Println("timed out!")
    }
}
```

1. What gets printed and why?
2. Change the goroutine's sleep to 500ms. What happens now?
3. Add a `default` case that prints "checking...". What happens and why might this be a problem?

### Answers

1. **"timed out!"** - The `time.After` channel fires after 1 second. The goroutine sleeps for 2 seconds, so the timeout wins.

2. **"received: 42"** - Now the goroutine (500ms) beats the timeout (1s), so the channel receive case executes.

3. **"checking..."** then exits immediately - `default` runs when no other case is ready. Since both channels need time before they're ready, `default` fires instantly. The select doesn't wait at all.

**When to use default:** Only when you intentionally want non-blocking behavior, typically inside a loop:
```go
for {
    select {
    case val := <-ch:
        process(val)
    default:
        // Do other work while waiting
        doSomethingElse()
        time.Sleep(10 * time.Millisecond)
    }
}
```

---

## Summary

| Pattern | Use Case |
|---------|----------|
| `select { case... case... }` | Wait for first of multiple channels |
| `select { case... default: }` | Non-blocking check |
| `case <-time.After(d):` | Timeout |
| `case <-done:` | Cancellation signal |
| `select {}` | Block forever |

Select is the **control flow** for concurrent Go - it's how you coordinate multiple goroutines.
