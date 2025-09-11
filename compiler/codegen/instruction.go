package codegen

// Type of operation
type OperationCode int

const (
	// Data movement operations
	Move  OperationCode = iota // Move data between registers or load immediate
	Load                       // Load from memory to register
	Store                      // Store from register to memory

	// Arithmetic operations
	Add      // Addition
	Subtract // Subtraction

	// Control flow
	Call   // Function call
	Return // Return from function

	// Meta operations
	DefineLabel // Define a label for jumps
	Global      // Global symbol declaration (.globl)
	Comment     // Comment in assembly
)

// Operand represents an instruction operand (what the operation works on)
type Operand interface {
	operandMarker()
}

// Register represents an abstract register
type Register int

const (
	// Argument/return registers (0-7)
	Register0 Register = iota
	Register1
	Register2
	Register3

	// Special purpose registers
	StackPointer
	FramePointer
	LinkRegister
)

type Immediate struct {
	Value int64
}

// Label represents a jump target or function name
type Label struct {
	Name string
}

// Memory represents a memory address
type Memory struct {
	Base   Register // base register (usually stack pointer)
	Offset int64    // offset from base register
}

func (r Register) operandMarker()  {}
func (i Immediate) operandMarker() {}
func (l Label) operandMarker()     {}
func (m Memory) operandMarker()    {}

type Instruction struct {
	Operation OperationCode
	Operands  []Operand // First operand is usually destination
	Comment   string    // Optional comment for debugging
}

// MoveImmediate creates a move instruction with immediate value
func MoveImmediate(destination Register, value int64) Instruction {
	return Instruction{
		Operation: Move,
		Operands:  []Operand{destination, Immediate{value}},
	}
}

// MoveRegister creates a move instruction between registers
func MoveRegister(destination Register, source Register) Instruction {
	return Instruction{
		Operation: Move,
		Operands:  []Operand{destination, source},
	}
}

// AddRegisters creates an add instruction
func AddRegisters(destination, source1, source2 Register) Instruction {
	return Instruction{
		Operation: Add,
		Operands:  []Operand{destination, source1, source2},
	}
}

// CallFunction creates a function call instruction
func CallFunction(functionName string, arguments ...Register) Instruction {
	operands := make([]Operand, 0, len(arguments)+1)

	// First operand is the function name
	operands = append(operands, Label{functionName})

	// Rest are the arguments
	for _, arg := range arguments {
		operands = append(operands, arg)
	}

	return Instruction{
		Operation: Call,
		Operands:  operands,
	}
}

// ReturnVoid creates a return instruction with no value
func ReturnVoid() Instruction {
	return Instruction{
		Operation: Return,
		Operands:  []Operand{},
	}
}

// ReturnValue creates a return instruction with a value
func ReturnValue(value Register) Instruction {
	return Instruction{
		Operation: Return,
		Operands:  []Operand{value},
	}
}

// LoadFromMemory creates a load instruction
func LoadFromMemory(destination Register, base Register, offset int64) Instruction {
	return Instruction{
		Operation: Load,
		Operands: []Operand{
			destination,
			Memory{Base: base, Offset: offset},
		},
	}
}

// StoreToMemory creates a store instruction
func StoreToMemory(source Register, base Register, offset int64) Instruction {
	return Instruction{
		Operation: Store,
		Operands: []Operand{
			source,
			Memory{Base: base, Offset: offset},
		},
	}
}

// MakeLabel creates a label definition instruction
func MakeLabel(name string) Instruction {
	return Instruction{
		Operation: DefineLabel,
		Operands:  []Operand{Label{name}},
	}
}

// MakeGlobal creates a global symbol declaration instruction
func MakeGlobal(name string) Instruction {
	return Instruction{
		Operation: Global,
		Operands:  []Operand{Label{name}},
	}
}

// MakeComment creates a comment instruction
func MakeComment(text string) Instruction {
	return Instruction{
		Operation: Comment,
		Operands:  []Operand{},
		Comment:   text,
	}
}
