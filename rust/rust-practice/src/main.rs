#[derive(Debug)]
struct Circle {
    radius: f64,
}

impl Circle {
    fn new(radius: f64) -> Circle {
        Circle { radius }
    }

    fn area(&self) -> f64 {
        self.radius.powi(2) * std::f64::consts::PI
    }

    fn circumference(&self) -> f64 {
        2.0 * std::f64::consts::PI * self.radius
    }
}

#[derive(Debug)]
struct Person {
    name: String,
    age: u32,
}

impl Person {
    fn new(name: &str, age: u32) -> Person {
        Person {
            name: name.to_string(),
            age,
        }
    }

    fn greet(&self) -> String {
        format!("Hi, I'm {} and I'm {} years old.", self.name, self.age)
    }
}

#[derive(Debug)]
struct Stack(Vec<i32>);

impl Stack {
    fn new() -> Stack {
        Stack(Vec::new())
    }

    fn push(&mut self, val: i32) {
        self.0.push(val);
    }

    fn pop(&mut self) -> Option<i32> {
        self.0.pop()
    }

    fn is_empty(&self) -> bool {
        self.0.is_empty()
    }
}

fn main() {
    let circle = Circle::new(5.0);
    println!("{:?}", circle);
    println!("{}", circle.area());
    println!("{}", circle.circumference());

    let person = Person::new("Luffy", 19);
    println!("{}", person.greet());

    let mut stack = Stack::new();
    stack.push(1);
    stack.push(2);
    stack.push(3);
    stack.pop();
    println!("{}", stack.is_empty());
}
