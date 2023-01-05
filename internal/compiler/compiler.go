package compiler

import (
	"os"
	"os/exec"
	"fmt"

	"github.com/LordOfTrident/russel/internal/errors"
	"github.com/LordOfTrident/russel/internal/parser"
	"github.com/LordOfTrident/russel/internal/node"
	"github.com/LordOfTrident/russel/internal/token"
)

const MainFuncName = "main"

type DeclaredFunc interface {
	Compile(c *Compiler, name string)
	GetCompiled() string

	IsCompiled()  bool
	IsIntrinsic() bool
	IsInline()    bool
	Where()       token.Where
}

type RusselFunc struct {
	node *node.Func

	compiled   string
	isCompiled bool

	c *Compiler
}

func (rf *RusselFunc) Compile(c *Compiler, name string) {
	rf.isCompiled = true

	rf.c = c

	if rf.node.Inline {
		rf.compiled += c.compileStatements(rf.node.Body)
	} else {
		rf.compiled += c.genFuncLabel(name)
		rf.compiled += c.compileStatements(rf.node.Body)
		rf.compiled += c.genFuncExit(name)
	}
}

func (rf *RusselFunc) IsInline() bool {
	return rf.node.Inline
}

func (rf *RusselFunc) GetCompiled() string {
	if rf.node.Inline {
		rf.compiled = ""
		rf.Compile(rf.c, "")
	}

	return rf.compiled
}

func (rf *RusselFunc) IsCompiled() bool {
	return rf.isCompiled
}

func (rf *RusselFunc) IsIntrinsic() bool {
	return false
}

func (rf *RusselFunc) Where() token.Where {
	return rf.node.Token.Where
}

type Intrinsic struct {
	body string

	compiled   string
	isCompiled bool

	inline bool
}

func (i *Intrinsic) Compile(c *Compiler, name string) {
	i.isCompiled = true

	if i.inline {
		i.compiled += i.body
	} else {
		i.compiled += c.genFuncLabel(name)
		i.compiled += i.body
		i.compiled += c.genFuncExit(name)
	}
}

func (i *Intrinsic) IsInline() bool {
	return i.inline
}

func (i *Intrinsic) GetCompiled() string {
	return i.compiled
}

func (i *Intrinsic) IsCompiled() bool {
	return i.isCompiled
}

func (i *Intrinsic) IsIntrinsic() bool {
	return true
}

func (i *Intrinsic) Where() token.Where {
	return token.Where{}
}

type Compiler struct {
	p *parser.Parser

	strCount int

	funcs map[string]DeclaredFunc
}

func New(input, path string) *Compiler {
	c := &Compiler{p: parser.New(input, path), funcs: make(map[string]DeclaredFunc)}
	c.addIntrinsics()

	return c
}

func (c *Compiler) genFuncExit(name string) (asm string) {
	if name == MainFuncName {
		asm += c.genInst("psh", "0")
		asm += c.genInst("hlt", "")
	} else {
		asm += c.genInst("ret", "")
	}

	return
}

func (c *Compiler) genInst(inst, format string, args... interface{}) string {
	return fmt.Sprintf("\t" + inst + " " + format + "\n", args...)
}

func (c *Compiler) genFuncLabel(name string) string {
	if name == MainFuncName {
		return ".entry\n"
	} else {
		return fmt.Sprintf(".f_%v\n", name)
	}
}

func (c *Compiler) genCallFunc(name string) string {
	if name == MainFuncName {
		return c.genInst("cal", "entry")
	} else {
		return c.genInst("cal", "f_%v", name)
	}
}

