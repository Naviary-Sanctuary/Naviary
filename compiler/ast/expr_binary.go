package ast

import (
	"bytes"
	"naviary/compiler/token"
)

type BinaryExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (binary *BinaryExpression) expressionNode() {}

func (binary *BinaryExpression) TokenLiteral() string {
	return binary.Token.Literal
}

func (binary *BinaryExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(binary.Left.String())
	out.WriteString(" " + binary.Operator + " ")
	out.WriteString(binary.Right.String())
	out.WriteString(")")

	return out.String()
}
