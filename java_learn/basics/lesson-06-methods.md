# Java Basics

## Lesson 6: Methods

### The Problem

Imagine you need to print a greeting for multiple users:

```java
System.out.println("Hello, Jason!");
System.out.println("Welcome back.");

System.out.println("Hello, Maria!");
System.out.println("Welcome back.");

System.out.println("Hello, Alex!");
System.out.println("Welcome back.");
```

Every time you add a user, you copy the same lines. If you want to change the greeting, you update it in three places and hope you don't miss one.

A **method** lets you name a block of code and reuse it:

```java
static void greet(String name) {
    System.out.println("Hello, " + name + "!");
    System.out.println("Welcome back.");
}

// Now call it anywhere:
greet("Jason");
greet("Maria");
greet("Alex");
```

---

## Anatomy of a Method

```java
static int add(int a, int b) {
    return a + b;
}
```

Breaking it down:

| Part         | Example      | Meaning                                      |
|--------------|--------------|----------------------------------------------|
| `static`     | `static`     | Belongs to the class, not an object (needed for now — explained below) |
| Return type  | `int`        | The type of value this method sends back      |
| Name         | `add`        | What you call it                              |
| Parameters   | `int a, int b` | Inputs the method receives                  |
| Body         | `{ return a + b; }` | The code that runs                   |
| `return`     | `return a + b` | Sends the value back to the caller          |

---

## `void` — Methods That Return Nothing

When a method doesn't send a value back, use `void` as the return type:

```java
static void printLine() {
    System.out.println("---");
}
```

You call it, it does its work, and nothing comes back:

```java
printLine();  // prints "---"
int x = printLine();  // compile error — void returns nothing
```

If the return type is `void`, you don't write a `return` statement (though you can write bare `return;` to exit early — more on that below).

---

## Parameters and Arguments

**Parameters** are the variables declared in the method signature.
**Arguments** are the actual values you pass when calling the method.

```java
// a and b are parameters
static int multiply(int a, int b) {
    return a * b;
}

// 6 and 7 are arguments
int result = multiply(6, 7);  // result = 42
```

### Multiple parameters

```java
static void describe(String name, int age, double height) {
    System.out.println(name + " is " + age + " years old.");
    System.out.println("Height: " + height);
}

describe("Jason", 25, 5.11);
```

Arguments must be passed **in the same order** as the parameters. Java matches by position, not name.

### No parameters

```java
static void sayHello() {
    System.out.println("Hello!");
}

sayHello();  // empty parentheses — still required
```

---

## Return Values

A method with a non-void return type **must** return a value on every possible path:

```java
static int max(int a, int b) {
    if (a > b) {
        return a;
    }
    return b;  // if a <= b
}
```

You can use the returned value however you like:

```java
int biggest = max(10, 25);          // store it
System.out.println(max(10, 25));    // print it directly
int doubled = max(10, 25) * 2;      // use it in an expression
```

### The method stops at `return`

`return` exits the method immediately — anything after it in the same path doesn't run:

```java
static int absolute(int n) {
    if (n < 0) {
        return -n;  // exits here if n is negative
    }
    return n;       // only reaches here if n >= 0
}
```

---

## Early Return from `void` Methods

You can use `return;` (no value) inside a `void` method to exit early:

```java
static void printPositive(int n) {
    if (n <= 0) {
        return;  // exit early, print nothing
    }
    System.out.println(n);
}
```

This is cleaner than wrapping the whole body in an `if`:

```java
// same result, but more nesting
static void printPositive(int n) {
    if (n > 0) {
        System.out.println(n);
    }
}
```

---

## Why `static`?

Right now all your methods are inside the `Main` class. When you call a method from `main`, both `main` and the method must be `static` — or the compiler complains:

```java
public class Main {
    public static void main(String[] args) {
        greet("Jason");  // works
    }

    static void greet(String name) {       // static — can be called from main
        System.out.println("Hello, " + name + "!");
    }
}
```

If you remove `static` from `greet`:

```java
void greet(String name) { ... }  // non-static
// main tries to call it:
greet("Jason");  // compile error: cannot make a static reference to non-static method
```

