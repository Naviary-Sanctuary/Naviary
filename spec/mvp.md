# Naviary MVP Roadmap (Revised)

## Implementation Design

### Core Decisions

- **Source Code Extension**: `.navi`
- **Compiler Language**: Go
- **Target**: LLVM IR → Native binary
- **Runtime**: Zig runtime library (linked via LLVM)
- **Memory Management**: Stack map-based GC
  - Phase 1: Mark & Sweep
  - Phase 2: Generational
  - Phase 3: Concurrent (Go-style)
- **Type System**: Static typing with local inference
- **OOP Model**: Single inheritance + interfaces
- **Build Pipeline**: `.navi` → AST → NIR (Naviary High-level Intermediate Representation) → LLVM IR → native executable

## Version Roadmap

### 0.0.1 - Minimal Working Compiler

#### Goal

```navi
func main() {
  let a = 1 + 2
  print(a)
}
```

#### Features

- Basic lexer (numbers, identifiers, operators)
- Simple parser (functions, let, expressions)
- AST generation
- Basic NIR lowering from AST
- LLVM IR code generation
- No type system yet (everything is int)
- Stack map infrastructure (prep for GC)

### 0.0.2 - Basic Classes + Mark & Sweep GC

#### Goal

```navi
class Person(name: string, age: int) {
    func greet() {
        print("Hello, {this.name}")
    }
}

func main() {
    let p = Person("Alice", 30)
    p.greet()
}
```

#### Features

- Class declaration and instantiation
- Methods with `this`
- Basic string type
- **Mark & Sweep GC implementation**
- Stack maps for precise collection

### 0.0.3 - Type System & Control Flow

#### Features

- Type annotations: `let x: int = 5`
- Basic types: `int`, `float`, `string`, `bool`
- Type checking in NIR
- Control flow: `if-else`, `for`, `while`
- Comparison and logical operators
- GC safepoints at loops and function calls

### 0.0.4 - Inheritance

#### Goal

```navi
class Animal(name: string) {
    func speak() -> string { "..." }
}

class Dog(name: string, breed: string): Animal(name) {
    override func speak() -> string { "Woof!" }
}
```

#### Features

- Single inheritance
- Method overriding
- Virtual method tables (vtable)
- `super` keyword

### 0.0.5 - Interfaces

#### Goal

```navi
interface Drawable {
    func draw()
}

class Button: Drawable {
    func draw() { /* ... */ }
}
```

#### Features

- Interface declaration
- Multiple interface implementation
- Interface method dispatch

### 0.0.6 - Arrays & GC Optimization

#### Features

- Dynamic arrays: `int[]`
- Array methods: `append`, `length`
- GC optimizations:
  - Bitmap marking
  - Free lists
  - Size classes

### 0.0.7 - Advanced Types

#### Features

- Optional types: `int?`
- Result type: `Result<T, E>`
- Tuples: `(x: int, y: int)`
- Nil-coalescing: `value ?? default`

### 0.0.8 - Pattern Matching

#### Goal

```navi
let result = match value {
    0 => "zero",
    1..10 => "small",
    _ => "large"
}
```

#### Features

- Match expressions
- Pattern guards
- Exhaustiveness checking

### 0.0.9 - Generics

#### Goal

```navi
class List<T> {
    mut items: T[]
    func add(item: T) { this.items.append(item) }
}
```

#### Features

- Generic classes
- Generic functions
- Type parameter constraints

### 0.1.0 - Modules & Standard Library

#### Features

- Module system
- Import/export
- Basic standard library:
  - Collections (List, Map, Set)
  - String utilities
  - File I/O
  - Math functions

## GC Implementation Details

### Phase 1: Mark & Sweep (0.0.2)

#### Zig Runtime Structure

```zig
const GCHeader = struct {
    next: ?*GCHeader,
    type_id: u32,
    size: usize,
    marked: bool,
};

pub fn collect() void {
    mark();  // Mark reachable objects
    sweep(); // Free unmarked objects
}
```

