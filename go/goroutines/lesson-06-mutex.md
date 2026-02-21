# Lesson 6: Mutex and Shared State

## The Problem: Race Conditions

When multiple goroutines access shared data, things break:

```go
func main() {
    counter := 0

    for i := 0; i < 1000; i++ {
        go func() {
            counter++  // READ, INCREMENT, WRITE - not atomic!
        }()
    }

    time.Sleep(time.Second)
    fmt.Println(counter)  // Not 1000! Different every run.
}
```

`counter++` looks atomic but it's actually three steps:
1. Read current value
2. Add 1
3. Write new value

Goroutines interleave these steps, overwriting each other's work.

---

## Solution 1: Mutex (Mutual Exclusion)

A mutex ensures only one goroutine accesses code at a time:

```go
func main() {
    var mu sync.Mutex
    counter := 0

    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()

            mu.Lock()         // Acquire lock (blocks if held)
            counter++         // Only one goroutine here at a time
            mu.Unlock()       // Release lock
        }()
    }

    wg.Wait()
    fmt.Println(counter)  // Always 1000
}
```

---

## Key Rules

### Always use defer for Unlock

```go
mu.Lock()
defer mu.Unlock()  // Guaranteed to run, even if panic
// ... do work ...
```

### Keep critical sections small

```go
// Bad - holding lock during slow operation
mu.Lock()
result := expensiveNetworkCall()  // Others blocked waiting!
data = result
mu.Unlock()

// Good - only lock for the shared data access
result := expensiveNetworkCall()
mu.Lock()
data = result
mu.Unlock()
```

### Never copy a mutex

```go
var mu sync.Mutex
mu2 := mu  // BUG! Copies lock state, breaks everything
```

---

## RWMutex: Multiple Readers, Single Writer

When reads are more common than writes:

```go
var rwmu sync.RWMutex
data := make(map[string]int)

// Multiple goroutines can read simultaneously
func read(key string) int {
    rwmu.RLock()         // Read lock - shared
    defer rwmu.RUnlock()
    return data[key]
}

// Only one goroutine can write (and no readers during write)
func write(key string, val int) {
    rwmu.Lock()          // Write lock - exclusive
    defer rwmu.Unlock()
    data[key] = val
}
```

| Lock Type | Blocks Readers | Blocks Writers |
|-----------|---------------|----------------|
| `RLock()` | No | Yes |
| `Lock()` | Yes | Yes |

---

## Common Pattern: Embed Mutex in Struct

```go
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}
```

---

## Channels vs Mutex: When to Use Each

| Use Channels | Use Mutex |
|--------------|-----------|
| Passing data between goroutines | Protecting shared state |
| Coordinating goroutines | Simple counters, caches |
| Building pipelines | When you need the data locally |

**Go proverb:** "Don't communicate by sharing memory; share memory by communicating."

But sometimes a mutex is simpler:

```go
// Overkill to use channels for a counter
counterCh := make(chan int)
// ... complex channel logic ...

// Just use a mutex
var mu sync.Mutex
counter++
```

---

## Detecting Races: The -race Flag

Go has a built-in race detector:

```bash
go run -race main.go
```

It instruments your code to detect concurrent access. Always test with this!

---

## Your Turn

1. Run this with `go run -race main.go`. What does the race detector report?

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var wg sync.WaitGroup
    count := 0

    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            count++
        }()
    }

    wg.Wait()
    fmt.Println(count)
}
```

2. Fix the code above using a mutex.

3. When would you use `RWMutex` instead of `Mutex`?

### Answers

1. **`WARNING: DATA RACE`** - The race detector shows:
   - Which goroutines are conflicting
   - The exact line (`count++`)
   - Stack traces for both read and write

2. **Fixed with mutex:**
```go
var wg sync.WaitGroup
var mu sync.Mutex
count := 0

for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        mu.Lock()
        defer mu.Unlock()
        count++
    }()
}

wg.Wait()
fmt.Println(count)  // Always 100, no race warning
```

3. **Use `RWMutex` when reads outnumber writes** - Multiple readers can hold `RLock()` simultaneously, improving throughput for read-heavy workloads. `Mutex` forces serial access even for reads.

---

## Summary

| Type | Use Case |
|------|----------|
| `sync.Mutex` | General mutual exclusion |
| `sync.RWMutex` | Read-heavy workloads (caches, config) |
| Channels | Passing data, coordination |

**Key practices:**
- Always `defer mu.Unlock()`
- Keep critical sections small
- Use `-race` flag in tests
- Never copy a mutex
