## Naviary MVP Roadmap

## Implementation Design

### Core Decisions

- **Source Code Extension**: .navi
- **Compiler Language**: **Go**
- **Backend**: **BEAM target** — emit **Erlang source**
- **Runtime/GC**: Use **BEAM runtime and GC**
- **Memory Management**: Managed by BEAM GC
- **Type Inference**: Bidirectional + HM (rank‑1) - from 0.0.3
- **OOP Model**: Single inheritance + nominal interfaces - from 0.0.4
- **Observability**: Leverage OTP tools (observer, tracing, profiling)
- **Concurrency Model**: Automatic Actor conversion on `async` usage

### Class & Concurrency Design

#### Philosophy
All classes compile to **Erlang Records** by default (lightweight). When `async` keyword is used, the runtime **automatically converts** them to Actors (processes) for safe concurrent access.

#### Default: Record-based
```navi
class Order(mut items: OrderItem[], mut status: OrderStatus) {
    func addItem(item: OrderItem) -> Result<(), string> {
        if this.status != OrderStatus.Draft {
            return Err("Cannot modify")
        }
        this.items.append(item)
        Ok(())
    }
}

// Synchronous → Record
let order = repo.find(orderId)
order.addItem(item)
```

Compiles to:
```erlang
-record(order, {items, status}).
add_item(Order, Item) -> Order#order{items = [Item | Order#order.items]}.
```

#### Automatic Actor Conversion
```navi
// Asynchronous → Automatic Actor
let order = repo.find(orderId)
async order.addItem(item)  // Runtime creates actor automatically
```

Runtime:
```erlang
Order = repo:find(OrderId),
OrderActor = naviary_runtime:ensure_actor(Order),  % Auto-convert
naviary_runtime:async_call(OrderActor, add_item, [Item]).
```

#### Performance Characteristics
- **100,000 inactive orders**: Records (~10MB)
- **100 active concurrent orders**: Actors (~200KB)
- **Auto-cleanup**: 5 minutes idle → Actor terminates

#### Key Rules
1. All classes → Records (default)
2. `async` keyword → Automatic actor conversion
3. Runtime manages actor lifecycle
4. Developer writes zero actor code

## Roadmap

### 0.0.1 - Minimal Working Compiler (Week 1)

#### Goal

Run this code successfully:

```navi
func main() {
  let a = 1 + 2
  print(a)
}
```

#### Lexer

- **Tokens**: Numbers (integers only), identifiers
- **Keywords**: `let`, `func`, `print`
- **Operators**: `+`, `-`, `*`, `/`, `=`
- **Symbols**: `(`, `)`, `{`, `}`

#### Parser

- Function definition (main only)
- Let statements (immutable only)
- Binary expressions (arithmetic)
- Function calls (print only)
- **NO type annotations** - everything is integer

#### Code Generation

- Emit Erlang source
- Map `main` → `start/0`
- Map `print` → `io:format("~p~n", [X])`
- Basic arithmetic operations

#### NO Type Checking

- Assume everything is integer
- Type system deferred to 0.0.2

### 0.0.2 - Type System & Control Flow (Week 2-3)

#### Type System

- **Basic types**: `int`, `float`, `string`, `bool`
- **Type annotations**: `let x: int = 5`
- **Type checking**: Full type verification
- **Function signatures**: `func add(a: int, b: int) -> int`

#### Control Flow

- `if-else` statements
- `for` loops: `for i in 0..10`
- Multiple function definitions
- Return statements

#### Extended Operators

- Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`
- Logical: `&&`, `||`, `!`

### 0.0.3 - Arrays & Error Handling (Week 4-5)

#### Type System Extension

- Arrays: `int[]`, `string[]` (mapped to Erlang lists)
- Optionals: `int?`, `string?` (mapped to `{ok, Value} | nil`)
- Tuples: `(int, string)` (mapped to Erlang tuples)
- **Local type inference** for `let` bindings

#### Control Flow

- `while` loops
- Basic `match` expressions (pattern matching)

#### Error Handling

- `Result<T, E>` type
- `?` operator for error propagation
- Panic as process exit

### 0.0.4 - Basic Classes (Month 2)

#### Classes (Record-based, No Concurrency)

- Class declaration with fields
- Constructor
- Methods with `this` reference
- Field visibility (`private`, `public`)
- **All classes compile to Erlang records**

#### Implementation Strategy

```navi
class Person(name: string, age: int) {
  func greet() -> string {
    "Hello, {this.name}"
  }
}
```

Compiles to:

```erlang
-module(person).
-record(person, {name, age}).

