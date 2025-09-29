package ast

import (
	"bytes"
	"compiler/token"
)

type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (r *ReturnStatement) statementNode() {}

func (r *ReturnStatement) TokenLiteral() string {
	return r.Token.Value
}

func (r *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(r.TokenLiteral())

	if r.ReturnValue != nil {
		out.WriteString(" ")
		out.WriteString(r.ReturnValue.String())
	}

	return out.String()
}
