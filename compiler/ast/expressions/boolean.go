package expressions

import "naviary/compiler/token"

// Example: true, false
type BooleanLiteral struct {
	Token token.Token
	value bool
}

func (boolean *BooleanLiteral) expressionNode() {}

func (boolean *BooleanLiteral) TokenLiteral() string {
	return boolean.Token.Literal
}

func (boolean *BooleanLiteral) String() string {
	return boolean.Token.Literal
}
