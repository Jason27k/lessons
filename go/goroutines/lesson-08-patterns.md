# Lesson 8: Common Concurrency Patterns

Now we put it all together. These patterns appear everywhere in Go code.

---

## Pattern 1: Worker Pool

Process jobs concurrently with a fixed number of workers:

```go
func worker(id int, jobs <-chan int, results chan<- int) {
    for job := range jobs {
        fmt.Printf("worker %d processing job %d\n", id, job)
        time.Sleep(time.Second)  // Simulate work
        results <- job * 2
    }
}

func main() {
    jobs := make(chan int, 100)
    results := make(chan int, 100)

    // Start 3 workers
    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }

    // Send 9 jobs
    for j := 1; j <= 9; j++ {
        jobs <- j
    }
    close(jobs)  // Signal no more jobs

    // Collect results
    for r := 1; r <= 9; r++ {
        fmt.Println(<-results)
    }
}
```

**Why use it:**
- Limit concurrency (don't spawn 10,000 goroutines)
- Control resource usage (DB connections, API rate limits)

---

## Pattern 2: Fan-Out

Distribute work from one source to multiple workers:

```go
func main() {
    jobs := make(chan int)

    // Fan-out: multiple workers reading from same channel
    for i := 1; i <= 3; i++ {
        go func(id int) {
            for job := range jobs {
                fmt.Printf("worker %d got %d\n", id, job)
            }
        }(i)
    }

    // Send work
    for i := 0; i < 10; i++ {
        jobs <- i
    }
    close(jobs)
    time.Sleep(time.Second)
}
```

Jobs are automatically distributed - each job goes to exactly one worker.

---

## Pattern 3: Fan-In

Merge multiple channels into one:

```go
func fanIn(channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for val := range c {
                out <- val
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}

func main() {
    ch1 := make(chan int)
    ch2 := make(chan int)

    go func() { ch1 <- 1; ch1 <- 2; close(ch1) }()
    go func() { ch2 <- 3; ch2 <- 4; close(ch2) }()

    merged := fanIn(ch1, ch2)
    for val := range merged {
        fmt.Println(val)  // 1, 2, 3, 4 (order may vary)
    }
}
```

**Use case:** Aggregate results from multiple sources.

---

## Pattern 4: Pipeline

Chain stages together, each running concurrently:

```go
// Stage 1: Generate numbers
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

// Stage 2: Square numbers
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

// Stage 3: Filter (keep only > 10)
func filter(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            if n > 10 {
                out <- n
            }
        }
        close(out)
    }()
    return out
}

func main() {
    // Chain: generate -> square -> filter
    nums := generate(1, 2, 3, 4, 5)
    squared := square(nums)
    filtered := filter(squared)

    for n := range filtered {
        fmt.Println(n)  // 16, 25
    }
}
```

**Benefits:**
- Each stage runs concurrently
- Memory efficient (streaming, not batching)
- Easy to add/remove stages

---

## Pattern 5: Semaphore (Limit Concurrency)

Use a buffered channel as a counting semaphore:

```go
func main() {
    sem := make(chan struct{}, 3)  // Max 3 concurrent
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            sem <- struct{}{}        // Acquire (blocks if 3 already running)
            defer func() { <-sem }() // Release

            fmt.Printf("task %d running\n", id)
            time.Sleep(time.Second)
        }(i)
    }

    wg.Wait()
}
```

---

## Pattern 6: Or-Done Channel

Stop reading when context cancels:

```go
func orDone(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for {
            select {
            case <-ctx.Done():
                return
            case val, ok := <-in:
                if !ok {
                    return
                }
                select {
                case out <- val:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()
    return out
}
```

Wraps any channel to respect context cancellation.

---

## Your Turn

1. In the worker pool pattern, why do we use buffered channels for jobs and results?

2. What happens in fan-out if one worker is slower than others?

3. You have 1000 URLs to fetch but want max 10 concurrent requests. Which pattern would you use?

### Answers

1. **Buffered channels decouple producers and consumers:**
   - Jobs buffer: sender can queue work without waiting for workers
   - Results buffer: workers can submit results without waiting for collector
   - Prevents blocking when speeds differ temporarily
   - Unbuffered would work but with more synchronization overhead

2. **Slow workers just process fewer jobs:**
   - Work naturally flows to whoever is ready
   - Fast workers pick up slack
   - All jobs still complete
   - This is a feature, not a bug - automatic load balancing

3. **Worker pool or Semaphore pattern:**
   ```go
   // Worker pool approach
   urls := make(chan string, 1000)
   results := make(chan Response, 1000)

   for i := 0; i < 10; i++ {  // 10 workers
       go fetcher(urls, results)
   }

   // OR Semaphore approach
   sem := make(chan struct{}, 10)  // Max 10 concurrent
   for _, url := range urls {
       sem <- struct{}{}
       go func(u string) {
           defer func() { <-sem }()
           fetch(u)
       }(url)
   }
   ```

---

## Summary

| Pattern | Use When |
|---------|----------|
| Worker Pool | Fixed concurrency, job queue |
| Fan-Out | Distribute to multiple workers |
| Fan-In | Merge multiple sources |
| Pipeline | Chain processing stages |
| Semaphore | Limit concurrent operations |
| Or-Done | Respect cancellation in pipelines |

These patterns compose - a real system might use fan-out into a worker pool, with fan-in to collect results, all wrapped in or-done for cancellation.
