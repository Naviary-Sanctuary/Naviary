package nir

import (
	"compiler/nir/instruction"
	"fmt"
	"strings"
)

// BasicBlock represents a sequence of instructions with single entry and exit
// the fundamental unit of control flow in NIR
type BasicBlock struct {
	Name         string
	Instructions []instruction.Instruction
	Terminator   instruction.Instruction
}

func NewBasicBlock(name string) *BasicBlock {
	return &BasicBlock{
		Name:         name,
		Instructions: make([]instruction.Instruction, 0),
		Terminator:   nil,
	}
}

func (block *BasicBlock) IsComplete() bool {
	return block.Terminator != nil
}

func (block *BasicBlock) String() string {
	var builder strings.Builder

	// Block header
	builder.WriteString(fmt.Sprintf("BasicBlock: %s\n", block.Name))

	// Instructions
	for _, inst := range block.Instructions {
		builder.WriteString(fmt.Sprintf("  %s\n", inst.String()))
	}

	// Terminator
	if block.Terminator != nil {
		builder.WriteString(fmt.Sprintf("  %s\n", block.Terminator.String()))
	} else {
		builder.WriteString("  <no terminator>\n")
	}

	return builder.String()
}

func (block *BasicBlock) AddInstruction(inst instruction.Instruction) {
	block.Instructions = append(block.Instructions, inst)
}
