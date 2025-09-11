package codegen

import (
	"bytes"
	"fmt"
	"naviary/compiler/ast"
)

// CodeGenerator generates assembly code from AST
type CodeGenerator struct {
	instructions    []Instruction          // Generated instructions
	emitter         Emitter                // Platform-specific emitter
	currentFunction *ast.FunctionStatement // Current function being compiled
	labelCounter    int                    // For unique label generation
}

// New creates a new code generator
func New(emitter Emitter) *CodeGenerator {
	return &CodeGenerator{
		instructions:    []Instruction{},
		emitter:         emitter,
		currentFunction: nil,
		labelCounter:    0,
	}
}

// Emit adds an instruction to the list
func (generator *CodeGenerator) Emit(instruction Instruction) {
	generator.instructions = append(generator.instructions, instruction)
}

// EmitMove adds a move instruction
func (generator *CodeGenerator) EmitMove(destination Register, value int64) {
	generator.Emit(MoveImmediate(destination, value))
}

// EmitMoveRegister adds a register move instruction
func (generator *CodeGenerator) EmitMoveRegister(destination, source Register) {
	generator.Emit(MoveRegister(destination, source))
}

// EmitCall adds a function call instruction
func (generator *CodeGenerator) EmitCall(functionName string, arguments ...Register) {
	generator.Emit(CallFunction(functionName, arguments...))
}

// EmitReturn adds a return instruction
func (generator *CodeGenerator) EmitReturn() {
	generator.Emit(ReturnVoid())
}

// EmitLabel adds a label definition
func (generator *CodeGenerator) EmitLabel(name string) {
	generator.Emit(MakeLabel(name))
}

// EmitGlobal adds a global symbol declaration
func (generator *CodeGenerator) EmitGlobal(name string) {
	generator.Emit(MakeGlobal(name))
}

// EmitComment adds a comment
func (generator *CodeGenerator) EmitComment(text string) {
	generator.Emit(MakeComment(text))
}

// NewLabel generates a unique label name
func (generator *CodeGenerator) NewLabel() string {
	label := fmt.Sprintf(".L%d", generator.labelCounter)
	generator.labelCounter++
	return label
}

// GenerateAssembly converts all instructions to assembly string
func (generator *CodeGenerator) GenerateAssembly() string {
	var buffer bytes.Buffer

	// File header
	buffer.WriteString(".text\n")
	buffer.WriteString(".align 2\n")
	buffer.WriteString("\n")

	// Convert each instruction
	for _, instruction := range generator.instructions {
		line := generator.emitter.EmitInstruction(instruction)
		buffer.WriteString(line)
		buffer.WriteString("\n")
	}

	return buffer.String()
}

// Generate compiles the AST to assembly
func (generator *CodeGenerator) Generate(program *ast.Program) {
	// Process each statement in the program
	for _, statement := range program.Statements {
		generator.generateStatement(statement)
	}
}

// generateStatement handles different statement types
func (generator *CodeGenerator) generateStatement(statement ast.Statement) {
	switch stmt := statement.(type) {
	case *ast.FunctionStatement:
		generator.generateFunction(stmt)
	case *ast.ExpressionStatement:
		generator.generateExpression(stmt.Expression)
	default:
		generator.EmitComment(fmt.Sprintf("TODO: %T", stmt))
	}
}

// generateFunction compiles a function definition
func (generator *CodeGenerator) generateFunction(function *ast.FunctionStatement) {
	// Track current function
	generator.currentFunction = function
	defer func() { generator.currentFunction = nil }()

	// Declare main function as global
	if function.Name.Value == "main" {
		generator.EmitGlobal("_main")
	}

	// Function label (with macOS prefix)
	generator.EmitLabel("_" + function.Name.Value)

	// Function prologue
	generator.EmitComment("Function prologue")
	generator.generatePrologue()

	// Function body
	for _, statement := range function.Body.Statements {
		generator.generateStatement(statement)
	}

	// Function epilogue
	generator.EmitComment("Function epilogue")
	generator.generateEpilogue()
}

// generatePrologue generates function entry code
func (generator *CodeGenerator) generatePrologue() {
	// Save frame pointer (x29) and link register (x30)
	// stp x29, x30, [sp, #-16]!
	generator.Emit(Instruction{
		Operation: Store, // Special store pair (we'll handle in emitter)
		Operands: []Operand{
			FramePointer,
			LinkRegister,
			Memory{Base: StackPointer, Offset: -16},
		},
		Comment: "stp x29, x30, [sp, #-16]!",
	})

	// Set up new frame pointer
	generator.EmitMoveRegister(FramePointer, StackPointer)
}

