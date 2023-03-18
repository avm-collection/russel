package compiler

import (
	"os"
	"math"

	"github.com/avm-collection/goerror"
	"github.com/avm-collection/agen"

	"github.com/LordOfTrident/russel/internal/parser"
	"github.com/LordOfTrident/russel/internal/node"
	"github.com/LordOfTrident/russel/internal/token"
)

// https://en.wikipedia.org/wiki/Levenshtein_distance
func levenDist(a, b string) int {
	d := make([][]int, len(a) + 1)

	for i := 0; i < len(a) + 1; i ++ {
		d[i] = make([]int, len(b) + 1)
	}

	for i := 1; i < len(a) + 1; i ++ {
		d[i][0] = i
	}

	for i := 1; i < len(b) + 1; i ++ {
		d[0][i] = i
	}

	for i := 1; i < len(a) + 1; i ++ {
		for j := 1; j < len(b) + 1; j ++ {
			subCost := 0
			if a[i - 1] != b[j - 1] {
				subCost = 1
			}

			a := float64(d[i - 1][j] + 1)
			b := float64(d[i][j - 1] + 1)
			c := float64(d[i - 1][j - 1] + subCost)

			d[i][j] = int(math.Min(math.Min(a, b), c))
		}
	}

	return d[len(a)][len(b)]
}

func getMostSimilarName(name string, names []string) string {
	smallest := -1
	which    := ""

	for _, n := range names {
		dist := levenDist(name, n)
		if dist < smallest || smallest == -1 {
			smallest = dist
			which    = n
		}
	}

	return which
}

const MainFuncName = "main"

type Func struct {
	Used  bool
	Addr  agen.Word
	Node *node.Func
}

type Var struct {
	Used bool
	Addr agen.Word
}

type Macro struct {
	Used bool
	Expr node.Expr
}

type Call struct {
	Name string
	Addr agen.Word
}

type Compiler struct {
	p *parser.Parser
	a *agen.AGEN

	funcs  map[string]Func
	vars   map[string]Var
	macros map[string]Macro

	toCompile     []Func
	deferredCalls []Call

	inLoop bool
	breaks, continues []agen.Word
}

func New(input, path string) *Compiler {
	c := &Compiler{
		p: parser.New(input, path),
		a: agen.New(),

		funcs:  make(map[string]Func),
		vars:   make(map[string]Var),
		macros: make(map[string]Macro),
	}

	return c
}

func (c *Compiler) CompileInto(path string, exec bool) error {
	program := c.p.Parse()
	if goerror.Happened() {
		os.Exit(1)
	}

	if c.compile(program); goerror.Happened() {
		os.Exit(1)
	}

	return c.a.CreateExecAVM(path, exec)
}

func (c *Compiler) compile(program *node.Stmts) {
	for _, stmt := range program.List {
		switch s := stmt.(type) {
		case *node.Func:  c.registerFunc(s)
		case *node.Macro: c.registerMacro(s)
		case *node.Let:   c.registerVar(s)

		default: panic("TODO: Unimplemented")
		}
	}

	main, ok := c.funcs[MainFuncName]
	if !ok {
		goerror.SimpleError("Missing entry function '%v'", MainFuncName)
		goerror.NoteSuggestNewCode(c.p.WhereFileEnd, "Suggestion: add", []string{
			"proc (main) -> int {",
			"    # Put your entry code here",
			"",
			"    return -> 0",
			"}",
		})
	}

	c.compileFunc(main)

	for name, func_ := range c.funcs {
		if !func_.Used {
			goerror.Warning(func_.Node.Where, "Unused function '%v'", name)
			continue
		}
	}

	c.a.SetEntryHere()
	c.a.AddInstWith("cal", main.Addr)
	c.a.AddInstWith("psh", 0)
	c.a.AddInst(    "hlt")
}

func (c *Compiler) checkNameExists(where token.Where, name string) bool {
	if prev, ok := c.funcs[name]; ok {
		goerror.Error(where, "Function '%v' redefined", name)
		goerror.Note(prev.Node.Where, "Previously defined here")
		return true
	}

	return false
}

