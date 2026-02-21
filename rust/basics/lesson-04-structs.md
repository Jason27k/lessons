# Rust Basics

## Lesson 4: Structs

### The Problem

Go has structs and methods, so this lesson will feel familiar — but Rust separates data (the `struct`) from behavior (`impl` block) more explicitly, and adds a few features Go doesn't have: struct update syntax, tuple structs, and associated functions that serve as constructors.

---

### Defining a Struct

```rust
struct Point {
    x: f64,
    y: f64,
}
```

**Go comparison:**
```go
type Point struct {
    X float64
    Y float64
}
```

Differences:
- Field visibility: in Go, uppercase = exported. In Rust, all struct fields are **private by default** — you make them public with `pub` (covered in the modules lesson). For now, within the same file, you can access all fields freely.
- Field types use `:` not a space: `x: f64` not `x f64`

---

### Instantiating a Struct

```rust
let p = Point { x: 1.0, y: 2.5 };
println!("{}", p.x); // 1
```

**Go comparison:**
```go
p := Point{X: 1.0, Y: 2.5}
fmt.Println(p.X)
```

To mutate fields, the **entire binding** must be `mut`:

```rust
let mut p = Point { x: 1.0, y: 2.5 };
p.x = 3.0; // OK

let p = Point { x: 1.0, y: 2.5 };
p.x = 3.0; // ERROR: p is not mut
```

You can't make individual fields mutable — it's all or nothing (unlike Go).

---

### Field Shorthand

When a variable name matches the field name, you can omit the repetition:

```rust
let x = 1.0;
let y = 2.5;

let p = Point { x, y }; // shorthand for Point { x: x, y: y }
```

**Go comparison:**
```go
p := Point{X: x, Y: y} // no shorthand in Go
```

---

### Struct Update Syntax

Create a new struct based on an existing one, overriding some fields:

```rust
let p1 = Point { x: 1.0, y: 2.5 };
let p2 = Point { x: 5.0, ..p1 }; // y is copied from p1

println!("{}", p2.y); // 2.5
```

**Go comparison:**
```go
p2 := p1      // copy the whole struct
p2.X = 5.0   // then change what you need
```

Rust's `..p1` syntax is cleaner when you're changing only a few fields.

---

### Printing Structs

The `println!("{}", p)` macro doesn't work on structs by default — structs don't implement the `Display` trait yet. Use `{:?}` for debug printing, but you need to opt in with a derive attribute:

```rust
#[derive(Debug)]
struct Point {
    x: f64,
    y: f64,
}

let p = Point { x: 1.0, y: 2.5 };
println!("{:?}", p);   // Point { x: 1.0, y: 2.5 }
println!("{:#?}", p);  // pretty-printed (multi-line)
```

`#[derive(Debug)]` is a macro that auto-generates the debug printing code. You'll use it constantly in Rust.

---

### impl Blocks — Methods

Methods go in a separate `impl` block:

```rust
#[derive(Debug)]
struct Rectangle {
    width: f64,
    height: f64,
}

impl Rectangle {
    fn area(&self) -> f64 {
        self.width * self.height
    }

    fn is_square(&self) -> bool {
        self.width == self.height
    }
}

let r = Rectangle { width: 10.0, height: 5.0 };
println!("{}", r.area());      // 50
println!("{}", r.is_square()); // false
```

**Go comparison:**
```go
type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}
```

Key differences:
- `&self` is the receiver in Rust (Go uses `r Rectangle` or `r *Rectangle`)
- The `&self` means "borrow this struct immutably" — more on borrowing in the ownership module
- Methods live in `impl TypeName { }` instead of being scattered as top-level functions

---

### &self vs &mut self

| Receiver | Meaning |
|----------|---------|
| `&self` | read-only access, struct is not consumed |
| `&mut self` | can mutate fields |
| `self` | takes ownership, struct is consumed after call |

```rust
impl Rectangle {
    fn area(&self) -> f64 {          // read-only — most common
        self.width * self.height
    }

    fn scale(&mut self, factor: f64) { // mutates self
        self.width *= factor;
        self.height *= factor;
    }
}

let mut r = Rectangle { width: 10.0, height: 5.0 };
r.scale(2.0);
println!("{:?}", r); // Rectangle { width: 20.0, height: 10.0 }
```

---

### Associated Functions (Constructors)

Functions in `impl` that don't take `self` are called **associated functions**. The most common use is constructors, conventionally named `new`:

```rust
impl Rectangle {
    fn new(width: f64, height: f64) -> Rectangle {
        Rectangle { width, height }
    }

    fn square(size: f64) -> Rectangle {
        Rectangle { width: size, height: size }
    }
}

let r = Rectangle::new(10.0, 5.0);   // :: syntax, not dot
let s = Rectangle::square(4.0);
```

**Go comparison:**
```go
func NewRectangle(width, height float64) Rectangle {
    return Rectangle{Width: width, Height: height}
}
```

The `::` syntax is how you call associated functions — same syntax as Go's package-level functions, but scoped to the type.

---

### Tuple Structs

Structs without named fields — useful for lightweight wrappers:

