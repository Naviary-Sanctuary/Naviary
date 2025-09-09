package expressions

import "naviary/compiler/token"

// Actual type is determined during type checking
// Example: 42, 100, 1_000_000
type IntegerLiteral struct {
	Token token.Token
	value string
}

func (integer *IntegerLiteral) expressionNode() {}

func (integer *IntegerLiteral) TokenLiteral() string {
	return integer.Token.Literal
}

func (integer *IntegerLiteral) String() string {
	return integer.value
}
