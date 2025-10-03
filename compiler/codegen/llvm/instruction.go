package llvm

import (
	llvmvalue "compiler/codegen/llvm/value"
	"compiler/nir/instruction"
	nirvalue "compiler/nir/value"
	"fmt"

	"tinygo.org/x/go-llvm"
)

type InstructionConverter struct {
	builder        llvm.Builder
	valueConverter *llvmvalue.Converter
	typeConverter  *TypeConverter
}

func NewInstructionConverter(
	builder llvm.Builder,
	valueConverter *llvmvalue.Converter,
	typeConverter *TypeConverter,
) *InstructionConverter {
	return &InstructionConverter{
		builder:        builder,
		valueConverter: valueConverter,
		typeConverter:  typeConverter,
	}
}

func (converter *InstructionConverter) ConvertAlloc(allocInstruction *instruction.AllocInstruction) error {
	allocateType := allocInstruction.GetAllocateType()

	llvmType, err := converter.typeConverter.Convert(allocateType)
	if err != nil {
		return fmt.Errorf("failed to convert allocate type: %w", err)
	}

	result := allocInstruction.GetResult()
	if result == nil {
		return fmt.Errorf("alloc instruction has no result")
	}

	variable, ok := result.(*nirvalue.Variable)
	if !ok {
		return fmt.Errorf("alloc result must be a variable, got %T", result)
	}

	allocaInstruction := converter.builder.CreateAlloca(llvmType, variable.String())

	converter.valueConverter.RegisterVariable(variable, allocaInstruction)

	return nil
}
