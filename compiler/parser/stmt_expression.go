package parser

import "compiler/lexer"

// ExpressionStatement wraps an expression used as a statement
type ExpressionStatement struct {
	Token      lexer.Token
	Expression Expression
}

func (expressionStatement *ExpressionStatement) String() string {
	return expressionStatement.Token.Value
}
func (expressionStatement *ExpressionStatement) statementNode() {}
