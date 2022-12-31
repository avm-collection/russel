package compiler

import (
	"os"
	"os/exec"
	"fmt"

	"github.com/LordOfTrident/russel/internal/errors"
	"github.com/LordOfTrident/russel/internal/parser"
	"github.com/LordOfTrident/russel/internal/node"
)

type Compiler struct {
	p *parser.Parser

	strCount int

	out string
}

func New(input, path string) *Compiler {
	return &Compiler{p: parser.New(input, path)}
}

func (c *Compiler) CompileInto(path string) error {
	program := c.p.Parse()
	if errors.Happend() {
		os.Exit(1)
	}

	c.compile(program)
	if errors.Happend() {
		os.Exit(1)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	f.Write([]byte(c.out))
	f.Close()

	fmt.Println("[CMD] anasm " + path)
	cmd := exec.Command("anasm", path)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin  = os.Stdin

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func (c *Compiler) compile(program *node.Statements) {
	hasEntry := false
	for _, statement := range program.List {
		switch s := statement.(type) {
		case *node.Func:
			isEntry := c.compileFunc(s)
			if !hasEntry && isEntry {
				hasEntry = true
			}
		}
	}

	if !hasEntry {
		errors.Simple("Missing 'main' function")
	}
}

func (c *Compiler) compileFunc(f *node.Func) bool {
	if f.Name.Value != "main" {
		panic("Only main function supported currently")
	}

	if f.Type == nil || f.Type.Value != "int" {
		panic("Main has to return an int")
	}

	c.out += ".entry\n"
	c.compileStatements(f.Body)

	return true
}

func (c *Compiler) compileStatements(ss *node.Statements) {
	for _, statement := range ss.List {
		c.compileStatement(statement)
	}
}

func (c *Compiler) compileStatement(statement node.Statement) {
	switch s := statement.(type) {
	case *node.ExprStatement: c.compileExpr(s.Expr)
	case *node.Return:        c.compileReturn(s)
	}
}

func (c *Compiler) compileExpr(expr node.Expr) {
	switch e := expr.(type) {
	case *node.String:   c.compileString(e)
	case *node.FuncCall: c.compileFuncCall(e)
	}
}

func (c *Compiler) compileString(str *node.String) {
	strId := fmt.Sprintf("STR%v", c.strCount)
	c.strCount ++

	c.out += fmt.Sprintf("\tlet %v char = ", strId)
	first := true
	for i, ch := range str.Value {
		if first {
			first = false
		} else {
			c.out += ", "
		}

		if i % 16 == 0 {
			c.out += "\n\t\t"
		}

		c.out += fmt.Sprintf("%v", int(ch))
	}
	c.out += "\n"

	c.genInst("psh", strId)
	c.genInst("psh", "(sizeof %v)", strId)
}

func (c *Compiler) genInst(inst, format string, args... interface{}) {
	c.out += fmt.Sprintf("\t" + inst + " " + format + "\n", args...)
}

func (c *Compiler) compileFuncCall(f *node.FuncCall) {
	if f.Name.Value != "writef" {
		panic("Only 'writef' intrinsic implemented currently")
	}

	if len(f.Args) != 1 {
		panic("Only one string argument allowed currently")
	}

	for _, arg := range f.Args {
		c.compileExpr(arg)
	}

	c.genInst("psh", "1")
	c.genInst("wrf", "")
}

func (c *Compiler) compileReturn(r *node.Return) {
	panic("Return not implemented yet")
}
