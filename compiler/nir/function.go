package nir

import (
	"compiler/types"
	"fmt"
	"strings"
)

type Parameter struct {
	Name string
	Type types.Type
}

func NewParameter(name string, paramType types.Type) Parameter {
	return Parameter{
		Name: name,
		Type: paramType,
	}
}

func (param *Parameter) String() string {
	return fmt.Sprintf("%s: %s", param.Name, param.Type.String())
}

type Function struct {
	Name        string
	Parameters  []Parameter
	ReturnType  types.Type
	BasicBlocks []*BasicBlock
}

func NewFunction(name string, parameters []Parameter, returnType types.Type) *Function {
	return &Function{
		Name:        name,
		Parameters:  parameters,
		ReturnType:  returnType,
		BasicBlocks: make([]*BasicBlock, 0),
	}
}

func (function *Function) AddBasicBlock(block *BasicBlock) {
	function.BasicBlocks = append(function.BasicBlocks, block)
}

func (function *Function) GetEntryBlock() *BasicBlock {
	if len(function.BasicBlocks) == 0 {
		return nil
	}
	return function.BasicBlocks[0]
}

func (function *Function) IsComplete() bool {
	if len(function.BasicBlocks) == 0 {
		return false
	}

	for _, block := range function.BasicBlocks {
		if !block.IsComplete() {
			return false
		}
	}

	return true
}

func (function *Function) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Function: %s(", function.Name))

	for i, param := range function.Parameters {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(param.String())
	}

	builder.WriteString(")")

	if function.ReturnType != nil {
		builder.WriteString(fmt.Sprintf(" -> %s", function.ReturnType.String()))
	}

	builder.WriteString("\n")

	for _, block := range function.BasicBlocks {
		blockStr := block.String()
		lines := strings.Split(blockStr, "\n")
		for _, line := range lines {
			if line != "" {
				builder.WriteString(fmt.Sprintf("  %s\n", line))
			}
		}
	}

	return builder.String()
}
