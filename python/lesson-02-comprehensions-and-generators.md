# Python — Comprehensions & Generators

---

### The Problem

Python's most idiomatic way to build and process collections looks nothing like Go, Rust, or JavaScript. Instead of writing a loop that builds up a list, you write a single expression that describes what the list should contain. This is more concise, often faster, and considered standard Python style.

---

### List Comprehensions

The basic pattern: `[expression for item in iterable]`

```python
# The loop way:
squares = []
for n in range(5):
    squares.append(n ** 2)

# The comprehension way:
squares = [n ** 2 for n in range(5)]

print(squares)  # [0, 1, 4, 9, 16]
```

You can add a filter condition at the end:

```python
# Only even squares
even_squares = [n ** 2 for n in range(10) if n % 2 == 0]
print(even_squares)  # [0, 4, 16, 36, 64]
```

The full pattern: `[expression for item in iterable if condition]`

You can iterate over anything — lists, strings, ranges, files, etc.:

```python
words = ["hello", "world", "python"]
upper = [w.upper() for w in words]
print(upper)  # ['HELLO', 'WORLD', 'PYTHON']

lengths = [len(w) for w in words]
print(lengths)  # [5, 5, 6]
```

---

### Dict Comprehensions

Same idea, but produces a dictionary:

```python
# { key: value for item in iterable }
words = ["hello", "world", "python"]
word_lengths = {w: len(w) for w in words}
print(word_lengths)  # {'hello': 5, 'world': 5, 'python': 6}
```

With a filter:

```python
# Only words longer than 4 characters
long_words = {w: len(w) for w in words if len(w) > 4}
print(long_words)  # {'hello': 5, 'world': 5, 'python': 6}
```

Reversing a dict (swap keys and values):

```python
original = {"a": 1, "b": 2, "c": 3}
flipped = {v: k for k, v in original.items()}
print(flipped)  # {1: 'a', 2: 'b', 3: 'c'}
```

---

### Set Comprehensions

Like a list comprehension but uses `{}` and produces a set (no duplicates):

```python
numbers = [1, 2, 2, 3, 3, 3, 4]
unique_squares = {n ** 2 for n in numbers}
print(unique_squares)  # {1, 4, 9, 16} — order not guaranteed
```

---

### Nested Comprehensions

You can nest loops inside a comprehension. Read the `for` clauses left to right (outer to inner):

```python
# Flatten a 2D list
matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
flat = [n for row in matrix for n in row]
print(flat)  # [1, 2, 3, 4, 5, 6, 7, 8, 9]

# All (x, y) pairs where x != y
pairs = [(x, y) for x in range(3) for y in range(3) if x != y]
print(pairs)  # [(0, 1), (0, 2), (1, 0), (1, 2), (2, 0), (2, 1)]
```

Don't nest more than two levels deep — it becomes hard to read.

---

### enumerate — Looping with an Index

When you need both the index and the value, use `enumerate` instead of `range(len(...))`:

```python
fruits = ["apple", "banana", "cherry"]

# The Go/C way — don't do this in Python:
for i in range(len(fruits)):
    print(i, fruits[i])

# The Python way:
for i, fruit in enumerate(fruits):
    print(i, fruit)
# 0 apple
# 1 banana
# 2 cherry
```

`enumerate` starts at 0 by default. You can change the start:

```python
for i, fruit in enumerate(fruits, start=1):
    print(f"{i}. {fruit}")
# 1. apple
# 2. banana
# 3. cherry
```

In a list comprehension:

```python
indexed = [(i, w.upper()) for i, w in enumerate(words)]
```

---

### zip — Looping Over Multiple Iterables Together

`zip` pairs up items from multiple iterables:

```python
names = ["Alice", "Bob", "Charlie"]
scores = [95, 82, 78]

for name, score in zip(names, scores):
    print(f"{name}: {score}")
# Alice: 95
# Bob: 82
# Charlie: 78
```

`zip` stops at the shortest iterable. If they're different lengths, extra items are dropped silently.

Building a dict from two lists:

```python
keys = ["a", "b", "c"]
values = [1, 2, 3]
d = dict(zip(keys, values))
print(d)  # {'a': 1, 'b': 2, 'c': 3}
```

In a comprehension:

```python
combined = {name: score for name, score in zip(names, scores)}
```

---

### Generators — Lazy Evaluation

A **generator** produces values one at a time, on demand, instead of building the whole list in memory at once. Use parentheses instead of brackets:

```python
# List comprehension — builds the whole list in memory:
squares_list = [n ** 2 for n in range(1_000_000)]

# Generator expression — computes each value only when asked:
squares_gen = (n ** 2 for n in range(1_000_000))
```

`squares_gen` uses almost no memory. Values are computed one at a time as you iterate:

```python
gen = (n ** 2 for n in range(5))
print(next(gen))  # 0
print(next(gen))  # 1
print(next(gen))  # 4
# ...or just loop over it:
for val in (n ** 2 for n in range(5)):
    print(val)
```

Generators are one-shot — once exhausted, they're done. You can't reset or re-use them.

When to use a generator vs a list:
- Need all values at once, or need to index/slice → **list comprehension**
- Processing a large sequence one item at a time, or just passing to another function like `sum`/`max` → **generator**

```python
# sum() works fine with a generator — no need to build a list first
total = sum(n ** 2 for n in range(1000))
```

---

### Generator Functions

You can write a function that generates values using `yield`:

```python
def count_up(start, end):
    n = start
    while n <= end:
        yield n       # pauses here, returns n, resumes next time
        n += 1

for n in count_up(1, 5):
    print(n)  # 1, 2, 3, 4, 5
```

