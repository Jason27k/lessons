# Rust Basics

## Lesson 8: Error Handling Basics

### The Problem

Go handles errors by returning `(value, error)` pairs and checking `if err != nil`. Rust takes a different approach: errors are encoded directly in the return type using `Result<T, E>`. You can't ignore them without being explicit about it, and the compiler enforces that you handle every case.

---

### The Vocabulary

| Term | Meaning |
|------|---------|
| `Result<T, E>` | An enum: either `Ok(T)` (success) or `Err(E)` (failure) |
| `Option<T>` | An enum: either `Some(T)` or `None` — for absence, not errors |
| `?` operator | Propagates an error up to the caller automatically |
| `panic!` | Crashes the program immediately — for unrecoverable bugs |
| `unwrap()` | Extracts the value, panics on `Err`/`None` |
| `expect()` | Like `unwrap()` but with a custom panic message |

**Go comparison:**

| Go | Rust |
|----|------|
| `func f() (int, error)` | `fn f() -> Result<i32, SomeError>` |
| `if err != nil { return err }` | `?` operator |
| `errors.New("msg")` | `Err("msg")` or a custom error type |
| `panic("msg")` | `panic!("msg")` |
| `nil` check for missing values | `Option<T>` |

---

### Result\<T, E\>

`Result` is an enum with two variants:

```rust
enum Result<T, E> {
    Ok(T),   // success — contains the value
    Err(E),  // failure — contains the error
}
```

A function that can fail returns `Result`:

```rust
fn divide(a: f64, b: f64) -> Result<f64, String> {
    if b == 0.0 {
        Err(String::from("division by zero"))
    } else {
        Ok(a / b)
    }
}

fn main() {
    match divide(10.0, 2.0) {
        Ok(result) => println!("Result: {}", result),
        Err(e)     => println!("Error: {}", e),
    }
}
```

**Go comparison:**
```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

result, err := divide(10, 2)
if err != nil {
    fmt.Println("Error:", err)
}
```

The key difference: in Go you can ignore the error return value without any complaint. In Rust, if you call a function returning `Result` and do nothing with it, the compiler warns you.

---

### Handling Results

Beyond `match`, there are several ergonomic ways to handle a `Result`:

```rust
let result = divide(10.0, 2.0);

// if let — when you only care about the Ok case
if let Ok(val) = result {
    println!("Got: {}", val);
}

// unwrap_or — provide a default on error
let val = divide(10.0, 0.0).unwrap_or(0.0); // 0.0

// unwrap_or_else — compute the default lazily
let val = divide(10.0, 0.0).unwrap_or_else(|e| {
    println!("Error: {}", e);
    -1.0
});

// map — transform the Ok value, leave Err unchanged
let doubled = divide(10.0, 2.0).map(|v| v * 2.0); // Ok(10.0)

// is_ok / is_err — simple boolean checks
if divide(10.0, 2.0).is_ok() {
    println!("Success!");
}
```

---

### The ? Operator — Propagating Errors

The `?` operator is Rust's equivalent of `if err != nil { return err }`. It unwraps `Ok` values and returns `Err` early:

```rust
use std::num::ParseIntError;

fn parse_and_double(s: &str) -> Result<i32, ParseIntError> {
    let n = s.parse::<i32>()?; // returns Err early if parse fails
    Ok(n * 2)
}

fn main() {
    println!("{:?}", parse_and_double("5"));   // Ok(10)
    println!("{:?}", parse_and_double("abc")); // Err(ParseIntError { ... })
}
```

Without `?`, you'd write:

```rust
fn parse_and_double(s: &str) -> Result<i32, ParseIntError> {
    let n = match s.parse::<i32>() {
        Ok(n)  => n,
        Err(e) => return Err(e),
    };
    Ok(n * 2)
}
```

**Go comparison:**
```go
func parseAndDouble(s string) (int, error) {
    n, err := strconv.Atoi(s)
    if err != nil {
        return 0, err  // manual propagation
    }
    return n * 2, nil
}
```

`?` does exactly what Go's `if err != nil { return 0, err }` does, in one character.

**Important:** `?` can only be used in functions that return `Result` (or `Option`). Using it in `main` requires changing `main`'s signature:

```rust
fn main() -> Result<(), Box<dyn std::error::Error>> {
    let n = "42".parse::<i32>()?;
    println!("{}", n);
    Ok(())
}
```

---

### Chaining with ?

