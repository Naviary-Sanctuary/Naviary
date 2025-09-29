package ast

import (
	"bytes"
	"compiler/token"
)

type LetStatement struct {
	Token   token.Token
	Name    *Identifier
	Value   Expression
	Mutable bool
}

func (let *LetStatement) statementNode() {}

func (let *LetStatement) TokenLiteral() string {
	return let.Token.Value
}

func (let *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(let.TokenLiteral() + " ")
	if let.Mutable {
		out.WriteString("mut ")
	}
	out.WriteString(let.Name.String())
	out.WriteString(" = ")

	if let.Value != nil {
		out.WriteString(let.Value.String())
	}

	return out.String()
}
