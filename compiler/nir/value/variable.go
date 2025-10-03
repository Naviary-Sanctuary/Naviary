package value

import "compiler/types"

// Variable represents a named variable in the source code
type Variable struct {
	name         string
	variableType types.Type
}

func NewVariable(name string, variableType types.Type) *Variable {
	return &Variable{
		name:         name,
		variableType: variableType,
	}
}

func (variable *Variable) Type() types.Type {
	return variable.variableType
}

func (variable *Variable) IsConstant() bool {
	return false
}

func (variable *Variable) String() string {
	return variable.name
}
