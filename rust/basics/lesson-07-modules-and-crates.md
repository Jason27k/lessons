# Rust Basics

## Lesson 7: Modules and Crates

### The Problem

As projects grow you need to split code across files and control what's visible to other parts of the program. Rust's module system handles this with `mod`, `use`, and `pub`. It also has Cargo for pulling in external libraries from `crates.io` — the Rust equivalent of `pkg.go.dev`.

---

### The Vocabulary

| Term | Meaning |
|------|---------|
| **crate** | A compiled unit — either a binary (`main.rs`) or a library (`lib.rs`). Your project is a crate. |
| **module** | A named scope inside a crate, declared with `mod`. |
| **path** | The address of an item: `crate::module::function` |
| **package** | A Cargo project containing one or more crates (defined by `Cargo.toml`). |

**Go comparison:**

| Go | Rust |
|----|------|
| package | module |
| module (go.mod) | crate / package |
| `import "fmt"` | `use std::fmt` |
| exported (capital letter) | `pub` |
| unexported (lowercase) | no `pub` (default) |

---

### Defining Modules Inline

The simplest form — modules declared directly in a file:

```rust
mod math {
    pub fn add(a: i32, b: i32) -> i32 {
        a + b
    }

    pub fn subtract(a: i32, b: i32) -> i32 {
        a - b
    }

    fn secret() -> i32 { // no pub — private to this module
        42
    }
}

fn main() {
    let result = math::add(5, 3);
    println!("{}", result); // 8

    // math::secret(); // ERROR: private
}
```

Everything in Rust is **private by default**. `pub` makes it accessible from outside the module.

**Go comparison:**
```go
// Go uses capitalization — no explicit pub keyword
func Add(a, b int) int { return a + b }  // exported
func secret() int { return 42 }          // unexported
```

---

### Nested Modules

Modules can be nested arbitrarily deep:

```rust
mod geometry {
    pub mod shapes {
        pub struct Circle {
            pub radius: f64,
        }

        impl Circle {
            pub fn area(&self) -> f64 {
                std::f64::consts::PI * self.radius * self.radius
            }
        }
    }

    pub mod utils {
        pub fn distance(x1: f64, y1: f64, x2: f64, y2: f64) -> f64 {
            ((x2 - x1).powi(2) + (y2 - y1).powi(2)).sqrt()
        }
    }
}

fn main() {
    let c = geometry::shapes::Circle { radius: 5.0 };
    println!("Area: {:.2}", c.area());

    let d = geometry::utils::distance(0.0, 0.0, 3.0, 4.0);
    println!("Distance: {}", d); // 5
}
```

**Note on struct fields:** even if a struct is `pub`, its fields are private by default. You need `pub` on each field you want accessible from outside the module.

---

### use — Bringing Paths into Scope

Typing full paths everywhere is verbose. `use` creates a local shorthand:

```rust
use geometry::shapes::Circle;
use geometry::utils::distance;

fn main() {
    let c = Circle { radius: 3.0 };
    let d = distance(0.0, 0.0, 3.0, 4.0);
}
```

You can bring in multiple items from the same path:

```rust
use std::collections::{HashMap, HashSet, BTreeMap};
```

Or everything from a module with `*` (the glob import — use sparingly):

```rust
use std::collections::*;
```

**Go comparison:**
```go
import (
    "fmt"
    "strings"
)
// Go imports whole packages; Rust imports specific items
```

---

### use with as — Renaming Imports

When two items have the same name, or you want a shorter alias:

```rust
use std::fmt::Result as FmtResult;
use std::io::Result as IoResult;

// Or just shorten a long name
use std::collections::HashMap as Map;
```

**Go comparison:**
```go
import (
    myfmt "fmt"  // aliased import
)
```

---

### super and crate in Paths

Rust provides two keywords for navigating the module tree:

```rust
mod outer {
    pub fn hello() {
        println!("hello from outer");
    }

    mod inner {
        pub fn call_outer() {
            super::hello(); // super = parent module
        }
    }
}

// crate:: refers to the root of the current crate
// use crate::some_module::SomeType;
```

**Go comparison:**
```go
// Go doesn't have relative imports — everything is by full import path
```

---

### File-Based Modules

For real projects you split modules into separate files. There are two conventions:

**Single-file module** (`src/math.rs`):

```
src/
  main.rs
  math.rs
```

In `main.rs`:
```rust
mod math; // tells Rust to look for src/math.rs

fn main() {
    println!("{}", math::add(2, 3));
}
```

In `math.rs`:
```rust
pub fn add(a: i32, b: i32) -> i32 {
    a + b
}
```

**Directory module** (when a module has submodules):

```
src/
  main.rs
  geometry/
    mod.rs      (or: geometry.rs at src level + geometry/ directory)
    shapes.rs
    utils.rs
```

In `main.rs`:
```rust
mod geometry;
```

In `geometry/mod.rs`:
```rust
pub mod shapes;
pub mod utils;
```

In `geometry/shapes.rs`:
```rust
pub struct Circle {
    pub radius: f64,
}
```

**Go comparison:**
```go
// Go: one package per directory, files in same dir share a package
// Rust: you explicitly declare each module with `mod`
```

The key difference from Go: Rust requires you to declare modules explicitly with `mod`. Files don't automatically become part of a package just by being in the same directory.

---

### pub(crate) and Visibility Levels

Beyond `pub` and private, Rust has fine-grained visibility:

```rust
pub struct Foo {
    pub name: String,          // visible everywhere
    pub(crate) id: u32,        // visible within this crate only
    secret: bool,              // visible only within this module
}
```

**`pub(crate)`** is very useful for items you want to share across modules in your crate but not expose to external users.

