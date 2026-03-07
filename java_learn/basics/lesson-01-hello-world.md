# Java Basics

## Lesson 1: Your First Java Program

### Why So Much Boilerplate?

If you've seen Python or Go, Java's "Hello World" looks verbose:

```python
# Python
print("Hello, World!")
```

```go
// Go
package main
import "fmt"
func main() { fmt.Println("Hello, World!") }
```

```java
// Java
package java_learn.test;

public class Main {
    public static void main(String[] args) {
        System.out.println("Hello, World!");
    }
}
```

Each piece exists for a reason. Let's go through them one by one.

---

## The Package Declaration

```java
package java_learn.test;
```

A **package** is a namespace — a way to group related files and avoid name collisions. The name must mirror the directory structure on disk:

```
java_learn/
  test/
    Main.java   ← package java_learn.test
```

If you have two classes both named `Main` in different packages, Java can tell them apart:
- `java_learn.test.Main`
- `com.company.Main`

**Rule:** The package name must exactly match the folder path from your source root. If the file is in `src/com/example/`, the package is `com.example`.

Files in the root directory (no subfolder) can omit the package line entirely.

---

## The Class

```java
public class Main {
    // everything goes inside here
}
```

In Java, **all code must live inside a class**. There are no top-level functions like in Go or Python. The class is the fundamental unit of organization.

Two strict rules:
1. **The filename must match the class name** — `Main.java` must contain `public class Main`. If you name the class `App`, the file must be `App.java`.
2. **Only one `public` class per file** — a file can have multiple classes, but only one can be `public`.

The `public` keyword means this class is visible to all other code. You'll learn more about visibility (access modifiers) in a later lesson.

---

## The Main Method

```java
public static void main(String[] args) {
```

This is the **entry point** — the one line Java looks for to start your program. Every keyword here has a job:

| Keyword        | Meaning                                                    |
|----------------|------------------------------------------------------------|
| `public`       | Accessible from outside this class (the JVM needs to call it) |
| `static`       | Belongs to the class itself, not to an object              |
| `void`         | Returns nothing                                            |
| `main`         | The specific name Java looks for as the entry point        |
| `String[] args`| An array of command-line arguments passed in at startup    |

### What does `static` actually mean?

Without `static`, you'd need to create an object before calling the method:

```java
Main obj = new Main();
obj.main();  // Without static, you'd have to do this
```

With `static`, the JVM can call it directly on the class:

```java
Main.main(args);  // This is what the JVM does internally
```

Since the JVM can't create an object before the program starts, `main` must be `static`.

### The `String[] args` parameter

`args` holds any arguments passed when running from the command line:

```bash
java Main hello world
```

```java
// args[0] = "hello"
// args[1] = "world"
System.out.println(args[0]);  // prints: hello
```

For now you won't use this, but the signature still needs to be there for the JVM.

---

## Printing Output

```java
System.out.println("Hello, World!");
```

Breaking this apart:

| Part     | What it is                                              |
|----------|---------------------------------------------------------|
| `System` | A built-in Java class (in `java.lang`, always available) |
| `out`    | A static field on `System` — the standard output stream |
| `println`| A method that prints and then moves to a new line       |

### `println` vs `print`

```java
System.out.println("Hello");  // prints "Hello" + newline
System.out.print("Hello");    // prints "Hello", cursor stays on same line
```

```java
System.out.print("one ");
System.out.print("two ");
System.out.println("three");
// Output: one two three
```

### Printing different types

`println` works with any type — it calls `.toString()` internally:

```java
System.out.println(42);       // integer
System.out.println(3.14);     // decimal
System.out.println(true);     // boolean
System.out.println('A');      // single character
```

---

## Putting It All Together

```java
package java_learn.test;

public class Main {
    public static void main(String[] args) {
        System.out.println("Hello, World!");
        System.out.println("My name is Jason.");
        System.out.print("This ");
        System.out.print("is ");
        System.out.println("one line.");
        System.out.println(2025);
    }
}
```

Output:
```
Hello, World!
My name is Jason.
This is one line.
2025
```

---

## Your Turn

1. What happens if you name your file `App.java` but write `public class Main` inside it?
2. What is the difference between `println` and `print`?
3. Why does `main` need to be `static`?
4. What does `String[] args` represent?

### Answers

1. **Compile error.** Java enforces that the filename matches the public class name. `App.java` must contain `public class App`.

2. **`println`** appends a newline after the output, moving the cursor to the next line. **`print`** leaves the cursor on the same line. If you call `print` twice in a row, the output appears on one line.

3. **Because the JVM needs to call it before any object exists.** `static` means it belongs to the class itself, not an instance. The JVM calls `Main.main(args)` directly — it can't create a `new Main()` first because that's what the program is supposed to set up.

4. **Command-line arguments.** When you run `java Main foo bar`, `args[0]` is `"foo"` and `args[1]` is `"bar"`. It's an array of `String` values passed in from the terminal.

---

## Summary

| Concept                     | What it is                                             |
|-----------------------------|--------------------------------------------------------|
| `package`                   | Declares the namespace; must match the folder path     |
| `public class Main`         | The container for all code; name must match filename   |
| `public static void main`   | Entry point; `static` so JVM can call it directly      |
| `String[] args`             | Command-line arguments passed at startup               |
| `System.out.println(...)`   | Prints a value followed by a newline                   |
| `System.out.print(...)`     | Prints a value with no newline                         |
