package token

import "fmt"

type Where struct {
	Row,  Col, Len int
	Path, Line     string
}

func (w Where) String() string {
	return fmt.Sprintf("%v:%v:%v", w.Path, w.Row, w.Col)
}

type Type int
const (
	EOF = Type(iota)

	Separator

	Id

	Dec
	Hex
	Oct
	Bin
	Char

	True
	False

	String

	LParen
	RParen

	LCurly
	RCurly

	LBracket
	RBracket

	Assign
	Increment
	Decrement

	Arrow
	Colon
	Dot

	Name
	Uses

	Macro
	Let
	Func
	Inline

	If
	Unless
	Else

	While
	Until
	For
	Break
	Continue

	Return

	Error
	count // Count of all token types
)

var tokTypeNames = map[Type]string{
	EOF: "end of file",

	Separator: ";",

	Id: "identifier",

	Dec:  "decimal number",
	Hex:  "hexadecimal number",
	Oct:  "octal number",
	Char: "character",

	True:  "true",
	False: "false",

	String: "string",

	LParen: "(",
	RParen: ")",

	LCurly: "{",
	RCurly: "}",

	LBracket: "[",
	RBracket: "]",

	Assign:    "=",
	Increment: "++",
	Decrement: "--",

	Dot:   ".",
	Arrow: "->",
	Colon: ":",

	Name: "keyword name",
	Uses: "keyword uses",

	Macro:  "keyword mac",
	Let:    "keyword let",
	Func:   "keyword fun",
	Inline: "keyword inline",

	If:     "keyword if",
	Unless: "keyword unless",
	Else:   "keyword else",

	While:    "keyword while",
	Until:    "keyword until",
	For:      "keyword for",
	Break:    "keyword break",
	Continue: "keyword continue",

	Return: "keyword return",

	Error: "error",
}

func AllTokensCoveredTest() {
	if count != 39 {
		panic("Cover all token types")
	}
}

func (t Type) String() string {
	name, ok := tokTypeNames[t]
	if !ok {
		panic("Unreachable")
	}

	return name
}

type Token struct {
	Type Type
	Data string

	Where Where
}

func (t Token) String() string {
	switch t.Type {
	case EOF: return "'end of file'"

	case Id, Dec, Hex, Oct, Char, String: return fmt.Sprintf("'%v' of type '%v'", t.Data, t.Type)

	default: return fmt.Sprintf("'%v'", t.Type)
	}
}

func NewEOF(where Where) Token {
	return Token{Type: EOF, Where: where}
}

func NewError(where Where, format string, args... interface{}) Token {
	return Token{Type: Error, Where: where, Data: fmt.Sprintf(format, args...)}
}
