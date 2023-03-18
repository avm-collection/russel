package lexer

import (
	"strings"

	"github.com/LordOfTrident/russel/internal/token"
)

const EOF = '\x00'

var Keywords = map[string]token.Type{
	"true":  token.True,
	"false": token.False,

	"->": token.Arrow,
	"=":  token.Assign,
	"++": token.Increment,
	"--": token.Decrement,

	"module": token.Module,
	"import": token.Import,

	"macro":  token.Macro,
	"let":    token.Let,
	"proc":   token.Proc,
	"inline": token.Inline,

	"if":     token.If,
	"unless": token.Unless,
	"else":   token.Else,

	"while":    token.While,
	"until":    token.Until,
	"for":      token.For,
	"break":    token.Break,
	"continue": token.Continue,

	"return": token.Return,
}

type Lexer struct {
	input string
	pos   int
	ch    byte

	lineStart int

	where token.Where
}

func New(input, path string) *Lexer {
	l := &Lexer{input: input, pos: -1}
	l.next()

	l.where.Row  = 1
	l.where.Path = path
	l.where.Line = l.getLine()

	return l
}

func isSeparatorCh(ch byte) bool {
	switch ch {
	case '(', ')', '{', '}', '[', ']', ',', '.', ':', ';': return true

	default: return isWhitespace(ch)
	}
}

func isIdCh(ch byte) bool {
	switch ch {
	case '+', '-', '*', '/', '%', '>', '<', '=', '_', '$': return true

	default: return (ch >= 'a' && ch <= 'z') ||
	                (ch >= 'A' && ch <= 'Z') ||
	                (ch >= '0' && ch <= '9')
	}
}

func isDecDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isOctDigit(ch byte) bool {
	return ch >= '0' && ch <= '7'
}

func isBinDigit(ch byte) bool {
	return ch == '0' || ch == '1'
}

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') ||
	       (ch >= 'a' && ch <= 'f') ||
	       (ch >= 'A' && ch <= 'F')
}

func isWhitespace(ch byte) bool {
	switch ch {
	case ' ', '\r', '\t', '\v', '\f', '\n': return true

	default: return false
	}
}

func (l *Lexer) NextToken() (tok token.Token) {
	for {
		start := l.where

		switch l.ch {
		case EOF: return token.NewEOF(l.where)

		case '#':
			l.skipComment()

			continue

		case ';': tok = l.lexSimpleSym(token.Separator)

		case '(': tok = l.lexSimpleSym(token.LParen)
		case ')': tok = l.lexSimpleSym(token.RParen)

		case '{': tok = l.lexSimpleSym(token.LCurly)
		case '}': tok = l.lexSimpleSym(token.RCurly)

		case '[': tok = l.lexSimpleSym(token.LSquare)
		case ']': tok = l.lexSimpleSym(token.RSquare)

		case ':': tok = l.lexSimpleSym(token.Colon)
		case '.': tok = l.lexSimpleSym(token.Dot)

		case '"': tok = l.lexString()

		default:
			if isDecDigit(l.ch) {
				tok = l.lexNum()
			} else if isIdCh(l.ch) {
				tok = l.lexId()
			} else if isWhitespace(l.ch) {
				l.next()

				continue
			} else {
				l.where.Len = 1
				return token.NewError(l.where, "Unexpected character '%v'", string(l.ch))
			}
		}

		tok.Where = start
		if l.where.Row != start.Row {
			tok.Where.Len = len(start.Line) - start.Col + 1
		} else {
			tok.Where.Len = l.where.Col - start.Col
		}

		break
	}

	return
}

