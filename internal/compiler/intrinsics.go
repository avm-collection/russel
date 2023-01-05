package compiler

func (c *Compiler) addIntrinsics() {
	c.funcs["writef"] = &Intrinsic{body: c.genIntrinsicWrf(), inline: true}

	c.funcs["+"] = &Intrinsic{body: c.genIntrinsicAdd(), inline: true}
	c.funcs["-"] = &Intrinsic{body: c.genIntrinsicSub(), inline: true}
	c.funcs["*"] = &Intrinsic{body: c.genIntrinsicMul(), inline: true}
	c.funcs["/"] = &Intrinsic{body: c.genIntrinsicDiv(), inline: true}

	c.funcs["print-int"]  = &Intrinsic{body: c.genIntrinsicPrt(), inline: true}
}

func (c *Compiler) genIntrinsicWrf() (asm string) {
	asm += c.genInst("psh", "1")
	asm += c.genInst("wrf", "")

	return
}

func (c *Compiler) genIntrinsicAdd() string {
	return c.genInst("add", "")
}

func (c *Compiler) genIntrinsicSub() string {
	return c.genInst("sub", "")
}

func (c *Compiler) genIntrinsicMul() string {
	return c.genInst("mul", "")
}

func (c *Compiler) genIntrinsicDiv() string {
	return c.genInst("div", "")
}

func (c *Compiler) genIntrinsicPrt() string {
	return c.genInst("prt", "")
}
