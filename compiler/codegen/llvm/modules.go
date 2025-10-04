package llvm

import (
	llvmvalue "compiler/codegen/llvm/value"
	"compiler/nir"
	"fmt"

	"tinygo.org/x/go-llvm"
)

type ModuleConverter struct {
	context           *Context
	module            llvm.Module
	typeConverter     *TypeConverter
	valueConverter    *llvmvalue.Converter
	functionConverter *FunctionConverter
}

func NewModuleConverter(context *Context, moduleName string) *ModuleConverter {
	module := context.GetRawContext().NewModule(moduleName)

	typeConverter := NewTypeConverter(context)
	valueConverter := llvmvalue.NewConverter(typeConverter)
	functionConverter := NewFunctionConverter(context, module, typeConverter, valueConverter)

	return &ModuleConverter{
		context:           context,
		module:            module,
		typeConverter:     typeConverter,
		valueConverter:    valueConverter,
		functionConverter: functionConverter,
	}
}

func (converter *ModuleConverter) Convert(nirModule *nir.Module) (string, error) {
	err := converter.declareRuntimeFunctions()
	if err != nil {
		return "", fmt.Errorf("failed to declare runtime functions: %w", err)
	}

	err = converter.convertFunctions(nirModule.Functions)
	if err != nil {
		return "", fmt.Errorf("failed to convert functions: %w", err)
	}

	if err := llvm.VerifyModule(converter.module, llvm.ReturnStatusAction); err != nil {
		return "", fmt.Errorf("failed to verify module: %w", err)
	}

	llvmIR := converter.module.String()

	return llvmIR, nil
}

func (converter *ModuleConverter) declareRuntimeFunctions() error {
	// TODO: currently we only support int64 type for print function
	printParamTypes := []llvm.Type{llvm.GlobalContext().Int64Type()}
	printFuncType := llvm.FunctionType(llvm.GlobalContext().VoidType(), printParamTypes, false)
	llvm.AddFunction(converter.module, "print", printFuncType)

	return nil
}

func (converter *ModuleConverter) convertFunctions(nirFunctions []*nir.Function) error {
	for _, nirFunction := range nirFunctions {
		err := converter.functionConverter.Convert(nirFunction)
		if err != nil {
			return fmt.Errorf("failed to convert function %s: %w", nirFunction.Name, err)
		}
	}

	return nil
}
