# Java Basics

## Lesson 4: If Statements

### Making Decisions

Programs need to do different things depending on conditions. The `if` statement runs a block of code only when a condition is `true`:

```java
int temperature = 30;

if (temperature > 25) {
    System.out.println("It's hot outside.");
}
```

The condition inside the parentheses must be a `boolean` expression — something that evaluates to `true` or `false`.

---

## Basic Structure

```java
if (condition) {
    // runs when condition is true
}
```

```java
if (condition) {
    // runs when condition is true
} else {
    // runs when condition is false
}
```

```java
if (condition1) {
    // runs when condition1 is true
} else if (condition2) {
    // runs when condition1 is false AND condition2 is true
} else {
    // runs when none of the above were true
}
```

Java checks conditions **top to bottom** and stops at the first one that is `true`. Only one branch ever runs.

```java
int score = 85;

if (score >= 90) {
    System.out.println("A");
} else if (score >= 80) {
    System.out.println("B");   // this runs
} else if (score >= 70) {
    System.out.println("C");
} else {
    System.out.println("F");
}
```

Even though `score >= 70` is also true, Java already matched `score >= 80` and stopped.

---

## Comparison Operators

These produce a `boolean` result and are used in conditions:

| Operator | Meaning                  | Example        | Result  |
|----------|--------------------------|----------------|---------|
| `==`     | Equal to                 | `5 == 5`       | `true`  |
| `!=`     | Not equal to             | `5 != 3`       | `true`  |
| `<`      | Less than                | `3 < 5`        | `true`  |
| `>`      | Greater than             | `5 > 3`        | `true`  |
| `<=`     | Less than or equal to    | `5 <= 5`       | `true`  |
| `>=`     | Greater than or equal to | `5 >= 6`       | `false` |

```java
int x = 10;

System.out.println(x == 10);  // true
System.out.println(x != 10);  // false
System.out.println(x > 5);    // true
System.out.println(x <= 9);   // false
```

---

## The `==` vs `=` Trap

This is the most common beginner mistake:

```java
int x = 5;

if (x = 10) {   // COMPILE ERROR in Java (unlike C/C++)
    ...
}
```

`=` is assignment. `==` is comparison. Java won't compile `if (x = 10)` because `x = 10` is an `int`, not a `boolean`. This is actually safer than some other languages.

However, **comparing Strings with `==` is a trap:**

```java
String name = "Jason";

if (name == "Jason") {         // unreliable — compares memory addresses
    System.out.println("hi");
}

if (name.equals("Jason")) {    // correct — compares the actual text
    System.out.println("hi");
}
```

For primitive types (`int`, `double`, `boolean`, `char`), use `==`. For objects (including `String`), use `.equals()`. We'll cover why in a later lesson on objects, but for now: **always use `.equals()` for Strings**.

---

## Logical Operators

Combine multiple conditions into one:

### `&&` — AND (both must be true)

```java
int age = 20;
boolean hasID = true;

if (age >= 18 && hasID) {
    System.out.println("You may enter.");
}
```

Both sides must be `true` for the whole expression to be `true`.

```java
true  && true  = true
true  && false = false
false && true  = false
false && false = false
```

### `||` — OR (at least one must be true)

```java
boolean isWeekend = true;
boolean isHoliday = false;

if (isWeekend || isHoliday) {
    System.out.println("No work today!");
}
```

Only one side needs to be `true`.

```java
true  || true  = true
true  || false = true
false || true  = true
false || false = false
```

### `!` — NOT (flips true/false)

```java
boolean isRaining = false;

if (!isRaining) {
    System.out.println("Go outside.");
}
```

`!true = false`, `!false = true`.

### Short-circuit evaluation

Java stops evaluating as soon as the result is determined:

```java
int x = 0;
if (x != 0 && 10 / x > 1) {  // safe — stops at x != 0 (false), never divides by zero
    ...
}
```

With `&&`, if the left side is `false`, the right side is never checked. With `||`, if the left side is `true`, the right side is never checked.

---

## Nested If Statements

You can put `if` statements inside other `if` statements:

```java
int age = 20;
boolean hasTicket = true;

if (age >= 18) {
    if (hasTicket) {
        System.out.println("Welcome to the show.");
    } else {
        System.out.println("You need a ticket.");
    }
} else {
    System.out.println("You must be 18 or older.");
}
```

Nesting is fine, but more than 2-3 levels deep is hard to read. Often you can flatten nested ifs using `&&`:

```java
if (age >= 18 && hasTicket) {
    System.out.println("Welcome to the show.");
} else if (age < 18) {
    System.out.println("You must be 18 or older.");
} else {
    System.out.println("You need a ticket.");
}
```

---

## Braces Are Optional (But Don't Skip Them)

For a single-line body, braces are technically optional:

```java
if (x > 0)
    System.out.println("positive");  // legal but risky
```

**Don't do this.** It causes bugs:

```java
if (x > 0)
    System.out.println("positive");
    System.out.println("done");    // this runs regardless! not part of the if
```

Always use braces. The second `println` looks like it's inside the `if`, but it isn't. This is a famous source of bugs.

---

## Your Turn

1. What is the difference between `=` and `==`?
2. What is wrong with using `==` to compare two Strings?
3. What does short-circuit evaluation mean for `&&`?
4. What does this print?
   ```java
   int x = 15;
   if (x > 20) {
       System.out.println("A");
   } else if (x > 10) {
       System.out.println("B");
   } else if (x > 5) {
       System.out.println("C");
   } else {
       System.out.println("D");
   }
   ```

### Answers

1. **`=` assigns a value.** `x = 5` sets `x` to `5`. **`==` checks equality.** `x == 5` evaluates to `true` or `false`. Using `=` inside an `if` condition is a compile error in Java.

2. **`==` compares memory addresses, not content.** Two `String` objects with the same text can be stored at different locations in memory, so `==` might return `false` even when the text matches. Use `.equals()` to compare String content.

3. **With `&&`, if the left side is `false`, Java skips the right side entirely** — the overall result must be `false` regardless of the right side. This matters when the right side has side effects or could throw an error (like dividing by zero).

4. **`"B"`** — `x = 15`, so `x > 20` is false, `x > 10` is true. Java matches the second branch and stops. Even though `x > 5` is also true, it's never reached.

---

## Summary

| Concept            | Syntax / Example                         |
|--------------------|------------------------------------------|
| Basic if           | `if (x > 0) { ... }`                    |
| if-else            | `if (...) { ... } else { ... }`          |
| else if chain      | `else if (condition) { ... }`            |
| Equal to           | `==` (not `=`)                           |
| Not equal          | `!=`                                     |
| AND                | `&&` (both true)                         |
| OR                 | `\|\|` (at least one true)               |
| NOT                | `!` (flips boolean)                      |
| String comparison  | `.equals()` not `==`                     |

Key rules:
- Only one branch in an if/else-if/else chain ever runs
- Always use `{` braces `}` even for single-line bodies
- Use `.equals()` for String comparison, `==` for primitives
