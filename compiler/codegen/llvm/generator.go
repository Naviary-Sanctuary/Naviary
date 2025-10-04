package llvm

import (
	"compiler/nir"
	"fmt"
)

type Generator struct {
	context *Context
}

func NewGenerator() *Generator {
	return &Generator{
		context: NewContext(),
	}
}

func (generator *Generator) Generate(nirModule *nir.Module) (string, error) {
	moduleConverter := NewModuleConverter(generator.context, nirModule.Name)

	llvmIr, err := moduleConverter.Convert(nirModule)
	if err != nil {
		return "", fmt.Errorf("failed to convert module: %w", err)
	}

	return llvmIr, nil
}

func (generator *Generator) Dispose() {
	if generator.context != nil {
		generator.context.Dispose()
		generator.context = nil
	}
}
