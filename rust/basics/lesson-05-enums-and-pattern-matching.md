# Rust Basics

## Lesson 5: Enums and Pattern Matching

### The Problem

Go's enums are just `iota` constants — they're integers with no enforcement and no data attached. Rust enums are a completely different beast. Each variant can carry its own data, and the compiler forces you to handle every variant when you use `match`. Combined with `Option<T>` (Rust's replacement for `nil`), this is one of the most powerful parts of the language.

---

### Basic Enums

```rust
enum Direction {
    North,
    South,
    East,
    West,
}

let d = Direction::North;
```

**Go comparison:**
```go
type Direction int
const (
    North Direction = iota
    South
    East
    West
)
```

Key differences:
- Rust variants are accessed with `::` (`Direction::North`) — they're namespaced to the enum
- Rust variants are not integers — they have no numeric value unless you give them one
- `#[derive(Debug)]` works the same as with structs

---

### match Expressions

`match` is how you work with enums. It's like Go's `switch` but exhaustive — the compiler errors if you miss a variant:

```rust
let d = Direction::North;

match d {
    Direction::North => println!("Going north"),
    Direction::South => println!("Going south"),
    Direction::East  => println!("Going east"),
    Direction::West  => println!("Going west"),
}
```

**Go comparison:**
```go
switch d {
case North:
    fmt.Println("Going north")
// forgot the others — Go doesn't care, Rust won't compile
}
```

The exhaustiveness check is a feature. It means if you add a new variant to an enum, the compiler will point out every `match` that needs updating.

`match` is also an expression — it returns a value:

```rust
let label = match d {
    Direction::North => "N",
    Direction::South => "S",
    Direction::East  => "E",
    Direction::West  => "W",
};
```

---

### Catch-all with `_`

When you don't care about every variant:

```rust
match d {
    Direction::North => println!("Heading north"),
    _ => println!("Going somewhere else"),
}
```

`_` matches anything, like Go's `default`. It must come last.

If you want the catch-all value:

```rust
match d {
    Direction::North => println!("Heading north"),
    other => println!("Going {:?}", other),  // binds the value to `other`
}
```

---

### Enums with Data

This is where Rust enums go far beyond Go. Each variant can hold different data:

```rust
#[derive(Debug)]
enum Shape {
    Circle(f64),           // radius
    Rectangle(f64, f64),   // width, height
    Triangle { base: f64, height: f64 }, // named fields
}
```

**Go comparison:**
```go
// Go can't do this — you'd need an interface + multiple structs
type Shape interface {
    Area() float64
}
```

In Rust you can model this in a single type. Creating variants:

```rust
let c = Shape::Circle(5.0);
let r = Shape::Rectangle(10.0, 4.0);
let t = Shape::Triangle { base: 6.0, height: 3.0 };
```

Extracting the data with `match`:

```rust
fn area(shape: &Shape) -> f64 {
    match shape {
        Shape::Circle(r)                      => std::f64::consts::PI * r * r,
        Shape::Rectangle(w, h)                => w * h,
        Shape::Triangle { base: b, height: h} => 0.5 * b * h,
    }
}
```

The `match` arms destructure the data out of the variant. The compiler ensures all variants are handled.

---

### Option\<T\> — Rust's Answer to nil

Rust has no `nil` or `null`. Instead, the possibility of "no value" is explicit in the type system using `Option<T>`:

```rust
enum Option<T> {
    Some(T),  // there is a value
    None,     // there is no value
}
```

You use it whenever a value might not exist:

```rust
fn find_first_even(nums: &[i32]) -> Option<i32> {
    for &n in nums {
        if n % 2 == 0 {
            return Some(n);
        }
    }
    None
}

let result = find_first_even(&[1, 3, 4, 7]);
```

**Go comparison:**
```go
func findFirstEven(nums []int) (int, bool) {
    for _, n := range nums {
        if n%2 == 0 {
            return n, true
        }
    }
    return 0, false
}
```

Go returns a zero value + boolean. Rust encodes the possibility of absence in the type itself. You **can't** accidentally use `None` as if it were a value — the compiler won't allow it.

Working with `Option` using `match`:

```rust
match result {
    Some(n) => println!("Found: {}", n),
    None    => println!("Not found"),
}
```

---

### Useful Option Methods

You'll use these constantly — they're more concise than a full `match`:

```rust
let maybe: Option<i32> = Some(42);

// Get the value or a default
let val = maybe.unwrap_or(0);        // 42  (or 0 if None)
let val = maybe.unwrap_or_else(|| expensive_default()); // lazy default

// Check if it has a value
maybe.is_some(); // true
maybe.is_none(); // false

// Transform the inner value (only runs if Some)
let doubled = maybe.map(|n| n * 2);  // Some(84)

// Chain operations that also return Option
let result = maybe.and_then(|n| if n > 10 { Some(n) } else { None });
```

**Avoid `unwrap()` in real code** — it panics on `None`. Use it only in examples or when you're certain the value exists.

---

### if let — Concise Single-Variant Matching

When you only care about one variant and want to ignore the rest:

```rust
let result = find_first_even(&[1, 3, 4, 7]);

// Full match (verbose if you only care about Some):
match result {
    Some(n) => println!("Found: {}", n),
    None    => {},
}

// if let (cleaner):
if let Some(n) = result {
    println!("Found: {}", n);
}
```

`if let` is syntactic sugar for a `match` with one arm and a `_ => {}` catch-all. Use it when you only care about one case.

**Go comparison:**
```go
if n, ok := findFirstEven(nums); ok {
    fmt.Println("Found:", n)
}
```

The intent is similar — do something only if the value exists — but Rust's approach encodes it in the type rather than using a separate boolean.

`if let` can have an `else`:

```rust
if let Some(n) = result {
    println!("Found: {}", n);
} else {
    println!("Not found");
}
```

---

### while let

Same idea as `if let`, but loops while the pattern matches:

```rust
let mut stack = vec![1, 2, 3];

while let Some(top) = stack.pop() {
    println!("{}", top); // 3, 2, 1
}
// loop ends when pop() returns None (empty vec)
```

---

### Matching Multiple Patterns and Ranges

```rust
let n = 5;

match n {
    1 | 2       => println!("one or two"),
    3..=5       => println!("three through five"),
    _           => println!("something else"),
}
```

---

### Destructuring in match

You can destructure structs and tuples directly in `match` arms:

```rust
struct Point { x: i32, y: i32 }

let p = Point { x: 3, y: -2 };

match p {
    Point { x: 0, y: 0 } => println!("origin"),
    Point { x, y: 0 }    => println!("on x-axis at {}", x),
    Point { x: 0, y }    => println!("on y-axis at {}", y),
    Point { x, y }       => println!("at ({}, {})", x, y),
}
```

---

### Match Guards

Add a condition to a `match` arm with `if`:

```rust
let n = 7;

match n {
    x if x < 0  => println!("negative: {}", x),
    x if x == 0 => println!("zero"),
    x           => println!("positive: {}", x),
}
```

---

## Common Mistakes

```rust
// WRONG: trying to use an Option value directly
let maybe: Option<i32> = Some(5);
let doubled = maybe * 2; // ERROR: can't multiply Option<i32>

// RIGHT: unwrap safely or use map
let doubled = maybe.map(|n| n * 2);
let doubled = maybe.unwrap_or(0) * 2;

// WRONG: non-exhaustive match
enum Color { Red, Green, Blue }
let c = Color::Red;
match c {
    Color::Red   => println!("red"),
    Color::Green => println!("green"),
    // ERROR: Blue not handled
}

// RIGHT: handle all variants or use _
match c {
    Color::Red   => println!("red"),
    Color::Green => println!("green"),
    Color::Blue  => println!("blue"),
}

// WRONG: using unwrap() carelessly
let val = some_option.unwrap(); // panics if None!

// RIGHT: handle None explicitly
let val = some_option.unwrap_or(default_value);
```

---

## Your Turn

Open `rust-practice/src/main.rs` and replace its contents with solutions to these exercises.

**1.** Define an enum `Coin` with variants `Penny`, `Nickel`, `Dime`, `Quarter`. Write a function `value_in_cents(coin: &Coin) -> u32` that returns the value of each coin using `match`. Create one of each coin and print their values.

**2.** Define an enum `Message` with these variants:
- `Quit`
- `Move { x: i32, y: i32 }`
- `Write(String)`
- `ChangeColor(u8, u8, u8)`

Write a function `process(msg: &Message)` that prints a description of each message. For example, `Move` should print `"Moving to (x, y)"`. Test it with one of each variant.

**3.** Write a function `divide(a: f64, b: f64) -> Option<f64>` that returns `None` if `b` is `0.0`, otherwise returns `Some(a / b)`. Call it twice — once with a valid divisor and once with `0.0`. Use `if let` to print the result of the first call, and `unwrap_or` to print a fallback for the second.

---

### Answers

<details>
<summary>Click to reveal</summary>

```rust
// Exercise 1
#[derive(Debug)]
enum Coin {
    Penny,
    Nickel,
    Dime,
    Quarter,
}

fn value_in_cents(coin: &Coin) -> u32 {
    match coin {
        Coin::Penny   => 1,
        Coin::Nickel  => 5,
        Coin::Dime    => 10,
        Coin::Quarter => 25,
    }
}

// Exercise 2
#[derive(Debug)]
enum Message {
    Quit,
    Move { x: i32, y: i32 },
    Write(String),
    ChangeColor(u8, u8, u8),
}

fn process(msg: &Message) {
    match msg {
        Message::Quit               => println!("Quitting"),
        Message::Move { x, y }     => println!("Moving to ({}, {})", x, y),
        Message::Write(text)        => println!("Writing: {}", text),
        Message::ChangeColor(r, g, b) => println!("Color: ({}, {}, {})", r, g, b),
    }
}

// Exercise 3
fn divide(a: f64, b: f64) -> Option<f64> {
    if b == 0.0 {
        None
    } else {
        Some(a / b)
    }
}

fn main() {
    // Exercise 1
    let coins = [Coin::Penny, Coin::Nickel, Coin::Dime, Coin::Quarter];
    for coin in &coins {
        println!("{:?} = {} cents", coin, value_in_cents(coin));
    }

    // Exercise 2
    let messages = vec![
        Message::Quit,
        Message::Move { x: 10, y: 20 },
        Message::Write(String::from("hello")),
        Message::ChangeColor(255, 0, 128),
    ];
    for msg in &messages {
        process(msg);
    }

    // Exercise 3
    let result = divide(10.0, 3.0);
    if let Some(val) = result {
        println!("10 / 3 = {:.4}", val);
    }

    let bad = divide(5.0, 0.0);
    println!("5 / 0 = {}", bad.unwrap_or(f64::INFINITY));
}
```

</details>

---

## Summary

| Concept | Syntax |
|---------|--------|
| Define enum | `enum Name { Variant1, Variant2 }` |
| Enum with data | `enum Name { Variant(Type), Variant { field: Type } }` |
| Create variant | `Name::Variant` or `Name::Variant(data)` |
| Match | `match val { Pattern => expr, _ => expr }` |
| Match is exhaustive | Compiler errors if any variant is unhandled |
| Match returns value | `let x = match val { ... };` |
| Option some | `Some(value)` |
| Option none | `None` |
| if let | `if let Some(x) = opt { }` |
| while let | `while let Some(x) = iter.next() { }` |
| Safe unwrap | `opt.unwrap_or(default)` |
| Transform Option | `opt.map(\|x\| x * 2)` |

Next up: common collections — `Vec`, `String` (in depth), and `HashMap`.
