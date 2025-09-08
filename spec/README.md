# Naviary

## Core Philosophy

High developer productivity, stable, powerful, and syntactically elegant language

## Goals

1. A compiled language producing lightweight binaries
2. A strongly-typed language with the flexibility to create anonymous objects
3. Support for structured types in anonymous objects
4. Extremely simple syntax
5. Highly powerful type inference system
6. Compiler written in its own language
7. Multithreaded language with robust concurrency support, where threads are lightweight
8. Object-oriented by default but supports functional programming pipelines ( |> )
9. Low learning curve
10. High productivity, High modeling capacity, low learning curve and low memory.

## Types

### Numeric Types

```
// integer types
i8, i16, i32, i64
// float types
f32, f64

// aliases (convenience)
int     // alias decided by build target (i64 on 64-bit targets; i32 on 32-bit)
float   // alias to f64

// Literals
let a = 42          // int
let b = 3.14        // float
let c = 0xFF        // int (hexadecimal)
let d = 0b1010      // int (binary)
let e = 0o755       // int (octal)
```

### Boolean

```
bool // true or false
```

### String

```
string  // UTF-8, immutable

let name = "Naviary"
let message = "Hello, {name}!"  // String interpolation
let multiline = "
    This is a
    multiline string
    "
let escaped = "\"Mr.arthur\" is good man"
```

### Special Types

```
nil // null value, default for all optional types
```

### Array (Dynamic Array)

```
int[]               // Array of integers
string[][]          // 2D array of strings
User[]              // Array of User objects

let numbers = [1, 2, 3, 4, 5]
let matrix = [[1, 2], [3, 4]]
```

### Map (HashMap)

```
// Invariant by default
Map<string, int>    // String key, int value
Map<int, User>      // Integer key, User value

let scores = Map::new<string, int>{
    "Math": 95,
    "Science": 87
}

scores["English"] = 92
let math = scores["Math"]
scores.set("History", 50)
let value = scores.get("History")?
```

### Tuple

```
(age:int, name:string) // Named tuple

let person = (age:10, name:"arthur") // Named tuple
let (age, name) = person // Destructuring
```

### Anonymous Struct

```
// Object literal - structure fixed at compile time
let person = {
    name: "Alice",
    age: 30
}
// Type: { name: string, age: int }
// person.email = "..." // Error! Cannot add fields

// Type alias definition
type Person = {
    name: string
    age: int
    email?: string  // Optional field
}
```

### Optional

```
int?                // nil or int
string?             // nil or string
User?               // nil or User

let maybe: int? = nil
let surely: int? = 42

// nil check
if maybe != nil {
    print(maybe)    // Auto-unwrapped
}

// Nil-coalescing operator
let value = maybe ?? 0

// Optional chaining
let name = user?.name
let length = user?.name?.length()

// Nil-coalescing operator (??) provides a default value when nil
let value = nothing ?? 0 // nothing is nil, so value is 0
```

### Variables

```
// Immutable variables (constants)
let pi = 3.14159
let name = "Alice"

// Mutable variables
let mut counter = 0
counter += 1

// Compile-time constants
const MAX_SIZE = 1000
const VERSION = "1.0.0"

// Type annotation
let age: int = 30
let mut score: float = 0.0
```

### Default

naviary has no default value assignment if there is no optional operator.

```naviary
let x:string? // x == nil
let x:string; // Compile Error
let mut x:string; // Compile error if you don't allocate it later.naviary has no default value assignment.
```

```naviary
let x:string; // Compile Error
let mut x:string; // Compile error if you don't allocate it later.
```

### Decorator

```naviary
@Service
class UserService {}
```

### Functions

#### Basic Functions

```
// Function declaration
func addNumbers(a: int, b: int) -> int {
    a + b  // Last expression is automatically returned
}

// Function with no return value
func printMessage(msg: string) {
    print(msg)
}

// Multiple return values
func divMod(a: int, b: int) -> (int, int) {
    (a / b, a % b)
}

// Early return
func safeDivide(a: int, b: int) -> int? {
    if b == 0 {
        return nil
    }
    a / b
}
```

#### Arrow Functions (Lambdas)

