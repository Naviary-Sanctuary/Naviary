package instruction

import "compiler/nir/value"

// Instruction represents a single operation in NIR
// All instructions operate on Values and may produce a Value
// Instructions are organized into BasicBlocks
type Instruction interface {
	GetResult() value.Value
	String() string // for debugging
}