new(Name, Age) -> #person{name=Name, age=Age}.
greet(This) -> "Hello, " ++ This#person.name.
```

### 0.0.5 - Inheritance & Interfaces (Month 3)

#### OOP Completion

- Single inheritance
- Interfaces (`implements`)
- Method overriding
- `super` keyword
- Virtual method dispatch via dispatch tables

#### Type System

- Subtyping rules
- Interface satisfaction checking
- Liskov substitution principle

### 0.0.6 - Generics & Advanced Features (Month 4)

#### Generics

- Generic functions: `func map<T, U>(items: T[], f: func(T) -> U) -> U[]`
- Generic classes: `class List<T>`
- Type constraints

#### Advanced Control Flow

- Full pattern matching with exhaustiveness checking
- Guards in match expressions

### 0.0.7 - Concurrency (Month 5)

#### Automatic Actor Model

**Core Feature**: `async` keyword triggers automatic Record→Actor conversion

```navi
class Counter(mut value: int) {
    func increment() {
        this.value += 1
    }
}

// Local usage → Record
let counter = Counter(0)
counter.increment()

// Concurrent usage → Automatic Actor
let counter = Counter(0)
async counter.increment()  // Runtime converts to actor
```

#### Runtime Implementation

- `naviary_runtime` module for actor management
- Lazy actor creation (spawn on first `async` call)
- Actor caching with automatic cleanup (5 min timeout)
- Message serialization for method calls

#### Concurrency Primitives

- `async` keyword for non-blocking calls
- `Task<T>` type wrapping actor messages
- `.await()` for synchronization
- Automatic serialization of concurrent access to same instance

#### DDD Support

```navi
// Domain aggregate - just business logic
class Order(id: OrderId, mut items: OrderItem[], mut status: OrderStatus) {
    func addItem(item: OrderItem) -> Result<(), string> {
        if this.status != OrderStatus.Draft {
            return Err("Cannot modify")
        }
        this.items.append(item)
        Ok(())
    }
}

// Concurrent access is automatically safe
let order = repo.find(orderId)
async order.addItem(item1)  // Serialized
async order.addItem(item2)  // Serialized
```

### 0.0.8+ - Future Features

- Module system & imports
- Package manager
- Standard library
- Lambda expressions & closures
- Enum types
- Pipeline operator (`|>`)
- Structured/anonymous objects

## Implementation Timeline

### Week-by-Week Breakdown

**MVP (0.0.1)**

- Lexer
- Parser
- Code Generator
- Testing & debugging

**Type System (0.0.2 Part 1)**

- Type checker architecture
- Basic type checking
- Function type checking

**Control Flow (0.0.2 Part 2)**

- if-else implementation
- for loop implementation
- Integration testing

**Classes**

- Class parsing
- Erlang record mapping
- Method dispatch
- Testing & optimization

**Month 5: Concurrency**

- Runtime actor management module
- `async` keyword implementation
- Automatic conversion logic
- Testing & optimization

## Key Technical Decisions

### Erlang Mapping Rules

| Naviary         | Erlang                          | Notes |
| --------------- | ------------------------------- | ----- |
| `let x = 5`     | `X = 5`                         | |
| `let mut x = 5` | `X1 = 5, X2 = ...` (versioning) | |
| `class`         | Erlang module + record          | Default |
| `async obj.method()` | Actor process (lazy spawn) | Automatic |
| `int[]`         | Erlang list                     | |
| `int?`          | `{ok, Value} \| nil`            | |
| `match`         | `case` expression               | |

### Concurrency Model

```
Class Definition → Always Record

Usage Context:
  Synchronous call → Record (function call)
  async call → Actor (automatic conversion)
    ↓
  Runtime:
    - Check actor cache
    - Spawn if needed
    - Send message
    - Auto-cleanup after 5min idle
```

### Naming Conventions

- Variables: `camelCase` → `CamelCase` (Erlang requires uppercase)
- Functions: `camelCase` → `camelCase`
- Classes: `PascalCase` → `pascalcase` (Erlang module names)

### Performance Targets

- **0.0.1-0.0.3**: Correctness over performance
- **0.0.4-0.0.6**: Basic optimizations (tail recursion, pattern match compilation)
- **0.0.7**: Concurrency performance (lazy actors, efficient caching)
- **0.0.8+**: Performance benchmarking and optimization

## Success Criteria

### 0.0.1 (MVP)

- ✅ Compiles simple arithmetic program
- ✅ Generates valid Erlang code
- ✅ Runs on BEAM successfully

### 0.0.2

- ✅ Type safety guaranteed
- ✅ All type errors caught at compile time
- ✅ Control flow works correctly

### 0.0.4

- ✅ Classes compile to efficient Erlang modules
- ✅ Method dispatch works correctly
- ✅ No runtime overhead compared to hand-written Erlang

### 0.0.7

- ✅ Developer writes zero actor code (just `async` keyword)
- ✅ Concurrent access is automatically safe
- ✅ Memory efficient (inactive = records, active = actors)
- ✅ No race conditions
- ✅ Auto-cleanup of idle actors
- ✅ DDD aggregates work naturally
- ✅ Good performance under concurrent load