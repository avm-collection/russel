package node

import (
	"strconv"

	"github.com/LordOfTrident/russel/internal/token"
)

// Identifier
type Id struct {
	Where token.Where

	Value string
}

func (n *Id) exprNode() {}
func (n *Id) NodeWhere() token.Where {return n.Where}
func (n *Id) String() string {return n.Value}

// Integer
type Int struct {
	Where token.Where

	Value int64
}

func (n *Int) exprNode() {}
func (n *Int) NodeWhere() token.Where {return n.Where}
func (n *Int) String() string {return strconv.Itoa(int(n.Value))}

// Bool
type Bool struct {
	Where token.Where

	Value bool
}

func (n *Bool) exprNode() {}
func (n *Bool) NodeWhere() token.Where {return n.Where}
func (n *Bool) String() string {
	if n.Value {
		return "true"
	} else {
		return "false"
	}
}

// String
type String struct {
	Where token.Where

	Value string
}

func (n *String) exprNode() {}
func (n *String) NodeWhere() token.Where {return n.Where}
func (n *String) String() string {return "\"" + n.Value + "\""}

// Func call
type FuncCall struct {
	Where token.Where

	Name *Id
	Args []Expr
}

func (n *FuncCall) exprNode() {}
func (n *FuncCall) NodeWhere() token.Where {return n.Where}
func (n *FuncCall) String() (str string) {
	str = "(" + n.Name.String()

	for _, s := range n.Args {
		str += " " + s.String()
	}

	str += ")"

	return
}
