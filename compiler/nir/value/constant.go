package value

import (
	"compiler/types"
	"fmt"
)

// Constant represents a compile-time constant value
type Constant struct {
	value     any
	valueType types.Type
}

func NewConstant(value any, valueType types.Type) *Constant {
	return &Constant{
		value:     value,
		valueType: valueType,
	}
}

func (constant *Constant) Type() types.Type {
	return constant.valueType
}

func (constant *Constant) IsConstant() bool {
	return true
}

func (constant *Constant) String() string {
	switch v := constant.value.(type) {
	case int:
		return fmt.Sprintf("Constant(%d)", v)
	case string:
		return fmt.Sprintf("Constant(\"%s\")", v)
	case float64:
		return fmt.Sprintf("Constant(%f)", v)
	case bool:
		return fmt.Sprintf("Constant(%t)", v)
	default:
		return "Constant(?)"
	}
}
