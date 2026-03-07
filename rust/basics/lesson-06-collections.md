# Rust Basics

## Lesson 6: Common Collections

### The Problem

Rust's standard library collections are similar to what you've seen in Go — growable arrays, maps, sets — but ownership rules shape how you use them. This lesson covers the three most common: `Vec<T>`, `String`, and `HashMap<K, V>`.

---

### Vec\<T\> — Growable Arrays

`Vec<T>` is Rust's equivalent of Go slices. It's heap-allocated, growable, and typed.

```rust
// Create an empty Vec and push to it
let mut v: Vec<i32> = Vec::new();
v.push(1);
v.push(2);
v.push(3);

// Create with the vec! macro (most common)
let v = vec![1, 2, 3];
```

**Go comparison:**
```go
s := []int{1, 2, 3}
s = append(s, 4)
```

The `vec!` macro is equivalent to a slice literal in Go. You don't need to specify the type if Rust can infer it.

---

### Accessing Elements

Two ways to access elements — one safe, one not:

```rust
let v = vec![10, 20, 30];

// Indexing — panics if out of bounds
let x = v[1]; // 20

// .get() — returns Option<&T>, safe
let x = v.get(1); // Some(&20)
let x = v.get(99); // None
```

**Go comparison:**
```go
s := []int{10, 20, 30}
x := s[1]   // 20 — also panics if out of bounds
// Go has no built-in safe-get for slices
```

Prefer `.get()` when the index might be out of range. Use `[]` only when you're certain it's valid.

---

### Iterating

```rust
let v = vec![1, 2, 3];

// Read-only iteration (most common)
for x in &v {
    println!("{}", x);
}

// Mutable iteration
let mut v = vec![1, 2, 3];
for x in &mut v {
    *x *= 2; // dereference to modify
}
// v is now [2, 4, 6]

// Consuming iteration (takes ownership of v)
for x in v {
    println!("{}", x);
}
// v is no longer usable after this
```

**Go comparison:**
```go
for _, x := range s {
    fmt.Println(x)
}
```

In Go, `range` always copies the value. In Rust, `&v` borrows the elements, `&mut v` gives mutable access, and plain `v` moves them out. The `*x` dereference in the mutable case is required — `x` is a `&mut i32`, so you need `*x` to get to the underlying value.

---

### Common Vec Methods

```rust
let mut v = vec![3, 1, 4, 1, 5, 9];

v.len()           // 6
v.is_empty()      // false
v.push(2)         // add to end
v.pop()           // remove from end → Some(2)
v.contains(&4)    // true
v.sort()          // sort in place: [1, 1, 3, 4, 5, 9]
v.reverse()       // reverse in place
v.dedup()         // remove consecutive duplicates (after sorting): [1, 3, 4, 5, 9]

// Slice operations
let first_three = &v[0..3]; // &[i32] — a slice
```

**Go comparison:**
```go
// Go: sort.Ints(s), len(s), append(s[:i], s[i+1:]...) for removal
// Rust has more built-in methods, fewer manual operations
```

---

### Storing Multiple Types with Enums

A `Vec` must hold a single type. When you need multiple types in one Vec, use an enum:

```rust
#[derive(Debug)]
enum Cell {
    Int(i32),
    Float(f64),
    Text(String),
}

let row = vec![
    Cell::Int(42),
    Cell::Float(3.14),
    Cell::Text(String::from("hello")),
];

for cell in &row {
    match cell {
        Cell::Int(n)    => println!("int: {}", n),
        Cell::Float(f)  => println!("float: {}", f),
        Cell::Text(s)   => println!("text: {}", s),
    }
}
```

**Go comparison:**
```go
// Go uses []interface{} (or []any) and type assertions — less safe
row := []any{42, 3.14, "hello"}
```

---

### String — A Quick Summary

Strings were covered in the mini-lesson, but here's what matters for collections:

```rust
let mut s = String::new();
s.push_str("hello");
s.push(' ');
s.push_str("world");
println!("{}", s); // "hello world"

// format! is the idiomatic way to combine strings
let s = format!("{} {}", "hello", "world");

// Iterating characters
for c in "hello".chars() {
    println!("{}", c);
}

// Useful methods
"hello world".split(' ').collect::<Vec<&str>>() // ["hello", "world"]
"  hello  ".trim()                               // "hello"
"hello".to_uppercase()                           // "HELLO"
"hello world".contains("world")                 // true
```

The main rule: use `&str` for reading, `String` for owning and mutating. Function parameters should take `&str` — it accepts both.

---

### HashMap\<K, V\>

`HashMap` is Rust's equivalent of Go maps. It lives in `std::collections` and must be imported.

```rust
use std::collections::HashMap;

let mut scores: HashMap<String, i32> = HashMap::new();

scores.insert(String::from("Alice"), 100);
scores.insert(String::from("Bob"), 85);
```

