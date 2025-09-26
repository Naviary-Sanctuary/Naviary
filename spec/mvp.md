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

#### Classes (No Inheritance)

- Class declaration with fields
- Constructor
- Methods with `this` reference
- Field visibility (`private`, `public`)

#### Implementation Strategy

```navi
class Person {
  name: string
  age: int

  func greet() -> string {
    return "Hello, " + this.name
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

#### Concurrency Model

- `async` keyword at call-site
- `Task<T>` type wrapping Erlang processes
- `.await()` for synchronization
- Channel abstraction over message passing

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

**Week 1: MVP (0.0.1)**

- Day 1-2: Lexer
- Day 3-4: Parser
- Day 5: Code Generator
- Weekend: Testing & debugging

**Week 2: Type System (0.0.2 Part 1)**

- Day 1-2: Type checker architecture
- Day 3-4: Basic type checking
- Day 5: Function type checking

**Week 3: Control Flow (0.0.2 Part 2)**

- Day 1-2: if-else implementation
- Day 3-4: for loop implementation
- Day 5: Integration testing

**Month 2: Classes**

- Week 1: Class parsing
- Week 2: Erlang record mapping
- Week 3: Method dispatch
- Week 4: Testing & optimization

## Key Technical Decisions

### Erlang Mapping Rules

| Naviary         | Erlang                          |
| --------------- | ------------------------------- |
| `let x = 5`     | `X = 5`                         |
| `let mut x = 5` | `X1 = 5, X2 = ...` (versioning) |
| `class`         | Erlang module + record          |
| `int[]`         | Erlang list                     |
| `int?`          | `{ok, Value} \| nil`            |
| `match`         | `case` expression               |
| `async f()`     | `spawn(fun() -> f() end)`       |

### Naming Conventions

- Variables: `camelCase` → `CamelCase` (Erlang requires uppercase)
- Functions: `camelCase` → `camelCase`
- Classes: `PascalCase` → `pascalcase` (Erlang module names)

### Performance Targets

- **0.0.1-0.0.3**: Correctness over performance
- **0.0.4-0.0.6**: Basic optimizations (tail recursion, pattern match compilation)
- **0.0.7+**: Performance benchmarking and optimization

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

- ✅ Concurrent programs work correctly
- ✅ No race conditions
- ✅ Good performance under load
