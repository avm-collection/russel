package node

import (
	"github.com/LordOfTrident/russel/internal/token"
)

type Node interface {
	NodeWhere() token.Where
	String()    string
}

type Stmt interface {
	Node
	stmtNode()
}

type Expr interface {
	Node
	exprNode()
}
