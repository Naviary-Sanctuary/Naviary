package parser

import "compiler/lexer"

// LetStatement represents 'let x = expression'
type LetStatement struct {
	Token lexer.Token
	Name  string
	Value Expression
}

func (letStatement *LetStatement) String() string {
	return letStatement.Token.Value
}
func (letStatement *LetStatement) statementNode() {}
