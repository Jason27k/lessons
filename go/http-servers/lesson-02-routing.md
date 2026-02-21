# Lesson 2: Routing

## The Problem

In lesson 1 you saw that `/` catches everything. You also had to manually check `r.Method` to reject non-POST requests on `/echo`. The default ServeMux is bare-bones:

- No path parameters (`/users/123`)
- No method-based routing (`GET /users` vs `POST /users`)
- `/` is a greedy catch-all

Go 1.22 (released Feb 2024) overhauled `ServeMux` to fix most of this. Before that, everyone reached for third-party routers. Now the standard library covers the common cases.

---

## Enhanced Patterns (Go 1.22+)

### Method Matching

You can now prefix a pattern with an HTTP method:

```go
mux.HandleFunc("GET /users", listUsers)
mux.HandleFunc("POST /users", createUser)
```

A `POST` to `/users` hits `createUser`. A `GET` hits `listUsers`. A `DELETE`? Automatic `405 Method Not Allowed`. No more manual `if r.Method != ...` checks.

### Path Parameters (Wildcards)

Curly braces define named wildcards:

```go
mux.HandleFunc("GET /users/{id}", getUser)

func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // Extract the value
    fmt.Fprintf(w, "User ID: %s", id)
}
```

`GET /users/42` → `id` is `"42"`.

### The `{$}` Exact Match

Remember how `/` catches everything? You can stop that:

```go
mux.HandleFunc("GET /{$}", home)  // ONLY matches exactly "/"
```

Without `{$}`, `GET /` matches `/`, `/anything`, `/foo/bar`, etc. With `{$}`, it only matches `/` exactly.

### Catch-All Wildcard

Grab the rest of the path:

```go
mux.HandleFunc("GET /files/{path...}", serveFile)

func serveFile(w http.ResponseWriter, r *http.Request) {
    path := r.PathValue("path")  // "docs/readme.md"
    fmt.Fprintf(w, "File: %s", path)
}
```

`GET /files/docs/readme.md` → `path` is `"docs/readme.md"`.

---

## Precedence Rules

When patterns overlap, the **most specific** one wins:

```go
mux.HandleFunc("GET /users/{id}", getUser)       // Matches /users/42
mux.HandleFunc("GET /users/me", getCurrentUser)   // Matches /users/me
```

`/users/me` is more specific than `/users/{id}`, so it wins. The rule: **literal segments beat wildcards, longer paths beat shorter ones.**

If two patterns conflict and neither is more specific, `ServeMux` panics at registration time — you'll catch it immediately, not at runtime.

---

## When You Still Need a Third-Party Router

The standard library covers most needs now, but some projects use routers like `chi`, `gorilla/mux`, or `httprouter` for:

- **Regex constraints** — `/users/{id:[0-9]+}` (stdlib wildcards match anything)
- **Route groups** — applying middleware to a set of routes
- **Named routes / URL generation** — building URLs from route names

For learning (and many production apps), the stdlib is plenty.

---

## Your Turn

Try running the code in `main.go`. It's a mini user API with proper method routing and path params.

1. Run `go run main.go`
2. Test these:
   - `curl localhost:8080/` — what about `curl localhost:8080/nope`?
   - `curl localhost:8080/users`
   - `curl -X POST localhost:8080/users -d '{"name": "gopher"}'`
   - `curl localhost:8080/users/42`
   - `curl -X DELETE localhost:8080/users/42` — what status code do you get?
   - `curl localhost:8080/users/me`

**Questions:**
1. Why does `GET /nope` return a 404 now instead of hitting the home handler?
2. What happens if you `DELETE /users`? Why?
3. Which handler wins for `GET /users/me` — the `{id}` wildcard or the literal `/users/me`?

### Answers

1. **`{$}` makes `/` exact-match only.** Without it, `/` is a prefix and catches everything unmatched. With `/{$}`, only an exact request to `/` matches. Everything else with no matching route gets a 404.

2. **`405 Method Not Allowed`** — the mux knows `/users` exists but only for `GET` and `POST`. A `DELETE` to that path gets an automatic 405 with an `Allow` header listing the valid methods.

3. **`/users/me` wins** — literal segments are always more specific than wildcards. The mux resolves this at registration time so there's zero runtime cost.

---

## Summary

| Pattern | Matches |
|---------|---------|
| `"/users"` | Any method to `/users` |
| `"GET /users"` | Only `GET /users` (others get 405) |
| `"GET /users/{id}"` | `GET /users/42`, `GET /users/abc`, etc. |
| `"GET /users/{id...}"` | `GET /users/a/b/c` (rest-of-path) |
| `"GET /{$}"` | Only `GET /` exactly (no catch-all) |

| Function | What It Does |
|----------|-------------|
| `r.PathValue("id")` | Get a wildcard value from the URL |
| `http.Error(w, msg, code)` | Write error response (you already know this one) |

Go 1.22's routing upgrades mean you can build real APIs with just the standard library.
