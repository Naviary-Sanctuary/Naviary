package llvm

import (
	"compiler/types"
	"fmt"

	"tinygo.org/x/go-llvm"
)

type TypeConverter struct {
	context *Context
}

func NewTypeConverter(context *Context) *TypeConverter {
	return &TypeConverter{
		context: context,
	}
}

func (converter *TypeConverter) Convert(naviaryType types.Type) (llvm.Type, error) {
	if naviaryType == nil {
		return llvm.Type{}, fmt.Errorf("cannot convert nil type")
	}

	switch t := naviaryType.(type) {
	case *types.PrimitiveType:
		return converter.convertPrimitiveType(t)
	case *types.NilType:
		return converter.convertNilType(t)
	case *types.FunctionType:
		return converter.convertFunctionType(t)
	default:
		return llvm.Type{}, fmt.Errorf("unsupported type: %s", naviaryType.String())
	}
}

func (converter *TypeConverter) convertPrimitiveType(primitiveType *types.PrimitiveType) (llvm.Type, error) {
	context := converter.context.GetRawContext()
	switch primitiveType.Name {
	case "int":
		return context.Int64Type(), nil
	case "float":
		return context.DoubleType(), nil
	case "string":
		return llvm.PointerType(context.Int8Type(), 0), nil
	case "bool":
		return context.Int1Type(), nil
	default:
		return llvm.Type{}, fmt.Errorf("unknown primitive type: %s", primitiveType.Name)
	}
}

func (converter *TypeConverter) convertNilType(nilType *types.NilType) (llvm.Type, error) {
	return converter.context.GetRawContext().VoidType(), nil
}

func (converter *TypeConverter) convertFunctionType(functionType *types.FunctionType) (llvm.Type, error) {
	parameterTypes := make([]llvm.Type, len(functionType.ParameterTypes))

	for i, paramType := range functionType.ParameterTypes {
		llvmType, err := converter.Convert(paramType)
		if err != nil {
			return llvm.Type{}, fmt.Errorf("failed to convert parameter type %d: %w", i, err)
		}
		parameterTypes[i] = llvmType
	}

	returnType, err := converter.Convert(functionType.ReturnType)
	if err != nil {
		return llvm.Type{}, fmt.Errorf("failed to convert return type: %w", err)
	}

	return llvm.FunctionType(returnType, parameterTypes, false), nil
}
