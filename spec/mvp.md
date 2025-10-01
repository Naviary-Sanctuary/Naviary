# Naviary MVP Roadmap (Revised)

## Implementation Design

### Core Decisions

- **Source Code Extension**: `.navi`
- **Compiler Language**: Go
- **Target**: C source code → Native binary
- **Runtime**: Zig runtime library
- **Memory Management**: Stack map-based GC
  - Phase 1: Mark & Sweep
  - Phase 2: Generational
  - Phase 3: Concurrent (Go-style)
- **Type System**: Static typing with local inference
- **OOP Model**: Single inheritance + interfaces
- **Build Pipeline**: `.navi` → `.c` → native executable

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
- C code generation
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
- Type checking
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

```c
// Generated for each function
static StackMapEntry main_stack_map[] = {
    {.pc = 0x10, .live_ptrs = 0b101},  // Bitmap of pointer locations
};
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

```c
// Added to all pointer writes
na_write_barrier(obj, field, value);
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
# Compile Naviary to C
naviary compile app.navi
# Output:
#   app.c            - Main program
#   app_gc_meta.c    - Stack maps & type info

# Build with Zig runtime
zig cc app.c app_gc_meta.c -lnaviary_runtime -o app

# Run
./app
```

### Runtime Options

```bash
NAVIARY_GC_THRESHOLD=10M ./app  # GC trigger threshold
NAVIARY_GC_VERBOSE=1 ./app      # Print GC statistics
```

## C Code Generation Examples

### Class to C Struct

```navi
class Point(x: int, y: int)
```

Generates:

```c
typedef struct {
    GCHeader gc_header;
    int x;
    int y;
} Point;

Point* Point_new(int x, int y) {
    Point* p = (Point*)na_gc_alloc(sizeof(Point), POINT_TYPE);
    p->x = x;
    p->y = y;
    return p;
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

```c
float Circle_area(Circle* this) {
    return 3.14 * this->radius * this->radius;
}
```

### Inheritance with VTable

```navi
class Animal { virtual func speak() -> string }
class Dog: Animal { override func speak() -> string }
```

Generates:

```c
typedef struct {
    char* (*speak)(void*);
} Animal_VTable;

typedef struct {
    GCHeader gc_header;
    Animal_VTable* vtable;
} Animal;

typedef struct {
    Animal base;
    // Dog-specific fields
} Dog;

static Animal_VTable dog_vtable = {
    .speak = (void*)Dog_speak
};
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
