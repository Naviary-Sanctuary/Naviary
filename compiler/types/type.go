package types

import "strings"

type Type interface {
	Equals(other Type) bool
	String() string // for debugging
}

// Primitive
// Int, Float, String, Bool
type PrimitiveType struct {
	Name string
}

func (primitive *PrimitiveType) String() string {
	return primitive.Name
}

func (primitive *PrimitiveType) Equals(other Type) bool {
	otherPrimitive, ok := other.(*PrimitiveType)
	if !ok {
		return false
	}
	return primitive.Name == otherPrimitive.Name
}

// Nil
// nil
type NilType struct{}

func (nilType *NilType) String() string {
	return "nil"
}

func (nilType *NilType) Equals(other Type) bool {
	return other == Nil
}

// Function
type FunctionType struct {
	Parameters []Type
	ReturnType Type
}

func (function *FunctionType) String() string {
	var result strings.Builder
	result.WriteString("func(")

	// Join parameter types
	for i, param := range function.Parameters {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(param.String())
	}

	result.WriteString(")")

	// Add return type if exists
	if function.ReturnType != nil {
		result.WriteString(" -> ")
		result.WriteString(function.ReturnType.String())
	}

	return result.String()
}

func (function *FunctionType) Equals(other Type) bool {
	otherFunc, ok := other.(*FunctionType)
	if !ok {
		return false
	}

	// Check parameter count
	if len(function.Parameters) != len(otherFunc.Parameters) {
		return false
	}

	// Check each parameter type
	for i, param := range function.Parameters {
		if !param.Equals(otherFunc.Parameters[i]) {
			return false
		}
	}

	// Check return type
	if function.ReturnType == nil && otherFunc.ReturnType == nil {
		return true
	}
	if function.ReturnType == nil || otherFunc.ReturnType == nil {
		return false
	}

	return function.ReturnType.Equals(otherFunc.ReturnType)
}

var (
	Int    = &PrimitiveType{Name: "int"}
	Float  = &PrimitiveType{Name: "float"}
	String = &PrimitiveType{Name: "string"}
	Bool   = &PrimitiveType{Name: "bool"}

	Nil = &NilType{}
)

func GetPrimitiveType(name string) Type {
	switch name {
	case "int":
		return Int
	case "float":
		return Float
	case "string":
		return String
	case "nil":
		return Nil
	default:
		return nil
	}
}
