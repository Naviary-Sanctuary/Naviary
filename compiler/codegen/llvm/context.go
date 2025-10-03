package llvm

import "tinygo.org/x/go-llvm"

type Context struct {
	context llvm.Context
}

func NewContext() *Context {
	return &Context{
		context: llvm.NewContext(),
	}
}

func (ctx *Context) Dispose() {
	ctx.context.Dispose()
}

func (ctx *Context) GetRawContext() llvm.Context {
	return ctx.context
}