func (c *Compiler) CompileInto(path string, anasm bool) error {
	program := c.p.Parse()
	if errors.Happend() {
		os.Exit(1)
	}

	compiled := c.compile(program)
	if errors.Happend() {
		os.Exit(1)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	f.Write([]byte(compiled))
	f.Close()

	if !anasm {
		if err := c.anasmToExec(path); err != nil {
			return err
		}

		fmt.Printf("Remove '%v'\n", path)
		if err := os.Remove(path); err != nil {
			return err
		}
	}

	return nil
}

func (c *Compiler) anasmToExec(path string) error {
	if !errors.NoOutput() {
		fmt.Println()
	}

	fmt.Printf("[CMD] anasm '%v'\n", path)
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

func (c *Compiler) compile(program *node.Statements) (asm string) {
	for _, statement := range program.List {
		switch s := statement.(type) {
		case *node.Func: c.registerFunc(s)

		default: panic("TODO: Unimplemented")
		}
	}

	main, ok := c.funcs[MainFuncName]
	if !ok {
		errors.Simple("Missing '%v' function", MainFuncName)
	}

	main.Compile(c, MainFuncName)
	asm += main.GetCompiled()

	for name, func_ := range c.funcs {
		if name == MainFuncName {
			continue
		}

		if !func_.IsCompiled() {
			if !func_.IsIntrinsic() {
				errors.Warning(func_.Where(), "Unused function '%v'", name)
			}

			continue
		}

		if !func_.IsInline() {
			asm += "\n" + func_.GetCompiled()
		}
	}

	return
}

func (c *Compiler) registerFunc(f *node.Func)  {
	if _, ok := c.funcs[f.Name.Value]; ok {
		panic("TODO: Function redefinition")
	}

	if f.Type != nil {
		panic("TODO: Return type not allowed for functions except 'main'")
	}

	c.funcs[f.Name.Value] = &RusselFunc{node: f, isCompiled: false}
}

func (c *Compiler) compileStatements(ss *node.Statements) (asm string) {
	for _, statement := range ss.List {
		asm += c.compileStatement(statement)
	}

	return
}

func (c *Compiler) compileStatement(statement node.Statement) string {
	switch s := statement.(type) {
	case *node.ExprStatement: return c.compileExpr(s.Expr)
	case *node.Return:        return c.compileReturn(s)

	default: panic("TODO: Unimplemented")
	}
}

func (c *Compiler) compileExpr(expr node.Expr) string {
	switch e := expr.(type) {
	case *node.Int:      return c.compileInt(e)
	case *node.Bool:     return c.compileBool(e)
	case *node.String:   return c.compileString(e)
	case *node.FuncCall: return c.compileFuncCall(e)

	default: panic("TODO: Unimplemented")
	}
}

func (c *Compiler) compileInt(int_ *node.Int) string {
	return c.genInst("psh", "%v", int_.Value)
}

func (c *Compiler) compileBool(bool_ *node.Bool) string {
	if bool_.Value {
		return c.genInst("psh", "1")
	} else {
		return c.genInst("psh", "0")
	}
}

func (c *Compiler) compileString(str *node.String) (asm string) {
	strId := fmt.Sprintf("STR%v", c.strCount)
	c.strCount ++

	asm += fmt.Sprintf("\tlet %v char = ", strId)
	first := true
	for i, ch := range str.Value {
		if first {
			first = false
		} else {
			asm += ", "
		}

		if i % 16 == 0 {
			asm += "\n\t\t"
		}

		asm += fmt.Sprintf("%v", int(ch))
	}
	asm += "\n"

	asm += c.genInst("psh", strId)
	asm += c.genInst("psh", "(sizeof %v)", strId)

	return
}


func (c *Compiler) compileFuncCall(f *node.FuncCall) (asm string) {
	name := f.Name.Value

	func_, ok := c.funcs[name]
	if !ok {
		panic("TODO: Function undefined")
	}

	if len(f.Args) != 0 && !func_.IsIntrinsic() {
		panic("TODO: Only intrinsics can take arguments")
	}

	if !func_.IsCompiled() {
		func_.Compile(c, name)
	}

	for _, arg := range f.Args {
		asm += c.compileExpr(arg)
	}

	if func_.IsInline() {
		asm += func_.GetCompiled()
	} else {
		asm += c.genCallFunc(name)
	}

	return
}

func (c *Compiler) compileReturn(r *node.Return) (asm string) {
	panic("TODO: Return not implemented yet")

	return
}
