# Python — Classes

---

### What is a class?

A class is a blueprint for creating objects. An **object** bundles data (attributes) and behavior (methods) together.

```python
class Dog:
    pass  # empty class, valid Python

rex = Dog()  # create an instance (object) of Dog
```

---

### `__init__` — The Constructor

`__init__` runs automatically when you create an instance. It sets up the object's initial data:

```python
class Dog:
    def __init__(self, name, breed):
        self.name = name    # self.name is an instance attribute
        self.breed = breed

rex = Dog("Rex", "Labrador")
print(rex.name)   # Rex
print(rex.breed)  # Labrador
```

`self` refers to the specific instance being created or used. Every method must have `self` as the first parameter — Python passes it automatically, you never provide it yourself.

---

### Instance Attributes

Attributes set with `self.` belong to that specific instance:

```python
class Dog:
    def __init__(self, name, breed):
        self.name = name
        self.breed = breed

rex = Dog("Rex", "Labrador")
fido = Dog("Fido", "Poodle")

print(rex.name)   # Rex
print(fido.name)  # Fido — different object, different data
```

You can also add attributes after creation, though it's generally cleaner to define them all in `__init__`:

```python
rex.age = 3      # valid, but not recommended as a pattern
print(rex.age)   # 3
```

---

### Methods

Methods are functions defined inside a class. They always take `self` as the first parameter:

```python
class Dog:
    def __init__(self, name, breed):
        self.name = name
        self.breed = breed

    def bark(self):
        print(f"{self.name} says: Woof!")

    def describe(self):
        return f"{self.name} is a {self.breed}"

rex = Dog("Rex", "Labrador")
rex.bark()                  # Rex says: Woof!
print(rex.describe())       # Rex is a Labrador
```

Methods can take extra parameters beyond `self`:

```python
class Dog:
    def __init__(self, name, breed):
        self.name = name
        self.breed = breed
        self.tricks = []

    def learn_trick(self, trick):
        self.tricks.append(trick)

    def show_tricks(self):
        if self.tricks:
            print(f"{self.name} knows: {', '.join(self.tricks)}")
        else:
            print(f"{self.name} knows no tricks yet")

rex = Dog("Rex", "Labrador")
rex.learn_trick("sit")
rex.learn_trick("shake")
rex.show_tricks()  # Rex knows: sit, shake
```

---

### Class Attributes

Class attributes are shared across **all** instances of the class:

```python
class Dog:
    species = "Canis lupus familiaris"  # class attribute

    def __init__(self, name):
        self.name = name                # instance attribute

rex = Dog("Rex")
fido = Dog("Fido")

print(rex.species)   # Canis lupus familiaris
print(fido.species)  # Canis lupus familiaris — same value, shared
print(Dog.species)   # can also access via the class itself
```

If you set it on an instance, that instance gets its own copy — the class attribute is not changed:

```python
rex.species = "wolf"   # creates an instance attribute that shadows the class one
print(rex.species)     # wolf
print(fido.species)    # Canis lupus familiaris — unchanged
print(Dog.species)     # Canis lupus familiaris — unchanged
```

---

### `__str__` — Readable Representation

By default, printing an object gives something useless like `<__main__.Dog object at 0x...>`. Define `__str__` to fix that:

```python
class Dog:
    def __init__(self, name, breed):
        self.name = name
        self.breed = breed

    def __str__(self):
        return f"Dog(name={self.name}, breed={self.breed})"

rex = Dog("Rex", "Labrador")
print(rex)  # Dog(name=Rex, breed=Labrador)
```

`__str__` is called by `print()` and `str()`. Methods with double underscores on both sides are called **dunder methods** (or magic methods) — Python calls them automatically in specific situations.

---

### Inheritance

A class can inherit from another class, getting all its attributes and methods:

```python
class Animal:
    def __init__(self, name):
        self.name = name

    def eat(self):
        print(f"{self.name} is eating")

class Dog(Animal):          # Dog inherits from Animal
    def __init__(self, name, breed):
        super().__init__(name)  # call Animal's __init__
        self.breed = breed

    def bark(self):
        print(f"{self.name} says: Woof!")

rex = Dog("Rex", "Labrador")
rex.eat()   # inherited from Animal: Rex is eating
rex.bark()  # defined on Dog: Rex says: Woof!
```

`super()` gives you access to the parent class. Calling `super().__init__(name)` runs the parent's constructor so you don't have to duplicate that setup.

