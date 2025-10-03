package value

import (
	"compiler/types"
)

type Value interface {
	Type() types.Type
	IsConstant() bool
	String() string // for debugging
}
