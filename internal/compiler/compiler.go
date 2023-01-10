package compiler

import (
	"os"
	"os/exec"
	"fmt"
	"math"

	"github.com/LordOfTrident/russel/internal/errors"
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

const MainFuncName = "main"

type Func struct {
	node *node.Func

	compiled bool
	asm      string
}

func NewFunc(node *node.Func) *Func {
	return &Func{node: node}
}

func (f *Func) compile(c *Compiler) {
	f.compiled = true
	f.asm      = c.compileFunc(f.node)
}

type Var struct {
	node *node.Let

	used bool
}

func NewVar(node *node.Let) *Var {
	return &Var{node: node}
}

type ids struct {
	prefix   string
	strCount int
	labelCount, labelNest int

	currentLoopLabel string
}

type Compiler struct {
	p *parser.Parser

	funcs      map[string]*Func
	intrinsics map[string]string

	macros map[string]*node.Macro

	vars map[string]*Var

	ids ids
}

func New(input, path string) *Compiler {
	c := &Compiler{
		p: parser.New(input, path),

		funcs:      make(map[string]*Func),
		intrinsics: make(map[string]string),

		macros: make(map[string]*node.Macro),

		vars: make(map[string]*Var),
	}

	c.addIntrinsics()

	return c
}

func (c *Compiler) addIntrinsics() {
	c.intrinsics["writef"] = c.genInst("wrf", "") + c.genInst("pop", "")
	c.intrinsics["iprint"] = c.genInst("prt", "")
	c.intrinsics["fprint"] = c.genInst("fpr", "")

	c.intrinsics["+"] = c.genInst("add", "")
	c.intrinsics["-"] = c.genInst("sub", "")
	c.intrinsics["*"] = c.genInst("mul", "")
	c.intrinsics["/"] = c.genInst("div", "")

	c.intrinsics["=="] = c.genInst("equ", "")
	c.intrinsics["/="] = c.genInst("neq", "")
	c.intrinsics[">"]  = c.genInst("grt", "")
	c.intrinsics[">="] = c.genInst("geq", "")
	c.intrinsics["<"]  = c.genInst("les", "")
	c.intrinsics["<="] = c.genInst("leq", "")

	c.intrinsics["not"] = c.genInst("not", "")
	c.intrinsics["and"] = c.genInst("and", "")
	c.intrinsics["or"]  = c.genInst("orr", "")

	c.intrinsics["exit"] = c.genInst("hlt", "")
}

func (c *Compiler) compileFunc(f *node.Func) (asm string) {
	if f.Inline {
		asm += c.compileStatements(f.Body)
	} else {
		prev := c.ids
		c.ids = ids{prefix: f.Name.Value}

		asm += fmt.Sprintf(".f_%v\n", f.Name.Value)
		asm += c.compileStatements(f.Body)
		asm += c.genInst("ret", "")

		c.ids = prev
	}

	return
}

func (c *Compiler) genInst(inst, format string, args... interface{}) string {
	return fmt.Sprintf("\t" + inst + " " + format + "\n", args...)
}

func (c *Compiler) genFuncCall(name string) string {
	return c.genInst("cal", "f_%v", name)
}

func (c *Compiler) genLabel(name string) string {
	return fmt.Sprintf(".%v\n", name)
}

func (c *Compiler) genVarRead(name string) (asm string) {
	asm += c.genInst("psh", "v_%v", name)
	asm += c.genInst("r64", "")

	return
}

func (c *Compiler) genVarWrite(name string) (asm string) {
	asm += c.genInst("psh", "v_%v", name)
	asm += c.genInst("swp", "0")
	asm += c.genInst("w64", "")

	return
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
		case *node.Func:  c.registerFunc(s)
		case *node.Macro: c.registerMacro(s)
		case *node.Let:   c.registerVar(s)

		default: panic("TODO: Unimplemented")
		}
	}

	main, ok := c.funcs[MainFuncName]
	if !ok {
		errors.Simple("Missing entry function '%v'", MainFuncName)
		errors.NoteSuggestNewCode(c.p.WhereFileEnd, "Suggestion: add", []string{
			"fun (main) -> int {",
			"    # Put your entry code here",
			"",
			"    return 0",
			"}",
		})
		return
	}

	main.compile(c)
	asm += main.asm

	for name, func_ := range c.funcs {
		if name == MainFuncName {
			continue
		}

		if !func_.compiled {
			errors.Warning(func_.node.Token.Where, "Unused function '%v'", name)
			continue
		}

		if !func_.node.Inline {
			asm += "\n" + func_.asm
		}
	}

	init := c.genLabel("entry")
	for name, var_ := range c.vars {
		if !var_.used {
			errors.Warning(var_.node.Token.Where, "Unused variable '%v'", name)
			continue
		}

		// TODO: initialize this using const expressions,
		//       do not allow non-const expressions like now
		init += fmt.Sprintf("let v_%v i64 = 0\n", name)
		if var_.node.Expr != nil {
			init += c.compileExpr(var_.node.Expr)
			init += c.genVarWrite(name)
		}
	}

	init += c.genFuncCall("main")
	init += c.genInst("psh", "0")
	init += c.genInst("hlt", "")
	init += "\n"

	asm = init + asm

	return
}