func (c *Compiler) registerFunc(n *node.Func) {
	name := n.Name.Value
	if (c.checkNameExists(n.Where, name)) {
		return
	}

	if name != MainFuncName && n.Type != nil {
		panic("TODO: Return type not allowed for functions except 'main'")
	}

	c.funcs[n.Name.Value] = Func{Node: n}
}

func (c *Compiler) registerMacro(n *node.Macro) {
}

func (c *Compiler) registerVar(n *node.Let) {
}

func (c *Compiler) compileFunc(f Func) {
	if !f.Used {
		f.Used = true;
		c.funcs[f.Node.Name.Value] = f
	}

	if f.Node.Attrs & node.AttrInline != 0 {
		c.compileStmts(f.Node.Body)
	} else {
		f.Addr = c.a.Label()
		c.funcs[f.Node.Name.Value] = f
		c.compileStmts(f.Node.Body)
		c.a.AddInst("ret")
	}

	toCompile := make([]Func, len(c.toCompile))
	copy(toCompile, c.toCompile)
	c.toCompile = c.toCompile[:0]

	deferredCalls := make([]Call, len(c.deferredCalls))
	copy(deferredCalls, c.deferredCalls)
	c.deferredCalls = c.deferredCalls[:0]

	for _, func_ := range toCompile {
		if func_ := c.funcs[func_.Node.Name.Value]; func_.Used {
			continue
		}

		c.compileFunc(func_)
	}

	for _, call := range deferredCalls {
		c.a.GetInstAt(call.Addr).Data = c.funcs[call.Name].Addr
	}
}

func (c *Compiler) compileStmts(n *node.Stmts) {
	for _, stmt := range n.List {
		c.compileStmt(stmt)
	}
}

func (c *Compiler) compileStmt(n node.Stmt) {
	switch s := n.(type) {
	case *node.ExprStmt:  c.compileExpr(s.Expr)
	case *node.Let:       c.compileLet(s)
	case *node.Return:    c.compileReturn(s)
	case *node.If:        c.compileIf(s)
	case *node.While:     c.compileWhile(s)
	case *node.For:       c.compileFor(s)
	case *node.Assign:    c.compileAssign(s)
	case *node.Increment: c.compileIncrement(s)
	case *node.Break:     c.compileBreak(s)
	case *node.Continue:  c.compileContinue(s)

	default: panic("TODO: Unimplemented")
	}
}

func (c *Compiler) compileExpr(n node.Expr) {
	switch e := n.(type) {
	case *node.Int:      c.compileInt(e)
	case *node.Bool:     c.compileBool(e)
	case *node.String:   c.compileString(e)
	case *node.FuncCall: c.compileFuncCall(e)
	case *node.Id:       c.compileId(e)

	default: panic("TODO: Unimplemented")
	}
}

func (c *Compiler) compileInt(n *node.Int) {
	c.a.AddInstWith("psh", agen.Word(n.Value))
}

func (c *Compiler) compileBool(n *node.Bool) {
	if n.Value {
		c.a.AddInstWith("psh", 1)
	} else {
		c.a.AddInstWith("psh", 0)
	}
}

func (c *Compiler) compileString(n *node.String) {
	addr := c.a.AddMemoryString(n.Value)
	c.a.AddInstWith("psh", addr)
	c.a.AddInstWith("psh", agen.Word(len(n.Value)))
}

func (c *Compiler) getFuncNames() (names []string) {
	for name, _ := range c.funcs {
		names = append(names, name)
	}
	return
}

func (c *Compiler) getVarAndMacroNames() (names []string) {
	for name, _ := range c.vars {
		names = append(names, name)
	}

	for name, _ := range c.macros {
		names = append(names, name)
	}
	return
}

