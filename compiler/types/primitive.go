package types

type PrimitiveType struct {
	Name string
}

func (primitiveType *PrimitiveType) String() string {
	return primitiveType.Name
}

func (primitiveType *PrimitiveType) Equals(other Type) bool {
	if otherPrimitive, ok := other.(*PrimitiveType); ok {
		return primitiveType.Name == otherPrimitive.Name
	}
	return false
}

var (
	Int    = &PrimitiveType{Name: "int"}
	Float  = &PrimitiveType{Name: "float"}
	String = &PrimitiveType{Name: "string"}
	Bool   = &PrimitiveType{Name: "bool"}
)

func GetPrimitiveType(name string) Type {
	switch name {
	case "int":
		return Int
	case "float":
		return Float
	case "string":
		return String
	case "bool":
		return Bool
	default:
		return nil
	}
}
