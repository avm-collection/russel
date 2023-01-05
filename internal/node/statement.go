package node

import (
	"fmt"

	"github.com/LordOfTrident/russel/internal/token"
	"github.com/LordOfTrident/russel/internal/value"
)

type Statements struct {
	Token token.Token

	List []Statement
}

func (s *Statements) statementNode() {}
func (s *Statements) TypeToString() string       {return "statements"}
func (s *Statements) NodeToken()    token.Token  {return s.Token}
func (s *Statements) String()       (str string) {
	str = "{\n"

	for _, s := range s.List {
		str += s.String() + "\n"
	}
	str += "}"

	return
}

type ExprStatement struct {
	Expr Expr
}

func (e *ExprStatement) statementNode() {}
func (e *ExprStatement) TypeToString() string      {return e.Expr.TypeToString()}
func (e *ExprStatement) NodeToken()    token.Token {return e.Expr.NodeToken()}
func (e *ExprStatement) String()       string      {return e.Expr.String()}

// Type node
type Type struct {
	Token token.Token

	Type value.Type
}

func (t *Type) statementNode() {}
func (t *Type) TypeToString() string      {return "type"}
func (t *Type) NodeToken()    token.Token {return t.Token}
func (t *Type) String()       string      {return string(t.Type)}

// Variable declaration
type Let struct {
	Token token.Token

	Name *Id
	Type *Id
	Expr Expr
}

func (l *Let) statementNode() {}
func (l *Let) TypeToString() string      {return "variable declaration"}
func (l *Let) NodeToken()    token.Token {return l.Token}
func (l *Let) String()       string      {
	if l.Type != nil {
		return fmt.Sprintf("%v %v = %v", l.Type.String(), l.Name.String(), l.Expr.String())
	} else {
		return fmt.Sprintf("auto %v = %v", l.Name.String(), l.Expr.String())
	}
}

// Return
type Return struct {
	Token token.Token

	Expr Expr
}

func (r *Return) statementNode() {}
func (r *Return) TypeToString() string      {return "return statement"}
func (r *Return) NodeToken()    token.Token {return r.Token}
func (r *Return) String()       string      {return r.Expr.String()}

// Func declaration
type Func struct {
	Token token.Token

	Name *Id
	Type *Id
	Body *Statements
}

func (f *Func) statementNode() {}
func (f *Func) TypeToString() string      {return "function declaration"}
func (f *Func) NodeToken()    token.Token {return f.Token}
func (f *Func) String()       string      {
	if f.Type != nil {
		return fmt.Sprintf("%v %v %v", f.Type.String(), f.Name.String(), f.Body.String())
	} else {
		return fmt.Sprintf("void %v %v", f.Name.String(), f.Body.String())
	}
}