func (c *Compiler) compileFuncCall(n *node.FuncCall) {
	name := n.Name.Value

	for _, expr := range n.Args {
		c.compileExpr(expr)
	}

	// TODO: Make an intrinsic system
	if name == "writef" {
		c.a.AddInst("wrf")
		return
	} else if name == "iprint" {
		c.a.AddInst("prt")
		return
	} else if name == "fprint" {
		c.a.AddInst("fpr")
		return
	} else if name == "halt" {
		c.a.AddInst("hlt")
		return
	} else if name == "+" {
		c.a.AddInst("add")
		return
	} else if name == "-" {
		c.a.AddInst("sub")
		return
	} else if name == "*" {
		c.a.AddInst("mul")
		return
	} else if name == "/" {
		c.a.AddInst("div")
		return
	} else if name == "%" {
		c.a.AddInst("mod")
		return
	} else if name == "not" {
		c.a.AddInst("not")
		return
	} else if name == "and" {
		c.a.AddInst("and")
		return
	} else if name == "or" {
		c.a.AddInst("orr")
		return
	} else if name == "==" {
		c.a.AddInst("equ")
		return
	} else if name == "/=" {
		c.a.AddInst("neq")
		return
	} else if name == ">" {
		c.a.AddInst("grt")
		return
	} else if name == ">=" {
		c.a.AddInst("geq")
		return
	} else if name == "<" {
		c.a.AddInst("les")
		return
	} else if name == "<=" {
		c.a.AddInst("leq")
		return
	}

	func_, ok := c.funcs[name]
	if !ok {
		goerror.Error(n.Name.Where, "Unknown function '%v'", name)

		similar := getMostSimilarName(name, c.getFuncNames())
		if len(similar) > 0 {
			goerror.NoteSuggestName(n.Name.Where, similar)
		}
		return
	}

	if func_.Node.Attrs & node.AttrInline != 0 {
		c.compileFunc(func_)
	} else {
		addr := c.a.AddInst("cal")

		if !func_.Used {
			c.toCompile     = append(c.toCompile, func_)
			c.deferredCalls = append(c.deferredCalls, Call{Name: name, Addr: addr})
		} else {
			c.a.GetInstAt(addr).Data = func_.Addr
		}
	}
}

func (c *Compiler) compileId(n *node.Id) {
	if macro, ok := c.macros[n.Value]; ok {
		if !macro.Used {
			macro.Used = true
			c.macros[n.Value] = macro
		}

		c.compileExpr(macro.Expr)
	} else if var_, ok := c.vars[n.Value]; ok {
		if !var_.Used {
			var_.Used = true
			c.vars[n.Value] = var_
		}

		c.compileReadVar(n.Value)
	}
}

func (c *Compiler) compileReadVar(name string) {
	c.a.AddInstWith("psh", c.vars[name].Addr)
	c.a.AddInst(    "r64")
}

func (c *Compiler) compileWriteVar(name string) {
	c.a.AddInstWith("psh", c.vars[name].Addr)
	c.a.AddInstWith("swp", 0)
	c.a.AddInst(    "w64")
}

func (c *Compiler) compileLet(n *node.Let) {
	panic("TODO: Unimplemented")
}

// TODO: Return does not work properly with inlined functions
func (c *Compiler) compileReturn(n *node.Return) {
	if n.Expr != nil {
		panic("TODO: Unimplemented")
	}

	c.a.AddInst("ret")
}

func (c *Compiler) compileIf(n *node.If) {
	/*
		if let x = 5; (== x 5) {
			(println "x is 5")
		} else {
			(println "x is not 5")
		}
	*/

	if n.Var != nil {
		c.compileLet(n.Var)                           //     INIT        # let x = 5
	}

	c.compileExpr(n.Cond)                             //     COND        # (== x 5)
	if !n.Invert {
		c.a.AddInst("not")                            //     not
	}
	                                                  // -------------------- if ... else ...
	if n.Else != nil {
		elseAddr := c.a.AddInst("jnz")                //     jnz else
		c.compileStmts(n.Then)                        //     IF_BODY     # { (println "x is 5") }

		endAddr := c.a.AddInst("jmp")                 //     jmp end
		c.a.GetInstAt(elseAddr).Data = c.a.Label()    // else:
		c.compileStmts(n.Else)                        //     ELSE_BODY   # { (println "x is not 5") }
		c.a.GetInstAt(endAddr).Data = c.a.Label()     // end:
	} else {                                          // -------------------- if ...
		endAddr := c.a.AddInst("jnz")                 //     jnz end
		c.compileStmts(n.Then)                        //     IF_BODY     # { (println "x is 5") }
		c.a.GetInstAt(endAddr).Data = c.a.Label()     // end:
	}
}