func (c *Compiler) nameExists(where token.Where, name string) bool {
	if prev, ok := c.macros[name]; ok {
		errors.Error(where, "Macro '%v' redefined", name)
		errors.Note(prev.Token.Where, "Previously defined here")
		return true
	}

	if _, ok := c.intrinsics[name]; ok {
		errors.Error(where, "Intrinsic '%v' redefined", name)
		return true
	}

	if prev, ok := c.funcs[name]; ok {
		errors.Error(where, "Function '%v' redefined", name)
		errors.Note(prev.node.Token.Where, "Previously defined here")
		return true
	}

	if prev, ok := c.vars[name]; ok {
		errors.Error(where, "Variable '%v' redefined", name)
		errors.Note(prev.node.Token.Where, "Previously defined here")
		return true
	}

	return false
}

func (c *Compiler) registerVar(l *node.Let) {
	name := l.Name.Value
	if c.nameExists(l.Token.Where, name) {
		return
	}

	c.vars[name] = NewVar(l)
}

func (c *Compiler) registerFunc(f *node.Func) {
	name := f.Name.Value
	if c.nameExists(f.Token.Where, name) {
		return
	}

	if f.Type != nil {
		panic("TODO: Return type not allowed for functions except 'main'")
	}

	c.funcs[name] = NewFunc(f)
}

func (c *Compiler) registerMacro(m *node.Macro) {
	name := m.Name.Value
	if c.nameExists(m.Token.Where, name) {
		return
	}

	c.macros[name] = m
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
	case *node.If:            return c.compileIf(s)
	case *node.While:         return c.compileWhile(s)
	case *node.For:           return c.compileFor(s)
	case *node.Assign:        return c.compileAssign(s)
	case *node.Increment:     return c.compileIncrement(s)
	case *node.Break:         return c.compileBreak(s)
	case *node.Continue:      return c.compileContinue(s)

	default: panic("TODO: Unimplemented")
	}
}

func (c *Compiler) compileExpr(expr node.Expr) string {
	switch e := expr.(type) {
	case *node.Int:      return c.compileInt(e)
	case *node.Bool:     return c.compileBool(e)
	case *node.String:   return c.compileString(e)
	case *node.FuncCall: return c.compileFuncCall(e)
	case *node.Id:       return c.compileId(e)

	default: panic("TODO: Unimplemented")
	}
}

func (c *Compiler) compileId(id *node.Id) string {
	if macro, ok := c.macros[id.Value]; ok {
		return c.compileExpr(macro.Expr)
	} else if var_, ok := c.vars[id.Value]; ok {
		var_.used = true
		c.vars[id.Value] = var_

		return c.genVarRead(id.Value)
	}

	errors.Error(id.Token.Where, "Unknown macro/variable '%v'", id.Value)

	names := []string{}
	for _, macro := range c.macros {
		names = append(names, macro.Name.Value)
	}

	similar := c.getMostSimilarName(id.Value, names)
	if len(similar) > 0 {
		errors.NoteSuggestName(id.Token.Where, similar)
	}

	return ""
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
	name := fmt.Sprintf("s_%v_%v", c.ids.prefix, c.ids.strCount)
	c.ids.strCount ++

	asm += fmt.Sprintf("\tlet %v char = ", name)
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

	asm += c.genInst("psh", name)
	asm += c.genInst("psh", "(sizeof %v)", name)

	return
}

