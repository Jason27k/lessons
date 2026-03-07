# Java Basics

## Lesson 3: Arithmetic Operators

### The Basics

The five arithmetic operators work on numbers:

```java
int a = 10;
int b = 3;

System.out.println(a + b);  // 13  (addition)
System.out.println(a - b);  // 7   (subtraction)
System.out.println(a * b);  // 30  (multiplication)
System.out.println(a / b);  // 3   (division — watch out!)
System.out.println(a % b);  // 1   (remainder / modulo)
```

Four of these behave exactly as you'd expect. Division is the dangerous one.

---

## Integer Division

When you divide two `int` values, Java **truncates** (cuts off) the decimal part — it does not round:

```java
int a = 7;
int b = 2;

System.out.println(a / b);   // 3, not 3.5
System.out.println(10 / 3);  // 3, not 3.333
System.out.println(1 / 4);   // 0, not 0.25
```

The remainder is simply thrown away. This is called **integer division**.

### Fixing It: Use `double`

If at least one side is a `double`, Java performs decimal division:

```java
System.out.println(7.0 / 2);    // 3.5
System.out.println(7 / 2.0);    // 3.5
System.out.println(7.0 / 2.0);  // 3.5
```

When you have `int` variables and need decimal division, **cast** one of them:

```java
int a = 7;
int b = 2;

double result = (double) a / b;
System.out.println(result);  // 3.5
```

`(double) a` converts `a` to a `double` for this expression only — the variable `a` itself stays an `int`.

### The Trap

```java
double result = a / b;  // WRONG — division happens as int first, then stored
```

Even though `result` is `double`, the division `a / b` is evaluated as integer division first (giving `3`), and then `3` is stored in `result` as `3.0`. Casting must happen **before** the division.

---

## The Modulo Operator `%`

`%` gives you the **remainder** after division:

```java
System.out.println(10 % 3);   // 1  (10 = 3*3 + 1)
System.out.println(15 % 5);   // 0  (15 divides evenly)
System.out.println(7 % 10);   // 7  (7 < 10, so quotient is 0, remainder is 7)
```

### What it's useful for

**Checking if a number is even or odd:**
```java
int n = 8;
System.out.println(n % 2);  // 0 = even, 1 = odd
```

**Wrapping around a range:**
```java
// Cycling through 0, 1, 2, 3, 4, 0, 1, 2, 3, 4, ...
int index = count % 5;
```

**Extracting digits:**
```java
int number = 12345;
int lastDigit = number % 10;  // 5
```

---

## Operator Precedence

Java follows standard math order of operations: multiplication, division, and modulo are evaluated before addition and subtraction. Left to right when tied.

```java
int result = 2 + 3 * 4;     // 14, not 20  (* before +)
int result = 10 - 4 / 2;    // 8, not 3   (/ before -)
int result = 10 % 3 + 1;    // 2, not 0   (% before +)
```

Use parentheses to control the order explicitly:

```java
int result = (2 + 3) * 4;   // 20
int result = (10 - 4) / 2;  // 3
```

When in doubt, add parentheses. They cost nothing and make intent clear.

---

## Compound Assignment Operators

These combine an operation and an assignment:

```java
int x = 10;

x += 5;   // same as: x = x + 5   → x is now 15
x -= 3;   // same as: x = x - 3   → x is now 12
x *= 2;   // same as: x = x * 2   → x is now 24
x /= 4;   // same as: x = x / 4   → x is now 6
x %= 4;   // same as: x = x % 4   → x is now 2
```

These are shorthand — use them to avoid repeating the variable name.

---

## Increment and Decrement

Adding or subtracting 1 is so common that Java has dedicated operators:

```java
int x = 5;

x++;  // x is now 6  (post-increment)
x--;  // x is now 5  (post-decrement)
++x;  // x is now 6  (pre-increment)
--x;  // x is now 5  (pre-decrement)
```

For simple statements like `x++`, pre vs post doesn't matter. The difference shows up when the expression is used as a value:

```java
int x = 5;
int a = x++;   // a = 5, then x becomes 6  (use THEN increment)
int b = ++x;   // x becomes 7, then b = 7  (increment THEN use)
```

**Recommendation:** Use `x++` and `x--` as standalone statements (like in for loops). Avoid embedding them inside larger expressions — it's confusing.

---

## Working with Mixed Types

When you mix `int` and `double` in an expression, Java automatically widens the `int` to `double`:

```java
int a = 5;
double b = 2.5;
double result = a + b;  // a is automatically treated as 5.0
// result = 7.5
```

The result type is always the "larger" type. `int` + `double` → `double`.

Going the other direction (narrowing) requires an explicit cast:

```java
double d = 9.99;
int i = (int) d;   // i = 9  (truncated, not rounded)
```

---

## Your Turn

1. What does `9 / 2` evaluate to in Java?
2. What does `9 % 2` evaluate to?
3. What is wrong with this code?
   ```java
   int a = 7;
   int b = 2;
   double result = a / b;
   System.out.println(result);  // expects 3.5
   ```
4. What does `x++` vs `++x` mean when used inside an expression?

### Answers

1. **`4`** — integer division truncates the decimal. `9 / 2 = 4.5` → `4`.

2. **`1`** — `9 = 2 * 4 + 1`, so the remainder is `1`.

3. **The division happens as integer division first.** `a / b` is evaluated as `7 / 2 = 3` (both ints), and then `3` is widened to `3.0` before being stored. The fix is to cast before dividing: `double result = (double) a / b;`

4. **`x++` uses the current value, then increments.** **`++x` increments first, then uses the new value.**
   ```java
   int x = 5;
   int a = x++;  // a = 5, x = 6
   int b = ++x;  // x = 7, b = 7
   ```

---

## Summary

| Operator | Meaning       | Watch out for                          |
|----------|---------------|----------------------------------------|
| `+`      | Addition      | Concatenation with Strings             |
| `-`      | Subtraction   |                                        |
| `*`      | Multiplication|                                        |
| `/`      | Division      | Integer division truncates the decimal |
| `%`      | Remainder     | Useful for even/odd checks, wrapping   |
| `++`     | Add 1         | Pre vs post matters inside expressions |
| `--`     | Subtract 1    | Pre vs post matters inside expressions |
| `+=`     | Add and assign| Shorthand for `x = x + n`             |

**Key rule:** `int / int` = `int` (truncated). To get decimal division, at least one operand must be `double` — use a cast `(double) x` if needed.
