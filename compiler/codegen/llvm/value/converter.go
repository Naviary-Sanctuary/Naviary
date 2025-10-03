package value

import (
	nirvalue "compiler/nir/value"
	"fmt"

	"tinygo.org/x/go-llvm"
)

type Converter struct {
	constantConverter *ConstantConverter
	registry          *Registry
}

func NewConverter(typeConverter TypeConverter) *Converter {
	return &Converter{
		constantConverter: NewConstantConverter(typeConverter),
		registry:          NewRegistry(),
	}
}

func (converter *Converter) RegisterTemporary(temporary *nirvalue.Temporary, llvmValue llvm.Value) {
	converter.registry.RegisterTemporary(temporary, llvmValue)
}

func (converter *Converter) GetTemporary(temporary *nirvalue.Temporary) (llvm.Value, error) {
	return converter.registry.GetTemporary(temporary)
}

func (converter *Converter) RegisterVariable(variable *nirvalue.Variable, llvmValue llvm.Value) {
	converter.registry.RegisterVariable(variable, llvmValue)
}

func (converter *Converter) GetVariable(variable *nirvalue.Variable) (llvm.Value, error) {
	return converter.registry.GetVariable(variable)
}

func (converter *Converter) Convert(nirVal nirvalue.Value) (llvm.Value, error) {
	if nirVal == nil {
		return llvm.Value{}, fmt.Errorf("cannot convert nil value")
	}

	switch val := nirVal.(type) {
	case *nirvalue.Constant:
		return converter.constantConverter.Convert(val)
	case *nirvalue.Temporary:
		return converter.GetTemporary(val)
	case *nirvalue.Variable:
		return converter.GetVariable(val)
	default:
		return llvm.Value{}, fmt.Errorf("unsupported NIR value type: %T", nirVal)
	}
}

func (converter *Converter) Reset() {
	converter.registry.Reset()
}