func (l *Lexer) lexString() token.Token {
	l.next()

	str    := ""
	escape := false

	for l.ch != '"' {
		switch l.ch {
		case '\\':
			if escape {
				str   += string('\\')
				escape = false
			} else {
				escape = true
			}

		default:
			if escape {
				switch l.ch {
				case 'e': str += string(27)
				case 'n': str += string('\n')
				case 'r': str += string('\r')
				case 't': str += string('\t')
				case 'v': str += string('\v')
				case 'b': str += string('\b')
				case 'f': str += string('\f')

				default: return token.NewError(l.where, "Unknown escape sequence '\\%v'",
				                               string(l.ch))
				}

				escape = false
			} else {
				str += string(l.ch)
			}
		}

		l.next()
	}

	l.next()

	return token.Token{Type: token.String, Data: str}
}

func (l *Lexer) lexSimpleSym(type_ token.Type) (tok token.Token) {
	tok = token.Token{Type: type_, Data: string(l.ch)}
	l.next()

	return
}

func (l *Lexer) lexNum() token.Token {
	if l.ch == '0' && (l.peek() == 'x' || l.peek() == 'X') {
		l.next()
		l.next()

		return l.lexHex()
	} else if l.ch == '0' && (l.peek() == 'o' || l.peek() == 'O') {
		l.next()
		l.next()

		return l.lexOct()
	} else if l.ch == '0' && (l.peek() == 'b' || l.peek() == 'B') {
		l.next()
		l.next()

		return l.lexBin()
	} else {
		return l.lexDec()
	}
}

func (l *Lexer) lexHex() token.Token {
	str := ""

	for !isSeparatorCh(l.ch) {
		if !isHexDigit(l.ch) {
			return token.NewError(l.where, "Unexpected character '%v' in hexadecimal number",
			                      string(l.ch))
		}

		str += string(l.ch)

		l.next()
	}

	return token.Token{Type: token.Hex, Data: str}
}

func (l *Lexer) lexDec() token.Token {
	str := ""

	for !isSeparatorCh(l.ch) {
		if !isDecDigit(l.ch) {
			return token.NewError(l.where, "Unexpected character '%v' in decimal number",
			                      string(l.ch))
		}

		str += string(l.ch)

		l.next()
	}

	return token.Token{Type: token.Dec, Data: str}
}

func (l *Lexer) lexOct() token.Token {
	str := ""

	for !isSeparatorCh(l.ch) {
		if !isOctDigit(l.ch) {
			return token.NewError(l.where, "Unexpected character '%v' in octal number",
			                      string(l.ch))
		}

		str += string(l.ch)

		l.next()
	}

	return token.Token{Type: token.Oct, Data: str}
}

func (l *Lexer) lexBin() token.Token {
	str := ""

	for !isSeparatorCh(l.ch) {
		if !isBinDigit(l.ch) {
			return token.NewError(l.where, "Unexpected character '%v' in binary number",
			                      string(l.ch))
		}

		str += string(l.ch)

		l.next()
	}

	return token.Token{Type: token.Bin, Data: str}
}

func (l *Lexer) lexId() token.Token {
	str := l.readId()

	type_, ok := Keywords[str]
	if !ok {
		type_ = token.Id
	}

	return token.Token{Type: type_, Data: str}
}

func (l *Lexer) readId() (str string) {
	for isIdCh(l.ch) {
		str += string(l.ch)

		l.next()
	}

	return
}

func (l *Lexer) skipComment() {
	for l.ch != EOF && l.ch != '\n' {
		l.next()
	}
}

func (l *Lexer) next() {
	l.pos ++
	if l.pos >= len(l.input) {
		l.ch = EOF
	} else {
		l.ch = l.input[l.pos]
	}

	if l.ch == '\n' {
		l.where.Col = 0
		l.where.Row ++
		l.where.Line = l.getLine()
	} else {
		l.where.Col ++
	}
}

func (l *Lexer) peek() byte {
	if l.pos + 1 >= len(l.input) {
		return EOF
	} else {
		return l.input[l.pos + 1]
	}
}

func (l *Lexer) getLine() (line string) {
	end := strings.Index(l.input[l.lineStart:], "\n")

	if end == -1 {
		line = l.input[l.lineStart:]
	} else {
		line = l.input[l.lineStart:l.lineStart + end]
	}

	l.lineStart += end + 1

	return
}
