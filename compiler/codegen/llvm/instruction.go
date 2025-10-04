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

func (converter *InstructionConverter) ConvertStore(storeInstruction *instruction.StoreInstruction) error {
	destination := storeInstruction.GetDestination()
	if destination == nil {
		return fmt.Errorf("store instruction has no destination")
	}

	destinationPointer, err := converter.valueConverter.Convert(destination)
	if err != nil {
		return fmt.Errorf("failed to convert store destination: %w", err)
	}

	value := storeInstruction.GetValue()
	if value == nil {
		return fmt.Errorf("store instruction has no value")
	}

	llvmValue, err := converter.valueConverter.Convert(value)
	if err != nil {
		return fmt.Errorf("failed to convert store value: %w", err)
	}

	converter.builder.CreateStore(llvmValue, destinationPointer)

	return nil
}

func (converter *InstructionConverter) ConvertLoad(loadInstruction *instruction.LoadInstruction) error {
	source := loadInstruction.GetSource()
	if source == nil {
		return fmt.Errorf("load instruction has no source")
	}

	sourcePointer, err := converter.valueConverter.Convert(source)
	if err != nil {
		return fmt.Errorf("failed to convert load source: %w", err)
	}

	result := loadInstruction.GetResult()
	if result == nil {
		return fmt.Errorf("load instruction has no result")
	}

	temporary, ok := result.(*nirvalue.Temporary)
	if !ok {
		return fmt.Errorf("load result must be a temporary, got %T", result)
	}

	loadType, err := converter.typeConverter.Convert(temporary.Type())
	if err != nil {
		return fmt.Errorf("failed to convert load type: %w", err)
	}

	loadedValue := converter.builder.CreateLoad(loadType, sourcePointer, "")

	converter.valueConverter.RegisterTemporary(temporary, loadedValue)

	return nil
}

func (converter *InstructionConverter) ConvertBinary(binaryInstruction *instruction.BinaryInstruction) error {
	left := binaryInstruction.GetLeft()
	if left == nil {
		return fmt.Errorf("binary instruction has no left operand")
	}

	llvmLeft, err := converter.valueConverter.Convert(left)
	if err != nil {
		return fmt.Errorf("failed to convert binary left operand: %w", err)
	}

	right := binaryInstruction.GetRight()
	if right == nil {
		return fmt.Errorf("binary instruction has no right operand")
	}

	llvmRight, err := converter.valueConverter.Convert(right)
	if err != nil {
		return fmt.Errorf("failed to convert binary right operand: %w", err)
	}

	result := binaryInstruction.GetResult()
	if result == nil {
		return fmt.Errorf("binary instruction has no result")
	}

	temporary, ok := result.(*nirvalue.Temporary)
	if !ok {
		return fmt.Errorf("binary result must be a temporary, got %T", result)
	}

	operator := binaryInstruction.GetOperator()
	var llvmResult llvm.Value
	switch operator {
	case instruction.BinaryAdd:
		llvmResult = converter.builder.CreateAdd(llvmLeft, llvmRight, "")
	case instruction.BinarySubtract:
		llvmResult = converter.builder.CreateSub(llvmLeft, llvmRight, "")
	case instruction.BinaryMultiply:
		llvmResult = converter.builder.CreateMul(llvmLeft, llvmRight, "")
	case instruction.BinaryDivide:
		llvmResult = converter.builder.CreateSDiv(llvmLeft, llvmRight, "")
	case instruction.BinaryModulo:
		llvmResult = converter.builder.CreateSRem(llvmLeft, llvmRight, "")
	default:
		return fmt.Errorf("unsupported binary operator: %v", operator)
	}

	converter.valueConverter.RegisterTemporary(temporary, llvmResult)

	return nil
}

func (converter *InstructionConverter) ConvertCall(callInstruction *instruction.CallInstruction) error {
	functionName := callInstruction.GetFunctionName()
	if functionName == "" {
		return fmt.Errorf("call instruction has no function name")
	}

	arguments := callInstruction.GetArguments()

	llvmArguments := make([]llvm.Value, len(arguments))
	for i, arg := range arguments {
		llvmArg, err := converter.valueConverter.Convert(arg)
		if err != nil {
			return fmt.Errorf("failed to convert call argument %d: %w", i, err)
		}
		llvmArguments[i] = llvmArg
	}

	function := converter.builder.GetInsertBlock().Parent()
	module := function.GlobalParent()
	calleeFunction := module.NamedFunction(functionName)

	if calleeFunction.IsNil() {
		return fmt.Errorf("function %s not found in module", functionName)
	}

	functionType := calleeFunction.Type().ElementType()

	result := callInstruction.GetResult()
	if result != nil {
		temporary, ok := result.(*nirvalue.Temporary)
		if !ok {
			return fmt.Errorf("call result must be a temporary, got %T", result)
		}

		llvmResult := converter.builder.CreateCall(functionType, calleeFunction, llvmArguments, "")
		converter.valueConverter.RegisterTemporary(temporary, llvmResult)
	} else {
		converter.builder.CreateCall(functionType, calleeFunction, llvmArguments, "")
	}

	return nil
}

func (converter *InstructionConverter) ConvertReturn(returnInstruction *instruction.ReturnInstruction) error {
	returnValue := returnInstruction.GetValue()

	if returnValue == nil {
		converter.builder.CreateRetVoid()
	} else {
		llvmValue, err := converter.valueConverter.Convert(returnValue)
		if err != nil {
			return fmt.Errorf("failed to convert return value: %w", err)
		}
		converter.builder.CreateRet(llvmValue)
	}

	return nil
}
