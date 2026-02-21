# Lesson 5: Response Patterns

## The Problem

You know how to read requests. Now you need to write responses — and there's more to it than `w.Write([]byte("hello"))`. Real APIs need consistent JSON envelopes, proper status codes, error responses that clients can parse, and sometimes streaming data. Getting this wrong means clients that can't handle your errors, silent failures, and debugging nightmares.

## The Solution

### A JSON Response Helper

You'll write the same `json.NewEncoder(w).Encode(v)` boilerplate over and over unless you extract a helper:

```go
func writeJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(data); err != nil {
        // Already wrote headers — can't change status code now.
        // Log it; don't try to write another response.
        log.Printf("writeJSON encode error: %v", err)
    }
}
```

Why `any` and not `interface{}`? Same thing — `any` is the alias since Go 1.18. Use whichever your team prefers.

### Consistent Error Responses

Clients need a predictable shape for errors. Pick a format and stick with it:

```go
type APIError struct {
    Status  int    `json:"-"`              // HTTP status (not in JSON body)
    Code    string `json:"code"`           // machine-readable
    Message string `json:"message"`        // human-readable
}

func writeError(w http.ResponseWriter, err APIError) {
    writeJSON(w, err.Status, err)
}
```

Usage in a handler:

```go
func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")

    user, err := findUser(id)
    if err != nil {
        writeError(w, APIError{
            Status:  http.StatusNotFound,
            Code:    "user_not_found",
            Message: fmt.Sprintf("no user with id %q", id),
        })
        return // <-- DON'T FORGET THIS
    }

    writeJSON(w, http.StatusOK, user)
}
```

The `return` after `writeError` is one of the most common bugs in Go HTTP handlers. Without it, execution continues and you write a second response.

### Status Code Groups — Know What They Mean

| Range | Category      | Common Ones                                                    |
|-------|---------------|----------------------------------------------------------------|
| 2xx   | Success       | 200 OK, 201 Created, 204 No Content                           |
| 3xx   | Redirect      | 301 Moved Permanently, 304 Not Modified                       |
| 4xx   | Client Error  | 400 Bad Request, 401 Unauthorized, 403 Forbidden, 404, 422    |
| 5xx   | Server Error  | 500 Internal Server Error, 502 Bad Gateway, 503 Unavailable   |

Use the `http` package constants (`http.StatusOK`, `http.StatusNotFound`) — they're self-documenting and prevent typos.

**204 No Content** is worth knowing: use it for successful DELETEs or updates where there's nothing to return. Don't write a body with 204 — clients may ignore it.

```go
func deleteUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    if err := removeUser(id); err != nil {
        writeError(w, APIError{Status: 500, Code: "delete_failed", Message: err.Error()})
        return
    }
    w.WriteHeader(http.StatusNoContent) // 204, no body
}
```

### Response Envelopes

Some APIs wrap everything in an envelope:

```go
type Response struct {
    Data  any    `json:"data,omitempty"`
    Error *APIError `json:"error,omitempty"`
}
```

This gives clients one shape to parse. Whether you use envelopes is a design choice — many modern APIs skip them and rely on status codes + a flat body. Both approaches work; consistency is what matters.

### Streaming Responses (Server-Sent Events)

Sometimes you need to push data to the client over time — progress updates, live feeds, etc. **Server-Sent Events (SSE)** are the simplest way:

```go
func eventsHandler(w http.ResponseWriter, r *http.Request) {
    // Check that we can flush
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "streaming not supported", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    for i := 0; i < 5; i++ {
        fmt.Fprintf(w, "data: message %d\n\n", i)
        flusher.Flush()

        select {
        case <-r.Context().Done():
            return // client disconnected
        case <-time.After(1 * time.Second):
        }
    }
}
```

Key details:
- **`http.Flusher`** — `ResponseWriter` implements this. `Flush()` sends buffered data immediately.
- **SSE format** — each message is `data: <content>\n\n` (double newline terminates a message).
- **`r.Context().Done()`** — fires when the client disconnects. Always check this in long-running handlers.

### http.Error — The Quick-and-Dirty Helper

For plain-text error responses, the stdlib has `http.Error`:

```go
http.Error(w, "something went wrong", http.StatusInternalServerError)
```

This sets `Content-Type: text/plain`, writes the status code, and writes the message. Fine for simple cases, but for JSON APIs, use your own `writeError`.

## Key Rules

