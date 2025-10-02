package ast

import "compiler/token"

type StringLiteral struct {
	Token token.Token
	Value string
}

func (s *StringLiteral) expressionNode() {}

func (s *StringLiteral) TokenLiteral() string {
	return s.Token.Value
}

func (s *StringLiteral) String() string {
	return "\"" + s.Value + "\""
}
