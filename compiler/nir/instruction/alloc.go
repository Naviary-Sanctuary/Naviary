package instruction

import (
	"compiler/nir/value"
	"compiler/types"
	"fmt"
)

// AllocInstruction allocates memory for a variable
// Example: %x = Alloc(int)
type AllocInstruction struct {
	result       value.Value
	allocateType types.Type
}

func NewAllocInstruction(result value.Value, allocateType types.Type) *AllocInstruction {
	return &AllocInstruction{
		result:       result,
		allocateType: allocateType,
	}
}

func (alloc *AllocInstruction) GetResult() value.Value {
	return alloc.result
}

func (alloc *AllocInstruction) String() string {
	return fmt.Sprintf("%s = Alloc(%s)", alloc.result.String(), alloc.allocateType.String())
}

func (alloc *AllocInstruction) GetAllocateType() types.Type {
	return alloc.allocateType
}
