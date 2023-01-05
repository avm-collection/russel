package parser

import (
	"os"
	"strconv"

	"github.com/LordOfTrident/russel/internal/errors"
	"github.com/LordOfTrident/russel/internal/lexer"
	"github.com/LordOfTrident/russel/internal/token"
	"github.com/LordOfTrident/russel/internal/node"
)

type Parser struct {
	tok token.Token

	l *lexer.Lexer
}

func New(input, path string) *Parser {
	return &Parser{l: lexer.New(input, path)}
}

func (p *Parser) Parse() *node.Statements {
	topLevel := &node.Statements{Token: p.tok}

	if p.tok = p.l.NextToken(); p.tok.Type == token.Error {
		errors.Error(p.tok.Where, p.tok.Data)
		os.Exit(1)
	}

	for p.tok.Type != token.EOF {
		var s node.Statement

		switch p.tok.Type {
		case token.Func: s = p.parseFunc()
		case token.Let:  s = p.parseLet()

		default:
			errors.Error(p.tok.Where, "Unexpected %v in top-level", p.tok)
			p.next()
		}

		topLevel.List = append(topLevel.List, s)
	}

	return topLevel
}

func (p *Parser) parseStatements() *node.Statements {
	statements := &node.Statements{Token: p.tok}

	// One-liners
	if p.tok.Type != token.LCurly {
		s := p.parseStatement()
		statements.List = append(statements.List, s)

		return statements
	}

	p.next()

	// Statement list
	for p.tok.Type != token.RCurly {
		if p.tok.Type == token.EOF {
			errors.Error(p.tok.Where, "Expected matching '}', got %v", p.tok)
			errors.Note(statements.NodeToken().Where, "Opened here")
			return nil
		}

		s := p.parseStatement()

		statements.List = append(statements.List, s)
	}

	p.next()

	return statements
}

func (p *Parser) parseStatement() node.Statement {
	switch p.tok.Type {
	case token.Let:    return p.parseLet()
	case token.Return: return p.parseReturn()

	default:
		expr := p.parseExpr()

		return &node.ExprStatement{Expr: expr}
	}
}

func (p *Parser) parseReturn() node.Statement {
	r := &node.Return{Token: p.tok}
	p.next()
	r.Expr = p.parseExpr()

	return r
}

func (p *Parser) parseLet() *node.Let {
	l := &node.Let{Token: p.tok}

	p.next()
	l.Name = p.parseId()

	if p.tok.Type == token.Colon {
		p.next()

		l.Type = p.parseId()
	}

	if p.tok.Type != token.Assign {
		errors.Error(l.NodeToken().Where, "Uninitialized variables are not allowed")
		p.next()
		return l
	}

	p.next()
	l.Expr = p.parseExpr()

	return l
}

func (p *Parser) parseId() *node.Id {
	if p.tok.Type != token.Id {
		errors.Error(p.tok.Where, "Expected identifier, got %v", p.tok)
	}

	tok := p.tok
	p.next()

	return &node.Id{Token: tok, Value: tok.Data}
}

func (p *Parser) parseExpr() (expr node.Expr) {
	tok := p.tok
	switch p.tok.Type {
	case token.LParen: return p.parseFuncCall()
	case token.Id:     return p.parseId()

	case token.Dec:
		num, err := strconv.ParseInt(p.tok.Data, 10, 64)
		if err != nil {
			panic(err)
		}

		expr = &node.Int{Token: tok, Value: num}

	case token.Hex:
		num, err := strconv.ParseInt(p.tok.Data, 16, 64)
		if err != nil {
			panic(err)
		}

		expr = &node.Int{Token: tok, Value: num}

	case token.Oct:
		num, err := strconv.ParseInt(p.tok.Data, 8, 64)
		if err != nil {
			panic(err)
		}

		expr = &node.Int{Token: tok, Value: num}

	case token.True:   expr = &node.Bool{Token: tok, Value: true}
	case token.False:  expr = &node.Bool{Token: tok, Value: false}
	case token.String: expr = &node.String{Token: tok, Value: tok.Data}

	default: errors.Error(p.tok.Where, "Unexpected %v", p.tok)
	}

	p.next()

	return
}

func (p *Parser) parseFuncCall() *node.FuncCall {
	fc := &node.FuncCall{Token: p.tok}

	start := p.tok.Where
	p.next()
	fc.Name = p.parseId()

	for p.tok.Type != token.RParen {
		if p.tok.Type == token.EOF {
			errors.Error(p.tok.Where, "Expected matching ')', got %v", p.tok)
			errors.Note(start, "Opened here")
			return nil
		}

		fc.Args = append(fc.Args, p.parseExpr())
	}
	p.next()

	return fc
}

func (p *Parser) parseFunc() *node.Func {
	f := &node.Func{Token: p.tok}

	if p.next(); p.tok.Type == token.Inline {
		f.Inline = true

		p.next()
	}

	if p.tok.Type != token.LParen {
		errors.Error(p.tok.Where, "Expected '(' to open function head definition, got %v", p.tok)
		p.next()
		return nil
	}

	p.next()
	f.Name = p.parseId()

	start := p.tok.Where
	for p.tok.Type != token.RParen {
		if p.tok.Type == token.EOF {
			errors.Error(p.tok.Where, "Expected matching ')', got %v", p.tok)
			errors.Note(start, "Opened here")
			return nil
		}

		// TODO: Function definition arguments

		p.next()
	}

	if p.next(); p.tok.Type == token.Arrow {
		p.next()

		f.Type = p.parseId()
	}

 	f.Body = p.parseStatements()

	return f
}

func (p *Parser) next() {
	if p.tok.Type == token.EOF {
		return
	}

	if p.tok = p.l.NextToken(); p.tok.Type == token.Error {
		errors.Error(p.tok.Where, p.tok.Data)
		os.Exit(1)
	}
}
