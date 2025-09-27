package parser

import "compiler/lexer"

// IntegerLiteral represents an integer literal
type IntegerLiteral struct {
	Token lexer.Token // The number token
	Value int64
}

func (numberLiteral *IntegerLiteral) String() string {
	return numberLiteral.Token.Value
}
func (numberLiteral *IntegerLiteral) expressionNode() {}
