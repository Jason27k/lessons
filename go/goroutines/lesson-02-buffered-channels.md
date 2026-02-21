# Lesson 2: Buffered vs Unbuffered Channels

## Quick Recap

You know channels let goroutines communicate:

```go
ch := make(chan int)    // Create a channel
ch <- 42                // Send
value := <-ch           // Receive
```

But there's a critical detail: **blocking behavior**.

---

## Unbuffered Channels (what you've been using)

```go
ch := make(chan int)  // No buffer size = unbuffered
```

**Unbuffered channels are synchronous:**
- A send blocks until another goroutine receives
- A receive blocks until another goroutine sends

Think of it as a **handoff** - both parties must be present.

```go
func main() {
    ch := make(chan int)

    ch <- 42  // DEADLOCK! No one is receiving yet

    fmt.Println(<-ch)
}
```

This works:

```go
func main() {
    ch := make(chan int)

    go func() {
        ch <- 42  // This goroutine waits here...
    }()

    fmt.Println(<-ch)  // ...until main receives
}
```

---

## Buffered Channels

```go
ch := make(chan int, 3)  // Buffer size of 3
```

**Buffered channels have capacity:**
- Send only blocks when the buffer is **full**
- Receive only blocks when the buffer is **empty**

Think of it as a **mailbox** - you can drop messages without waiting.

```go
func main() {
    ch := make(chan int, 2)

    ch <- 1  // Doesn't block (buffer has space)
    ch <- 2  // Doesn't block (buffer has space)
    // ch <- 3  // Would block! Buffer is full

    fmt.Println(<-ch)  // 1
    fmt.Println(<-ch)  // 2
}
```

---

## Checking Channel State

```go
ch := make(chan int, 3)
ch <- 1
ch <- 2

fmt.Println(len(ch))  // 2 - items currently in buffer
fmt.Println(cap(ch))  // 3 - total buffer capacity
```

---

## When to Use Which?

| Type | Use Case |
|------|----------|
| **Unbuffered** | When you need synchronization. Sender knows receiver got the message. |
| **Buffered** | When sender shouldn't wait. Decouples producer from consumer speed. |

### Common Use Cases

**Unbuffered:**
- Request/response patterns
- Signaling (done channels)
- When timing matters

**Buffered:**
- Rate limiting
- Batch processing
- When producer is faster than consumer (temporarily)

---

## Your Turn

Try this in your main.go:

```go
func main() {
    ch := make(chan int, 1)  // Buffer of 1

    ch <- 1
    ch <- 2  // What happens here?

    fmt.Println(<-ch)
}
```

1. What happens when you run this?
2. What if you change the buffer size to 2?
3. Can you make it work with a goroutine instead of changing the buffer?

### Answers

1. **Deadlock** - Buffer of 1 fills with first send, second send blocks forever.

2. **Works** - Buffer of 2 holds both values. Would block on a third send.

3. **Goroutine solution:**
```go
ch := make(chan int, 1)
go func() {
    fmt.Println(<-ch)  // Receives 1, frees buffer
}()
ch <- 1  // Into buffer
ch <- 2  // Blocks until goroutine receives 1
fmt.Println(<-ch)  // Receives 2
```

**Note:** For production code, add a WaitGroup to ensure the goroutine completes before main exits.

---

## Summary

| Channel Type | Blocks on Send | Blocks on Receive |
|--------------|----------------|-------------------|
| `make(chan T)` | Until received | Until sent |
| `make(chan T, n)` | When buffer full | When buffer empty |

Key insight: Unbuffered channels **synchronize** goroutines. Buffered channels **decouple** them.
