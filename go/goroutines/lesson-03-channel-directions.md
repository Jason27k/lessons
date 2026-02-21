# Lesson 3: Channel Directions

## The Problem

Look at this function signature:

```go
func worker(ch chan int) {
    // Can this function send? Receive? Both?
}
```

You can't tell from the signature what `worker` does with the channel. This makes code harder to understand and easier to misuse.

---

## Channel Direction Syntax

Go lets you restrict channels to **send-only** or **receive-only**:

```go
chan int      // Bidirectional (send and receive)
chan<- int    // Send-only (arrow points INTO channel)
<-chan int    // Receive-only (arrow points OUT OF channel)
```

Memory trick: The arrow shows which way data flows relative to `chan`.

---

## Using Directions in Functions

```go
// producer can ONLY send to ch
func producer(ch chan<- int) {
    ch <- 42
    // <-ch  // Compile error! Can't receive from send-only
}

// consumer can ONLY receive from ch
func consumer(ch <-chan int) {
    val := <-ch
    // ch <- 1  // Compile error! Can't send to receive-only
}

func main() {
    ch := make(chan int)  // Bidirectional

    go producer(ch)  // Automatically converts to chan<-
    consumer(ch)     // Automatically converts to <-chan
}
```

**Key insight:** A bidirectional channel automatically converts to a directional one when passed to a function. But you can't go the other way.

---

## Why Bother?

### 1. Self-documenting code

```go
// Immediately clear what this function does with the channel
func logger(messages <-chan string) { ... }
```

### 2. Compile-time safety

```go
func producer(ch chan<- int) {
    <-ch  // Compiler catches this mistake!
}
```

### 3. Better API design

```go
// Return a receive-only channel - caller can't accidentally send
func startWorker() <-chan Result {
    ch := make(chan Result)
    go func() {
        // ... do work ...
        ch <- result
        close(ch)
    }()
    return ch  // Caller can only receive
}
```

---

## Common Pattern: Returning Receive-Only Channels

This is idiomatic Go:

```go
func generator(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out  // Caller can only read from this
}

func main() {
    ch := generator(1, 2, 3)

    for val := range ch {
        fmt.Println(val)
    }
}
```

---

## Your Turn

1. What's wrong with this code? (Don't run it, just read)

```go
func sender(ch <-chan int) {
    ch <- 42
}
```

2. Fix this function signature to be more precise:

```go
func processData(input chan string, output chan string) {
    for msg := range input {
        output <- strings.ToUpper(msg)
    }
}
```

3. Why might you return `<-chan Error` instead of `chan Error` from a function?

### Answers

1. **Sending to a receive-only channel** - `<-chan int` means receive-only, but the function tries to send with `ch <- 42`. This won't compile.

2. **Fixed signature:**
```go
func processData(input <-chan string, output chan<- string)
```
- `input` is only read from → `<-chan`
- `output` is only written to → `chan<-`

3. **Prevents caller from accidentally sending** - The function owns the channel and controls when it closes. Returning receive-only ensures callers can only consume values, not interfere with the producer.

---

## Summary

| Syntax | Direction | Use in function params |
|--------|-----------|----------------------|
| `chan T` | Both | When function needs full control |
| `chan<- T` | Send-only | Producer functions |
| `<-chan T` | Receive-only | Consumer functions, return types |

Channel directions are a form of **documentation that the compiler enforces**.
