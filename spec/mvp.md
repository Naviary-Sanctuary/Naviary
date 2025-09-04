# Naviary MVP Roadmap

## Implementation Design

### Core Decisions

- **Compiler Language**: Rust
- **Backend**: LLVM (with monomorphization for generics)
- **Memory Management**: Pure GC
- **GC Strategy**: Generational FIRST → Concurrent later
- **Type Inference**: Bidirectional + HM (rank-1)
- **OOP Model**: Single inheritance + nominal interfaces

## Roadmap

### 0.0.1 - Frontend Skeleton + Minimal Runtime

#### Lexer & Parser

##### Tokens

- Keywords: `let`, `func`, `if`, `for`, `return`, `class`
- Types: `int`, `float`, `string`, `bool`
- Operators: `+`, `-`, `*`, `/`, `==`, `!=`, `=`

##### Grammar

- Variable declaration: `let`, `let mut`
- Function definition & calling
- Control flow: `if-else`, `for` loop
- Main entry point

#### Type System (Explicit Only)

- Basic types: `int`, `float`, `string`, `bool`
- Type checking (no inference yet)
- Function signature verification

#### Code Generation

- LLVM IR generation
- Basic operations compilation
- Function calling convention
- Main function entry point
- **Precise stack maps from day one** (for future GC)

#### Runtime

- Stop-the-world mark & sweep GC
- Single-thread bump allocator
- Print function (stdout)
- **Object header design locked in**

#### Goals

- Compile 1-10K LOC samples
- ASAN/UBSAN clean
- Root maps validated

### 0.0.2 - Language Expansion + Infrastructure

#### Type System Extension

- Arrays: `int[]`, `string[]`
- Optionals: `int?`, `string?`
- Tuples: `(int, string)`
- **Local type inference**: `let x = 42` (bidirectional for let bindings only)
- **Method boundary annotations required**

#### Control Flow

- `while` loop
- Basic `match` expression
- Error handling: `Result<T, E>` + `?` operator
- No panic unwinding (panic = abort)

#### Classes (Basic)

- Class declaration with fields
- Method implementation
- `this` reference
- Field visibility (`private`, `public`)
- **No inheritance yet**
- **Fixed vtable layout**

#### Testing Infrastructure

- Golden tests for parser/typing
- IR diff tests
- Runtime property tests

#### Goals

- 100+ compiler tests
- Stable IR baselines

### 0.0.3 - Generational GC (Big Win!)

#### Young Generation GC

- Copying collector for young gen
- TLAB (Thread-Local Allocation Buffer)
- Survivor spaces & promotion rules
- **Card table** (512B cards)
- Precise minor root scanning

#### Basic Optimizations

- Simple inlining
- Light SROA (Scalar Replacement)
- Loop strength reduction

#### Performance Targets

- Minor GC avg < 1ms
- Minor GC P95 < 3ms
- Promotion rate < 10% on alloc-heavy benchmarks
- Single-threaded performance baseline

### 0.0.4 - Polymorphism + Generics

#### OOP Completion

- **Single inheritance** (`extends`)
- **Nominal interfaces** (`implements`)
- Method overriding
- `super` keyword
- Abstract classes
- **Vtable dispatch** (measured overhead)

#### Generics

- Generic functions (monomorphization)
- Generic classes
- Type constraints (`where` clauses)
- **No ad-hoc overloading** (keep inference simple)

#### Type System Rules

- Annotations required at dynamic boundaries
- Explicit `implements` declarations
- Liskov substitution principle enforced

#### Goals

- Dynamic dispatch overhead documented
- Code size vs performance trade-offs clear

### 0.0.5 - Concurrency + Old-gen Concurrent GC + Modules

#### Language Features

- `spawn` keyword + minimal channels
- Memory model documentation
- Basic synchronization primitives

#### Old Generation Concurrent GC

- SATB (Snapshot-At-The-Beginning) marking
- Concurrent marking for old gen
- Background sweep
- Full write barrier
- Remembered sets

#### Module System

- Import/export mechanism
- Package manager (basic)
- Minimal standard library

#### Performance Targets

- P99 STW (remark) < 5ms
- GC CPU < 10% on simple web server

### 0.0.6+ (Future)

- Pipeline operator (`|>`)
- Anonymous/structural objects
- Full pattern matching with exhaustiveness
- Richer standard library
- Lambda & closures
- Enum types

## Key Design Invariants

### Object Layout

- Fixed object header format
- ABI-stable `this` layout
- Vtable slot ordering frozen

### Type System

- Method bodies check with annotations
- No global inference across OOP boundaries
- Subtyping: no strengthening preconditions, no weakening postconditions

### GC Integration

- Precise stack maps from day 1
- Card size: 512B
- SATB write barrier for concurrent GC

## Benchmarking Targets

### Micro Benchmarks

- Allocation rate (M/s)
- Minor pause avg/P95/P99
- Promotion rate
- Card scan cost
- Marking throughput
- RSS

### E2E Benchmarks

- Echo server
- JSON parser
- Template renderer

### Success Criteria

- After 0.0.3: Minor avg < 1ms, P95 < 3ms
- After 0.0.5: Remark STW < 5ms, GC CPU < 10%