1. **Set headers before WriteHeader, WriteHeader before Write** — you already know this from lesson 1, but it's the #1 source of bugs
2. **Always `return` after writing an error** — otherwise you double-write
3. **Use `http.Status*` constants** — `http.StatusNotFound` > `404`
4. **Pick one error shape and be consistent** — clients will parse it programmatically
5. **Don't write a body with 204** — it's technically allowed but many clients ignore it
6. **Flush for streaming** — without `Flush()`, data buffers and the client sees nothing until the handler returns

## Common Mistakes

- **Forgetting `return` after error response** — handler keeps running, writes a second response, you get "superfluous response.WriteHeader call"
- **Setting Content-Type after Write** — too late, Go already sniffed it (usually `text/plain` or `application/octet-stream`)
- **Writing 200 then trying to change to 500** — first status wins, `WriteHeader` is a one-shot
- **Using `Encode` on a nil value** — `json.Encode(nil)` writes `null\n`, which might surprise clients expecting `{}`

## Your Turn

### Exercise 1: Build a mini CRUD API

Build a complete in-memory user API with proper response patterns. The API should:

- `GET /users` — return all users (200 with JSON array)
- `GET /users/{id}` — return one user (200) or error (404)
- `POST /users` — create user from JSON body, return 201 with the created user
- `DELETE /users/{id}` — delete user, return 204 or 404

Use the `writeJSON` and `writeError` helpers from above. Store users in a `map[string]User`.

### Exercise 2: Add an SSE endpoint

Add `GET /users/feed` that streams a message every second for 5 seconds, reporting the current user count. The client should see:

```
data: {"count": 3, "time": "2024-01-15T10:30:00Z"}

data: {"count": 3, "time": "2024-01-15T10:30:01Z"}
```

Test it with: `curl -N http://localhost:8080/users/feed`

---

<details>
<summary><strong>Exercise 1 Answer</strong></summary>

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

var (
	users   = make(map[string]User)
	mu      sync.Mutex
	nextID  = 1
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

func writeError(w http.ResponseWriter, e APIError) {
	writeJSON(w, e.Status, e)
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	list := make([]User, 0, len(users))
	for _, u := range users {
		list = append(list, u)
	}
	writeJSON(w, http.StatusOK, list)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		writeError(w, APIError{http.StatusBadRequest, "bad_json", "invalid JSON body"})
		return
	}

	mu.Lock()
	u.ID = fmt.Sprintf("%d", nextID)
	nextID++
	users[u.ID] = u
	mu.Unlock()

	writeJSON(w, http.StatusCreated, u)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	mu.Lock()
	u, ok := users[id]
	mu.Unlock()

	if !ok {
		writeError(w, APIError{http.StatusNotFound, "not_found", fmt.Sprintf("user %q not found", id)})
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	mu.Lock()
	_, ok := users[id]
	if ok {
		delete(users, id)
	}
	mu.Unlock()

	if !ok {
		writeError(w, APIError{http.StatusNotFound, "not_found", fmt.Sprintf("user %q not found", id)})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", handleUsers)
	mux.HandleFunc("POST /users", handleCreateUser)
	mux.HandleFunc("GET /users/{id}", handleGetUser)
	mux.HandleFunc("DELETE /users/{id}", handleDeleteUser)

	fmt.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

</details>

<details>
<summary><strong>Exercise 2 Answer</strong></summary>

```go
func handleUserFeed(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for i := 0; i < 5; i++ {
		mu.Lock()
		count := len(users)
		mu.Unlock()

		msg := map[string]any{
			"count": count,
			"time":  time.Now().UTC().Format(time.RFC3339),
		}
		data, _ := json.Marshal(msg)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()

		select {
		case <-r.Context().Done():
			return
		case <-time.After(1 * time.Second):
		}
	}
}

// Register BEFORE the {id} route so it doesn't get caught as a path param:
// mux.HandleFunc("GET /users/feed", handleUserFeed)
```

**Important:** Register `/users/feed` before `/users/{id}` — Go 1.22's router gives literal segments higher precedence than wildcards, so it actually works either way, but it's good practice to be explicit about ordering.

</details>

## Summary

| Pattern                | When to Use                            | Example                                    |
|------------------------|----------------------------------------|--------------------------------------------|
| `writeJSON` helper     | Every JSON response                    | `writeJSON(w, 200, user)`                  |
| `writeError` helper    | Every error response                   | `writeError(w, APIError{404, ...})`        |
| `204 No Content`       | Successful DELETE, no body to return   | `w.WriteHeader(http.StatusNoContent)`      |
| `http.Error`           | Quick plain-text errors                | `http.Error(w, "bad", 400)`               |
| SSE / `http.Flusher`   | Streaming real-time data to clients    | `flusher.Flush()` after each `Fprintf`     |
| Response envelope      | When clients need one consistent shape | `{"data": ..., "error": ...}`              |
