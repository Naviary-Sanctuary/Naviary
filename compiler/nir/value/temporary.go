package value

import (
	"compiler/types"
	"fmt"
)

// temporary represents a compiler-generated temporary value
// unnamed value created during expression evaluation
// Examples: %0, %1, %2 in LLVM IR
type Temporary struct {
	id            int
	temporaryType types.Type
}

func NewTemporary(id int, temporaryType types.Type) *Temporary {
	return &Temporary{
		id:            id,
		temporaryType: temporaryType,
	}
}

func (temporary *Temporary) Type() types.Type {
	return temporary.temporaryType
}
func (temporary *Temporary) String() string {
	return fmt.Sprintf("%%%d: %s", temporary.id, temporary.temporaryType.String())
}

func (temporary *Temporary) IsConstant() bool {
	return false
}

func (temporary *Temporary) GetID() int {
	return temporary.id
}
