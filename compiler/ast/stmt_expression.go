package ast

import (
	"compiler/token"
)

// Example: print(42), x + 5
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (expression *ExpressionStatement) statementNode() {}

func (expression *ExpressionStatement) TokenLiteral() string {
	return expression.Token.Value
}

func (expression *ExpressionStatement) String() string {
	if expression.Expression != nil {
		return expression.Expression.String()
	}
	return ""
}