---

### Overriding Methods

A child class can replace a parent's method:

```python
class Animal:
    def speak(self):
        print("...")

class Dog(Animal):
    def speak(self):          # overrides Animal.speak
        print("Woof!")

class Cat(Animal):
    def speak(self):          # overrides Animal.speak
        print("Meow!")

animals = [Dog(), Cat(), Animal()]
for a in animals:
    a.speak()
# Woof!
# Meow!
# ...
```

Each object uses its own version of `speak`. This is called **polymorphism** — the same method call behaves differently depending on the object's type.

---

### `isinstance` — Type Checking

```python
rex = Dog("Rex", "Labrador")

print(isinstance(rex, Dog))     # True
print(isinstance(rex, Animal))  # True  — Dog is a subclass of Animal
print(isinstance(rex, Cat))     # False
```

---

## Common Mistakes

```python
# WRONG: forgetting self in method definition
class Dog:
    def bark():         # missing self
        print("Woof!")

rex = Dog()
rex.bark()  # TypeError: bark() takes 0 positional arguments but 1 was given

# RIGHT:
class Dog:
    def bark(self):
        print("Woof!")

# WRONG: forgetting to call super().__init__ in child class
class Dog(Animal):
    def __init__(self, name, breed):
        # forgot super().__init__(name)
        self.breed = breed

rex = Dog("Rex", "Labrador")
rex.eat()  # AttributeError: 'Dog' object has no attribute 'name'

# RIGHT:
class Dog(Animal):
    def __init__(self, name, breed):
        super().__init__(name)
        self.breed = breed
```

---

## Your Turn

Create a file `practice.py` and work through these:

**1.** Create a `BankAccount` class with:
- `__init__(self, owner, balance=0)` — `balance` defaults to 0
- `deposit(self, amount)` — adds to balance
- `withdraw(self, amount)` — subtracts from balance, but print `"Insufficient funds"` and do nothing if the amount exceeds the balance
- `__str__` that returns `"owner's account: $balance"`

**2.** Create a `Shape` base class with an `area(self)` method that returns `0`. Then create:
- `Rectangle(Shape)` with `width` and `height` — `area` returns `width * height`
- `Circle(Shape)` with `radius` — `area` returns `π * radius²` (use `import math` and `math.pi`)

Make a list of one `Rectangle` and one `Circle` and print each shape's area.

**3.** Add a `__str__` to `Rectangle` and `Circle` that returns something like `"Rectangle(4x5)"` and `"Circle(r=3)"`.

---

### Answers

<details>
<summary>Click to reveal</summary>

```python
import math

# Exercise 1
class BankAccount:
    def __init__(self, owner, balance=0):
        self.owner = owner
        self.balance = balance

    def deposit(self, amount):
        self.balance += amount

    def withdraw(self, amount):
        if amount > self.balance:
            print("Insufficient funds")
        else:
            self.balance -= amount

    def __str__(self):
        return f"{self.owner}'s account: ${self.balance}"

account = BankAccount("Alice", 100)
account.deposit(50)
account.withdraw(30)
account.withdraw(200)  # Insufficient funds
print(account)         # Alice's account: $120


# Exercise 2 & 3
class Shape:
    def area(self):
        return 0

class Rectangle(Shape):
    def __init__(self, width, height):
        self.width = width
        self.height = height

    def area(self):
        return self.width * self.height

    def __str__(self):
        return f"Rectangle({self.width}x{self.height})"

class Circle(Shape):
    def __init__(self, radius):
        self.radius = radius

    def area(self):
        return math.pi * self.radius ** 2

    def __str__(self):
        return f"Circle(r={self.radius})"

shapes = [Rectangle(4, 5), Circle(3)]
for shape in shapes:
    print(f"{shape} — area: {shape.area():.2f}")
# Rectangle(4x5) — area: 20.00
# Circle(r=3) — area: 28.27
```

</details>

---

## Summary

| Concept | Syntax |
|---------|--------|
| Define class | `class Name:` |
| Constructor | `def __init__(self, ...):` |
| Instance attribute | `self.attr = value` |
| Class attribute | defined at class level, outside methods |
| Method | `def method(self, ...):` |
| Readable print | `def __str__(self): return "..."` |
| Inheritance | `class Child(Parent):` |
| Call parent init | `super().__init__(...)` |
| Override method | redefine method in child class |
| Type check | `isinstance(obj, ClassName)` |
