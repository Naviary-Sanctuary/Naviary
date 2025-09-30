package types

type Type interface {
	String() string
	Equals(other Type) bool
}
