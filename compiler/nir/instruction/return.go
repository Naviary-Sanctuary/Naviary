package instruction

import (
	"compiler/nir/value"
	"fmt"
)

type ReturnInstruction struct {
	value value.Value
}

func NewReturnInstruction(value value.Value) *ReturnInstruction {
	return &ReturnInstruction{
		value: value,
	}
}

func (returnInst *ReturnInstruction) String() string {
	if returnInst.value != nil {
		return fmt.Sprintf("Return(%s)", returnInst.value.String())
	}
	return "Return(void)"
}

func (returnInst *ReturnInstruction) GetResult() value.Value {
	return nil
}

func (returnInst *ReturnInstruction) GetValue() value.Value {
	return returnInst.value
}
