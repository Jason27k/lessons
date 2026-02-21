# Rust Basics

## Lesson 1: Cargo and Hello World

### The Problem

You want to write and run Rust code, but Rust projects aren't just single files. Rust uses **Cargo** — its build system and package manager — for everything. Understanding Cargo first means everything else just works.

If you're coming from Go: Cargo is what `go` is to Go. One tool handles building, running, testing, and dependency management.

---

### Creating a Project

```bash
cargo new hello-world
cd hello-world
```

This creates:

```
hello-world/
├── Cargo.toml    ← like go.mod (project metadata + dependencies)
└── src/
    └── main.rs   ← your entry point (like main.go)
```

**Cargo.toml** looks like:

```toml
[package]
name = "hello-world"
version = "0.1.0"
edition = "2021"

[dependencies]
# external crates go here (like go.mod require blocks)
```

**src/main.rs** starts as:

```rust
fn main() {
    println!("Hello, world!");
}
```

---

### The Essential Cargo Commands

| Command | What it does | Go equivalent |
|---------|-------------|---------------|
| `cargo new <name>` | Create a new project | `go mod init` + manual setup |
| `cargo build` | Compile (debug mode) | `go build` |
| `cargo build --release` | Compile with optimizations | `go build` (Go always optimizes) |
| `cargo run` | Build + run | `go run .` |
| `cargo test` | Run all tests | `go test ./...` |
| `cargo check` | Type-check without producing binary | (no direct equivalent) |
| `cargo add <crate>` | Add a dependency | `go get <package>` |
| `cargo doc --open` | Build and open docs | `godoc` |

> **`cargo check` is your friend.** It's much faster than `cargo build` because it skips code generation. Use it constantly while writing code to catch errors early.

---

### println! — Your First Macro

```rust
fn main() {
    println!("Hello, world!");           // basic
    println!("Hello, {}!", "Jason");     // positional placeholder
    println!("x = {}, y = {}", 1, 2);   // multiple values
    println!("{:?}", vec![1, 2, 3]);     // debug formatting
}
```

The `!` makes `println!` a **macro**, not a function. Macros are expanded at compile time. You'll use them often in Rust but you won't write many until much later.

**Go comparison:**
```go
fmt.Println("Hello, world!")
fmt.Printf("Hello, %s!\n", "Jason")
```

```rust
println!("Hello, world!");
println!("Hello, {}!", "Jason");  // no \n needed, println! adds it
```

---

### Debug Output

When you want to print any value to inspect it, use `{:?}` (debug format) or `{:#?}` (pretty-printed):

```rust
fn main() {
    let nums = vec![1, 2, 3];
    println!("{:?}", nums);    // [1, 2, 3]
    println!("{:#?}", nums);   // pretty-printed, one item per line
}
```

You'll see `{:?}` everywhere. It's the equivalent of `fmt.Println` with `%v` in Go — just dump the value.

---

### Running the Project

```bash
cargo run
```

Output:
```
   Compiling hello-world v0.1.0
    Finished dev [unoptimized + debuginfo] target(s) in 0.42s
     Running `target/debug/hello-world`
Hello, world!
```

The compiled binary lives at `target/debug/hello-world`. You can run it directly too.

---

### Library vs Binary Crates

```bash
cargo new my-app          # binary (has src/main.rs, produces executable)
cargo new my-lib --lib    # library (has src/lib.rs, no main, used by others)
```

A project can have both — `src/main.rs` for the binary and `src/lib.rs` for reusable code. For now, stick with binary crates.

---

## Your Turn

Before we continue, try this:

1. Create a new project called `rust-practice` with `cargo new rust-practice`.
2. Open `src/main.rs` and print your name, the current year, and a vec of 3 numbers using `{:?}`.
3. Run it with `cargo run`.
4. Run `cargo check` — what does it output when there are no errors?

### Answers

1. Standard `cargo new` — creates the directory structure described above.

2. Something like:
```rust
fn main() {
    println!("Jason");
    println!("Year: {}", 2026);
    println!("{:?}", vec![1, 2, 3]);
}
```

3. `cargo run` compiles and executes — you see the Compiling/Finished/Running lines then your output.

4. When there are no errors, `cargo check` just prints:
```
    Checking rust-practice v0.1.0
    Finished dev [unoptimized + debuginfo] target(s) in 0.15s
```
No binary is produced — it just validates your code compiles.

---

## Summary

| Concept | Key point |
|---------|-----------|
| Cargo | The one tool for everything — build, run, test, deps |
| `Cargo.toml` | Project manifest (like `go.mod`) |
| `src/main.rs` | Entry point (like `main.go`) |
| `cargo run` | Build and execute |
| `cargo check` | Fast type-check, no binary |
| `println!("{}", val)` | Print with `{}` placeholder |
| `println!("{:?}", val)` | Debug-print any value |

Next up: variables and Rust's type system.