func (c *Compiler) startLoop() (isFirst bool) {
	isFirst = !c.inLoop
	if isFirst {
		c.inLoop = true
	}
	return
}

func (c *Compiler) endLoop(isFirst bool, endLabel agen.Word) {
	if isFirst {
		c.inLoop = false
	}

	for _, break_ := range c.breaks {
		c.a.GetInstAt(break_).Data = endLabel
	}
	c.breaks = c.breaks[:0]

	for _, continue_ := range c.continues {
		c.a.GetInstAt(continue_).Data = endLabel
	}
	c.continues = c.continues[:0]
}

func (c *Compiler) compileWhile(n *node.While) {
	/*
		while (< i 10) {
			(println i)
			++ i
		}
	*/

	isFirst := c.startLoop()

	condLabelAddr := c.a.Label()                  // cond:
	c.compileExpr(n.Cond)                         //     COND       # (< i 10)
	if !n.Invert {
		c.a.AddInst("not")                        //     not
	}

	endAddr := c.a.AddInst("jnz")                 //     jnz end
	c.compileStmts(n.Body)                        //     BODY       # { (println i) ++ i }
	c.a.AddInstWith("jmp", condLabelAddr)         //     jmp cond
	endLabelAddr := c.a.Label()
	c.a.GetInstAt(endAddr).Data = endLabelAddr    // end:

	c.endLoop(isFirst, endLabelAddr)
}

func (c *Compiler) compileFor(n *node.For) {
	/*
		for let i = 0; (< i 10); ++ i {
			(println i)
		}
	*/

	isFirst := c.startLoop()

	if n.Var != nil {
		c.compileStmt(n.Var)                      //     INIT       # let i = 0
	}

	skipAddr := c.a.AddInst("jmp")                //     jmp skip
	condLabel := c.a.Label()                      // cond:
	c.compileStmt(n.Last)                         //     LAST       # ++ i
	c.a.GetInstAt(skipAddr).Data = c.a.Label()    // skip:
	c.compileExpr(n.Cond)                         //     COND       # (< i 10)
	c.a.AddInst("not")                            //     not
	endAddr := c.a.AddInst("jnz")                 //     jnz end
	c.compileStmts(n.Body)                        //     BODY       # { (println i) }
	c.a.AddInstWith("jmp", condLabel)             //     jmp cond
	endLabelAddr := c.a.Label()
	c.a.GetInstAt(endAddr).Data = endLabelAddr    // end:

	c.endLoop(isFirst, endLabelAddr)
}

func (c *Compiler) compileAssign(n *node.Assign) {
	c.compileExpr(n.Expr)
	c.compileWriteVar(n.Name.Value)
}

func (c *Compiler) compileIncrement(n *node.Increment) {
	c.compileReadVar(n.Name.Value)

	if n.Negative {
		c.a.AddInst("dec")
	} else {
		c.a.AddInst("inc")
	}

	c.compileWriteVar(n.Name.Value)
}

func (c *Compiler) compileBreak(n *node.Break) {
	if !c.inLoop {
		goerror.Error(n.Where, "'break' outside of a loop")
		return
	}

	c.breaks = append(c.breaks, c.a.AddInst("jmp"))
}

func (c *Compiler) compileContinue(n *node.Continue) {
	if !c.inLoop {
		goerror.Error(n.Where, "'continue' outside of a loop")
		return
	}

	c.continues = append(c.continues, c.a.AddInst("jmp"))
}
