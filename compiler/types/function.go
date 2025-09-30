package types

import "fmt"

type FunctionType struct {
	ParameterTypes []Type
	ReturnType     Type
}

func (function *FunctionType) String() string {
	params := ""
	for i, paramType := range function.ParameterTypes {
		if i > 0 {
			params += ", "
		}

		params += paramType.String()
	}

	return fmt.Sprintf("func(%s) -> %s", params, function.ReturnType.String())
}

func (function *FunctionType) Equals(other Type) bool {
	if otherFunction, ok := other.(*FunctionType); ok {
		if len(function.ParameterTypes) != len(otherFunction.ParameterTypes) {
			return false
		}

		for i, paramType := range function.ParameterTypes {
			if !paramType.Equals(otherFunction.ParameterTypes[i]) {
				return false
			}
		}

		return function.ReturnType.Equals(otherFunction.ReturnType)
	}

	return false
}
