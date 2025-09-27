package parser

import "compiler/lexer"

// CallExpression represents a function call like 'print(x)'
type CallExpression struct {
	Token     lexer.Token
	Function  string // For MVP, just the function name as string
	Arguments []Expression
}

func (callExpression *CallExpression) String() string {
	return callExpression.Token.Value
}
func (callExpression *CallExpression) expressionNode() {}