| Visibility | Accessible from |
|-----------|-----------------|
| (none) | Current module only |
| `pub(super)` | Parent module |
| `pub(crate)` | Anywhere in the crate |
| `pub` | Anywhere (including external crates) |

---

### External Crates — Using Cargo

To use a library from `crates.io`, add it to `Cargo.toml`:

```toml
[dependencies]
rand = "0.8"
serde = { version = "1", features = ["derive"] }
```

Then run `cargo build` (or `cargo add rand` to add it automatically). After that, bring it into scope with `use`:

```rust
use rand::Rng;

fn main() {
    let mut rng = rand::thread_rng();
    let n: u32 = rng.gen_range(1..=100);
    println!("Random number: {}", n);
}
```

**Go comparison:**
```go
// go.mod + go get golang.org/x/...
// Cargo.toml + cargo add rand
```

The workflow is nearly identical to Go modules. `crates.io` is the registry, `Cargo.toml` is your `go.mod`, and `Cargo.lock` is your `go.sum`.

---

### The Standard Library

You've been using `std` already — it's always available without adding anything to `Cargo.toml`:

```rust
use std::collections::HashMap;
use std::fs;
use std::io;
use std::fmt;
```

Some items are in the **prelude** — automatically imported so you never need to `use` them:

- Primitive types (`i32`, `bool`, etc.)
- `String`, `Vec`, `Option`, `Result`
- Common traits like `Clone`, `Copy`, `Iterator`

Everything else needs an explicit `use`.

---

### re-exporting with pub use

You can re-export items to flatten your public API:

```rust
// In src/lib.rs or a module's mod.rs:
mod shapes;
mod utils;

pub use shapes::Circle;    // consumers can use crate::Circle directly
pub use utils::distance;   // instead of crate::shapes::Circle
```

**Go comparison:**
```go
// Go has no re-export mechanism — callers import the original package path
```

This is a common pattern in Rust libraries: keep files organized internally but expose a clean, flat public API.

---

## Common Mistakes

```rust
// WRONG: forgetting pub on a function you need elsewhere
mod greet {
    fn hello() { println!("hello"); } // private!
}
greet::hello(); // ERROR: function is private

// RIGHT:
mod greet {
    pub fn hello() { println!("hello"); }
}

// WRONG: pub struct but private fields
mod point {
    pub struct Point {
        x: f64, // private!
        y: f64,
    }
}
let p = point::Point { x: 1.0, y: 2.0 }; // ERROR: fields are private

// RIGHT: mark fields pub too
pub struct Point {
    pub x: f64,
    pub y: f64,
}

// WRONG: declaring a file module without the file existing
mod network; // ERROR if src/network.rs doesn't exist

// WRONG: using use without mod
use math::add; // ERROR if you haven't declared `mod math;` first
// mod declares the module, use just makes the path shorter
```

---

## Your Turn

Open `rust-practice/src/main.rs` and work through these exercises.

**1.** In `main.rs`, define a module `temperature` with:
- A function `celsius_to_fahrenheit(c: f64) -> f64`
- A function `fahrenheit_to_celsius(f: f64) -> f64`
- A private helper constant or function (anything not marked `pub`)

Call both public functions from `main` and print the results.

**2.** Add a nested module `temperature::display` with a `pub fn show(label: &str, value: f64)` that prints `"label: value°"`. Call it from `main` using the full path, then add a `use` statement so you can call it with just `show(...)`.

**3.** Add `rand` to your `Cargo.toml` (run `cargo add rand` in the `rust-practice` directory). Use it in `main` to generate and print 3 random numbers between 1 and 10.

---

### Answers

<details>
<summary>Click to reveal</summary>

```rust
// Cargo.toml needs: rand = "0.8" (or run `cargo add rand`)
use rand::Rng;

mod temperature {
    const OFFSET: f64 = 32.0; // private to this module

    pub fn celsius_to_fahrenheit(c: f64) -> f64 {
        c * 9.0 / 5.0 + OFFSET
    }

    pub fn fahrenheit_to_celsius(f: f64) -> f64 {
        (f - OFFSET) * 5.0 / 9.0
    }

    pub mod display {
        pub fn show(label: &str, value: f64) {
            println!("{}: {:.1}°", label, value);
        }
    }
}

use temperature::display::show;

fn main() {
    // Exercise 1
    let boiling_f = temperature::celsius_to_fahrenheit(100.0);
    let freezing_c = temperature::fahrenheit_to_celsius(32.0);
    println!("100°C = {}°F", boiling_f);   // 212
    println!("32°F  = {}°C", freezing_c);  // 0

    // Exercise 2
    temperature::display::show("Boiling", boiling_f); // full path
    show("Freezing", freezing_c);                     // via use

    // Exercise 3
    let mut rng = rand::thread_rng();
    println!("Random numbers:");
    for _ in 0..3 {
        let n: u32 = rng.gen_range(1..=10);
        println!("  {}", n);
    }
}
```

</details>

---

## Summary

| Concept | Syntax |
|---------|--------|
| Declare module (inline) | `mod name { ... }` |
| Declare module (file) | `mod name;` (reads `name.rs`) |
| Make item public | `pub fn` / `pub struct` / `pub mod` |
| Crate-internal visibility | `pub(crate)` |
| Import an item | `use crate::module::Item;` |
| Import multiple | `use std::collections::{HashMap, HashSet};` |
| Rename import | `use std::io::Result as IoResult;` |
| Parent module | `super::` |
| Crate root | `crate::` |
| Re-export | `pub use module::Item;` |
| Add external crate | Add to `[dependencies]` in `Cargo.toml` |

Next up: error handling — `Result<T, E>`, the `?` operator, and when to use `panic!`.
