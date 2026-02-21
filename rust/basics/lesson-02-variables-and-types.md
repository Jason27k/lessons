# Rust Basics

## Lesson 2: Variables and Types

### The Problem

In Go, you can reassign variables freely. In Rust, variables are **immutable by default**. This isn't a limitation — it's a deliberate design that helps the compiler catch bugs and enable safe concurrency. Understanding this early prevents a lot of confusion.

---

### Variables: Immutable by Default

```rust
let x = 5;
x = 6; // ERROR: cannot assign twice to immutable variable
```

To make a variable mutable, you must say so explicitly:

```rust
let mut x = 5;
x = 6; // fine
```

**Go comparison:**
```go
x := 5    // mutable by default
x = 6     // fine
```

```rust
let mut x = 5;  // must opt in to mutability
x = 6;
```

This feels annoying at first. It becomes natural quickly, and the compiler tells you exactly where to add `mut`.

---

### Type Inference

Rust infers types, just like Go:

```rust
let x = 5;        // Rust infers i32
let y = 3.14;     // Rust infers f64
let name = "hi";  // Rust infers &str
```

You can annotate explicitly:

```rust
let x: i32 = 5;
let y: f64 = 3.14;
```

---

### Integer Types

Rust has many integer types. The key ones:

| Type | Size | Range |
|------|------|-------|
| `i8` | 8-bit | -128 to 127 |
| `i16` | 16-bit | -32768 to 32767 |
| `i32` | 32-bit | -2 billion to 2 billion |
| `i64` | 64-bit | very large |
| `i128` | 128-bit | enormous |
| `isize` | pointer-sized | platform-dependent |
| `u8` | 8-bit unsigned | 0 to 255 |
| `u32` | 32-bit unsigned | 0 to 4 billion |
| `u64` | 64-bit unsigned | very large |
| `usize` | pointer-sized unsigned | used for indexing |

**Default is `i32`** when Rust infers an integer type. Use `i32` for general arithmetic, `usize` for array/vec indices.

**Go comparison:**
```go
var x int = 5    // Go's int is platform-sized (64-bit on 64-bit systems)
var y int64 = 5
```

```rust
let x: i32 = 5;   // Rust's default integer
let y: i64 = 5;
let z: usize = 5; // use this for indexing into Vec/arrays
```

---

### Float Types

```rust
let x = 3.14;      // f64 (default)
let y: f32 = 3.14; // f32 if you need it
```

Go only has `float32` and `float64`. Same in Rust. Default to `f64` unless you have a specific reason.

---

### Booleans and Characters

```rust
let t: bool = true;
let f = false;

let c: char = 'z';       // single quotes
let emoji: char = '😀';  // Rust chars are Unicode (4 bytes)
```

Note: Rust's `char` is 4 bytes (Unicode scalar value). Go's `rune` is also 4 bytes — same concept.

---

### Tuples

Tuples group multiple values of **different types**:

```rust
let tup: (i32, f64, bool) = (500, 6.4, true);

// Destructure:
let (x, y, z) = tup;
println!("x = {}", x);

// Access by index:
println!("{}", tup.0);  // 500
println!("{}", tup.1);  // 6.4
```

**Go comparison:**
```go
// Go doesn't have first-class tuples, but multiple returns are similar:
func swap(a, b int) (int, int) { return b, a }
x, y := swap(1, 2)
```

```rust
// Rust uses tuples for multiple returns too:
fn swap(a: i32, b: i32) -> (i32, i32) { (b, a) }
let (x, y) = swap(1, 2);
```

---

### Arrays

Arrays in Rust have a **fixed length known at compile time**:

```rust
let arr: [i32; 5] = [1, 2, 3, 4, 5];
let first = arr[0];
let len = arr.len();

// Fill with same value:
let zeros = [0; 10];  // [0, 0, 0, 0, 0, 0, 0, 0, 0, 0]
```

**Go comparison:**
```go
var arr [5]int = [5]int{1, 2, 3, 4, 5}  // fixed-size array
slice := []int{1, 2, 3}                  // dynamic slice
```

```rust
let arr = [1, 2, 3, 4, 5];     // fixed-size array
let vec = vec![1, 2, 3];       // dynamic (like Go slice) — covered in collections lesson
```

You'll use `Vec<T>` (Rust's growable array) far more often than arrays. Arrays are for when the size truly never changes.

---

### Shadowing

Rust allows **shadowing** — declaring a new `let` with the same name:

```rust
let x = 5;
let x = x + 1;  // new variable, shadows the old x
let x = x * 2;  // shadows again

println!("{}", x); // 12
```

You can even shadow with a different type:

```rust
let spaces = "   ";        // &str
let spaces = spaces.len(); // usize — different type, same name
```

This is different from `mut`. With `mut`, the type must stay the same. With shadowing, you're creating a new variable that happens to share a name.

**Go has no equivalent** — you can't redeclare a variable in the same scope in Go. Shadowing is a Rust-specific concept.

---

### Constants

Constants are evaluated at compile time and must have an explicit type:

```rust
const MAX_POINTS: u32 = 100_000;
const PI: f64 = 3.14159;
```

Note the underscore in `100_000` — Rust allows underscores in numeric literals for readability. `100_000` = `100000`.

**Go comparison:**
```go
const MaxPoints = 100000
```

```rust
const MAX_POINTS: u32 = 100_000;  // SCREAMING_SNAKE_CASE by convention
```

---

## Common Mistakes

```rust
// WRONG: forgetting mut
let x = 5;
x = 10; // error: cannot assign twice to immutable variable

// WRONG: integer type mismatch
let x: i32 = 5;
let y: i64 = x; // error: expected i64, found i32

// RIGHT: explicit cast needed
let y: i64 = x as i64;
```

Rust does **not** do implicit type coercions. You always cast explicitly with `as`. This is stricter than Go.

---

## Your Turn

Before we continue:

1. What's the difference between `let x = 5` and `let mut x = 5`?
2. What type does Rust infer for `let x = 42`? For `let x = 3.14`?
3. What does this print?
   ```rust
   let x = 5;
   let x = x + 1;
   let x = x * 2;
   println!("{}", x);
   ```
4. Why does this fail, and how do you fix it?
   ```rust
   let a: i32 = 10;
   let b: i64 = a;
   ```

### Answers

1. Without `mut`, the variable cannot be reassigned after the initial binding. With `mut`, it can be reassigned (but the type must stay the same).

2. `i32` for integers, `f64` for floats — these are Rust's defaults.

3. Prints `12`. Shadowing: `5 + 1 = 6`, then `6 * 2 = 12`.

4. Rust doesn't do implicit casts. Fix: `let b: i64 = a as i64;`

---

## Summary

| Concept | Syntax |
|---------|--------|
| Immutable variable | `let x = 5;` |
| Mutable variable | `let mut x = 5;` |
| Explicit type | `let x: i32 = 5;` |
| Default integer | `i32` |
| Default float | `f64` |
| Index type | `usize` |
| Shadowing | `let x = x + 1;` (new variable, same name) |
| Constant | `const MAX: u32 = 100;` |
| Type cast | `x as i64` |
| Tuple | `let t = (1, 2.0, true);` |
| Array | `let a = [1, 2, 3];` |

Next up: functions and control flow.