// generateEpilogue generates function exit code
func (generator *CodeGenerator) generateEpilogue() {
	// Return 0 for main function
	if generator.currentFunction != nil &&
		generator.currentFunction.Name.Value == "main" {
		generator.EmitMove(Register0, 0)
	}

	// Restore frame pointer and link register
	// ldp x29, x30, [sp], #16
	generator.Emit(Instruction{
		Operation: Load, // Special load pair
		Operands: []Operand{
			FramePointer,
			LinkRegister,
			Memory{Base: StackPointer, Offset: 16},
		},
		Comment: "ldp x29, x30, [sp], #16",
	})

	// Return to caller
	generator.EmitReturn()
}

// generateExpression compiles an expression
func (generator *CodeGenerator) generateExpression(expression ast.Expression) Register {
	switch expr := expression.(type) {
	case *ast.IntegerLiteral:
		return generator.generateIntegerLiteral(expr)
	case *ast.CallExpression:
		return generator.generateCallExpression(expr)
	case *ast.Identifier:
		return generator.generateIdentifier(expr)
	case *ast.BinaryExpression:
		return generator.generateBinaryExpression(expr)
	default:
		generator.EmitComment(fmt.Sprintf("TODO: expression %T", expr))
		return Register0
	}
}

// generateIntegerLiteral compiles an integer literal
func (generator *CodeGenerator) generateIntegerLiteral(literal *ast.IntegerLiteral) Register {
	// Parse the integer value
	value := int64(0)
	fmt.Sscanf(literal.Value, "%d", &value)

	// Move to register 0
	generator.EmitMove(Register0, value)
	return Register0
}

// generateCallExpression compiles a function call
func (generator *CodeGenerator) generateCallExpression(call *ast.CallExpression) Register {
	// Get function name
	funcIdent, ok := call.Function.(*ast.Identifier)
	if !ok {
		generator.EmitComment("ERROR: function call target is not an identifier")
		return Register0
	}

	// Evaluate arguments and put in argument registers
	argRegisters := []Register{}
	for i, arg := range call.Arguments {
		if i >= 4 {
			// TODO: Handle more than 4 arguments (need stack)
			generator.EmitComment("WARNING: only first 4 arguments supported")
			break
		}

		// Evaluate argument
		resultReg := generator.generateExpression(arg)

		// Move to argument register if not already there
		argReg := Register(i) // Register0, Register1, etc.
		if resultReg != argReg {
			generator.EmitMoveRegister(argReg, resultReg)
		}
		argRegisters = append(argRegisters, argReg)
	}

	// Map built-in function names to runtime equivalents
	functionName := generator.mapBuiltinFunction(funcIdent.Value)

	// Call the function
	generator.EmitCall(functionName, argRegisters...)

	// Result is in Register0
	return Register0
}

// mapBuiltinFunction maps built-in function names to their runtime equivalents
func (generator *CodeGenerator) mapBuiltinFunction(name string) string {
	switch name {
	case "print":
		return "navi_print_int"
	default:
		return name
	}
}

// generateIdentifier compiles an identifier reference
func (generator *CodeGenerator) generateIdentifier(ident *ast.Identifier) Register {
	// TODO: Implement variable lookup
	generator.EmitComment(fmt.Sprintf("TODO: load variable %s", ident.Value))
	return Register0
}

// generateBinaryExpression compiles a binary operation
func (generator *CodeGenerator) generateBinaryExpression(binary *ast.BinaryExpression) Register {
	// Evaluate left side
	leftReg := generator.generateExpression(binary.Left)

	// Save left result if needed (using a temp register)
	if leftReg == Register0 {
		generator.EmitMoveRegister(Register1, Register0)
		leftReg = Register1
	}

	// Evaluate right side
	rightReg := generator.generateExpression(binary.Right)

	// Perform operation
	switch binary.Operator {
	case "+":
		generator.Emit(AddRegisters(Register0, leftReg, rightReg))
	case "-":
		generator.Emit(Instruction{
			Operation: Subtract,
			Operands:  []Operand{Register0, leftReg, rightReg},
		})
	default:
		generator.EmitComment(fmt.Sprintf("TODO: operator %s", binary.Operator))
	}

	return Register0
}
