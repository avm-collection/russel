package node

import (
	"github.com/LordOfTrident/russel/internal/token"
)

type Node interface {
	NodeToken()    token.Token
	TypeToString() string
	String()       string
}

// Statements
type Statement interface {
	Node

	statementNode()
}
