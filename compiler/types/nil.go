package types

type NilType struct{}

func (nilType *NilType) String() string {
	return "nil"
}

func (nilType *NilType) Equals(other Type) bool {
	return other == Nil
}

var Nil = &NilType{}
