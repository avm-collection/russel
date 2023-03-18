package node

import (
	"fmt"

	"github.com/LordOfTrident/russel/internal/token"
	"github.com/LordOfTrident/russel/internal/value"
)

type Stmts struct {
	Where token.Where

	List []Stmt
}

func (n *Stmts) stmtNode() {}
func (n *Stmts) NodeWhere() token.Where {return n.Where}
func (n *Stmts) String() (str string) {
	str = "{\n"

	for _, s := range n.List {
		str += s.String() + "\n"
	}

	str += "}"
	return
}

type ExprStmt struct {
	Expr Expr
}

func (n *ExprStmt) stmtNode() {}
func (n *ExprStmt) NodeWhere() token.Where {return n.Expr.NodeWhere()}
func (n *ExprStmt) String() string {return n.Expr.String()}

// Type node
type Type struct {
	Where token.Where

	Type value.Type
}

func (n *Type) stmtNode() {}
func (n *Type) NodeWhere() token.Where {return n.Where}
func (n *Type) String() string {return string(n.Type)}

// Variable declaration
type Let struct {
	Where token.Where

	Name *Id
	Type *Id
	Expr  Expr
}

func (n *Let) stmtNode() {}
func (n *Let) NodeWhere() token.Where {return n.Where}
func (n *Let) String() string {
	if n.Type != nil {
		return fmt.Sprintf("%v %v = %v", n.Type.String(), n.Name.String(), n.Expr.String())
	} else {
		return fmt.Sprintf("auto %v = %v", n.Name.String(), n.Expr.String())
	}
}

// Variable assignment
type Assign struct {
	Where token.Where

	Name *Id
	Expr  Expr
}

func (n *Assign) stmtNode() {}
func (n *Assign) NodeWhere() token.Where {return n.Where}
func (n *Assign) String() string {
	return fmt.Sprintf("%v = %v", n.Name.String(), n.Expr.String())
}

// Variable increment/decrement
type Increment struct {
	Where token.Where

	Name    *Id
	Negative bool
}

func (n *Increment) stmtNode() {}
func (n *Increment) NodeWhere() token.Where {return n.Where}
func (n *Increment) String() string {
	if n.Negative {
		return fmt.Sprintf("-- %v", n.Name.String())
	} else {
		return fmt.Sprintf("++ %v", n.Name.String())
	}
}

// Macro declaration
type Macro struct {
	Where token.Where

	Name *Id
	Expr  Expr
}

func (n *Macro) stmtNode() {}
func (n *Macro) NodeWhere() token.Where {return n.Where}
func (n *Macro) String() string {
	return fmt.Sprintf("macro %v = %v", n.Name.String(), n.Expr.String())
}

// Return
type Return struct {
	Where token.Where

	Expr Expr
}

func (n *Return) stmtNode() {}
func (n *Return) NodeWhere() token.Where {return n.Where}
func (n *Return) String() string {return n.Expr.String()}

// If
type If struct {
	Where token.Where

	Var *Let

	Cond  Expr
	Then *Stmts
	Else *Stmts

	Invert bool
}

func (n *If) stmtNode() {}
func (n *If) NodeWhere() token.Where {return n.Where}
func (n *If) String()    (str string) {
	if n.Invert {
		str += "unless"
	} else {
		str += "if"
	}

	if n.Var != nil {
		str += " " + n.Var.String() + ";"
	}

	if n.Else != nil {
		str += fmt.Sprintf(" %v %v else %v", n.Cond.String(), n.Then.String(), n.Else.String())
	} else {
		str += fmt.Sprintf(" %v %v", n.Cond.String(), n.Then.String())
	}
	return
}

// While
type While struct {
	Where token.Where

	Cond  Expr
	Body *Stmts

	Invert bool
}

func (n *While) stmtNode() {}
func (n *While) NodeWhere() token.Where {return n.Where}
func (n *While) String() string {
	return fmt.Sprintf("while %v %v", n.Cond.String(), n.Body.String())
}

// For
type For struct {
	Where token.Where

	Var  *Let
	Cond  Expr
	Last  Stmt
	Body *Stmts

	Invert bool
}

func (n *For) stmtNode() {}
func (n *For) NodeWhere() token.Where {return n.Where}
func (n *For) String() (str string) {
	str += "for "

	if n.Var != nil {
		str += n.Var.String() + "; "
	}

	str += fmt.Sprintf("%v; %v %v", n.Cond.String(), n.Last.String(), n.Body.String())
	return
}

// Break
type Break struct {
	Where token.Where
}

func (n *Break) stmtNode() {}
func (n *Break) NodeWhere() token.Where {return n.Where}
func (n *Break) String() string {return "break"}

// Continue
type Continue struct {
	Where token.Where
}

func (n *Continue) stmtNode() {}
func (n *Continue) NodeWhere() token.Where {return n.Where}
func (n *Continue) String() string {return "continue"}

const (
	AttrInline = 1 << iota
	AttrInterrupt
)

// Func declaration
type Func struct {
	Where token.Where

	Attrs int

	Name *Id
	Type *Id
	Body *Stmts
}

func (n *Func) stmtNode() {}
func (n *Func) NodeWhere() token.Where {return n.Where}
func (n *Func) String() string {
	if n.Type != nil {
		return fmt.Sprintf("%v %v %v", n.Type.String(), n.Name.String(), n.Body.String())
	} else {
		return fmt.Sprintf("void %v %v", n.Name.String(), n.Body.String())
	}
}
