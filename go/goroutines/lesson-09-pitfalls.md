# Lesson 9: Pitfalls and Debugging

Things that will bite you, and how to fix them.

---

## Pitfall 1: Goroutine Leaks

Goroutines that never exit consume memory forever.

### The Bug

```go
func leaky() {
    ch := make(chan int)

    go func() {
        val := <-ch  // Blocks forever - nothing sends!
        fmt.Println(val)
    }()

    // Function returns, but goroutine lives on...
}
```

### Common Causes

1. **Blocked on channel with no sender/receiver**
2. **Infinite loop without exit condition**
3. **Waiting on context that's never canceled**

### The Fix: Always provide an exit path

```go
func notLeaky(ctx context.Context) {
    ch := make(chan int)

    go func() {
        select {
        case val := <-ch:
            fmt.Println(val)
        case <-ctx.Done():
            return  // Exit path!
        }
    }()
}
```

### Detection

```go
import "runtime"

fmt.Println(runtime.NumGoroutine())  // Monitor this over time
```

If it keeps growing, you have a leak.

---

## Pitfall 2: Loop Variable Capture

Classic Go gotcha (fixed in Go 1.22+, but still common in older code):

### The Bug (Go < 1.22)

```go
for i := 0; i < 3; i++ {
    go func() {
        fmt.Println(i)  // Captures variable, not value!
    }()
}
// Prints: 3, 3, 3 (not 0, 1, 2)
```

All goroutines see the final value of `i`.

### The Fix

```go
// Option 1: Pass as argument
for i := 0; i < 3; i++ {
    go func(n int) {
        fmt.Println(n)
    }(i)  // Copy value here
}

// Option 2: Shadow the variable
for i := 0; i < 3; i++ {
    i := i  // New variable each iteration
    go func() {
        fmt.Println(i)
    }()
}
```

**Note:** Go 1.22+ fixed this - each iteration gets its own variable.

---

## Pitfall 3: Deadlock

All goroutines blocked, no progress possible.

### The Bug

```go
func main() {
    ch := make(chan int)
    ch <- 1  // Blocks forever - no receiver!
    fmt.Println(<-ch)
}
```

### Classic Deadlock: Two locks, wrong order

```go
var mu1, mu2 sync.Mutex

// Goroutine 1       // Goroutine 2
mu1.Lock()           mu2.Lock()
mu2.Lock()  // Wait  mu1.Lock()  // Wait - DEADLOCK!
```

### The Fix: Consistent lock ordering

```go
// Always acquire mu1 before mu2
mu1.Lock()
mu2.Lock()
// ...
mu2.Unlock()
mu1.Unlock()
```

### Detection

Go detects simple deadlocks:
```
fatal error: all goroutines are asleep - deadlock!
```

But complex deadlocks may not be detected. Use `-race` and careful design.

---

## Pitfall 4: Race Conditions

Already covered, but worth repeating common cases:

### Map concurrent access

```go
m := make(map[string]int)

go func() { m["a"] = 1 }()
go func() { m["b"] = 2 }()  // RACE: concurrent map writes

// Fix: use sync.Map or protect with mutex
```

### Slice append

```go
var slice []int

go func() { slice = append(slice, 1) }()
go func() { slice = append(slice, 2) }()  // RACE!

// Fix: use mutex or channel to collect values
```

### Detection

```bash
go run -race main.go
go test -race ./...
```

**Always run tests with -race in CI.**

---

## Pitfall 5: Forgetting WaitGroup.Add Before go

### The Bug

```go
var wg sync.WaitGroup

go func() {
    wg.Add(1)  // Too late! Main might call Wait() first
    defer wg.Done()
    // ...
}()

wg.Wait()  // Might exit before Add() runs
```

### The Fix

```go
var wg sync.WaitGroup

wg.Add(1)  // BEFORE go
go func() {
    defer wg.Done()
    // ...
}()

wg.Wait()
```

---

## Pitfall 6: Closing Channel from Wrong Side

### The Bug

```go
func consumer(ch <-chan int) {
    // Process values...
    close(ch)  // Compiler error with <-chan, but possible with chan
}

func producer(ch chan<- int) {
    for {
        ch <- work()  // Panic if consumer closed!
    }
}
```

### The Rule

**Only the sender closes.** Receiver doesn't know if sender is done.

---

## Debugging Tools

### 1. Race Detector

```bash
go run -race main.go
go build -race
go test -race ./...
```

### 2. pprof for Goroutine Dumps

```go
import _ "net/http/pprof"

go func() {
    http.ListenAndServe("localhost:6060", nil)
}()
```

Then: `go tool pprof http://localhost:6060/debug/pprof/goroutine`

### 3. Stack Traces

```go
import "runtime/debug"

debug.PrintStack()  // Print current goroutine's stack
```

Or send SIGQUIT (Ctrl+\) to dump all goroutine stacks.

### 4. Goroutine Count Monitoring

```go
import "runtime"

ticker := time.NewTicker(time.Second)
for range ticker.C {
    fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())
}
```

---

## Your Turn

1. What's wrong with this code?

```go
func process(items []int) {
    results := make(chan int)

    for _, item := range items {
        go func() {
            results <- item * 2
        }()
    }

    for range items {
        fmt.Println(<-results)
    }
}
```

2. How would you detect a goroutine leak in a long-running server?

3. You have a test that passes alone but fails when run with other tests using `-race`. What's likely happening?

### Answers

1. **Two bugs:**
   - **Loop variable capture** - all goroutines share `item`, likely all process the last value
   - **Unbuffered channel risk** - if `process` is called with empty slice, no issues. But the real fix:

   ```go
   for _, item := range items {
       go func(val int) {         // Pass as argument
           results <- val * 2
       }(item)                    // Copy value here
   }
   ```

   Note: WaitGroup isn't needed because the receive loop (`for range items`) synchronizes - we receive exactly as many times as we send.

2. **Monitor `runtime.NumGoroutine()` over time:**
   - Log it periodically or expose via metrics endpoint
   - If count keeps growing, you have a leak
   - Use pprof (`/debug/pprof/goroutine`) to see stack traces of stuck goroutines
   - In tests, check goroutine count before/after

3. **Shared state between tests:**
   - Global variables modified without synchronization
   - Tests running in parallel access same resources
   - The race exists but only manifests with specific timing
   - Fix: isolate test state, use `t.Parallel()` carefully, protect shared resources

---

## Summary: Concurrency Checklist

Before shipping concurrent code, verify:

- [ ] No goroutine leaks (all goroutines have exit paths)
- [ ] No loop variable capture (or using Go 1.22+)
- [ ] `wg.Add()` called before `go`
- [ ] Only senders close channels
- [ ] Locks acquired in consistent order
- [ ] Shared state protected (mutex or channels)
- [ ] Tests pass with `-race` flag
- [ ] Context used for cancellation in long operations