Each call to `yield` pauses the function and returns a value. The function resumes from that point when the next value is requested. When the function returns (or falls off the end), the generator is exhausted.

A more useful example — reading a large file line by line without loading it all into memory:

```python
def read_lines(filename):
    with open(filename) as f:
        for line in f:
            yield line.strip()

for line in read_lines("data.txt"):
    process(line)  # only one line in memory at a time
```

---

### map and filter (and why comprehensions are usually preferred)

Python has `map()` and `filter()` for functional-style transforms, but comprehensions are generally clearer:

```python
numbers = [1, 2, 3, 4, 5]

# map — apply a function to every element
doubled = list(map(lambda x: x * 2, numbers))
# same as:
doubled = [x * 2 for x in numbers]

# filter — keep elements that match a condition
evens = list(filter(lambda x: x % 2 == 0, numbers))
# same as:
evens = [x for x in numbers if x % 2 == 0]
```

`map` and `filter` return lazy iterators (like generators), so you need `list()` to materialize them. Most Python developers prefer comprehensions because they're easier to read — but `map`/`filter` are common enough that you'll see them in other people's code.

---

## Common Mistakes

```python
# WRONG: confusing [] and () for comprehension vs generator
result = [x * 2 for x in range(5)]   # list — you can index it, reuse it
result = (x * 2 for x in range(5))   # generator — one-shot, no indexing

gen = (x for x in range(3))
for val in gen:
    print(val)          # works: 0, 1, 2
for val in gen:
    print(val)          # prints nothing — generator is exhausted

# WRONG: trying to index a generator
gen = (x for x in range(5))
print(gen[2])  # TypeError: 'generator' object is not subscriptable

# WRONG: nested comprehension read order confusion
# This is NOT a list of lists — it's a flat list:
flat = [n for row in matrix for n in row]  # outer loop first, then inner

# WRONG: using a comprehension when a generator would do
total = sum([x ** 2 for x in range(1_000_000)])  # builds full list unnecessarily
# RIGHT:
total = sum(x ** 2 for x in range(1_000_000))    # generator, much less memory

# WRONG: mutating a list while iterating over it
nums = [1, 2, 3, 4, 5]
for n in nums:
    if n % 2 == 0:
        nums.remove(n)  # skips elements — unpredictable behavior

# RIGHT: build a new list with a comprehension
nums = [n for n in nums if n % 2 != 0]
```

---

## Your Turn

Create a file `practice.py` and work through these:

**1.** Given a list of temperatures in Celsius:
```python
celsius = [0, 20, 37, 100, -10, 25]
```
Use a list comprehension to create a list of Fahrenheit equivalents (formula: `f = c * 9/5 + 32`). Then use another comprehension to filter only the temperatures above freezing (> 32°F).

**2.** Given two lists:
```python
students = ["Alice", "Bob", "Charlie", "Diana"]
grades   = [88, 72, 95, 81]
```
Use `zip` and a dict comprehension to create `{"Alice": 88, "Bob": 72, ...}`. Then use another comprehension to build a new dict containing only students who passed (grade >= 80).

**3.** Write a generator function `fibonacci()` that yields Fibonacci numbers indefinitely (0, 1, 1, 2, 3, 5, 8, ...). Use it with `enumerate` to print the first 10 Fibonacci numbers with their index, like `"F(0) = 0"`, `"F(1) = 1"`, etc.

---

### Answers

<details>
<summary>Click to reveal</summary>

```python
# Exercise 1
celsius = [0, 20, 37, 100, -10, 25]

fahrenheit = [c * 9/5 + 32 for c in celsius]
print(fahrenheit)  # [32.0, 68.0, 98.6, 212.0, 14.0, 77.0]

above_freezing = [f for f in fahrenheit if f > 32]
print(above_freezing)  # [68.0, 98.6, 212.0, 77.0]


# Exercise 2
students = ["Alice", "Bob", "Charlie", "Diana"]
grades   = [88, 72, 95, 81]

grade_book = {student: grade for student, grade in zip(students, grades)}
print(grade_book)  # {'Alice': 88, 'Bob': 72, 'Charlie': 95, 'Diana': 81}

passed = {name: grade for name, grade in grade_book.items() if grade >= 80}
print(passed)  # {'Alice': 88, 'Charlie': 95, 'Diana': 81}


# Exercise 3
def fibonacci():
    a, b = 0, 1
    while True:
        yield a
        a, b = b, a + b

gen = fibonacci()
for i, val in enumerate(gen):
    if i >= 10:
        break
    print(f"F({i}) = {val}")
# F(0) = 0
# F(1) = 1
# F(2) = 1
# F(3) = 2
# F(4) = 3
# F(5) = 5
# F(6) = 8
# F(7) = 13
# F(8) = 21
# F(9) = 34
```

</details>

---

## Summary

| Concept | Syntax |
|---------|--------|
| List comprehension | `[expr for x in iterable]` |
| With filter | `[expr for x in iterable if condition]` |
| Dict comprehension | `{k: v for x in iterable}` |
| Set comprehension | `{expr for x in iterable}` |
| Nested loops | `[expr for a in outer for b in inner]` |
| Index + value | `for i, x in enumerate(iterable)` |
| Pair two iterables | `for a, b in zip(list1, list2)` |
| Generator expression | `(expr for x in iterable)` — lazy, one-shot |
| Generator function | `def f(): yield value` |
| Materialize iterator | `list(map(...))`, `list(filter(...))` |
