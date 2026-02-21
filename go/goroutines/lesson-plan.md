# Go Routines - Lesson Plan

## Prerequisites
- Basic goroutine creation (`go func()`)
- Basic channel usage (send/receive)

## Today's Topics

### 1. Synchronization with WaitGroups
- Why we need WaitGroups
- `sync.WaitGroup` basics: Add, Done, Wait

### 2. Buffered vs Unbuffered Channels
- Blocking behavior differences
- When to use each

### 3. Channel Directions
- Send-only channels (`chan<-`)
- Receive-only channels (`<-chan`)
- Type safety benefits

### 4. The Select Statement
- Multiplexing channels
- Non-blocking operations with `default`
- Timeouts with `time.After`

### 5. Closing Channels
- When and why to close
- Range over channels
- Detecting closed channels

### 6. Mutex and Shared State
- `sync.Mutex` and `sync.RWMutex`
- Protecting shared data
- Channels vs Mutexes

### 7. Context for Cancellation
- `context.Context` basics
- Cancellation propagation
- Timeouts and deadlines

### 8. Common Patterns
- Worker pools
- Fan-in / Fan-out
- Pipeline pattern

### 9. Pitfalls and Debugging
- Race conditions
- Deadlocks
- Using `-race` flag
