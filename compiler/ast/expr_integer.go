package ast

import "compiler/token"

// Actual type is determined during type checking. For now, we just store the string value.
// Example: 42, 100, 1_000, 1_000_000
type IntegerLiteral struct {
	Token token.Token
	Value string
}

func (integer *IntegerLiteral) expressionNode() {}

func (integer *IntegerLiteral) TokenLiteral() string {
	return integer.Token.Value
}

func (integer *IntegerLiteral) String() string {
	return integer.Token.Value
}
