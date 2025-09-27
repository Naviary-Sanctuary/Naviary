package parser

import "compiler/lexer"

// BinaryExpression represents operations like 'a + b'
type BinaryExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (binaryExpression *BinaryExpression) String() string {
	return binaryExpression.Token.Value
}
func (binaryExpression *BinaryExpression) expressionNode() {}