func (c *Compiler) getMostSimilarName(name string, names []string) string {
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

func (c *Compiler) compileAssign(a *node.Assign) (asm string) {
	asm += c.compileExpr(a.Expr)
	asm += c.genVarWrite(a.Name.Value)

	return
}

func (c *Compiler) compileIncrement(i *node.Increment) (asm string) {
	asm += c.genVarRead(i.Name.Value)

	if (i.Decrement) {
		asm += c.genInst("dec", "")
	} else {
		asm += c.genInst("inc", "")
	}

	asm += c.genVarWrite(i.Name.Value)

	return
}

func (c *Compiler) compileIf(i *node.If) (asm string) {
	name := fmt.Sprintf("l_%v_%v_%v", c.ids.prefix, c.ids.labelCount, c.ids.labelNest)
	c.ids.labelNest ++

	labelElse := name + "_else"
	labelEnd  := name + "_end_if"

	asm += c.compileExpr(i.Cond)
	if (!i.Invert) {
		asm += c.genInst("not", "")
	}

	if i.Else != nil {
		asm += c.genInst("jnz", labelElse)
	} else {
		asm += c.genInst("jnz", labelEnd)
	}

	asm += c.compileStatements(i.Then)
	if i.Else != nil {
		asm += c.genInst("jmp", labelEnd)
		asm += c.genLabel(labelElse)
		asm += c.compileStatements(i.Else)
	}

	asm += c.genLabel(labelEnd)

	c.ids.labelNest  --
	c.ids.labelCount ++

	return
}

func (c *Compiler) compileWhile(w *node.While) (asm string) {
	name := fmt.Sprintf("l_%v_%v_%v", c.ids.prefix, c.ids.labelCount, c.ids.labelNest)
	c.ids.labelNest ++

	prev := c.ids.currentLoopLabel
	c.ids.currentLoopLabel = name

	labelCond := name + "_loop"
	labelEnd  := name + "_end_loop"

	asm += c.genLabel(labelCond)
	asm += c.compileExpr(w.Cond)
	if (!w.Invert) {
		asm += c.genInst("not", "")
	}

	asm += c.genInst("jnz", labelEnd)
	asm += c.compileStatements(w.Body)
	asm += c.genInst("jmp", labelCond)
	asm += c.genLabel(labelEnd)

	c.ids.labelNest  --
	c.ids.labelCount ++

	c.ids.currentLoopLabel = prev

	return
}

func (c *Compiler) compileFor(f *node.For) (asm string) {
	name := fmt.Sprintf("l_%v_%v_%v", c.ids.prefix, c.ids.labelCount, c.ids.labelNest)
	c.ids.labelNest ++

	prev := c.ids.currentLoopLabel
	c.ids.currentLoopLabel = name

	labelCond := name + "_loop"
	labelSkip := name + "_loop_skip"
	labelEnd  := name + "_end_loop"

	if f.Init != nil {
		asm += c.compileStatement(f.Init)
	}

	asm += c.genInst("jmp", labelSkip)
	asm += c.genLabel(labelCond)

	if f.Last != nil {
		asm += c.compileStatement(f.Last)
	}

	asm += c.genLabel(labelSkip)

	if f.Cond != nil {
		asm += c.compileExpr(f.Cond)
	} else {
		asm += c.genInst("psh", "1")
	}

	asm += c.genInst("not", "")
	asm += c.genInst("jnz", labelEnd)
	asm += c.compileStatements(f.Body)
	asm += c.genInst("jmp", labelCond)
	asm += c.genLabel(labelEnd)

	c.ids.labelNest  --
	c.ids.labelCount ++

	c.ids.currentLoopLabel = prev

	return
}

func (c *Compiler) compileFuncCall(f *node.FuncCall) (asm string) {
	name := f.Name.Value

	if intrinsic, ok := c.intrinsics[name]; ok {
		for _, arg := range f.Args {
			asm += c.compileExpr(arg)
		}

		asm += intrinsic

		return
	}

	func_, ok := c.funcs[name]
	if !ok {
		errors.Error(f.Name.Token.Where, "Unknown function '%v'", name)

		names := []string{}
		for _, func_ := range c.funcs {
			names = append(names, func_.node.Name.Value)
		}

		similar := c.getMostSimilarName(name, names)
		if len(similar) > 0 {
			errors.NoteSuggestName(f.Name.Token.Where, similar)
		}

		return
	}

	if !func_.compiled || func_.node.Inline {
		func_.compile(c)
	}

	for _, arg := range f.Args {
		asm += c.compileExpr(arg)
	}

	if func_.node.Inline {
		asm += func_.asm
	} else {
		asm += c.genFuncCall(name)
	}

	return
}

func (c *Compiler) compileBreak(b *node.Break) (asm string) {
	if len(c.ids.currentLoopLabel) == 0 {
		errors.Error(b.Token.Where, "Unexpected 'break' outside of a loop")
		return
	}

	asm += c.genInst("jmp", "%v_end_loop", c.ids.currentLoopLabel)

	return
}

func (c *Compiler) compileContinue(b *node.Continue) (asm string) {
	if len(c.ids.currentLoopLabel) == 0 {
		errors.Error(b.Token.Where, "Unexpected 'continue' outside of a loop")
		return
	}

	asm += c.genInst("jmp", "%v_loop", c.ids.currentLoopLabel)

	return
}

func (c *Compiler) compileReturn(r *node.Return) (asm string) {
	panic("TODO: Return not implemented yet")

	return
}
