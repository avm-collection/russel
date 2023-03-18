package parser

import (
	"os"
	"strconv"

	"github.com/avm-collection/goerror"

	"github.com/LordOfTrident/russel/internal/lexer"
	"github.com/LordOfTrident/russel/internal/token"
	"github.com/LordOfTrident/russel/internal/node"
)

type Parser struct {
	WhereFileEnd token.Where

	tok token.Token

	l *lexer.Lexer
}

func New(input, path string) *Parser {
	return &Parser{l: lexer.New(input, path)}
}

func (p *Parser) Parse() *node.Stmts {
	topLevel := &node.Stmts{Where: p.tok.Where}

	if p.tok = p.l.NextToken(); p.tok.Type == token.Error {
		goerror.Error(p.tok.Where, p.tok.Data)
		os.Exit(1)
	}

	for p.tok.Type != token.EOF {
		var s node.Stmt

		switch p.tok.Type {
		case token.Proc:  s = p.parseFunc()
		case token.Let:   s = p.parseLet()
		case token.Macro: s = p.parseMacro()

		default:
			goerror.Error(p.tok.Where, "Unexpected %v in top-level", p.tok)
			p.next()
		}

		topLevel.List = append(topLevel.List, s)
	}

	return topLevel
}

func (p *Parser) parseStmts() *node.Stmts {
	n := &node.Stmts{Where: p.tok.Where}

	// One-liners
	if p.tok.Type != token.LCurly {
		s := p.parseStmt()
		n.List = append(n.List, s)

		return n
	}

	p.next()

	// Stmt list
	for p.tok.Type != token.RCurly {
		if p.tok.Type == token.EOF {
			goerror.Error(p.tok.Where, "Expected matching '%v', got %v", token.RCurly, p.tok)
			goerror.Note(n.Where, "Opened here")
			return nil
		}

		s := p.parseStmt()

		n.List = append(n.List, s)
	}

	p.next()
	return n
}

func (p *Parser) parseStmt() node.Stmt {
	tok := p.tok
	switch tok.Type {
	case token.Let:      return p.parseLet()
	case token.Macro:    return p.parseMacro()
	case token.Return:   return p.parseReturn()
	case token.If:       return p.parseIf(false)
	case token.Unless:   return p.parseIf(true)
	case token.While:    return p.parseWhile(false)
	case token.Until:    return p.parseWhile(true)
	case token.For:      return p.parseFor()

	case token.Break:
		p.next()
		return &node.Break{Where: tok.Where}

	case token.Continue:
		p.next()
		return &node.Continue{Where: tok.Where}

	case token.Increment:
		p.next()
		return &node.Increment{Where: tok.Where, Name: p.parseId()}

	case token.Decrement:
		p.next()
		return &node.Increment{Where: tok.Where, Name: p.parseId(), Negative: true}

	case token.Id:
		id := p.parseId()
		switch p.tok.Type {
		case token.Assign:
			p.next()
			return &node.Assign{Where: p.tok.Where, Name: id, Expr: p.parseExpr()}

		default: return &node.ExprStmt{Expr: id}
		}

	default:
		expr := p.parseExpr()

		return &node.ExprStmt{Expr: expr}
	}
}

func (p *Parser) parseReturn() node.Stmt {
	n := &node.Return{Where: p.tok.Where}

	p.next()
	if p.tok.Type == token.Arrow {
		p.next()
		n.Expr = p.parseExpr()
	}
	return n
}

func (p *Parser) parseIf(invert bool) node.Stmt {
	n := &node.If{Where: p.tok.Where, Invert: invert}

	p.next()
	if p.tok.Type == token.Let {
		n.Var = p.parseLet()

		if p.tok.Type != token.Separator {
			goerror.Error(p.tok.Where, "Expected '%v', got %v", token.Separator, p.tok)
			return n
		}
		p.next()
	}

	n.Cond = p.parseExpr()
	n.Then = p.parseStmts()

	if p.tok.Type == token.Else {
		p.next();
		n.Else = p.parseStmts()
	}
	return n
}

func (p *Parser) parseWhile(invert bool) node.Stmt {
	n := &node.While{Where: p.tok.Where, Invert: invert}

	p.next()
	n.Cond = p.parseExpr()
	n.Body = p.parseStmts()
	return n
}

func (p *Parser) parseFor() node.Stmt {
	n := &node.For{Where: p.tok.Where}

	p.next()
	if p.tok.Type == token.Let {
		n.Var = p.parseLet();

		if p.tok.Type != token.Separator {
			goerror.Error(p.tok.Where, "Expected '%v', got %v", token.Separator, p.tok)
			return n
		}
		p.next()
	}

	n.Cond = p.parseExpr()
	if p.tok.Type != token.Separator {
		goerror.Error(p.tok.Where, "Expected '%v', got %v", token.Separator, p.tok)
		return n
	}
	p.next()

	n.Last = p.parseStmt()
	n.Body = p.parseStmts()
	return n
}

