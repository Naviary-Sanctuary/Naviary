package codegen

import "fmt"

// DarwinARM64Emitter emits assembly for macOS on ARM64
type DarwinARM64Emitter struct{}

// NewDarwinARM64Emitter creates a new emitter for macOS ARM64
func NewDarwinARM64Emitter() *DarwinARM64Emitter {
	return &DarwinARM64Emitter{}
}

// GetPlatformName returns the platform identifier
func (emitter *DarwinARM64Emitter) GetPlatformName() string {
	return "darwin-arm64"
}

// MapRegister maps abstract register to ARM64 register name
func (emitter *DarwinARM64Emitter) MapRegister(register Register) string {
	switch register {
	case Register0:
		return "x0"
	case Register1:
		return "x1"
	case Register2:
		return "x2"
	case Register3:
		return "x3"
	case StackPointer:
		return "sp"
	case FramePointer:
		return "x29" // ARM64 frame pointer
	case LinkRegister:
		return "x30" // ARM64 link register
	default:
		// Should not happen if we defined all registers
		panic("unknown register")
	}
}

// EmitInstruction에 DefineLabel과 Comment 추가
func (emitter *DarwinARM64Emitter) EmitInstruction(instruction Instruction) string {
	switch instruction.Operation {
	case Move:
		return emitter.emitMove(instruction)
	case Load:
		return emitter.emitLoad(instruction)
	case Store:
		return emitter.emitStore(instruction)
	case Add:
		return emitter.emitAdd(instruction)
	case Call:
		return emitter.emitCall(instruction)
	case Return:
		return emitter.emitReturn(instruction)
	case DefineLabel:
		return emitter.emitDefineLabel(instruction)
	case Global:
		return emitter.emitGlobal(instruction)
	case Comment:
		return emitter.emitComment(instruction)
	case Subtract:
		return emitter.emitSubtract(instruction)
	default:
		return fmt.Sprintf("    # TODO: %v", instruction.Operation)
	}
}

// emitDefineLabel handles label definition
func (emitter *DarwinARM64Emitter) emitDefineLabel(instruction Instruction) string {
	if len(instruction.Operands) != 1 {
		panic("DefineLabel requires exactly 1 operand")
	}

	label, ok := instruction.Operands[0].(Label)
	if !ok {
		panic("DefineLabel operand must be a Label")
	}

	// Labels have no indentation and end with colon
	return fmt.Sprintf("%s:", label.Name)
}

// emitGlobal handles global symbol declaration
func (emitter *DarwinARM64Emitter) emitGlobal(instruction Instruction) string {
	if len(instruction.Operands) != 1 {
		panic("Global requires exactly 1 operand")
	}

	label, ok := instruction.Operands[0].(Label)
	if !ok {
		panic("Global operand must be a Label")
	}

	return fmt.Sprintf("    .globl %s", label.Name)
}

// emitComment handles comment
func (emitter *DarwinARM64Emitter) emitComment(instruction Instruction) string {
	// Comment text is in the Comment field, not Operands
	return fmt.Sprintf("    # %s", instruction.Comment)
}

// emitReturn handles Return instruction
func (emitter *DarwinARM64Emitter) emitReturn(instruction Instruction) string {
	// Return can have 0 or 1 operand
	// 0: void return
	// 1: return value (should already be in x0)

	// ARM64 uses 'ret' to return to address in Link Register (x30)
	return "    ret"
}

// emitCall handles Call instruction
func (emitter *DarwinARM64Emitter) emitCall(instruction Instruction) string {
	if len(instruction.Operands) < 1 {
		panic("Call requires at least 1 operand (function name)")
	}

	// First operand is the function name
	label, ok := instruction.Operands[0].(Label)
	if !ok {
		panic("Call first operand must be a Label")
	}

	// macOS requires underscore prefix for C functions
	functionName := "_" + label.Name

	// ARM64 uses 'bl' (Branch with Link) for function calls
	return fmt.Sprintf("    bl %s", functionName)
}

// emitAdd handles Add instruction
func (emitter *DarwinARM64Emitter) emitAdd(instruction Instruction) string {
	if len(instruction.Operands) != 3 {
		panic("Add requires exactly 3 operands")
	}

	destination := instruction.Operands[0]
	source1 := instruction.Operands[1]
	source2 := instruction.Operands[2]

	// All must be registers for basic add
	destReg, ok := destination.(Register)
	if !ok {
		panic("Add destination must be a register")
	}

	src1Reg, ok := source1.(Register)
	if !ok {
		panic("Add source1 must be a register")
	}

	src2Reg, ok := source2.(Register)
	if !ok {
		panic("Add source2 must be a register")
	}

	destName := emitter.MapRegister(destReg)
	src1Name := emitter.MapRegister(src1Reg)
	src2Name := emitter.MapRegister(src2Reg)

	return fmt.Sprintf("    add %s, %s, %s", destName, src1Name, src2Name)
}

