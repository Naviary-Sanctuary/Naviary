package instruction

import (
	"compiler/nir/value"
	"fmt"
)

// CallInstruction calls a function with arguments
// Example: %result = Call(%functionName, [%arg1, %arg2])
type CallInstruction struct {
	result       value.Value
	functionName string
	arguments    []value.Value
}

func NewCallInstruction(result value.Value, functionName string, arguments []value.Value) *CallInstruction {
	return &CallInstruction{
		result:       result,
		functionName: functionName,
		arguments:    arguments,
	}
}

func (call *CallInstruction) String() string {
	args := "["
	for i, arg := range call.arguments {
		if i > 0 {
			args += ", "
		}
		args += arg.String()
	}
	args += "]"

	if call.result != nil {
		return fmt.Sprintf("%s = Call(%s, %s)", call.result.String(), call.functionName, args)
	}
	return fmt.Sprintf("Call(%s, %s)", call.functionName, args)
}

func (call *CallInstruction) GetResult() value.Value {
	return call.result
}

func (call *CallInstruction) GetFunctionName() string {
	return call.functionName
}

func (call *CallInstruction) GetArguments() []value.Value {
	return call.arguments
}
