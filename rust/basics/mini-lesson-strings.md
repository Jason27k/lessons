# Rust Basics

## Mini-Lesson: Strings

### Why Rust Has Two String Types

In Python, Go, and JavaScript, there's essentially one string type you think about. Rust has two:

| Type | What it is |
|------|------------|
| `&str` | A reference to string data stored somewhere in memory (often the binary). Fixed size, immutable. |
| `String` | A heap-allocated, growable, owned string. Like `StringBuilder` in Java or `strings.Builder` in Go. |

The reason for the split comes down to **ownership** (the big upcoming topic). For now, the practical rule:

- Use `&str` for string data you're just reading
- Use `String` when you need to own, build, or mutate a string

---

### String Literals are `&str`

```rust
let s = "hello"; // type is &str
```

This string is baked into the compiled binary. `s` is just a reference (a pointer + length) pointing into that data. You can't grow it or mutate it.

**Go comparison:**
```go
s := "hello" // Go's string is also immutable and backed by read-only memory
```

Go's `string` type actually behaves a lot like Rust's `&str` — it's an immutable view into bytes. The difference is Go only has this one type.

---

### Creating a `String`

```rust
let s = String::from("hello");  // from a string literal
let s = "hello".to_string();    // same thing, different syntax
```

`String` is heap-allocated and owned. You can grow it, mutate it, pass ownership around.

**Python comparison:**
```python
s = "hello"  # Python strings are immutable — no direct equivalent to String
```

Python strings are always immutable. Rust's `String` is the mutable, heap-owned version.

---

### Going Between `&str` and `String`

```rust
// &str -> String
let owned: String = "hello".to_string();
let owned: String = String::from("hello");

// String -> &str (borrow it)
let owned = String::from("hello");
let borrowed: &str = &owned;        // borrow the whole string
let slice: &str = &owned[0..3];     // borrow a slice: "hel"
```

The `&` in front of a `String` coerces it into a `&str`. This is extremely common.

---

### Function Parameters: Which to Use?

Prefer `&str` for function parameters — it accepts both `&str` and `&String`:

```rust
fn print_greeting(name: &str) {
    println!("Hello, {}!", name);
}

print_greeting("Alice");                      // &str literal — works
print_greeting(&String::from("Alice"));       // &String coerces to &str — works
```

If you wrote `name: &String` instead, it would only accept `&String`, not `&str` literals. So `&str` is the more flexible choice.

**Go comparison:**
```go
func printGreeting(name string) { ... } // Go has one type, no choice to make
```

---

### Common String Operations

```rust
let mut s = String::from("hello");

// Append
s.push(' ');           // append a char
s.push_str("world");   // append a &str
println!("{}", s);     // "hello world"

// Concatenation with +
let s1 = String::from("hello ");
let s2 = String::from("world");
let s3 = s1 + &s2;    // note: s1 is MOVED here, s2 is borrowed
// s1 is no longer valid after this line

// Concatenation with format! (no ownership issues)
let s1 = String::from("hello");
let s2 = String::from("world");
let s3 = format!("{} {}", s1, s2); // s1 and s2 still valid
```

The `+` operator is awkward because of ownership rules. Prefer `format!` when combining multiple strings.

**JavaScript comparison:**
```js
const s = "hello" + " " + "world"; // JS strings are simple — no ownership
```

---

### Length and Indexing

```rust
let s = String::from("hello");
println!("{}", s.len()); // 5 — byte count, not character count
```

**You cannot index a Rust string with `s[0]`** — this is a compile error:

```rust
let s = String::from("hello");
let c = s[0]; // ERROR: Rust won't let you do this
```

Why? Because Rust strings are UTF-8 encoded. A single character might be 1–4 bytes. Indexing by byte position could land in the middle of a character.

To get characters:

```rust
let s = String::from("hello");

// Get a slice (by byte range — be careful with multi-byte chars)
let slice = &s[0..3]; // "hel"

// Iterate over characters
for c in s.chars() {
    println!("{}", c);
}

// Get the nth character (less common)
let third = s.chars().nth(2); // Some('l')
```

**Go comparison:**
```go
s := "hello"
fmt.Println(s[0])       // prints 104 (the byte value) — similar gotcha
for _, c := range s { } // range on a Go string iterates runes (like .chars() in Rust)
```

---

### Checking Contents

```rust
let s = String::from("hello world");

println!("{}", s.contains("world"));    // true
println!("{}", s.starts_with("hello")); // true
println!("{}", s.ends_with("world"));   // true

// Trim whitespace
let padded = "  hello  ";
println!("{}", padded.trim()); // "hello"

// Split
for word in s.split(' ') {
    println!("{}", word);
}

// Replace
let replaced = s.replace("world", "Rust");
println!("{}", replaced); // "hello Rust"
```

These all work on both `&str` and `String` — they're defined on `&str` and `String` coerces automatically.

---

### The Key Mental Model

Think of it like this:

| | Rust | Go | Python | JS |
|---|---|---|---|---|
| Immutable string view | `&str` | `string` | `str` | `string` |
| Owned/growable string | `String` | `strings.Builder` | N/A (immutable) | N/A (immutable) |

Most of the time:
- Hardcoded text → `&str`
- Building or owning a string → `String`
- Function parameters → `&str` (accepts both)

---

## Common Mistakes

```rust
// WRONG: trying to index a string
let s = String::from("hello");
let c = s[0]; // ERROR

// RIGHT: use .chars().nth() or slice with &s[0..1]
let c = &s[0..1]; // "h" as &str

// WRONG: using + with multiple strings (ownership issues)
let s = s1 + s2 + s3; // s1 and s2 are moved, gets messy

// RIGHT: use format!
let s = format!("{}{}{}", s1, s2, s3);

// WRONG: expecting .len() to be character count
let s = String::from("héllo"); // 'é' is 2 bytes
println!("{}", s.len()); // prints 6, not 5

// RIGHT: use .chars().count() for character count
println!("{}", s.chars().count()); // 5
```

---

## Summary

| Operation | Code |
|-----------|------|
| String literal | `let s = "hello"` → `&str` |
| Owned string | `String::from("hello")` or `"hello".to_string()` |
| `&str` from `String` | `&my_string` |
| Append char | `s.push('!')` |
| Append string | `s.push_str("world")` |
| Combine strings | `format!("{}{}", s1, s2)` |
| Length in bytes | `s.len()` |
| Length in chars | `s.chars().count()` |
| Iterate chars | `for c in s.chars()` |
| Contains | `s.contains("text")` |
| Trim | `s.trim()` |
| Split | `s.split(' ')` |
| Replace | `s.replace("old", "new")` |
