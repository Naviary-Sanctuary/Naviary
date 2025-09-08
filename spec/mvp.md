# Naviary MVP Roadmap

## Implementation Design

### Core Decisions

- **Source Code Extension**: .navi
- **Compiler Language**: **Go**
- **Backend**: **Direct Assembly emission** _Generics are monomorphized._
- **Runtime/GC Implementation**: **C++20/23 (exceptions/RTTI disabled)**, exposing only **C ABI** (`extern "C"`).
- **Memory Management**: GC-based (initially STW mark‑sweep → later generational → later concurrent)
- **GC Strategy**: **Generational first**, followed by **Concurrent (SATB) old‑gen**
- **Type Inference**: Bidirectional + HM (rank‑1)
- **OOP Model**: Single inheritance + nominal interfaces

> Note: Backend is ASM, so stack maps and debug info are emitted directly in **DWARF (.file/.loc, .cfi\_\*)**.

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

- **Direct assembly emission** (ARM64)
- Prologue/epilogue, calling convention, 16B stack alignment
- IR → ASM mapping for `add/cmp/br/call/ret/load/store`
- **DWARF** line/unwind info (`.file/.loc`, `.cfi_*`)
- **Precise stack maps from day one** (foundation for future GC)

#### Runtime (C++)

- Implemented in **C++ (exceptions/RTTI disabled)**, exposing **C ABI** functions (`rt_*`)
- Initial GC: **Stop‑the‑world mark & sweep**, single-thread bump/arena allocator
- Basic I/O: `rt_print*` (stdout)
- **Object header format locked in**

#### Goals

- Compile 1‑10K LOC samples
- ASAN/UBSAN clean builds
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
- **ASM output diff tests** (IR→ASM snapshot)
- Runtime property tests

#### x86-64 support (cross compile)

- add x86-64 assembly

#### Goals

- 100+ compiler tests
- Stable ASM baselines

### 0.0.3 - Generational GC (Big Win!)

#### Young Generation GC

- Copying collector for young gen
- TLAB (Thread‑Local Allocation Buffer)
- Survivor spaces & promotion rules
- **Card table** (512B cards)
- Precise minor root scanning (stack maps & object maps)

#### Basic Optimizations

- Simple inlining
- Light SROA (Scalar Replacement)
- Loop strength reduction

#### Performance Targets

- Minor GC avg < 1ms
- Minor GC P95 < 3ms
- Promotion rate < 10% on alloc‑heavy benchmarks
- Single‑threaded performance baseline

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
- **No ad‑hoc overloading** (keep inference simple)

#### Type System Rules

- Annotations required at dynamic boundaries
- Explicit `implements` declarations
- Liskov substitution principle enforced

#### Goals

- Dynamic dispatch overhead documented
- Code size vs performance trade‑offs clear

### 0.0.5 - Concurrency + Old‑gen Concurrent GC + Modules

#### Language Features

- `spawn` keyword + minimal channels
- Memory model documentation
- Basic synchronization primitives

#### Old Generation Concurrent GC (C++)

- SATB (Snapshot‑At‑The‑Beginning) marking
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
- ABI‑stable `this` layout
- Vtable slot ordering frozen

### Type System

- Method bodies check with annotations
- No global inference across OOP boundaries
- Subtyping: no strengthening preconditions, no weakening postconditions

### GC Integration

- **Precise stack maps from day 1**
- Card size: 512B
- SATB write barrier for concurrent GC (0.0.5)

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
