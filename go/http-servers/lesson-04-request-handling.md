# Lesson 4: Request Handling

## The Problem

Your server can route requests and wrap them in middleware. But your handlers are still just writing static strings. Real APIs need to:

- Parse JSON from request bodies (`POST /users` with `{"name": "gopher"}`)
- Read query parameters (`GET /users?page=2&limit=10`)
- Extract path variables (`GET /users/42`)
- Read headers and cookies
- Reject bad input before it reaches your business logic

Go doesn't give you magic bindings like some frameworks — you handle all of this explicitly. That's more code, but zero hidden behavior.

---

## Parsing JSON Request Bodies

The most common case: a client `POST`s JSON and you need to turn it into a struct.

```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func createUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }

    fmt.Fprintf(w, "Created user: %s (%s)", req.Name, req.Email)
}
```

Key points:
- `r.Body` is an `io.ReadCloser` — you read it once and it's consumed. No rewinding.
- `json.NewDecoder` streams directly from the body — no need to read into a `[]byte` first.
- Always check the error. Malformed JSON, wrong types, empty bodies — all return errors here.

### What About `json.Unmarshal`?

You'll see both approaches:

```go
// Decoder — streams from reader, preferred for HTTP bodies
json.NewDecoder(r.Body).Decode(&req)

// Unmarshal — needs the full []byte in memory first
body, _ := io.ReadAll(r.Body)
json.Unmarshal(body, &req)
```

Use `NewDecoder` for HTTP bodies. Use `Unmarshal` when you already have bytes (e.g., from a file or cache).

### Limiting Body Size

A client could send a 10GB body and blow your memory. Protect yourself:

```go
r.Body = http.MaxBytesReader(w, r.Body, 1_048_576) // 1MB limit
```

If the body exceeds the limit, `Decode` returns an error and the connection is closed. Put this at the top of handlers that accept bodies.

---

## Query Parameters

`GET /users?page=2&limit=10&active=true`

```go
func listUsers(w http.ResponseWriter, r *http.Request) {
    page := r.URL.Query().Get("page")     // "2" (string!)
    limit := r.URL.Query().Get("limit")   // "10" (string!)
    active := r.URL.Query().Get("active") // "true" (string!)

    // Missing params return ""
    sort := r.URL.Query().Get("sort")     // ""
}
```

Everything is a string. You have to convert:

```go
page, err := strconv.Atoi(r.URL.Query().Get("page"))
if err != nil {
    page = 1 // default
}
```

For parameters that can appear multiple times (`?tag=go&tag=web`):

```go
tags := r.URL.Query()["tag"] // []string{"go", "web"}
```

Note: `.Get()` returns the *first* value. The map index `["tag"]` returns *all* values.

---

## Path Parameters

You already know this from lesson 2:

```go
mux.HandleFunc("GET /users/{id}", getUser)

func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    // Convert if needed
    userID, err := strconv.Atoi(id)
    if err != nil {
        http.Error(w, "invalid user ID", http.StatusBadRequest)
        return
    }
    fmt.Fprintf(w, "User %d", userID)
}
```

---

## Headers

Read request headers:

```go
contentType := r.Header.Get("Content-Type")
apiKey := r.Header.Get("X-API-Key")
```

Headers are case-insensitive — `r.Header.Get("content-type")` works too. Go canonicalizes them internally.

Set response headers (before writing the body):

```go
w.Header().Set("Content-Type", "application/json")
w.Header().Set("X-Custom", "value")
```

---

## Cookies

Read a cookie:

```go
cookie, err := r.Cookie("session_id")
if err != nil {
    // http.ErrNoCookie if it doesn't exist
    http.Error(w, "no session", http.StatusUnauthorized)
    return
}
fmt.Println(cookie.Value)
```

Set a cookie:

```go
http.SetCookie(w, &http.Cookie{
    Name:     "session_id",
    Value:    "abc123",
    Path:     "/",
    HttpOnly: true,
    MaxAge:   86400, // 1 day in seconds
})
```

---

## Input Validation

Go doesn't have built-in validation annotations. You write it yourself:

```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

func (r CreateUserRequest) validate() error {
    if r.Name == "" {
        return fmt.Errorf("name is required")
    }
    if r.Email == "" {
        return fmt.Errorf("email is required")
    }
    if r.Age < 0 || r.Age > 150 {
        return fmt.Errorf("age must be between 0 and 150")
    }
    return nil
}
```