```
// Single expression
let double = (x: int) -> int => x * 2
let add = (a, b) => a + b  // Type inference

// Multiple statements
let process = (data: string) -> string => {
    let trimmed = data.trim()
    let upper = trimmed.toUpper()
    upper  // Last expression returned
}

// In higher-order functions
numbers.map(x => x * 2)
      .filter(x => x > 10)
```

#### Parameter Specification

```
func createConnection(host: string, port: int, timeout: int) -> Connection {
    // ...
}

// Various calling methods
let conn1 = createConnection("localhost", 8080, 30)  // Positional

let conn2 = createConnection(
    host: "localhost",
    port: 8080,
    timeout: 30
)  // Named parameters

let host = "localhost"
let port = 8080
let conn3 = createConnection(host, port, timeout: 30)  // Mixed
```

#### Accepting Anonymous Objects

```naviary
func greet(x: { name: string }) -> string {
    "hello, {x.name}!"
}
```

### Operators

#### Arithmetic Operators

```naviary
// Basic arithmetic
+   // Addition: 3 + 2 = 5
-   // Subtraction: 5 - 2 = 3
*   // Multiplication: 3 * 2 = 6
/   // Division: 6 / 2 = 3
%   // Modulus: 7 % 3 = 1
**  // Exponentiation: 2 ** 3 = 8

// Unary operators
-x  // Negation: -5
```

#### Comparison Operators

```naviary
==  // Equal: a == b
!=  // Not equal: a != b
<   // Less than: a < b
>   // Greater than: a > b
<=  // Less than or equal: a <= b
>=  // Greater than or equal: a >= b
```

#### Logical Operators

```naviary
&&  // AND: true && false = false
||  // OR: true || false = true
!   // NOT: !true = false
```

#### Bitwise Operators

```naviary
&   // AND: 5 & 3 = 1
|   // OR: 5 | 3 = 7
^   // XOR: 5 ^ 3 = 6
~   // NOT: ~5 = -6
<<  // Left shift: 5 << 2 = 20
>>  // Right shift: 20 >> 2 = 5
>>> // Unsigned right shift
```

#### Assignment Operators

```naviary
=   // Basic assignment: x = 5

// Compound assignment
+=  // x += 3  (x = x + 3)
-=  // x -= 3  (x = x - 3)
*=  // x *= 3  (x = x * 3)
/=  // x /= 3  (x = x / 3)
%=  // x %= 3  (x = x % 3)
**= // x **= 3 (x = x ** 3)

// Bitwise compound assignment
&=  // x &= 3
|=  // x |= 3
^=  // x ^= 3
<<= // x <<= 3
>>= // x >>= 3
```

#### Special Operators

```naviary
// Optional-related
?   // Optional type: int?
?.  // Optional chaining: user?.name
??  // Nil-coalescing: value ?? defaultValue

// Range
..  // Exclusive range: 0..10 (0 to 9)
..= // Inclusive range: 0..=10 (0 to 10)

// Member access
.   // Member access: object.field
::  // Static access: Class::method()

// Functional
|>  // Pipeline: value |> func
=>  // Arrow function: x => x * 2

// Channel
<-  // Channel send/receive: ch <- value, value = <-ch

// Pattern matching
_   // Wildcard: match x { _ => "default" }

// Spread
... // Variadic arguments, spread: func(...args), [...arr]
```

#### Operator Precedence

```naviary
1.  ()  []  .  ?.  ::           // Grouping, access
2.  !  ~  -  +                  // Unary
3.  **                          // Exponentiation
4.  *  /  %                     // Multiplication, division
5.  +  -                        // Addition, subtraction
6.  <<  >>  >>>                 // Shift
7.  ..  ..=                     // Range
8.  <  >  <=  >=                // Comparison
9.  ==  !=                      // Equality
10. &                           // Bitwise AND
11. ^                           // Bitwise XOR
12. |                           // Bitwise OR
13. is  as                      // Type
14. &&                          // Logical AND
15. ||                          // Logical OR
16. ??                          // Nil-coalescing
17. |>                          // Pipeline
18. =  +=  -=  *=  etc.         // Assignment
19. <-                          // Channel

// Examples
a + b * c    // (a + (b * c))
x || y && z  // (x || (y && z))
```

