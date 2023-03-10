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
	Expr  Expr
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

// Variable assignment
type Assign struct {
	Token token.Token

	Name *Id
	Expr  Expr
}

func (a *Assign) statementNode() {}
func (a *Assign) TypeToString() string      {return "variable declaration"}
func (a *Assign) NodeToken()    token.Token {return a.Token}
func (a *Assign) String()       string      {
	return fmt.Sprintf("%v = %v", a.Name.String(), a.Expr.String())
}

// Variable increment/decrement
type Increment struct {
	Token token.Token

	Name     *Id
	Decrement bool
}

func (i *Increment) statementNode() {}
func (i *Increment) TypeToString() string      {return "variable increment"}
func (i *Increment) NodeToken()    token.Token {return i.Token}
func (i *Increment) String()       string      {
	if i.Decrement {
		return fmt.Sprintf("-- %v", i.Name.String())
	} else {
		return fmt.Sprintf("++ %v", i.Name.String())
	}
}

// Macro declaration
type Macro struct {
	Token token.Token

	Name *Id
	Expr  Expr
}

func (m *Macro) statementNode() {}
func (m *Macro) TypeToString() string      {return "macro declaration"}
func (m *Macro) NodeToken()    token.Token {return m.Token}
func (m *Macro) String()       string      {
	return fmt.Sprintf("macro %v = %v", m.Name.String(), m.Expr.String())
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

// If
type If struct {
	Token token.Token

	Cond  Expr
	Then *Statements
	Else *Statements

	Invert bool
}

func (i *If) statementNode() {}
func (i *If) TypeToString() string      {return "if statement"}
func (i *If) NodeToken()    token.Token {return i.Token}
func (i *If) String()       string      {
	if i.Else != nil {
		return fmt.Sprintf("if %v %v else %v", i.Cond.String(), i.Then.String(), i.Else.String())
	} else {
		return fmt.Sprintf("if %v %v", i.Cond.String(), i.Then.String())
	}
}

// While
type While struct {
	Token token.Token

	Cond  Expr
	Body *Statements

	Invert bool
}

func (w *While) statementNode() {}
func (w *While) TypeToString() string      {return "while statement"}
func (w *While) NodeToken()    token.Token {return w.Token}
func (w *While) String()       string      {
	return fmt.Sprintf("while %v %v", w.Cond.String(), w.Body.String())
}

// For
type For struct {
	Token token.Token

	Init  Statement
	Cond  Expr
	Last  Statement
	Body *Statements

	Invert bool
}

func (f *For) statementNode() {}
func (f *For) TypeToString() string      {return "for statement"}
func (f *For) NodeToken()    token.Token {return f.Token}
func (f *For) String()       string      {
	return fmt.Sprintf("for %v; %v; %v %v",
	                   f.Init.String(), f.Cond.String(), f.Last.String(), f.Body.String())
}

// Break
type Break struct {
	Token token.Token
}

func (b *Break) statementNode() {}
func (b *Break) TypeToString() string      {return "break statement"}
func (b *Break) NodeToken()    token.Token {return b.Token}
func (b *Break) String()       string      {return "break"}

// Continue
type Continue struct {
	Token token.Token
}

func (c *Continue) statementNode() {}
func (c *Continue) TypeToString() string      {return "continue statement"}
func (c *Continue) NodeToken()    token.Token {return c.Token}
func (c *Continue) String()       string      {return "continue"}

// Func declaration
type Func struct {
	Token token.Token

	Name  *Id
	Type  *Id
	Body  *Statements
	Inline bool
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
