package codegen

// Emitter converts abstract instructions to target-specific assembly
type Emitter interface {
	// EmitInstruction converts one instruction to assembly string
	EmitInstruction(instruction Instruction) string

	// MapRegister maps abstract register to physical register name
	MapRegister(register Register) string

	// GetPlatformName returns the target platform name
	GetPlatformName() string
}
