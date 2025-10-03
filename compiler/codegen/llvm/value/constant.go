package value

import (
	nirvalue "compiler/nir/value"
	"compiler/types"
	"fmt"

	"tinygo.org/x/go-llvm"
)

type TypeConverter interface {
	Convert(naviaryType types.Type) (llvm.Type, error)
}

type ConstantConverter struct {
	typeConverter TypeConverter
}

func NewConstantConverter(typeConverter TypeConverter) *ConstantConverter {
	return &ConstantConverter{
		typeConverter: typeConverter,
	}
}

func (converter *ConstantConverter) Convert(naviaryConstant *nirvalue.Constant) (llvm.Value, error) {
	if naviaryConstant == nil {
		return llvm.Value{}, fmt.Errorf("cannot convert nil constant")
	}

	switch naviaryConstant.Type() {
	case types.Int:
		return converter.convertIntConstant(naviaryConstant)
	case types.String:
		return converter.convertStringConstant(naviaryConstant)
	case types.Float:
		return converter.convertFloatConstant(naviaryConstant)
	case types.Bool:
		return converter.convertBoolConstant(naviaryConstant)
	default:
		return llvm.Value{}, fmt.Errorf("unsupported constant type: %s", naviaryConstant.Type().String())

	}
}

func (converter *ConstantConverter) convertIntConstant(naviaryConstant *nirvalue.Constant) (llvm.Value, error) {
	llvmType, err := converter.typeConverter.Convert(types.Int)

	if err != nil {
		return llvm.Value{}, fmt.Errorf("failed to convert int type: %w", err)
	}

	constantString := naviaryConstant.String()
	var value int64
	_, err = fmt.Sscanf(constantString, "Constant(%d)", &value)
	if err != nil {
		return llvm.Value{}, fmt.Errorf("failed to parse integer constant: %w", err)
	}

	return llvm.ConstInt(llvmType, uint64(value), false), nil
}

func (converter *ConstantConverter) convertFloatConstant(naviaryConstant *nirvalue.Constant) (llvm.Value, error) {
	llvmType, err := converter.typeConverter.Convert(types.Float)

	if err != nil {
		return llvm.Value{}, fmt.Errorf("failed to convert float type: %w", err)
	}

	constantString := naviaryConstant.String()
	var value float64
	_, err = fmt.Sscanf(constantString, "Constant(%f)", &value)
	if err != nil {
		return llvm.Value{}, fmt.Errorf("failed to parse float constant: %w", err)
	}

	return llvm.ConstFloat(llvmType, value), nil
}

func (converter *ConstantConverter) convertStringConstant(naviaryConstant *nirvalue.Constant) (llvm.Value, error) {
	constantString := naviaryConstant.String()

	var value string
	_, err := fmt.Sscanf(constantString, "Constant(\"%s\")", &value)
	if err != nil {
		value = converter.extractStringValue(constantString)
	}

	return llvm.ConstString(value, false), nil
}

func (converter *ConstantConverter) extractStringValue(constantString string) string {
	startIndex := len("Constant(\"")
	endIndex := len(constantString) - 2 // Remove "))

	if startIndex >= len(constantString) || endIndex <= startIndex {
		return ""
	}

	return constantString[startIndex:endIndex]
}

func (converter *ConstantConverter) convertBoolConstant(naviaryConstant *nirvalue.Constant) (llvm.Value, error) {
	llvmType, err := converter.typeConverter.Convert(types.Bool)

	if err != nil {
		return llvm.Value{}, fmt.Errorf("failed to convert bool type: %w", err)
	}

	constantString := naviaryConstant.String()
	var value bool
	_, err = fmt.Sscanf(constantString, "Constant(%t)", &value)
	if err != nil {
		return llvm.Value{}, fmt.Errorf("failed to parse bool constant: %w", err)
	}

	var intValue uint64
	if value {
		intValue = 1
	} else {
		intValue = 0
	}

	return llvm.ConstInt(llvmType, intValue, false), nil
}
