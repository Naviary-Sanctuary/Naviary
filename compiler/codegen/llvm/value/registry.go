package value

import (
	nirvalue "compiler/nir/value"
	"fmt"

	"tinygo.org/x/go-llvm"
)

type Registry struct {
	temporaryMap map[int]llvm.Value
	variableMap  map[string]llvm.Value
}

func NewRegistry() *Registry {
	return &Registry{
		temporaryMap: make(map[int]llvm.Value),
		variableMap:  make(map[string]llvm.Value),
	}
}

func (registry *Registry) RegisterTemporary(naviaryTemporary *nirvalue.Temporary, llvmValue llvm.Value) {
	registry.temporaryMap[naviaryTemporary.GetID()] = llvmValue
}

func (registry *Registry) GetTemporary(naviaryTemporary *nirvalue.Temporary) (llvm.Value, error) {
	llvmValue, exists := registry.temporaryMap[naviaryTemporary.GetID()]

	if !exists {
		return llvm.Value{}, fmt.Errorf("temporary %%%d not found in registry", naviaryTemporary.GetID())
	}

	return llvmValue, nil
}

func (registry *Registry) RegisterVariable(naviaryVariable *nirvalue.Variable, llvmValue llvm.Value) {
	registry.variableMap[naviaryVariable.String()] = llvmValue
}

func (registry *Registry) GetVariable(naviaryVariable *nirvalue.Variable) (llvm.Value, error) {
	llvmValue, exists := registry.variableMap[naviaryVariable.String()]

	if !exists {
		return llvm.Value{}, fmt.Errorf("variable %s not found in registry", naviaryVariable.String())
	}

	return llvmValue, nil
}

func (registry *Registry) Reset() {
	registry.temporaryMap = make(map[int]llvm.Value)
	registry.variableMap = make(map[string]llvm.Value)
}
