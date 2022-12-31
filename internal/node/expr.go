package node

import (
	"strconv"

	"github.com/LordOfTrident/russel/internal/token"
)

type Expr interface {
	Node

	exprNode()
}

// Identifier
type Id struct {
	Token token.Token

	Value string
}

func (i *Id) exprNode() {}
func (i *Id) TypeToString() string      {return "identifier"}
func (i *Id) NodeToken()    token.Token {return i.Token}
func (i *Id) String()       string      {return i.Value}

// Integer
type Int struct {
	Token token.Token

	Value int64
}

func (i *Int) exprNode() {}
func (i *Int) TypeToString() string      {return "integer"}
func (i *Int) NodeToken()    token.Token {return i.Token}
func (i *Int) String()       string      {return strconv.Itoa(int(i.Value))}

// Bool
type Bool struct {
	Token token.Token

	Value bool
}

func (b *Bool) exprNode() {}
func (b *Bool) TypeToString() string      {return "bool"}
func (b *Bool) NodeToken()    token.Token {return b.Token}
func (b *Bool) String()       string {
	if b.Value {
		return "true"
	} else {
		return "false"
	}
}

// String
type String struct {
	Token token.Token

	Value string
}

func (s *String) exprNode() {}
func (s *String) TypeToString() string      {return "string"}
func (s *String) NodeToken()    token.Token {return s.Token}
func (s *String) String()       string      {return "\"" + s.Value + "\""}

// Func call
type FuncCall struct {
	Token token.Token

	Name *Id
	Args []Expr
}

func (fc *FuncCall) exprNode() {}
func (fc *FuncCall) TypeToString() string       {return "function call"}
func (fc *FuncCall) NodeToken()    token.Token  {return fc.Token}
func (fc *FuncCall) String()       (str string) {
	str = "(" + fc.Name.String()

	for _, s := range fc.Args {
		str += " " + s.String()
	}

	str += ")"

	return
}
