# Rust Basics

## Lesson 3: Functions and Control Flow

### The Problem

Rust's function and control flow syntax looks familiar if you know Go — but there's a key conceptual difference: **Rust is expression-based**. Almost everything returns a value. This changes how you write functions, if statements, and loops in ways that feel strange at first but become very natural.

---

### Functions

```rust
fn greet(name: &str) -> String {
    format!("Hello, {}!", name)
}

fn add(x: i32, y: i32) -> i32 {
    x + y  // no semicolon = this is the return value
}
```

**Go comparison:**
```go
func greet(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}

func add(x, y int) int {
    return x + y
}
```

```rust
fn greet(name: &str) -> String {
    format!("Hello, {}!", name)  // last expression is implicitly returned
}

fn add(x: i32, y: i32) -> i32 {
    x + y  // same — no return keyword needed
}
```

Key differences:
- Parameters: `name: &str` not `name string` (type comes after, but with a colon)
- Return type: `-> String` not `string`
- No explicit `return` needed — the last expression without a semicolon is the return value

---

### Expressions vs Statements

This is the most important conceptual shift in Rust.

**Statement**: performs an action, returns nothing.
**Expression**: evaluates to a value.

```rust
let x = 5;           // statement (the let binding itself)
let y = x + 1;       // x + 1 is an expression; the let binding is a statement

// A block {} is an expression:
let z = {
    let a = 3;
    a * 2       // no semicolon — this is what the block evaluates to
};
println!("{}", z); // 6
```

**The semicolon rule:**
- No semicolon at end of line → expression (has a value)
- Semicolon at end of line → statement (value is discarded, becomes `()` — Rust's "unit" type, like void)

```rust
fn returns_five() -> i32 {
    5       // expression, returns 5
}

fn returns_nothing() {
    5;      // statement, returns () — the 5 is discarded
}
```

This trips everyone up at first. If your function isn't returning what you expect, look for a stray semicolon on the last line.

---

### Explicit return

You can still use `return` when you want to return early:

```rust
fn first_even(nums: &[i32]) -> Option<i32> {
    for &n in nums {
        if n % 2 == 0 {
            return Some(n);  // early return
        }
    }
    None  // implicit return at end (Option covered in lesson 5)
}
```

---

### if / else

`if` in Rust is an **expression** — it returns a value:

```rust
let x = 5;

// Use as statement (like Go):
if x > 3 {
    println!("big");
} else {
    println!("small");
}

// Use as expression (unlike Go):
let label = if x > 3 { "big" } else { "small" };
println!("{}", label);
```

**Go has no equivalent** — Go's `if` is always a statement. In Rust, the ternary operator (`? :`) doesn't exist because `if` itself can be an expression.

Both branches must return the same type:

```rust
let n = 5;
let result = if n > 0 { "positive" } else { 42 }; // ERROR: types don't match
```

---

### loop

`loop` runs forever until you `break`. The `break` can return a value:

```rust
let mut counter = 0;
let result = loop {
    counter += 1;
    if counter == 10 {
        break counter * 2;  // break with a value — result = 20
    }
};
println!("{}", result); // 20
```

**Go comparison:**
```go
for {   // Go's infinite loop
    break
}
```

```rust
loop {  // Rust's infinite loop
    break;
}
```

---

### while

```rust
let mut n = 3;
while n != 0 {
    println!("{}!", n);
    n -= 1;
}
```

Same as Go's `for n != 0 {}`. Rust's `while` doesn't have an init or post clause.

---

### for and Ranges

```rust
// Range (exclusive end):
for i in 0..5 {
    println!("{}", i);  // 0, 1, 2, 3, 4
}

// Range (inclusive end):
for i in 0..=5 {
    println!("{}", i);  // 0, 1, 2, 3, 4, 5
}

// Iterate over a collection:
let fruits = vec!["apple", "banana", "cherry"];
for fruit in &fruits {
    println!("{}", fruit);
}
```

**Go comparison:**
```go
for i := 0; i < 5; i++ { }      // C-style — Rust doesn't have this
for i, v := range slice { }     // range — similar to Rust's for..in
```

```rust
for i in 0..5 { }               // Rust range instead of C-style loop
for (i, v) in slice.iter().enumerate() { } // with index — covered in collections
```

Rust has **no C-style `for` loop** (`for i := 0; i < n; i++`). Use ranges instead.

---

### Loop Labels (breaking nested loops)

```rust
'outer: for i in 0..5 {
    for j in 0..5 {
        if i == 2 && j == 2 {
            break 'outer;  // breaks the outer loop, not just inner
        }
    }
}
```

**Go comparison:**
```go
outer:
for i := 0; i < 5; i++ {
    for j := 0; j < 5; j++ {
        break outer
    }
}
```

Same concept, slightly different syntax (Rust uses `'label:` with a tick).

---

### match (preview)

`match` is Rust's most powerful control flow construct. Full coverage in lesson 5, but here's the basic idea:

```rust
let x = 3;
match x {
    1 => println!("one"),
    2 => println!("two"),
    3 => println!("three"),
    _ => println!("something else"),  // _ is the catch-all (like Go's default)
}
```

`match` is exhaustive — the compiler forces you to handle every possible case.

**Go comparison:**
```go
switch x {
case 1: fmt.Println("one")
case 2: fmt.Println("two")
default: fmt.Println("other")
}
```

`match` is more powerful than `switch` and you'll use it everywhere in Rust.

---

## Common Mistakes

```rust
// WRONG: semicolon on the return expression
fn add(x: i32, y: i32) -> i32 {
    x + y;  // this returns () not i32 — compile error!
}

// RIGHT:
fn add(x: i32, y: i32) -> i32 {
    x + y   // no semicolon
}

// WRONG: mismatched types in if expression
let x = if true { 5 } else { "hello" }; // error: types must match

// WRONG: no C-style for loop
for i := 0; i < 5; i++ { }  // this is Go, not valid Rust

// RIGHT: use ranges
for i in 0..5 { }
```

---

## Your Turn

1. What does this function return, and why?
   ```rust
   fn mystery() -> i32 {
       let x = 5;
       x + 1;
       x
   }
   ```

2. Rewrite this Go code in Rust:
   ```go
   func abs(x int) int {
       if x < 0 {
           return -x
       }
       return x
   }
   ```

3. What does this print?
   ```rust
   let x = loop {
       break 42;
   };
   println!("{}", x);
   ```

### Answers

1. Returns `5`. The line `x + 1;` is a statement (semicolon discards the value). The last expression is `x`, which is `5`.

2. Rust idiomatic version using if as expression:
   ```rust
   fn abs(x: i32) -> i32 {
       if x < 0 { -x } else { x }
   }
   ```

3. Prints `42`. `loop { break 42; }` evaluates to `42`.

---

## Summary

| Concept | Rust syntax |
|---------|-------------|
| Function | `fn name(param: Type) -> ReturnType { }` |
| Implicit return | Last expression without semicolon |
| Explicit return | `return value;` |
| if expression | `let x = if cond { val1 } else { val2 };` |
| Infinite loop | `loop { }` |
| Loop with value | `let x = loop { break value; };` |
| While loop | `while condition { }` |
| For range | `for i in 0..5 { }` |
| For collection | `for item in &vec { }` |
| No semicolon | expression (has a value) |
| Semicolon | statement (value is `()`) |

Next up: structs and `impl` blocks.