// emitMove handles Move instruction
func (emitter *DarwinARM64Emitter) emitMove(instruction Instruction) string {
	if len(instruction.Operands) != 2 {
		panic("Move requires exactly 2 operands")
	}

	destination := instruction.Operands[0]
	source := instruction.Operands[1]

	// Get destination register name
	destReg, ok := destination.(Register)
	if !ok {
		panic("Move destination must be a register")
	}
	destName := emitter.MapRegister(destReg)

	// Handle different source types
	switch src := source.(type) {
	case Register:
		// Register to register
		srcName := emitter.MapRegister(src)
		return fmt.Sprintf("    mov %s, %s", destName, srcName)

	case Immediate:
		// Immediate to register
		return fmt.Sprintf("    mov %s, #%d", destName, src.Value)

	default:
		panic(fmt.Sprintf("Invalid source type for Move: %T", src))
	}
}

// emitLoad handles Load instruction
func (emitter *DarwinARM64Emitter) emitLoad(instruction Instruction) string {
	// Check for load pair (special case for epilogue)
	if instruction.Comment == "ldp x29, x30, [sp], #16" {
		return "    ldp x29, x30, [sp], #16"
	}

	// Normal load requires exactly 2 operands
	if len(instruction.Operands) != 2 {
		panic("Load requires exactly 2 operands")
	}

	destination := instruction.Operands[0]
	source := instruction.Operands[1]

	// Get destination register
	destReg, ok := destination.(Register)
	if !ok {
		panic("Load destination must be a register")
	}
	destName := emitter.MapRegister(destReg)

	// Get memory address
	memory, ok := source.(Memory)
	if !ok {
		panic("Load source must be a memory address")
	}
	baseName := emitter.MapRegister(memory.Base)

	// Format based on offset
	if memory.Offset == 0 {
		return fmt.Sprintf("    ldr %s, [%s]", destName, baseName)
	}
	return fmt.Sprintf("    ldr %s, [%s, #%d]", destName, baseName, memory.Offset)
}

// emitStore handles Store instruction
func (emitter *DarwinARM64Emitter) emitStore(instruction Instruction) string {
	// Check for store pair (special case for prologue)
	if instruction.Comment == "stp x29, x30, [sp, #-16]!" {
		return "    stp x29, x30, [sp, #-16]!"
	}

	// Normal store requires exactly 2 operands
	if len(instruction.Operands) != 2 {
		panic("Store requires exactly 2 operands")
	}

	source := instruction.Operands[0]
	destination := instruction.Operands[1]

	// Get source register
	srcReg, ok := source.(Register)
	if !ok {
		panic("Store source must be a register")
	}
	srcName := emitter.MapRegister(srcReg)

	// Get memory address
	memory, ok := destination.(Memory)
	if !ok {
		panic("Store destination must be a memory address")
	}
	baseName := emitter.MapRegister(memory.Base)

	// Format based on offset
	if memory.Offset == 0 {
		return fmt.Sprintf("    str %s, [%s]", srcName, baseName)
	}
	return fmt.Sprintf("    str %s, [%s, #%d]", srcName, baseName, memory.Offset)
}

// emitSubtract handles Subtract instruction
func (emitter *DarwinARM64Emitter) emitSubtract(instruction Instruction) string {
	if len(instruction.Operands) != 3 {
		panic("Subtract requires exactly 3 operands")
	}

	destination := instruction.Operands[0]
	source1 := instruction.Operands[1]
	source2 := instruction.Operands[2]

	destReg, ok := destination.(Register)
	if !ok {
		panic("Subtract destination must be a register")
	}

	src1Reg, ok := source1.(Register)
	if !ok {
		panic("Subtract source1 must be a register")
	}

	src2Reg, ok := source2.(Register)
	if !ok {
		panic("Subtract source2 must be a register")
	}

	destName := emitter.MapRegister(destReg)
	src1Name := emitter.MapRegister(src1Reg)
	src2Name := emitter.MapRegister(src2Reg)

	return fmt.Sprintf("    sub %s, %s, %s", destName, src1Name, src2Name)
}