**Go comparison:**
```go
scores := map[string]int{
    "Alice": 100,
    "Bob":   85,
}
```

You can also build a HashMap from two Vecs using `zip` and `collect`:

```rust
use std::collections::HashMap;

let names = vec!["Alice", "Bob", "Carol"];
let scores = vec![100, 85, 92];

let map: HashMap<&str, i32> = names.into_iter().zip(scores).collect();
```

---

### Accessing Values

```rust
use std::collections::HashMap;

let mut scores = HashMap::new();
scores.insert(String::from("Alice"), 100);

// .get() returns Option<&V>
let alice = scores.get("Alice"); // Some(&100)
let dave  = scores.get("Dave");  // None

// Direct index — panics if key missing
let score = scores["Alice"]; // 100
```

**Go comparison:**
```go
score, ok := scores["Alice"] // Go's two-value lookup
```

In Go you use `_, ok` to check presence. In Rust, `.get()` returns `Option<&V>`, so you use `match` or `if let`.

---

### Iterating Over a HashMap

```rust
use std::collections::HashMap;

let mut scores = HashMap::new();
scores.insert("Alice", 100);
scores.insert("Bob", 85);

for (name, score) in &scores {
    println!("{}: {}", name, score);
}
```

**Important:** HashMap iteration order is not guaranteed — just like Go maps.

**Go comparison:**
```go
for name, score := range scores {
    fmt.Printf("%s: %d\n", name, score)
}
```

---

### Updating a HashMap

```rust
use std::collections::HashMap;

let mut scores = HashMap::new();

// Overwrite existing value
scores.insert("Alice", 100);
scores.insert("Alice", 200); // Alice is now 200

// Insert only if key is absent (very common pattern)
scores.entry("Bob").or_insert(50);   // inserts 50
scores.entry("Bob").or_insert(999);  // does nothing — Bob already exists

// Modify existing value in-place
let count = scores.entry("Carol").or_insert(0);
*count += 1; // dereference the &mut V returned by or_insert
```

**Go comparison:**
```go
// Go: if _, ok := m[k]; !ok { m[k] = v }
// Rust's entry API is cleaner
```

The `entry` API is one of Rust's most useful HashMap patterns. It's the idiomatic way to "insert or update" and replaces the verbose check-then-insert pattern from Go.

---

### Common HashMap Methods

```rust
use std::collections::HashMap;

let mut m: HashMap<&str, i32> = HashMap::new();
m.insert("a", 1);
m.insert("b", 2);

m.len()           // 2
m.is_empty()      // false
m.contains_key("a") // true
m.remove("a")     // removes "a", returns Some(1)
m.get("b")        // Some(&2)
```

---

### Iterators and collect()

Rust's iterator system is very powerful. You'll use these methods constantly:

```rust
let v = vec![1, 2, 3, 4, 5];

// map — transform each element
let doubled: Vec<i32> = v.iter().map(|x| x * 2).collect();
// [2, 4, 6, 8, 10]

// filter — keep elements matching a predicate
let evens: Vec<&i32> = v.iter().filter(|x| *x % 2 == 0).collect();
// [2, 4]

// fold — reduce to a single value (like Go's loop accumulation)
let sum: i32 = v.iter().fold(0, |acc, x| acc + x);
// 15

// sum and product — shortcuts for common folds
let sum: i32 = v.iter().sum();     // 15
let product: i32 = v.iter().product(); // 120

// any / all
let has_even = v.iter().any(|x| x % 2 == 0);  // true
let all_pos  = v.iter().all(|x| *x > 0);       // true

// find — returns first matching element
let first_even = v.iter().find(|x| *x % 2 == 0); // Some(&2)

// Chain multiple operations (lazy — nothing runs until collect/sum/etc.)
let result: Vec<i32> = v.iter()
    .filter(|x| *x % 2 != 0) // keep odds: [1, 3, 5]
    .map(|x| x * 10)          // multiply: [10, 30, 50]
    .collect();
```

**Go comparison:**
```go
// Go: manual loops
doubled := make([]int, len(v))
for i, x := range v { doubled[i] = x * 2 }

// Rust iterators replace most of these loops cleanly
```

Rust iterators are lazy — no work is done until you call a terminal method like `.collect()`, `.sum()`, or `.for_each()`. Chaining filters and maps costs nothing extra.

---

## Common Mistakes

