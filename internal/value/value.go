package value

import "fmt"

type Type int
const (
	Void = iota
	Int
	Bool
	String
)

func (t Type) String() string {
	switch t {
	case Int:    return "int"
	case Bool:   return "bool"
	case String: return "string"

	default: panic(fmt.Errorf("Unknown type value %v", int(t)))
	}
}
