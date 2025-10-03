package instruction

import (
	"compiler/nir/value"
	"fmt"
)

type BinaryOperator int

const (
	BinaryAdd BinaryOperator = iota
	BinarySubtract
	BinaryMultiply
	BinaryDivide
	BinaryModulo
)

func (operator BinaryOperator) String() string {
	switch operator {
	case BinaryAdd:
		return "Add"
	case BinarySubtract:
		return "Subtract"
	case BinaryMultiply:
		return "Multiply"
	case BinaryDivide:
		return "Divide"
	case BinaryModulo:
		return "Modulo"
	default:
		return "Unknown"
	}
}

// BinaryInstruction performs binary operations (like +, -, *, /, %)
// Example: %result = Add(%left, %right)
type BinaryInstruction struct {
	result   value.Value
	operator BinaryOperator
	left     value.Value
	right    value.Value
}

func NewBinaryInstruction(result value.Value, operator BinaryOperator, left value.Value, right value.Value) *BinaryInstruction {
	return &BinaryInstruction{
		result:   result,
		operator: operator,
		left:     left,
		right:    right,
	}
}

func (binary *BinaryInstruction) GetResult() value.Value {
	return binary.result
}

func (binary *BinaryInstruction) String() string {
	return fmt.Sprintf("%s = %s(%s, %s)",
		binary.result.String(),
		binary.operator.String(),
		binary.left.String(),
		binary.right.String())
}
func (binary *BinaryInstruction) GetOperator() BinaryOperator {
	return binary.operator
}

func (binary *BinaryInstruction) GetLeft() value.Value {
	return binary.left
}

func (binary *BinaryInstruction) GetRight() value.Value {
	return binary.right
}