### Classes

#### Basic Class

```naviary
class Person(
    name: string
    mut age: int        // Mutable field
    email: string?      // Optional field
) {
    // secondary constructor
    constructor(name: string, age: int, email: string?) {
        this.name = name
        this.age = age
        this.email = email
    }

    // Method, public by default
    func greet() -> string {
        "Hello, I'm {this.name}"
    }

    protected func birthday() {
        this.age += 1
    }

    // Static method
    static func create(name: string) -> Person {
        Person(name, 0, nil)
    }
}

// Usage
let person = Person("Alice", 30)
let greeting = person.greet()
let anonymous = Person::create("Anonymous")
```

#### Inheritance

```naviary
class Animal(name: string) {
    func speak() -> string {
        "Some sound"
    }

    func move() {
        print("{this.name} is moving")
    }
}

class Dog(name:string, breed: string): Animal(name) {

    // secondary constructor
    constructor(name: string, breed: string) {
        super(name)
        this.breed = breed
    }

    // Method override
    override func speak() -> string {
        "Woof!"
    }

    func wagTail() {
        print("{this.name} is wagging tail")
    }
}

// Polymorphism
let animals: Animal[] = [
    Dog("Buddy", "Golden"),
    Animal("Generic")
]

for animal in animals {
    print(animal.speak())
}
```

#### Abstract Class

```naviary
abstract class Shape(x: int, y: int) {
    // Abstract methods
    abstract func area() -> float
    abstract func perimeter() -> float

    // Regular method
    func move(dx: int, dy: int) {
        this.x += dx
        this.y += dy
    }

    func getPosition() -> (int, int) {
        (this.x, this.y)
    }
}

class Circle(x:int, y:int, radius: float): Shape(x,y) {
    override func area() -> float {
        this.getPI() * this.radius * this.radius
    }

    override func perimeter() -> float {
        2 * this.getPI() * this.radius
    }

    private func getPI() -> float {
        3.14159
    }
}

class Rectangle(x: int, y: int, width: float, height: float): Shape(x,y) {
    override func area() -> float {
        this.width * this.height
    }

    override func perimeter() -> float {
        2 * (this.width + this.height)
    }
}
```

#### Interface

```naviary
interface Drawable {
    func draw()
    func getBounds() -> Rectangle
}

interface Clickable {
    func onClick(x: int, y: int)
}

// Interface implementation
class Button(x: int, y: int, label: string, width:int, height:int): Drawable, Clickable {
    func draw() {
        // Access fields using this
        drawRect(this.x, this.y, this.width, this.height)
        drawText(this.label, this.x + 10, this.y + 15)
    }

    func getBounds() -> Rectangle {
        Rectangle(this.x, this.y, this.width, this.height)
    }

    func onClick(x: int, y: int) {
        if this.contains(x, y) {
            print("Button {this.label} clicked at ({x}, {y})")
        }
    }

    func contains(x: int, y: int) -> bool {
        x >= this.x && x <= this.x + this.width &&
        y >= this.y && y <= this.y + this.height
    }
}
```

#### Generic Class

```naviary
class Container<T>(mut items: T[]) {
    func add(item: T) {
        this.items.append(item)
    }

    func get(index: int) -> T? {
        if index < this.items.length() {
            this.items[index]
        } else {
            nil
        }
    }

    func map<U>(transform: func(T) -> U) -> Container<U> {
        let result = Container<U>()
        for item in this.items {
            result.add(transform(item))
        }
        result
    }

    func forEach(action: func(T)) {
        for item in this.items {
            action(item)
        }
    }
}
```

#### Method Chaining

```naviary
class StringBuilder(mut buffer: string) {
    func append(text: string) -> StringBuilder {
        this.buffer += text
        this  // Return this for chaining
    }

    func appendLine(text: string) -> StringBuilder {
        this.buffer += text + "\n"
        this
    }

    func toString() -> string {
        this.buffer
    }
}

// Usage
let result = StringBuilder("")
    .append("Hello")
    .append(" ")
    .append("World")
    .appendLine("!")
    .toString()
```

### Control Flow

#### Conditionals

