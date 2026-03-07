# Java Basics

## Lesson 2: Variables and Data Types

### What Is a Variable?

A variable is a named storage location that holds a value. In Java, every variable has a **type** that you declare upfront — and that type cannot change.

```java
int age = 25;       // age will always be an int
age = "hello";      // compile error! Can't change type
```

This is called **static typing**. Java checks your types at compile time, before the program even runs.

---

## Declaring Variables

The syntax is:

```
type name = value;
```

```java
int score = 100;
double price = 9.99;
boolean isActive = true;
String name = "Jason";
```

You can also declare without assigning immediately:

```java
int score;       // declared but not assigned
score = 100;     // assigned later
```

**Warning:** You cannot *use* a variable before assigning it. Java will refuse to compile:

```java
int x;
System.out.println(x);  // compile error: variable x might not have been initialized
```

---

## Primitive Types

Java has 8 primitive types. These are the most fundamental building blocks — they hold the value directly in memory (no object overhead).

You'll use four of them constantly:

### `int` — Whole Numbers

```java
int age = 25;
int year = 2025;
int temperature = -10;
```

- Range: roughly -2.1 billion to 2.1 billion
- No decimal point — `int score = 9.5;` is a compile error

If you need a bigger whole number, use `long`:

```java
long population = 8_000_000_000L;  // note the L suffix
```

The `_` in number literals is just for readability — Java ignores it.

### `double` — Decimal Numbers

```java
double price = 9.99;
double pi = 3.14159;
double ratio = 0.5;
```

- 64-bit floating point — handles very large and very small decimals
- Use this by default when you need decimals

There's also `float` (32-bit, less precision), but `double` is preferred:

```java
float f = 3.14f;   // needs the f suffix
double d = 3.14;   // no suffix needed
```

### `boolean` — True or False

```java
boolean isLoggedIn = true;
boolean hasError = false;
```

- Only ever `true` or `false` — nothing else
- Used heavily in `if` statements (covered next lesson)
- No shorthand like `1` or `0` — unlike C, `if (1)` is a compile error in Java

### `char` — A Single Character

```java
char grade = 'A';
char symbol = '$';
char newline = '\n';
```

- Always single quotes `'A'` — not double quotes
- Holds exactly one Unicode character
- `char letter = 'AB';` is a compile error

---

## String — Text

`String` is not a primitive — it's a **class** (object type). But it's so common that Java gives it special treatment.

```java
String name = "Jason";
String greeting = "Hello, World!";
String empty = "";
```

- Always double quotes `"..."` — never single quotes
- Can be any length: zero characters to millions
- Strings are **immutable** — once created, the value cannot be changed in memory

### String Concatenation

The `+` operator joins strings together:

```java
String first = "Hello";
String second = "World";
String result = first + ", " + second + "!";
System.out.println(result);  // Hello, World!
```

You can also concatenate strings with other types:

```java
int age = 25;
System.out.println("Age: " + age);   // Age: 25
System.out.println("Pi: " + 3.14);  // Pi: 3.14
```

**Gotcha:** `+` with two numbers adds them, but with a String it concatenates:

```java
System.out.println(1 + 2 + " hello");   // 3 hello  (1+2 first, then concat)
System.out.println("hello " + 1 + 2);   // hello 12 (left to right, both concat)
```

---

## Type Inference with `var`

Since Java 10, you can use `var` and let Java infer the type:

```java
var age = 25;          // Java sees int
var price = 9.99;      // Java sees double
var name = "Jason";    // Java sees String
```

The type is still fixed at compile time — `var` just saves you from writing it. You can't assign a different type later:

```java
var x = 10;
x = "hello";  // still a compile error
```

Use `var` when the type is obvious from the right side. Avoid it when the type would be unclear to someone reading the code.

---

## Common Mistakes

### 1. Confusing `int` and `double`

```java
int result = 7 / 2;      // result = 3, NOT 3.5! (integer division)
double result = 7.0 / 2; // result = 3.5 (decimal division)
```

When both sides are `int`, Java does integer division and throws away the remainder. This trips up beginners constantly — we'll cover it in detail in the arithmetic lesson.

### 2. Mixing up `String` and `char` quotes

```java
char letter = "A";    // compile error — double quotes = String
String name = 'Jason'; // compile error — single quotes = char only
```

### 3. Forgetting to initialize

```java
int x;
int y = x + 1;  // compile error: x not initialized
```

---

## Your Turn

1. What is the difference between `int` and `double`?
2. Why does `"hello " + 1 + 2` print `hello 12` instead of `hello 3`?
3. What is wrong with this code?
   ```java
   char grade = "A";
   ```
4. What does `var` actually do?

### Answers

1. **`int` holds whole numbers only** (no decimal point). **`double` holds decimal numbers** with fractional parts. `int count = 3` is fine, but `int price = 9.99` is a compile error.

2. **Java evaluates `+` left to right.** When the left side is a `String`, `+` becomes string concatenation instead of addition. So `"hello " + 1` produces `"hello 1"`, then `"hello 1" + 2` produces `"hello 12"`. Compare with `1 + 2 + " hello"` where `1 + 2 = 3` first (both ints), then `3 + " hello" = "3 hello"`.

3. **Double quotes are for `String`, not `char`.** A `char` uses single quotes: `char grade = 'A';`

4. **`var` tells Java to infer the type** from the right-hand side. The type is still fixed — it's just shorthand so you don't have to write it explicitly. `var x = 10` is exactly the same as `int x = 10` at compile time.

---

## Summary

| Type      | What it holds              | Example              |
|-----------|----------------------------|----------------------|
| `int`     | Whole numbers              | `42`, `-7`, `0`      |
| `double`  | Decimal numbers            | `3.14`, `-0.5`       |
| `boolean` | True or false only         | `true`, `false`      |
| `char`    | A single character         | `'A'`, `'$'`         |
| `String`  | Text (any length)          | `"hello"`, `""`      |

Key rules:
- Declare type before name: `int age = 25;`
- Type cannot change after declaration
- `String` uses double quotes; `char` uses single quotes
- Use `var` when the type is obvious from the assigned value
