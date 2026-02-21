# Rust Basics - Lesson Plan

## Prerequisites
- Familiarity with another systems or statically-typed language (Go experience applies well)
- Cargo installed (`curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh`)

## Module Goals

By the end of this module you will be able to:
- Create, build, and run Rust projects with Cargo
- Understand Rust's type system and variable semantics
- Write functions and use control flow
- Model data with structs and enums
- Work with common collections (Vec, String, HashMap)
- Organize code with modules

## Lessons

### 1. Cargo and Hello World
- What is Cargo (Rust's build tool + package manager)
- `cargo new`, `cargo build`, `cargo run`, `cargo test`
- Project structure: `src/main.rs`, `Cargo.toml`
- The `println!` macro

### 2. Variables and Types
- `let` bindings and immutability by default
- `mut` keyword
- Scalar types: integers, floats, booleans, chars
- Compound types: tuples, arrays
- Type inference and explicit annotations
- Shadowing

### 3. Functions and Control Flow
- Function syntax, parameters, return types
- Expressions vs statements (Rust is expression-based)
- `if` / `else` as expressions
- `loop`, `while`, `for` loops
- Ranges (`0..5`, `0..=5`)

### 4. Structs
- Defining and instantiating structs
- Field shorthand and struct update syntax
- `impl` blocks and methods
- Associated functions (like Go's package-level constructors)
- Tuple structs

### 5. Enums and Pattern Matching
- Defining enums
- Enums with data (more powerful than Go)
- `match` expressions
- `if let` shorthand
- The `Option<T>` type (Rust's nil-safety mechanism)

### 6. Common Collections
- `Vec<T>` — Rust's growable array (like Go slices)
- `String` vs `&str`
- `HashMap<K, V>` — like Go maps
- Iteration patterns

### 7. Modules and Crates
- `mod`, `use`, `pub`
- File-based modules
- External crates via Cargo.toml
- `crates.io`

### 8. Error Handling Basics
- `panic!` — when things go really wrong
- `Result<T, E>` — the normal way to handle errors
- The `?` operator
- Intro to `unwrap()` and when NOT to use it

## What Comes Next

After basics, the most important Rust-specific concept is **ownership** — how Rust manages memory without a garbage collector. That's covered in the next module and is what makes Rust unique.

## Go Comparison Notes

| Go | Rust |
|----|------|
| `go build` | `cargo build` |
| `go run .` | `cargo run` |
| `go test ./...` | `cargo test` |
| `go.mod` | `Cargo.toml` |
| `var x int = 5` | `let x: i32 = 5;` |
| `x := 5` | `let x = 5;` |
| `struct` | `struct` + `impl` block |
| `interface` | `trait` (covered later) |
| No built-in enum data | Enums with data |
| `error` interface | `Result<T, E>` |
| `nil` | `None` (inside `Option<T>`) |