func (p *Parser) parseLet() *node.Let {
	n := &node.Let{Where: p.tok.Where}

	p.next()
	n.Name = p.parseId()

	if p.tok.Type == token.Colon {
		p.next()

		n.Type = p.parseId()
	}

	if p.tok.Type != token.Assign {
		return n
	}

	p.next()
	n.Expr = p.parseExpr()
	return n
}

func (p *Parser) parseMacro() *node.Macro {
	n := &node.Macro{Where: p.tok.Where}

	p.next()
	n.Name = p.parseId()

	if p.tok.Type != token.Assign {
		goerror.Error(n.Where, "Macro expression expected")
		return n
	}

	p.next()
	n.Expr = p.parseExpr()
	return n
}

func (p *Parser) parseId() *node.Id {
	if p.tok.Type != token.Id {
		goerror.Error(p.tok.Where, "Expected identifier, got %v", p.tok)
	}

	tok := p.tok
	p.next()
	return &node.Id{Where: tok.Where, Value: tok.Data}
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

		expr = &node.Int{Where: tok.Where, Value: num}

	case token.Hex:
		num, err := strconv.ParseInt(p.tok.Data, 16, 64)
		if err != nil {
			panic(err)
		}

		expr = &node.Int{Where: tok.Where, Value: num}

	case token.Oct:
		num, err := strconv.ParseInt(p.tok.Data, 8, 64)
		if err != nil {
			panic(err)
		}

		expr = &node.Int{Where: tok.Where, Value: num}

	case token.Bin:
		num, err := strconv.ParseInt(p.tok.Data, 2, 64)
		if err != nil {
			panic(err)
		}

		expr = &node.Int{Where: tok.Where, Value: num}

	case token.True:   expr = &node.Bool{Where: tok.Where, Value: true}
	case token.False:  expr = &node.Bool{Where: tok.Where, Value: false}
	case token.String: expr = &node.String{Where: tok.Where, Value: tok.Data}

	default: goerror.Error(p.tok.Where, "Unexpected %v", p.tok)
	}

	p.next()
	return
}

func (p *Parser) parseFuncCall() *node.FuncCall {
	n := &node.FuncCall{Where: p.tok.Where}

	start := p.tok.Where
	p.next()
	n.Name = p.parseId()

	for p.tok.Type != token.RParen {
		if p.tok.Type == token.EOF {
			goerror.Error(p.tok.Where, "Expected matching '%v', got %v", token.RParen, p.tok)
			goerror.Note(start, "Opened here")
			return nil
		}

		n.Args = append(n.Args, p.parseExpr())
	}
	p.next()
	return n
}

var attrsMap = map[token.Type]int{
	token.Inline:    node.AttrInline,
	token.Interrupt: node.AttrInterrupt,
}

func (p *Parser) parseAttrs() (attrs int) {
	p.next()
	for p.tok.Type != token.RSquare {
		attr, ok := attrsMap[p.tok.Type]
		if !ok {
			goerror.Error(p.tok.Where, "Expected an attribute, got %v", p.tok)
		}

		attrs |= attr
		p.next()
	}
	p.next()
	return
}

func (p *Parser) parseFunc() *node.Func {
	n := &node.Func{Where: p.tok.Where}

	if p.next(); p.tok.Type != token.LParen {
		goerror.Error(p.tok.Where, "Expected '%v' to open function head definition, got %v",
		              token.LParen, p.tok)
		p.next()
		return nil
	}

	p.next()
	n.Name = p.parseId()

	start := p.tok.Where
	for p.tok.Type != token.RParen {
		if p.tok.Type == token.EOF {
			goerror.Error(p.tok.Where, "Expected matching '%v', got %v", token.RParen, p.tok)
			goerror.Note(start, "Opened here")
			return nil
		}

		// TODO: Function definition arguments

		p.next()
	}
	p.next()

	if p.tok.Type == token.LSquare {
		n.Attrs = p.parseAttrs()
	}

	if p.tok.Type == token.Arrow {
		p.next()
		n.Type = p.parseId()
	}

 	n.Body = p.parseStmts()
	return n
}

func (p *Parser) next() {
	if p.tok.Type == token.EOF {
		return
	}

	if p.tok = p.l.NextToken(); p.tok.Type == token.Error {
		goerror.Error(p.tok.Where, p.tok.Data)
		os.Exit(1)
	}

	if p.tok.Type == token.EOF {
		p.WhereFileEnd = p.tok.Where
	}
}
