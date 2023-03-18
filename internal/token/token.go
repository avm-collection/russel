package token

import "fmt"

type Where struct {
	Row,  Col, Len int
	Path, Line     string
}

func (w Where) AtRow()   int    {return w.Row}
func (w Where) AtCol()   int    {return w.Col}
func (w Where) GetLen()  int    {return w.Len}
func (w Where) InFile()  string {return w.Path}
func (w Where) GetLine() string {return w.Line}
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

	LSquare
	RSquare

	Assign
	Increment
	Decrement

	Arrow
	Colon
	Dot

	Module
	Import

	Macro
	Let
	Proc

	Inline
	Interrupt

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

	LSquare: "[",
	RSquare: "]",

	Assign:    "=",
	Increment: "++",
	Decrement: "--",

	Dot:   ".",
	Arrow: "->",
	Colon: ":",

	Module: "keyword module",
	Import: "keyword import",

	Macro: "keyword mac",
	Let:   "keyword let",
	Proc:  "keyword proc",

	Inline:    "keyword inline",
	Interrupt: "keyword interrupt",

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
	if count != 40 {
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