#### Stack Maps

```llvm
; Generated for each function
@main_stack_map = global [1 x { i64, i32 }] [
  { i64 16, i32 5 } ; PC offset 0x10, live ptrs bitmap 0b101
]
```

### Phase 2: Generational GC (0.2.0)

#### Structure

```zig
const GenerationalGC = struct {
    nursery: Region,     // Young generation
    mature: Region,      // Old generation
    remembered_set: Set, // Old→Young references
};
```

#### Write Barrier

```llvm
; Added to all pointer writes
call void @na_write_barrier(ptr %obj, i32 %field, ptr %value)
```

### Phase 3: Concurrent GC (0.3.0)

#### Tri-color Marking

- **White**: Unvisited
- **Gray**: Visited but children unprocessed
- **Black**: Fully processed

#### Phases

1. **STW Mark Start** (< 100μs)
2. **Concurrent Mark** (runs with program)
3. **STW Mark End** (< 100μs)
4. **Concurrent Sweep**

#### Performance Goals

- Max pause: < 500μs
- Throughput overhead: < 10%
- Scales to GB+ heaps

## Build System

### Compilation Pipeline

```bash
# Compile Naviary to LLVM IR
naviary compile app.navi
# Output:
#   app.ll           - Main program IR
#   app_gc_meta.ll   - Stack maps & type info

# Build with Zig runtime and LLVM
clang app.ll app_gc_meta.ll -lnaviary_runtime -o app

# Run
./app
```

### Runtime Options

```bash
NAVIARY_GC_THRESHOLD=10M ./app  # GC trigger threshold
NAVIARY_GC_VERBOSE=1 ./app      # Print GC statistics
```

## LLVM IR Generation Examples

### Class to LLVM Struct

```navi
class Point(x: int, y: int)
```

Generates:

```llvm
%GCHeader = type { ptr, i32, i64, i1 }

%Point = type { %GCHeader, i32, i32 }

define ptr @Point_new(i32 %x, i32 %y) {
  %p = call ptr @na_gc_alloc(i64 16, i32 1) ; sizeof(Point), POINT_TYPE
  %x_ptr = getelementptr %Point, ptr %p, i32 0, i32 1
  store i32 %x, ptr %x_ptr
  %y_ptr = getelementptr %Point, ptr %p, i32 0, i32 2
  store i32 %y, ptr %y_ptr
  ret ptr %p
}
```

### Method to Function

```navi
class Circle(radius: float) {
    func area() -> float {
        3.14 * this.radius * this.radius
    }
}
```

Generates:

```llvm
define double @Circle_area(ptr %this) {
  %radius_ptr = getelementptr %Circle, ptr %this, i32 0, i32 1
  %radius = load double, ptr %radius_ptr
  %sq = fmul double %radius, %radius
  %area = fmul double 3.140000e+00, %sq
  ret double %area
}
```

### Inheritance with VTable

```navi
class Animal { virtual func speak() -> string }
class Dog: Animal { override func speak() -> string }
```

Generates:

```llvm
%Animal_VTable = type { ptr } ; ptr to speak function

%Animal = type { %GCHeader, ptr } ; vtable ptr

%Dog = type { %Animal, ; Dog-specific fields }

@dog_vtable = global %Animal_VTable { ptr @Dog_speak }
```

## Success Metrics

### MVP (0.0.1 - 0.1.0)

- ✅ Compiles OOP code correctly
- ✅ No memory leaks
- ✅ Type safe
- ✅ < 100ms GC pause for small programs

### Production Ready (0.2.0+)

- ✅ Generational GC working
- ✅ < 10ms minor GC
- ✅ Handles 100MB+ heaps

### Enterprise Ready (0.3.0+)

- ✅ Concurrent GC
- ✅ < 500μs pause times
- ✅ Scales to GB heaps
- ✅ Comparable to Go GC performance
