package statements

import (
	ast "naviary/compiler/ast"
	token "naviary/compiler/token"
)

// Example: print(42), x + 5
type ExpressionStatement struct {
	Token      token.Token
	Expression ast.Expression
}

func (expression *ExpressionStatement) statementNode() {}

func (expression *ExpressionStatement) TokenLiteral() string {
	return expression.Token.Literal
}

func (expression *ExpressionStatement) String() string {
	if expression.Expression != nil {
		return expression.Expression.String()
	}
	return ""
}
