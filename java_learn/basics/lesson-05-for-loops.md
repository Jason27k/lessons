# Java Basics

## Lesson 5: For Loops

### Why Loops?

Without loops, repeating work means copying code:

```java
System.out.println("Count: 1");
System.out.println("Count: 2");
System.out.println("Count: 3");
// ... 97 more lines
```

With a loop:

```java
for (int i = 1; i <= 100; i++) {
    System.out.println("Count: " + i);
}
```

---

## The Basic `for` Loop

```java
for (initialization; condition; update) {
    // body — runs each iteration
}
```

Three parts, separated by semicolons:

| Part             | Purpose                                     | Runs            |
|------------------|---------------------------------------------|-----------------|
| `initialization` | Declare and set the loop variable           | Once at the start |
| `condition`      | Keep looping while this is `true`           | Before each iteration |
| `update`         | Change the variable after each iteration    | After each iteration |

```java
for (int i = 0; i < 5; i++) {
    System.out.println(i);
}
// Output: 0 1 2 3 4
```

Step by step:
1. `int i = 0` — create `i`, set to `0`
2. `i < 5` — `0 < 5` is true, enter body, print `0`, then `i++` → `i = 1`
3. `i < 5` — `1 < 5` is true, print `1`, then `i++` → `i = 2`
4. ...continues until `i = 5`
5. `i < 5` — `5 < 5` is false, loop ends

### Common starting points

```java
for (int i = 0; i < 5; i++)  // 0 to 4  (5 iterations)
for (int i = 1; i <= 5; i++) // 1 to 5  (5 iterations)
for (int i = 10; i >= 0; i--) // 10 down to 0 (counting backwards)
for (int i = 0; i < 10; i += 2) // 0, 2, 4, 6, 8 (step by 2)
```

### The loop variable is local

`i` only exists inside the loop. Using it after the loop is a compile error:

```java
for (int i = 0; i < 5; i++) {
    System.out.println(i);
}
System.out.println(i);  // compile error: i is out of scope
```

If you need the value after the loop, declare the variable outside:

```java
int i;
for (i = 0; i < 5; i++) {
    // ...
}
System.out.println(i);  // 5
```

---

## The `while` Loop

A `while` loop repeats as long as a condition is `true`. Use it when you don't know in advance how many iterations you need.

```java
while (condition) {
    // body
}
```

```java
int n = 1;
while (n <= 5) {
    System.out.println(n);
    n++;
}
// Output: 1 2 3 4 5
```

Any `for` loop can be rewritten as a `while` loop:

```java
// for loop
for (int i = 0; i < 5; i++) {
    System.out.println(i);
}

// equivalent while loop
int i = 0;
while (i < 5) {
    System.out.println(i);
    i++;
}
```

**When to use which:**
- `for` — when you know the number of iterations upfront (counting, iterating a range)
- `while` — when the exit condition depends on something that changes unpredictably

### The infinite loop trap

If the condition never becomes `false`, the loop runs forever:

```java
int n = 1;
while (n > 0) {
    n++;  // n keeps growing, never stops
}
```

Always make sure something in the body will eventually make the condition false.

---

## The `do-while` Loop

Like `while`, but the body runs **at least once** before checking the condition:

```java
do {
    // body — always runs at least once
} while (condition);
```

```java
int n = 10;
do {
    System.out.println(n);  // prints 10 even though n > 5 is false
    n++;
} while (n < 5);
// Output: 10
```

The condition is checked **after** the first execution. Useful when you always want to run the body once (like asking for user input and then validating it).

---

## `break` — Exit the Loop Early

`break` immediately stops the loop and jumps past it:

```java
for (int i = 0; i < 10; i++) {
    if (i == 5) {
        break;
    }
    System.out.println(i);
}
// Output: 0 1 2 3 4
```

Once `i == 5`, `break` exits the loop. The remaining iterations (5 through 9) never happen.

### Common use: searching

```java
int target = 7;
boolean found = false;

for (int i = 0; i < 100; i++) {
    if (i == target) {
        found = true;
        break;  // no need to keep going
    }
}

if (found) {
    System.out.println("Found it!");
}
```

---

## `continue` — Skip to the Next Iteration

`continue` skips the rest of the current iteration and jumps to the update step:

```java
for (int i = 0; i < 10; i++) {
    if (i % 2 == 0) {
        continue;  // skip even numbers
    }
    System.out.println(i);
}
// Output: 1 3 5 7 9
```

When `i` is even, `continue` fires and goes straight to `i++`, skipping the `println`.

### `break` vs `continue`

| Keyword    | What it does                                        |
|------------|-----------------------------------------------------|
| `break`    | Exits the loop entirely                             |
| `continue` | Skips the rest of this iteration, starts the next   |

---

## Nested Loops

Loops can be nested inside each other. The inner loop runs completely for each iteration of the outer loop:

```java
for (int row = 1; row <= 3; row++) {
    for (int col = 1; col <= 3; col++) {
        System.out.print(row + "," + col + "  ");
    }
    System.out.println();  // new line after each row
}
```

Output:
```
1,1  1,2  1,3
2,1  2,2  2,3
3,1  3,2  3,3
```

`break` inside a nested loop only exits the **innermost** loop:

```java
for (int i = 0; i < 3; i++) {
    for (int j = 0; j < 3; j++) {
        if (j == 1) break;        // only breaks out of inner loop
        System.out.println(i + " " + j);
    }
}
// Output:
// 0 0
// 1 0
// 2 0
```

---

## Your Turn

1. What are the three parts of a `for` loop and what does each do?
2. What is the difference between `while` and `do-while`?
3. What does `continue` do inside a loop?
4. What does this print?
   ```java
   for (int i = 1; i <= 5; i++) {
       if (i == 3) continue;
       System.out.println(i);
   }
   ```

### Answers

1. **Initialization** — runs once at the start, declares and sets the loop variable. **Condition** — checked before each iteration; loop runs while `true`. **Update** — runs after each iteration, typically increments or decrements the variable.

2. **`while` checks the condition before the first run.** If the condition is `false` from the start, the body never executes. **`do-while` runs the body first, then checks the condition.** The body always executes at least once.

3. **`continue` skips the rest of the current iteration** and jumps to the update step (`i++`), starting the next iteration. It does not exit the loop.

4. **`1, 2, 4, 5`** — when `i == 3`, `continue` skips the `println` for that iteration only. The loop keeps going, printing `4` and `5`.

---

## Summary

| Loop type    | When to use                                        |
|--------------|----------------------------------------------------|
| `for`        | Known number of iterations, counting a range       |
| `while`      | Unknown iterations, exit depends on changing state |
| `do-while`   | Must run at least once regardless of condition     |

| Keyword    | Effect                                                   |
|------------|----------------------------------------------------------|
| `break`    | Exit the loop immediately                                |
| `continue` | Skip remaining body, go to next iteration                |

Key rules:
- `for` loop variable only exists inside the loop body
- Always make sure a `while` loop's condition will eventually become `false`
- `break` in a nested loop only exits the innermost loop
