package nir

import (
	"compiler/nir/instruction"
	"compiler/nir/value"
	"compiler/types"
)

type Builder struct {
	nextTemporaryID int
	currentBlock    *BasicBlock
}

func NewBuilder() *Builder {
	return &Builder{
		nextTemporaryID: 0,
		currentBlock:    nil,
	}
}

func (builder *Builder) SetInsertBlock(block *BasicBlock) {
	builder.currentBlock = block
}

func (builder *Builder) GetInsertBlock() *BasicBlock {
	return builder.currentBlock
}
func (builder *Builder) CreateTemporary(temporaryType types.Type) value.Value {
	temp := value.NewTemporary(builder.nextTemporaryID, temporaryType)
	builder.nextTemporaryID++
	return temp
}

func (builder *Builder) CreateConstantInt(val int) value.Value {
	return value.NewConstant(val, types.Int)
}

func (builder *Builder) CreateConstantString(val string) value.Value {
	return value.NewConstant(val, types.String)
}

func (builder *Builder) CreateVariable(name string, variableType types.Type) value.Value {
	return value.NewVariable(name, variableType)
}

func (builder *Builder) BuildAlloc(name string, allocateType types.Type) value.Value {
	variable := builder.CreateVariable(name, allocateType)
	allocInstruction := instruction.NewAllocInstruction(variable, allocateType)

	if builder.currentBlock != nil {
		builder.currentBlock.AddInstruction(allocInstruction)
	}

	return variable
}

func (builder *Builder) BuildStore(destination value.Value, val value.Value) {
	storeInstruction := instruction.NewStoreInstruction(destination, val)

	if builder.currentBlock != nil {
		builder.currentBlock.AddInstruction(storeInstruction)
	}
}

func (builder *Builder) BuildLoad(source value.Value) value.Value {
	temporary := builder.CreateTemporary(source.Type())
	loadInstruction := instruction.NewLoadInstruction(temporary, source)

	if builder.currentBlock != nil {
		builder.currentBlock.AddInstruction(loadInstruction)
	}

	return temporary
}

func (builder *Builder) BuildBinary(left value.Value, right value.Value, operator instruction.BinaryOperator) value.Value {
	temporary := builder.CreateTemporary(left.Type())

	binaryInstruction := instruction.NewBinaryInstruction(temporary, operator, left, right)

	if builder.currentBlock != nil {
		builder.currentBlock.AddInstruction(binaryInstruction)
	}

	return temporary
}

func (builder *Builder) BuildCall(functionName string, arguments []value.Value, returnType types.Type) value.Value {
	var result value.Value = nil

	if returnType != nil {
		result = builder.CreateTemporary(returnType)
	}

	callInstruction := instruction.NewCallInstruction(result, functionName, arguments)

	if builder.currentBlock != nil {
		builder.currentBlock.AddInstruction(callInstruction)
	}

	return result
}

func (builder *Builder) BuildReturn(val value.Value) {
	returnInst := instruction.NewReturnInstruction(val)

	if builder.currentBlock != nil {
		builder.currentBlock.Terminator = returnInst
	}
}

func (builder *Builder) Reset() {
	builder.nextTemporaryID = 0
	builder.currentBlock = nil
}