Then in your handler:

```go
func createUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }
    if err := req.validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    // req is safe to use
}
```

This pattern — decode then validate — is how most Go APIs handle input. Libraries like `go-playground/validator` exist if you want struct tag validation, but explicit methods are fine for most cases.

---

## A Gotcha: Unknown Fields

By default, `json.Decoder` silently ignores fields that don't match your struct. A client sends `{"naem": "gopher"}` (typo) and you get an empty `Name` with no error.

To catch this:

```go
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()
if err := dec.Decode(&req); err != nil {
    http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
    return
}
```

Now `{"naem": "gopher"}` returns an error about the unknown field.

---

## Key Rules

1. **`r.Body` is read-once** — if you read it, it's gone. Don't try to read it twice.
2. **Query params are always strings** — convert with `strconv`.
3. **Always validate after decoding** — the decoder only checks JSON syntax, not your business rules.
4. **Limit body size** — `http.MaxBytesReader` prevents memory abuse.
5. **Set headers before writing** — once you call `w.Write()` or `w.WriteHeader()`, headers are sent.

---

## Your Turn

Run `main.go`. It's a small user API that demonstrates all these patterns.

1. `go run main.go`
2. Try these:
   - `curl "localhost:8080/users?page=2&limit=5"` — query params
   - `curl -X POST localhost:8080/users -d '{"name": "gopher", "email": "go@example.com", "age": 25}'` — valid JSON
   - `curl -X POST localhost:8080/users -d '{"name": ""}'` — missing required fields
   - `curl -X POST localhost:8080/users -d 'not json'` — malformed body
   - `curl -X POST localhost:8080/users -d '{"naem": "gopher"}'` — typo in field name
   - `curl localhost:8080/users/42` — path param
   - `curl localhost:8080/users/abc` — invalid ID
   - `curl -v --cookie "session_id=abc123" localhost:8080/me` — cookie reading

**Exercises:**

1. Add a `PUT /users/{id}` handler that accepts a JSON body to update a user. It should validate that `id` is a number and that the body has at least a `name`.
2. Add a `GET /search?q=...` handler that returns 400 if `q` is empty or missing.
3. The current `POST /users` doesn't limit body size. Add `http.MaxBytesReader` with a 1KB limit. Test it by sending a large body: `curl -X POST localhost:8080/users -d "$(python3 -c 'print("{\"name\":\"" + "a"*2000 + "\"}")')"`

### Answers

1. **PUT handler:**
```go
mux.HandleFunc("PUT /users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil {
        http.Error(w, "invalid user ID", http.StatusBadRequest)
        return
    }

    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }
    if req.Name == "" {
        http.Error(w, "name is required", http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "id":   id,
        "name": req.Name,
        "updated": true,
    })
})
```

2. **Search handler:**
```go
mux.HandleFunc("GET /search", func(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query().Get("q")
    if q == "" {
        http.Error(w, "q parameter is required", http.StatusBadRequest)
        return
    }
    fmt.Fprintf(w, "Searching for: %s", q)
})
```

3. **Body size limit:**
```go
// Add as the first line inside createUser:
r.Body = http.MaxBytesReader(w, r.Body, 1024) // 1KB
```
With the oversized body, you'll get an error from `Decode` because the reader cuts off after 1KB.

---

## Summary

| Source | How to Read | Returns |
|--------|------------|---------|
| JSON body | `json.NewDecoder(r.Body).Decode(&v)` | Error if malformed |
| Query param | `r.URL.Query().Get("key")` | `""` if missing |
| Multi-value query | `r.URL.Query()["key"]` | `[]string` |
| Path param | `r.PathValue("name")` | `string` |
| Header | `r.Header.Get("Name")` | `""` if missing |
| Cookie | `r.Cookie("name")` | `*Cookie, error` |

| Safety Pattern | What It Does |
|----------------|-------------|
| `http.MaxBytesReader` | Limits body size |
| `dec.DisallowUnknownFields()` | Rejects unknown JSON keys |
| `req.validate()` | Custom business rule checks |
| `strconv.Atoi` | Safe string-to-int conversion |
