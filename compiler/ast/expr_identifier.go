package ast

import "compiler/token"

// Identifier represents a variable name
type Identifier struct {
	Token token.Token
	Value string
}

func (identifier *Identifier) expressionNode() {}

func (identifier *Identifier) String() string {
	return identifier.Token.Value
}

func (identifier *Identifier) TokenLiteral() string {
	return identifier.Value
}