```rust
// WRONG: indexing a Vec that might be out of bounds
let v = vec![1, 2, 3];
let x = v[10]; // panics at runtime

// RIGHT: use .get() and handle None
if let Some(x) = v.get(10) {
    println!("{}", x);
}

// WRONG: modifying a Vec while iterating over it
let mut v = vec![1, 2, 3];
for x in &v {
    v.push(*x); // ERROR: can't borrow mutably while borrowed immutably
}

// RIGHT: collect into a new Vec, or use indices
let extras: Vec<i32> = v.iter().map(|x| x * 2).collect();
v.extend(extras);

// WRONG: forgetting to use `entry` and doing a redundant lookup
if !scores.contains_key("Alice") {
    scores.insert("Alice", 0);
}

// RIGHT: entry API does this atomically
scores.entry("Alice").or_insert(0);

// WRONG: using a String key but passing a &str to .get()
let mut m: HashMap<String, i32> = HashMap::new();
m.insert(String::from("key"), 1);
let v = m.get("key"); // This actually works — Rust is smart about this via Borrow
// But: m["key".to_string()] works; m[&String::from("key")] works
// Rust's Borrow trait makes m.get("key") work for HashMap<String, _>
```

---

## Your Turn

Open `rust-practice/src/main.rs` and replace its contents with solutions to these exercises.

**1.** Create a `Vec<String>` containing the names `"Alice"`, `"Bob"`, `"Carol"`, and `"Dave"`. Use an iterator to:
- Print only names that are longer than 3 characters.
- Collect the names as uppercase strings into a new `Vec<String>`.
- Print the new Vec.

**2.** Write a function `word_count(text: &str) -> HashMap<&str, usize>` that counts how many times each word appears in the input. Split on spaces. Test it with `"the cat sat on the mat the cat"` and print each word and its count.

**3.** Create a `Vec<i32>` with at least 8 numbers. Using iterator methods (no manual loops), compute and print:
- The sum of all even numbers
- The largest number (hint: `.max()` or `.fold`)
- A new Vec containing only the odd numbers multiplied by 3

---

### Answers

<details>
<summary>Click to reveal</summary>

```rust
use std::collections::HashMap;

// Exercise 1
fn exercise_one() {
    let names = vec![
        String::from("Alice"),
        String::from("Bob"),
        String::from("Carol"),
        String::from("Dave"),
    ];

    // Print names longer than 3 chars
    println!("Longer than 3:");
    for name in names.iter().filter(|n| n.len() > 3) {
        println!("  {}", name);
    }

    // Collect uppercase versions
    let upper: Vec<String> = names.iter().map(|n| n.to_uppercase()).collect();
    println!("Uppercase: {:?}", upper);
}

// Exercise 2
fn word_count<'a>(text: &'a str) -> HashMap<&'a str, usize> {
    let mut counts = HashMap::new();
    for word in text.split(' ') {
        let count = counts.entry(word).or_insert(0);
        *count += 1;
    }
    counts
}

// Exercise 3
fn exercise_three() {
    let numbers = vec![1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

    let even_sum: i32 = numbers.iter().filter(|&&x| x % 2 == 0).sum();
    println!("Sum of evens: {}", even_sum);

    let max = numbers.iter().fold(i32::MIN, |acc, &x| acc.max(x));
    println!("Largest: {}", max);

    let odd_tripled: Vec<i32> = numbers.iter()
        .filter(|&&x| x % 2 != 0)
        .map(|&x| x * 3)
        .collect();
    println!("Odd numbers tripled: {:?}", odd_tripled);
}

fn main() {
    println!("--- Exercise 1 ---");
    exercise_one();

    println!("\n--- Exercise 2 ---");
    let counts = word_count("the cat sat on the mat the cat");
    let mut pairs: Vec<(&&str, &usize)> = counts.iter().collect();
    pairs.sort_by_key(|&(word, _)| *word);
    for (word, count) in pairs {
        println!("  {}: {}", word, count);
    }

    println!("\n--- Exercise 3 ---");
    exercise_three();
}
```

</details>

---

## Summary

| Concept | Syntax |
|---------|--------|
| Create Vec | `vec![1, 2, 3]` or `Vec::new()` |
| Push to Vec | `v.push(x)` |
| Pop from Vec | `v.pop()` → `Option<T>` |
| Safe access | `v.get(i)` → `Option<&T>` |
| Iterate (borrow) | `for x in &v` |
| Iterate (mutate) | `for x in &mut v { *x = ... }` |
| Create HashMap | `HashMap::new()` |
| Insert | `m.insert(key, value)` |
| Safe lookup | `m.get(&key)` → `Option<&V>` |
| Insert-or-keep | `m.entry(key).or_insert(val)` |
| Iterate map | `for (k, v) in &m` |
| Map iterator | `v.iter().map(\|x\| ...)` |
| Filter iterator | `v.iter().filter(\|x\| ...)` |
| Collect | `.collect::<Vec<_>>()` |
| Sum | `v.iter().sum()` |

Next up: modules and crates — how to organize code across files and use external libraries.
