# Lesson 5: Closing Channels

## Why Close Channels?

Closing a channel signals "no more values will be sent." This lets receivers know when to stop waiting.

```go
close(ch)
```

---

## What Happens When You Close

```go
ch := make(chan int, 3)
ch <- 1
ch <- 2
close(ch)

fmt.Println(<-ch)  // 1
fmt.Println(<-ch)  // 2
fmt.Println(<-ch)  // 0 (zero value, channel closed and empty)
fmt.Println(<-ch)  // 0 (keeps returning zero value forever)
```

**Key points:**
- Buffered values are still received after close
- Once empty, receives return the zero value immediately
- Receives never block on a closed channel

---

## Detecting a Closed Channel: Comma-Ok Idiom

```go
val, ok := <-ch
if !ok {
    fmt.Println("channel closed!")
}
```

- `ok == true`: value received normally
- `ok == false`: channel is closed and empty

---

## Range Over Channels

`range` automatically stops when the channel closes:

```go
func main() {
    ch := make(chan int)

    go func() {
        for i := 0; i < 5; i++ {
            ch <- i
        }
        close(ch)  // MUST close or range blocks forever
    }()

    for val := range ch {
        fmt.Println(val)
    }
    // Exits cleanly when ch closes
}
```

**This is the idiomatic way to consume all values from a channel.**

---

## Rules for Closing

### Only the sender should close

```go
// Good - sender controls the channel
func producer(ch chan<- int) {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch)  // Producer knows when it's done
}
```

### Never close from the receiver side

```go
// Bad - receiver doesn't know if sender is done
func consumer(ch <-chan int) {
    close(ch)  // DON'T DO THIS
}
```

### Never close a channel twice

```go
close(ch)
close(ch)  // panic: close of closed channel
```

### Never send on a closed channel

```go
close(ch)
ch <- 1  // panic: send on closed channel
```

---

## Close as Broadcast

Closing a channel wakes up ALL receivers simultaneously:

```go
func worker(id int, done <-chan struct{}) {
    <-done  // Blocks until done is closed
    fmt.Printf("worker %d: shutting down\n", id)
}

func main() {
    done := make(chan struct{})

    for i := 1; i <= 3; i++ {
        go worker(i, done)
    }

    time.Sleep(time.Second)
    close(done)  // All 3 workers wake up!
    time.Sleep(100 * time.Millisecond)
}
```

This is better than sending 3 separate signals - you don't need to know how many receivers exist.

---

## Channels Don't Need to Be Closed

Closing is only necessary when the receiver needs to know no more values are coming (like with `range`). Unused channels are garbage collected.

```go
// This is fine - no close needed
func main() {
    ch := make(chan int)
    go func() {
        ch <- 42
    }()
    fmt.Println(<-ch)
    // ch is garbage collected, no close needed
}
```

---

## Your Turn

1. What does this print?

```go
ch := make(chan int)
close(ch)
val, ok := <-ch
fmt.Println(val, ok)
```

2. What's wrong with this code?

```go
func main() {
    ch := make(chan int)

    go func() {
        for val := range ch {
            fmt.Println(val)
        }
    }()

    ch <- 1
    ch <- 2
    ch <- 3
}
```

3. How would you signal 10 worker goroutines to stop using a single operation?

### Answers

1. **`0 false`** - Channel is closed and empty, so receive returns:
   - Zero value for the type (`0` for int)
   - `false` indicating channel is closed

2. **Two problems:**
   - Channel never closed → `range` blocks forever waiting for more values
   - No WaitGroup → main might exit before goroutine prints anything

   Fixed:
   ```go
   func main() {
       ch := make(chan int)
       var wg sync.WaitGroup
       wg.Add(1)

       go func() {
           defer wg.Done()
           for val := range ch {
               fmt.Println(val)
           }
       }()

       ch <- 1
       ch <- 2
       ch <- 3
       close(ch)  // Signal no more values
       wg.Wait()  // Wait for goroutine to finish
   }
   ```

3. **Close a done channel:**
   ```go
   done := make(chan struct{})
   // ... start 10 workers that select on done ...
   close(done)  // All 10 wake up simultaneously
   ```
   Closing broadcasts to all receivers - no need to send 10 separate signals.

---

## Summary

| Operation | Behavior |
|-----------|----------|
| `close(ch)` | Signal no more sends; wakes all receivers |
| `<-closedCh` | Returns zero value immediately |
| `val, ok := <-ch` | `ok=false` means closed and empty |
| `for v := range ch` | Loops until channel closes |

**Rules:** Only sender closes. Never close twice. Never send on closed.
