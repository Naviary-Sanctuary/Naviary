package ast

import (
	"bytes"
	"compiler/token"
)

type BinaryExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (binary *BinaryExpression) expressionNode() {}

func (binary *BinaryExpression) TokenLiteral() string {
	return binary.Token.Value
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
