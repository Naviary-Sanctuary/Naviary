package typechecker

import "compiler/types"

type EntryKind int

const (
	Variable EntryKind = iota
	Function
)

func (entryKind EntryKind) String() string {
	switch entryKind {
	case Variable:
		return "variable"
	case Function:
		return "function"
	default:
		return "unknown"
	}
}

// Entry represents a registered item in the type checker
// It stores information about variables, functions, classes, etc.
type Entry struct {
	Name string
	kind EntryKind
	Type types.Type
}

func NewVariableEntry(name string, variableType types.Type) *Entry {
	return &Entry{
		Name: name,
		kind: Variable,
		Type: variableType,
	}
}

func NewFunctionEntry(name string, functionType types.Type) *Entry {
	return &Entry{
		Name: name,
		kind: Function,
		Type: functionType,
	}
}
