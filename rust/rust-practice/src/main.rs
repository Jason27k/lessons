#[derive(Debug)]
enum Coin {
    Penny,
    Nickel,
    Dime,
    Quarter,
}

fn value_in_cents(coin: &Coin) -> u32 {
    match coin {
        Coin::Penny => return 1,
        Coin::Nickel => return 5,
        Coin::Dime => return 10,
        Coin::Quarter => return 25,
    }
}

#[derive(Debug)]
enum Message {
    Quit,
    Move { x: i32, y: i32 },
    Write(String),
    ChangeColor(u8, u8, u8),
}

fn process(msg: &Message) {
    match msg {
        Message::Quit => println!("Quitting"),
        Message::Move { x, y } => println!("Moving to ({}, {})", x, y),
        Message::Write(str) => println!("Writing: {}", str),
        Message::ChangeColor(r, g, b) => println!("Changing color to rgb({}, {}, {})", r, g, b),
    }
}

fn divide(a: f64, b: f64) -> Option<f64> {
    if b == 0.0 { None } else { Some(a / b) }
}

fn main() {
    let coins = [Coin::Penny, Coin::Nickel, Coin::Dime, Coin::Quarter];
    for coin in &coins {
        println!("{:?} = {} cents", coin, value_in_cents(coin));
    }

    // Exercise 2
    let messages = vec![
        Message::Quit,
        Message::Move { x: 10, y: 20 },
        Message::Write(String::from("hello")),
        Message::ChangeColor(255, 0, 128),
    ];
    for msg in &messages {
        process(msg);
    }

    if let Some(i) = divide(10.0, 2.0) {
        println!("{}", i)
    } else {
        println!("None")
    }
    if let Some(i) = divide(10.0, 0.0) {
        println!("{}", i)
    } else {
        println!("None")
    }
}
