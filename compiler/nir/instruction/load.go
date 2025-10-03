package instruction

import (
	"compiler/nir/value"
	"fmt"
)

// LoadInstruction loads a value from a memory location
// Example: %temp = Load(%x)
type LoadInstruction struct {
	result value.Value
	source value.Value
}

func NewLoadInstruction(result value.Value, source value.Value) *LoadInstruction {
	return &LoadInstruction{
		result: result,
		source: source,
	}
}

func (load *LoadInstruction) GetResult() value.Value {
	return load.result
}

func (load *LoadInstruction) String() string {
	return fmt.Sprintf("%s = Load(%s)", load.result.String(), load.source.String())
}

func (load *LoadInstruction) GetSource() value.Value {
	return load.source
}