`?` really shines when multiple operations can fail:

```rust
use std::fs;
use std::num::ParseIntError;

fn read_number_from_file(path: &str) -> Result<i32, Box<dyn std::error::Error>> {
    let contents = fs::read_to_string(path)?;   // IO error?
    let trimmed = contents.trim();
    let number = trimmed.parse::<i32>()?;        // parse error?
    Ok(number * 2)
}
```

**Go comparison:**
```go
func readNumberFromFile(path string) (int, error) {
    data, err := os.ReadFile(path)
    if err != nil { return 0, err }
    n, err := strconv.Atoi(strings.TrimSpace(string(data)))
    if err != nil { return 0, err }
    return n * 2, nil
}
```

Each `?` is one less `if err != nil` block.

---

### Option\<T\>

`Option` is for values that might not exist — not errors. Think of it as a type-safe alternative to returning zero values or `-1` as sentinels.

```rust
fn first_even(numbers: &[i32]) -> Option<i32> {
    for &n in numbers {
        if n % 2 == 0 {
            return Some(n);
        }
    }
    None
}

fn main() {
    let nums = vec![1, 3, 4, 7];
    match first_even(&nums) {
        Some(n) => println!("First even: {}", n),
        None    => println!("No even numbers"),
    }
}
```

`Option` has the same ergonomic methods as `Result`:

```rust
let v: Vec<i32> = vec![1, 2, 3];

v.get(0)                    // Some(&1)
v.get(99)                   // None
v.get(0).unwrap_or(&0)      // &1
v.get(0).map(|x| x * 2)    // Some(2)

// ? works with Option too (returns None early)
fn get_first(v: &Vec<i32>) -> Option<i32> {
    let first = v.get(0)?; // returns None if v is empty
    Some(*first * 10)
}
```

**Go comparison:**
```go
// Go: return -1 or a bool to signal absence
func firstEven(nums []int) (int, bool) {
    for _, n := range nums {
        if n%2 == 0 { return n, true }
    }
    return 0, false
}
```

---

### unwrap and expect

`unwrap()` and `expect()` extract the value from `Ok`/`Some` or panic:

```rust
let val = "42".parse::<i32>().unwrap();          // 42
let val = "42".parse::<i32>().expect("not a number"); // 42

let val = "abc".parse::<i32>().unwrap();  // PANIC: called `Result::unwrap()` on an `Err` value
let val = "abc".parse::<i32>().expect("not a number"); // PANIC: not a number: ...
```

**When to use them:**
- `expect()` is fine in tests, quick scripts, and cases where you've already verified the value is valid
- `unwrap()` is acceptable when it's truly impossible to fail (e.g., parsing a hardcoded literal)
- In production code, prefer `?`, `match`, or `unwrap_or` so errors are handled gracefully

**Go comparison:**
```go
// Go equivalent of unwrap: ignore the error return
n, _ := strconv.Atoi("abc") // n = 0, error silently ignored
// Rust's unwrap at least panics loudly rather than silently giving you a wrong value
```

---

### panic! — For Bugs, Not Errors

`panic!` terminates the program immediately with a message and a stack trace. Use it for **programming errors** — situations that should never happen:

```rust
fn get_index(v: &Vec<i32>, i: usize) -> i32 {
    if i >= v.len() {
        panic!("index {} out of bounds for len {}", i, v.len());
    }
    v[i]
}
```

**When panic is appropriate:**
- Violating a contract that callers are responsible for upholding
- Unrecoverable states (corrupted data, broken invariants)
- Tests (`assert!`, `assert_eq!` use panic internally)

**When panic is NOT appropriate:**
- User input is bad → return `Err`
- A file doesn't exist → return `Err`
- A network request fails → return `Err`

**Go comparison:**
```go
panic("something went wrong") // same idea, same use cases
```

---

### Box\<dyn Error\> — Mixing Error Types

When a function can return different error types (e.g., IO errors and parse errors), use `Box<dyn std::error::Error>` as the error type. It's a trait object that can hold any error:

```rust
use std::error::Error;
use std::fs;

fn run() -> Result<(), Box<dyn Error>> {
    let s = fs::read_to_string("data.txt")?; // std::io::Error
    let n: i32 = s.trim().parse()?;          // std::num::ParseIntError
    println!("Number: {}", n);
    Ok(())
}
```

This is the quick solution. For libraries and larger projects you'd define a custom error type or use a crate like `anyhow`, but `Box<dyn Error>` is fine for binaries and scripts.