```naviary
// if-else chain
let grade = if score >= 90 {
    "A"
} else if score >= 80 {
    "B"
} else if score >= 70 {
    "C"
} else {
    "F"
}

// Regular if statement
if user.isValid() {
    process(user)
}
```

#### Loops

```naviary
// for loop
for i in 0..10 {        // 0 to 9
    print(i)
}

for i in 0..=10 {       // 0 to 10 (inclusive)
    print(i)
}

// Collection iteration
for item in items {
    process(item)
}

// With index
for value, index in items.enumerate() {
    print("{index}: {value}")
}

// while
while condition {
    doWork()
}

// while let
while let Ok(value) = someOperation() {
    doWork()
}

// Infinite loop
loop {
    if shouldStop() {
        break
    }
}
```

#### Pattern Matching

```naviary
// match expression
let description = match value {
    0 => "zero",
    1..10 => "small",
    10..=100 => "medium",
    _ => "large"
}

// Struct matching
match user {
    { name: "admin", ... } => "Administrator",
    { age: 0..18, ... } => "Minor",
    { age: 65.., ... } => "Senior",
    _ => "Regular"
}

// if let
if let Ok(value) = tryOperation() {
    print("Success: {value}")
}

match value {
    x if x > 0 && x % 2 == 0 => "positive even",
    x if x > 0 => "positive odd",
    x if x < 0 => "negative",
    _ => "zero"
}

// Struct and guard
match user {
    User{age, ..} if age >= 18 => "adult",
    User{age, ..} if age >= 13 => "teenager",
    User{..} => "child"
}

match value {
    [first, ...rest] => process(first, rest),  // Array destructuring
    {x, y, ...} if x > 0 => "positive x",      // Guard and destructuring
}

// Array patterns
match array {
    [] => "empty",
    [single] => "one element: {single}",
    [first, second] => "two elements: {first}, {second}",
    [first, ...rest] => "first: {first}, rest: {rest}",
    [first, ...middle, last] => "first: {first}, last: {last}",
}

// Character ranges
match char {
    'a'..'z' => "lowercase",
    'A'..'Z' => "uppercase",
    '0'..'9' => "digit",
    _ => "other"
}

// Alias
match value {
    // Bind the entire value during pattern matching
    Point{x: 0, ..} as point => "y-axis point: {point}",
    [1, ..] as list => "list starting with 1: {list}",
    "hello" as s | "hi" as s => "greeting: {s}",
}

// With ranges
match age {
    13..20 as teen => "teenager of age {teen}",
    20.. as adult => "adult of age {adult}",
}

match value {
    [1 as one, ...] if value == 1 => "one"
    _ => "other"
}
```

### Error Handling

#### Result Type

```naviary
enum Result<T, E> {
    Ok(T),    // Success value
    Err(E)    // Error value
}

// Basic usage
func divide(a: int, b: int) -> Result<int, string> {
    if b == 0 {
        Err("Division by zero")
    } else {
        Ok(a / b)
    }
}

// Usage
let result = divide(10, 2)

// Handle with pattern matching
match result {
    Ok(value) => print("Result: {value}"),
    Err(error) => print("Error: {error}")
}

// Simplified with if let
if let Ok(value) = divide(10, 2) {
    print("Success: {value}")
}
```

#### ? Operator (Error Propagation)

- Propagates only for Result types, not allowed otherwise.

```naviary
func calculate() -> Result<int, string> {
    let x = divide(10, 2)?  // Returns immediately if error
    let y = divide(x, 2)?
    Ok(y * 10)
}

// ? operator is equivalent to
let x = match divide(10, 2) {
    Ok(v) => v,
    Err(e) => return Err(e)
}
```

### Panic

```naviary
// panic - terminates the program immediately
func assertPositive(x: int) {
    if x <= 0 {
        panic("Expected positive number, got {x}")
    }
}
```

### Defer

```naviary
// defer - executes at function exit (LIFO order)
func processFile(path: string) -> Result<Data, Error> {
    let file = openFile(path)?
    defer file.close()  // Executes at function exit

    let lock = acquireLock()
    defer lock.release()  // Executes before close()

    // defer executes even if error occurs
    processData(file.read()?)
}

// Multiple defers - stack order (LIFO)
func example() {
    defer print("3")  // Third to execute
    defer print("2")  // Second to execute
    defer print("1")  // First to execute
    print("0")        // Executes immediately
}
// Output: 0 1 2 3

// defer if
func transaction() {
    let tx = db.begin()
    let mut success = false

    defer if !success {
        tx.rollback()
    }

    // Perform operations
    tx.execute("...")
    success = true
    tx.commit()
}
```

