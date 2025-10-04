package llvm

import (
	llvmvalue "compiler/codegen/llvm/value"
	"compiler/nir"
	nirinstruction "compiler/nir/instruction"
	nirvalue "compiler/nir/value"
	"fmt"

	"tinygo.org/x/go-llvm"
)

type FunctionConverter struct {
	context              *Context
	module               llvm.Module
	typeConverter        *TypeConverter
	valueConverter       *llvmvalue.Converter
	instructionConverter *InstructionConverter
}

func NewFunctionConverter(
	context *Context,
	module llvm.Module,
	typeConverter *TypeConverter,
	valueConverter *llvmvalue.Converter,
) *FunctionConverter {
	return &FunctionConverter{
		context:              context,
		module:               module,
		typeConverter:        typeConverter,
		valueConverter:       valueConverter,
		instructionConverter: nil,
	}
}

func (converter *FunctionConverter) Convert(nirFunction *nir.Function) error {
	parameterTypes, err := converter.convertParameterTypes(nirFunction.Parameters)
	if err != nil {
		return fmt.Errorf("failed to convert parameter types: %w", err)
	}

	returnType, err := converter.typeConverter.Convert(nirFunction.ReturnType)
	if err != nil {
		return fmt.Errorf("failed to convert return type: %w", err)
	}

	functionType := llvm.FunctionType(returnType, parameterTypes, false)

	llvmFunction := llvm.AddFunction(converter.module, nirFunction.Name, functionType)

	converter.valueConverter.Reset()

	err = converter.registerParameters(nirFunction, llvmFunction)
	if err != nil {
		return fmt.Errorf("failed to register parameters: %w", err)
	}

	err = converter.convertBasicBlocks(nirFunction, llvmFunction)
	if err != nil {
		return fmt.Errorf("failed to convert basic blocks: %w", err)
	}

	return nil
}

func (converter *FunctionConverter) convertParameterTypes(parameters []nir.Parameter) ([]llvm.Type, error) {
	llvmTypes := make([]llvm.Type, len(parameters))

	for i, param := range parameters {
		llvmType, err := converter.typeConverter.Convert(param.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to convert parameter type %d: %w", i, err)
		}
		llvmTypes[i] = llvmType
	}

	return llvmTypes, nil
}

func (converter *FunctionConverter) registerParameters(nirFunction *nir.Function, llvmFunction llvm.Value) error {
	if len(nirFunction.Parameters) == 0 {
		return nil
	}

	builder := converter.context.GetRawContext().NewBuilder()
	defer builder.Dispose()

	entryBlock := nirFunction.GetEntryBlock()
	if entryBlock == nil {
		return fmt.Errorf("function %s has no entry block", nirFunction.Name)
	}

	llvmEntryBlock := llvm.AddBasicBlock(llvmFunction, entryBlock.Name)
	builder.SetInsertPointAtEnd(llvmEntryBlock)

	for i, param := range nirFunction.Parameters {
		llvmParam := llvmFunction.Param(i)
		llvmType, err := converter.typeConverter.Convert(param.Type)
		if err != nil {
			return fmt.Errorf("failed to convert parameter %s type: %w", param.Name, err)
		}

		allocaInstruction := builder.CreateAlloca(llvmType, param.Name)

		builder.CreateStore(llvmParam, allocaInstruction)

		paramVariable := nirvalue.NewVariable(param.Name, param.Type)
		converter.valueConverter.RegisterVariable(paramVariable, allocaInstruction)
	}

	return nil
}

func (converter *FunctionConverter) convertBasicBlocks(nirFunction *nir.Function, llvmFunction llvm.Value) error {
	builder := converter.context.GetRawContext().NewBuilder()
	defer builder.Dispose()

	converter.instructionConverter = NewInstructionConverter(
		builder,
		converter.valueConverter,
		converter.typeConverter,
	)

	for _, nirBlock := range nirFunction.BasicBlocks {
		var llvmBlock llvm.BasicBlock

		if nirBlock.Name == "entry" {
			llvmBlock = llvmFunction.FirstBasicBlock()
		} else {
			llvmBlock = llvm.AddBasicBlock(llvmFunction, nirBlock.Name)
		}

		builder.SetInsertPointAtEnd(llvmBlock)

		for _, instruction := range nirBlock.Instructions {
			err := converter.convertInstruction(instruction)
			if err != nil {
				return fmt.Errorf("failed to convert instruction %s: %w", instruction.String(), err)
			}
		}

		if nirBlock.Terminator != nil {
			err := converter.convertInstruction(nirBlock.Terminator)
			if err != nil {
				return fmt.Errorf("failed to convert terminator %s: %w", nirBlock.Terminator.String(), err)
			}
		}
	}

	return nil
}

func (converter *FunctionConverter) convertInstruction(instruction nirinstruction.Instruction) error {
	switch instruction := instruction.(type) {
	case *nirinstruction.AllocInstruction:
		return converter.instructionConverter.ConvertAlloc(instruction)

	case *nirinstruction.StoreInstruction:
		return converter.instructionConverter.ConvertStore(instruction)

	case *nirinstruction.LoadInstruction:
		return converter.instructionConverter.ConvertLoad(instruction)

	case *nirinstruction.BinaryInstruction:
		return converter.instructionConverter.ConvertBinary(instruction)

	case *nirinstruction.CallInstruction:
		return converter.instructionConverter.ConvertCall(instruction)

	case *nirinstruction.ReturnInstruction:
		return converter.instructionConverter.ConvertReturn(instruction)

	default:
		return fmt.Errorf("unsupported instruction type: %T", instruction)
	}
}