---

## Common Mistakes

```rust
// WRONG: using unwrap() in production code where errors are expected
let contents = fs::read_to_string("config.txt").unwrap(); // panics if file missing

// RIGHT: propagate the error
let contents = fs::read_to_string("config.txt")?;

// WRONG: using Option where Result would be better
fn parse(s: &str) -> Option<i32> {
    s.parse().ok() // silently discards the error message
}

// RIGHT: return Result so callers know why it failed
fn parse(s: &str) -> Result<i32, std::num::ParseIntError> {
    s.parse()
}

// WRONG: using ? in a function that returns ()
fn main() {
    let n = "42".parse::<i32>()?; // ERROR: ? requires Result return type
}

// RIGHT: change the return type
fn main() -> Result<(), Box<dyn std::error::Error>> {
    let n = "42".parse::<i32>()?;
    Ok(())
}

// WRONG: ignoring a Result with let _ = ...
let _ = std::fs::remove_file("temp.txt"); // silently swallows errors

// RIGHT: handle it, even if minimally
if let Err(e) = std::fs::remove_file("temp.txt") {
    eprintln!("Warning: could not remove temp file: {}", e);
}
```

---

## Your Turn

Open `rust-practice/src/main.rs` and work through these exercises.

**1.** Write a function `parse_age(s: &str) -> Result<u32, String>` that:
- Parses `s` as a `u32`
- Returns `Err("not a number".to_string())` if parsing fails
- Returns `Err("age too large".to_string())` if the number is greater than 150
- Returns `Ok(age)` otherwise

Test it with `"25"`, `"200"`, and `"abc"`.

**2.** Write a function `safe_divide(a: f64, b: f64) -> Option<f64>` that returns `None` if `b` is zero, otherwise `Some(a / b)`. Use it to compute `10 / 2` and `10 / 0`, printing a message for each result using `unwrap_or`.

**3.** Write a function `first_word(s: &str) -> Option<&str>` that returns the first word in a string (split on spaces), or `None` if the string is empty. Then write `first_word_len(s: &str) -> Option<usize>` that uses `?` to call `first_word` and returns the length. Test with `"hello world"` and `""`.

---

### Answers

<details>
<summary>Click to reveal</summary>

```rust
// Exercise 1
fn parse_age(s: &str) -> Result<u32, String> {
    let n: u32 = s.parse().map_err(|_| "not a number".to_string())?;
    if n > 150 {
        return Err("age too large".to_string());
    }
    Ok(n)
}

// Exercise 2
fn safe_divide(a: f64, b: f64) -> Option<f64> {
    if b == 0.0 {
        None
    } else {
        Some(a / b)
    }
}

// Exercise 3
fn first_word(s: &str) -> Option<&str> {
    s.split_whitespace().next()
}

fn first_word_len(s: &str) -> Option<usize> {
    let word = first_word(s)?;
    Some(word.len())
}

fn main() {
    // Exercise 1
    println!("{:?}", parse_age("25"));   // Ok(25)
    println!("{:?}", parse_age("200"));  // Err("age too large")
    println!("{:?}", parse_age("abc"));  // Err("not a number")

    // Exercise 2
    let a = safe_divide(10.0, 2.0).unwrap_or(f64::NAN);
    println!("10 / 2 = {}", a); // 5

    let b = safe_divide(10.0, 0.0).unwrap_or(f64::NAN);
    println!("10 / 0 = {}", b); // NaN

    // Exercise 3
    println!("{:?}", first_word_len("hello world")); // Some(5)
    println!("{:?}", first_word_len(""));            // None
}
```

</details>

---

## Summary

| Concept | Syntax |
|---------|--------|
| Success value | `Ok(value)` |
| Error value | `Err(error)` |
| Match on Result | `match result { Ok(v) => ..., Err(e) => ... }` |
| Propagate error | `value?` |
| Default on error | `.unwrap_or(default)` |
| Transform Ok value | `.map(\|v\| ...)` |
| Panic with message | `.expect("message")` |
| Present value | `Some(value)` |
| Absent value | `None` |
| First or default | `.unwrap_or(default)` |
| Early return None | `value?` (in Option-returning fn) |
| Mixed error types | `Box<dyn std::error::Error>` |
| Crash on bug | `panic!("message")` |

Next up: ownership deep dive — lifetimes, borrowing rules, and why the borrow checker works the way it does.
