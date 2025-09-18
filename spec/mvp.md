# Naviary MVP Roadmap

## Implementation Design

### Core Decisions

- **Source Code Extension**: .navi
- **Compiler Language**: **Gleam**
- **Backend**: **BEAM target** — initially emit **Erlang source**, optional **Core Erlang** later for optimizations
- **Runtime/GC**: Use **BEAM runtime and GC**; no custom native runtime
- **Memory Management**: Managed by BEAM GC
- **Type Inference**: Bidirectional + HM (rank‑1)
- **OOP Model**: Single inheritance + nominal interfaces; OO by default
- **Observability**: Leverage OTP tools (observer, tracing, profiling)

> Note: No direct ASM/DWARF/stack maps; rely on BEAM’s stack and tooling.

## Roadmap

### 0.0.1 - Frontend Skeleton + BEAM Codegen

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

- Emit **Erlang source** for functions/calls/arith/return/print
- Map `print` to `io:format/2` or a small wrapper module
- Preserve calling conventions compatible with OTP

#### Runtime

- No custom native runtime; run on BEAM
- Provide a tiny standard module for I/O bridging if needed

#### Goals

- Build and run via `erlc`/escript or `rebar3`
- Hello world, arithmetic, simple function call samples

### 0.0.2 - Language Expansion + Infrastructure

#### Type System Extension

- Arrays: `int[]`, `string[]` (mapped to lists or array lib)
- Optionals: `int?`, `string?`
- Tuples: `(int, string)`
- **Local type inference** for `let` bindings

#### Control Flow

- `while` loop
- Basic `match` expression
- Error handling: `Result<T, E>` + `?` operator (as `case` lowering)
- Panic = process exit (no unwinding)

#### Classes (Basic)

- Class declaration with fields
- Method implementation
- `this` reference
- Field visibility (`private`, `public`)
- Implementation strategy: compile to modules and records/maps

#### Testing & Tooling

- Golden tests for parser/typing
- E2E tests on BEAM (escript/rebar3)
- Baseline snapshots for codegen (Erlang source)

#### Packaging

- OTP release/escript packaging

### 0.0.3 - Optimization & Interop

#### Optimizations

- Pattern‑match compilation improvements

- Tail‑recursion style lowering
- Simple inlining and constant folding

#### Interop

- FFI story: prefer Ports; NIF only for tiny, non‑blocking ops
- Erlang/Elixir/Gleam interop examples and guidelines

#### Performance Targets

- Reduce allocations in hot paths
- Avoid mailbox back‑pressure; maintain stable reductions

### 0.0.4 - Polymorphism + Generics

#### OOP Completion

- Single inheritance

- Nominal interfaces (`implements`)
- Method overriding
- `super` keyword
- Dynamic dispatch semantics over BEAM terms

#### Generics

- Generic functions and classes

- Monomorphization at compile‑time where profitable; fallback to erased code when necessary
- Type constraints (`where` clauses)
- No ad‑hoc overloading

#### Type System Rules

- Annotations required at dynamic boundaries
- Explicit `implements` declarations
- Liskov substitution enforced

#### Goals

- Dispatch overhead documented
- Code size vs performance trade‑offs on BEAM

### 0.0.5 - Concurrency + Modules

#### Language Features

- `async` at call‑site → `spawn` processes; `Task<T>` → pid + reply protocol
- Basic channels via message passing conventions
- Memory model documentation (leveraging BEAM guarantees)

#### Module System

- Import/export mechanism
- Package manager (basic)
- Minimal standard library

#### Performance Targets

- Stable latencies under load (P95 < target)
- Scheduler‑friendly codegen (no long‑running NIFs)

### 0.0.6+ (Future)

- Pipeline operator (`|>`)
- Anonymous/structural objects
- Full pattern matching with exhaustiveness
- Richer standard library
- Lambda & closures
- Enum types

## Key Design Invariants

### OOP & Representation

- Language‑level OO semantics stable (classes, `this`, interfaces)
- Representation on BEAM via modules + records/maps; dispatch tables defined

### Type System

- Method bodies check with annotations
- No global inference across OOP boundaries
- Subtyping: no strengthening preconditions, no weakening postconditions

### BEAM Integration

- Memory management fully delegated to BEAM GC
- No custom stack maps/write barriers
- NIFs must be short and non‑blocking; prefer Ports/Tasks

## Benchmarking Targets

### Micro Benchmarks

- Reductions/sec on micro kernels
- Message passing latency
- Process heap/GC stats
- Pattern‑match dispatch cost

### E2E Benchmarks

- Echo server
- JSON parser
- Template renderer

### Success Criteria

- 0.0.3: Allocation reductions and tail‑rec recursion in hot paths
- 0.0.5: Stable latencies; GC CPU within target on simple web server
