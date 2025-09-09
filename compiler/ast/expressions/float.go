package expressions

import "naviary/compiler/token"

// Example: 42.00, 100.00, 1_000.00
type FloatLiteral struct {
	Token token.Token
	value string
}

func (float *FloatLiteral) expressionNode() {}

func (float *FloatLiteral) TokenLiteral() string {
	return float.Token.Literal
}

func (float *FloatLiteral) String() string {
	return float.value
}
