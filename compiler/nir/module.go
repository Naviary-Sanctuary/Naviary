package nir

import (
	"fmt"
	"strings"
)

type Module struct {
	Name      string
	Functions []*Function
}

func NewModule(name string) *Module {
	return &Module{
		Name:      name,
		Functions: make([]*Function, 0),
	}
}

func (module *Module) AddFunction(function *Function) {
	module.Functions = append(module.Functions, function)
}

func (module *Module) GetFunction(name string) *Function {
	for _, function := range module.Functions {
		if function.Name == name {
			return function
		}
	}

	return nil
}

func (module *Module) IsComplete() bool {
	if len(module.Functions) == 0 {
		return false
	}

	for _, function := range module.Functions {
		if !function.IsComplete() {
			return false
		}
	}

	return true
}

func (module *Module) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Module: %s\n", module.Name))

	if len(module.Functions) == 0 {
		builder.WriteString("  <no functions>\n")
		return builder.String()
	}

	for i, function := range module.Functions {
		if i > 0 {
			builder.WriteString("\n")
		}

		functionStr := function.String()
		lines := strings.Split(functionStr, "\n")
		for _, line := range lines {
			if line != "" {
				builder.WriteString(fmt.Sprintf("  %s\n", line))
			}
		}
	}

	return builder.String()
}