### Concurrency

All functions in Naviary are declared synchronous. Concurrency is a call-site choice.  
The language introduces a single keyword `async` (call-site only) and a standard library type `Task<T>`.  
**No implicit awaits** are allowed — blocking occurs only when the developer explicitly calls `await()` on a `Task`.

```naviary
let t = async fetch(3) // act like goroutine
let u2 = t.await()
```

### Pipeline Operator

```naviary
// Basic pipeline
let result = value |> function

// Equivalent to
let result = function(value)

// Chaining
let result = getData()
    |> filter(x => x > 0)
    |> map(x => x * 2)
    |> reduce((a, b) => a + b, 0)

// With partial application
let addOne = x => x + 1
let double = x => x * 2

5 |> addOne |> double  // 12

// With lambdas
getData()
    |> (data => data.filter(x => x > 0))
    |> (filtered => filtered.map(x => x * 2))

// Multi-argument functions
let divide = (x, y) => x / y
10 |> divide(2)  // divide(10, 2) = 5

// Curried functions
let add = x => y => x + y
5 |> add(3)  // 8

// Mixed with methods
"hello world"
    |> toUpper
    |> split(" ")
    |> join("-")  // "HELLO-WORLD"

// With error handling
readFile("data.txt")
    |> parseJson
    |> validate
    |> match {
        Ok(data) => process(data),
        Err(e) => handleError(e)
    }

// Custom pipeline-compatible functions
func transform(data: int[]) -> int[] {
    data.map(x => x * 2)
}

func summarize(data: int[]) -> Summary {
    Summary{
        total: data.sum(),
        average: data.sum() / data.length(),
        count: data.length()
    }
}

getData()
    |> transform
    |> summarize
    |> display
```

#### Pipeline Operator Rules

```naviary
// 1. Left value becomes the first argument of the right function
x |> f         // f(x)
x |> f(y)      // f(x, y)
x |> f(y, z)   // f(x, y, z)

// 2. Low precedence (evaluated after most operators)
1 + 2 |> double    // (1 + 2) |> double = 6
1 |> double + 2    // (1 |> double) + 2 = 4

// 3. Non-function values cause an error
let x = 5
x |> 10  // ❌ Compile error: "10 is not a function"

// 4. Void functions are allowed (ends chaining)
getData()
    |> process
    |> save
    |> print  // void return

// 5. Type inference
// Compiler infers types across the entire pipeline
let result = 5           // int
    |> double            // int -> int
    |> toString          // int -> string
    |> toUpper          // string -> string
// result: string
```

### Module System

#### Structure

```naviary
// File path determines module path
// src/math.nv → math module
// src/utils/strings.nv → utils.strings module
// src/models/user.nv → models.user module

// No package declaration - file location determines module
```

#### Export

```naviary
// Direct export
export func add(a: int, b: int) -> int { a + b }
export class Vector { x: float, y: float }
export const PI = 3.14159

// Group export
func multiply(a: int, b: int) -> int { a * b }
export { multiply, divide }

// Re-export (same syntax as import)
export geometry;
export std.math.{sin, cos};
export models.user.User as UserModel;

// Composite re-export
export {
    utils.strings.*,
    utils.arrays.{sort, filter},
    models.user.User as UserModel
};
```

#### Import

```naviary
// Basic import
import math;
import utils.strings;
import models.user;

// Selective import
import math.{add, Vector, PI};
import utils.strings.{toUpper, toLower};

// Import all exports
import math;
import utils.strings;

// Alias
import math.Vector as Vec;
import utils.strings as str;

// Nested import
import {
    math.{add, multiply},
    utils.strings,
    models.{user.User, product.Product}
};

// Usage
func main() {
    // Regular import
    let sum = math.add(1, 2);

    // Selective import
    let v = Vector(3, 4);

    // Alias
    let text = str.toUpper("hello");
}
```
