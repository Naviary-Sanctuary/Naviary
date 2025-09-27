package parser

import "compiler/lexer"

// Identifier represents a variable name
type Identifier struct {
	Token lexer.Token // The identifier token
	Value string
}

func (identifier *Identifier) String() string {
	return identifier.Token.Value
}
func (identifier *Identifier) expressionNode() {}
