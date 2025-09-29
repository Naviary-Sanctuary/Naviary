package ast

import (
	"bytes"
	"compiler/token"
	"strings"
)

type FunctionStatement struct {
	Token      token.Token
	Name       *Identifier
	Parameters []*FunctionParameter
	ReturnType *TypeAnnotation
	Body       *BlockStatement
}

type FunctionParameter struct {
	Name *Identifier
	Type TypeAnnotation
}

type TypeAnnotation struct {
	Token token.Token
	Value string
}

func (function *FunctionStatement) statementNode() {}

func (function *FunctionStatement) TokenLiteral() string {
	return function.Token.Value
}

func (function *FunctionStatement) String() string {
	var out bytes.Buffer

	out.WriteString("func ")
	out.WriteString(function.Name.String())
	out.WriteString("(")

	// Join parameters with comma
	params := []string{}
	for _, param := range function.Parameters {
		params = append(params, param.Name.String()+": "+param.Type.Value)
	}
	out.WriteString(strings.Join(params, ", "))

	out.WriteString(")")

	// Add return type if exists
	if function.ReturnType != nil {
		out.WriteString(" -> ")
		out.WriteString(function.ReturnType.Value)
	}

	out.WriteString(" ")
	out.WriteString(function.Body.String())

	return out.String()
}