```rust
struct Color(u8, u8, u8);   // RGB
struct Point2D(f64, f64);

let red = Color(255, 0, 0);
println!("{}", red.0); // 255 — access by index

let origin = Point2D(0.0, 0.0);
```

**Go comparison:**
```go
// Go doesn't have this — you'd just use a regular struct or a type alias
type Color [3]uint8
```

Tuple structs are useful when you want a distinct type (not just a type alias) but don't need named fields.

---

### Unit Structs

Structs with no fields at all:

```rust
struct Marker;
let m = Marker;
```

You won't use these much until you're working with traits (a later lesson). They exist to attach behavior to a type that holds no data.

---

### Multiple impl Blocks

You can have multiple `impl` blocks for the same type — they all apply:

```rust
impl Rectangle {
    fn area(&self) -> f64 { self.width * self.height }
}

impl Rectangle {
    fn perimeter(&self) -> f64 { 2.0 * (self.width + self.height) }
}
```

This is valid but usually unnecessary. It becomes useful with traits.

---

## Common Mistakes

```rust
// WRONG: trying to mutate a field on an immutable binding
let r = Rectangle { width: 10.0, height: 5.0 };
r.width = 20.0; // ERROR: r is not mut

// RIGHT:
let mut r = Rectangle { width: 10.0, height: 5.0 };
r.width = 20.0;

// WRONG: calling associated function with dot syntax
let r = Rectangle.new(10.0, 5.0); // ERROR

// RIGHT: use :: syntax
let r = Rectangle::new(10.0, 5.0);

// WRONG: forgetting #[derive(Debug)] when using {:?}
struct Point { x: f64, y: f64 }
println!("{:?}", Point { x: 1.0, y: 2.0 }); // ERROR: Debug not implemented

// RIGHT: add the derive
#[derive(Debug)]
struct Point { x: f64, y: f64 }
```

---

## Your Turn

Open `rust-practice/src/main.rs` and replace its contents with solutions to these exercises.

**1.** Define a `Circle` struct with a `radius: f64` field. Add an `impl` block with:
- `new(radius: f64) -> Circle`
- `area(&self) -> f64` (π × r²; use `std::f64::consts::PI`)
- `circumference(&self) -> f64` (2 × π × r)

Then create a circle with radius `5.0` and print its area and circumference.

**2.** Define a `Person` struct with `name: String` and `age: u32`. Add:
- `new(name: &str, age: u32) -> Person` (hint: `name.to_string()` converts `&str` to `String`)
- `greet(&self) -> String` that returns `"Hi, I'm NAME and I'm AGE years old."`

Create a `Person` and print their greeting.

**3.** Define a `Stack<i32>` (a named struct wrapping a `Vec<i32>`) with:
- `new() -> Stack`
- `push(&mut self, val: i32)`
- `pop(&mut self) -> Option<i32>` (hint: `Vec` already has `.pop()`)
- `is_empty(&self) -> bool`

Push three values, pop one, then print whether it's empty.

---

### Answers

<details>
<summary>Click to reveal</summary>

```rust
use std::f64::consts::PI;

// Exercise 1
struct Circle {
    radius: f64,
}

impl Circle {
    fn new(radius: f64) -> Circle {
        Circle { radius }
    }

    fn area(&self) -> f64 {
        PI * self.radius * self.radius
    }

    fn circumference(&self) -> f64 {
        2.0 * PI * self.radius
    }
}

// Exercise 2
struct Person {
    name: String,
    age: u32,
}

impl Person {
    fn new(name: &str, age: u32) -> Person {
        Person { name: name.to_string(), age }
    }

    fn greet(&self) -> String {
        format!("Hi, I'm {} and I'm {} years old.", self.name, self.age)
    }
}

// Exercise 3
struct Stack {
    data: Vec<i32>,
}

impl Stack {
    fn new() -> Stack {
        Stack { data: Vec::new() }
    }

    fn push(&mut self, val: i32) {
        self.data.push(val);
    }

    fn pop(&mut self) -> Option<i32> {
        self.data.pop()
    }

    fn is_empty(&self) -> bool {
        self.data.is_empty()
    }
}

fn main() {
    let c = Circle::new(5.0);
    println!("Area: {:.2}", c.area());
    println!("Circumference: {:.2}", c.circumference());

    let p = Person::new("Alice", 30);
    println!("{}", p.greet());

    let mut s = Stack::new();
    s.push(1);
    s.push(2);
    s.push(3);
    s.pop();
    println!("Is empty: {}", s.is_empty()); // false
}
```

</details>

---

## Summary

| Concept | Syntax |
|---------|--------|
| Define struct | `struct Name { field: Type }` |
| Instantiate | `Name { field: value }` |
| Field shorthand | `Name { field }` when variable name matches |
| Struct update | `Name { field: val, ..other }` |
| Debug print | `#[derive(Debug)]` + `{:?}` |
| Method | `fn method(&self) -> T { }` in `impl Name { }` |
| Mutable method | `fn method(&mut self)` |
| Constructor | `fn new(...) -> Name { }` — called with `Name::new(...)` |
| Tuple struct | `struct Name(Type1, Type2);` |

Next up: enums and pattern matching — where Rust really starts to diverge from Go.
