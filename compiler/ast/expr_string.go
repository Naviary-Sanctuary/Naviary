package ast

import "naviary/compiler/token"

// Example: "hello", "world"
type StringLiteral struct {
	Token token.Token
	Value string
}

func (str *StringLiteral) expressionNode() {}

func (str *StringLiteral) TokenLiteral() string {
	return str.Token.Literal
}

func (str *StringLiteral) String() string {
	return `"` + str.Value + `"`
}