`static` means the method belongs to the class itself, not to an object. Since `main` is `static`, it can only directly call other `static` methods. When you learn about objects in the classes lesson, this will make more sense — for now, just put `static` on every method you write.

---

## Method Overloading

You can have multiple methods with the **same name** as long as their parameter lists differ:

```java
static int add(int a, int b) {
    return a + b;
}

static double add(double a, double b) {
    return a + b;
}

static int add(int a, int b, int c) {
    return a + b + c;
}
```

Java picks the right one based on the argument types and count:

```java
add(1, 2);          // calls add(int, int)
add(1.5, 2.5);      // calls add(double, double)
add(1, 2, 3);       // calls add(int, int, int)
```

This is called **overloading**. The method name is the same; the signature (parameter types/count) is different. Return type alone is not enough to distinguish overloads.

---

## Parameters Are Copies

When you pass a primitive to a method, Java passes a **copy** of the value. Changing it inside the method does not affect the original:

```java
static void doubleIt(int x) {
    x = x * 2;
    System.out.println("Inside: " + x);  // 20
}

int n = 10;
doubleIt(n);
System.out.println("Outside: " + n);  // still 10
```

The method gets its own copy of `x`. `n` in the caller is untouched. This is called **pass by value**.

---

## Putting It All Together

```java
public class Main {
    public static void main(String[] args) {
        System.out.println(add(3, 4));       // 7
        System.out.println(max(10, 25));     // 25
        System.out.println(isEven(8));       // true

        printBanner("Java Methods");
    }

    static int add(int a, int b) {
        return a + b;
    }

    static int max(int a, int b) {
        if (a > b) return a;
        return b;
    }

    static boolean isEven(int n) {
        return n % 2 == 0;
    }

    static void printBanner(String title) {
        System.out.println("=================");
        System.out.println("  " + title);
        System.out.println("=================");
    }
}
```

Output:
```
7
25
true
=================
  Java Methods
=================
```

Notice: methods can be defined **after** `main` — Java reads the whole class before running, so order doesn't matter.

---

## Your Turn

1. What is the difference between a parameter and an argument?
2. What happens if a non-void method doesn't return a value on every path?
3. What does `static` on a method mean, and why do your methods need it right now?
4. What is wrong with this code?
   ```java
   static int add(int a, int b) {
       int result = a + b;
   }
   ```
5. What does this print?
   ```java
   static void change(int x) {
       x = 99;
   }

   int n = 5;
   change(n);
   System.out.println(n);
   ```

### Answers

1. **Parameters** are the variable names in the method's definition — they're placeholders. **Arguments** are the actual values passed when calling the method. In `static int add(int a, int b)`, `a` and `b` are parameters. In `add(3, 7)`, `3` and `7` are arguments.

2. **Compile error.** Java requires that every possible path through a non-void method ends with a `return` statement of the correct type. If the compiler detects any path without a return, it refuses to compile.

3. **`static` means the method belongs to the class itself, not to an instance (object).** Since `main` is `static`, it can only directly call other `static` methods without creating an object first. Once you learn about classes and objects, you'll have non-static methods that belong to specific instances.

4. **Missing `return` statement.** The method signature says it returns `int`, but no value is returned. It should be:
   ```java
   static int add(int a, int b) {
       int result = a + b;
       return result;
   }
   ```

5. **`5`** — Java passes primitives by value. `change` receives a copy of `n`. Setting `x = 99` inside the method has no effect on `n` in the caller.

---

## Summary

| Concept          | Example                                |
|------------------|----------------------------------------|
| `void` method    | `static void greet(String name) { }` |
| Return a value   | `static int add(int a, int b) { return a + b; }` |
| Call a method    | `greet("Jason");` / `int x = add(3, 4);` |
| Early return     | `return;` inside a `void` method       |
| Overloading      | Same name, different parameter types/count |

Key rules:
- Every non-`void` method must `return` a value on every possible path
- Parameters are copies of the arguments — changing them doesn't affect the caller
- All methods need `static` while working inside `main` (until you learn objects)
- Methods can be defined in any order inside the class
