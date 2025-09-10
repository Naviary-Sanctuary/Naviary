package ast

import (
	"bytes"
	"naviary/compiler/token"
	"strings"
)

type CallExpression struct {
	Token     token.Token
	Function  Expression
	Arguments []Expression
}

func (call *CallExpression) expressionNode() {}

func (call *CallExpression) TokenLiteral() string {
	return call.Token.Literal
}

func (call *CallExpression) String() string {
	var out bytes.Buffer

	// Function name or expression
	out.WriteString(call.Function.String())
	out.WriteString("(")

	// Join arguments with comma
	args := []string{}
	for _, arg := range call.Arguments {
		args = append(args, arg.String())
	}
	out.WriteString(strings.Join(args, ", "))

	out.WriteString(")")

	return out.String()
}
