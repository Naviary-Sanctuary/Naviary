package ast

import "naviary/compiler/token"

// Example: true, false
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (boolean *BooleanLiteral) expressionNode() {}

func (boolean *BooleanLiteral) TokenLiteral() string {
	return boolean.Token.Literal
}

func (boolean *BooleanLiteral) String() string {
	return boolean.Token.Literal
}
