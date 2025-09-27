package parser

import "compiler/lexer"

// Program is the root node of the AST
type Program struct {
	Functions []FunctionDeclaration
}

// FunctionDeclaration represents a function definition
type FunctionDeclaration struct {
	Token      lexer.Token // The 'func' token
	Name       string
	Parameters []string // For MVP, we only support main() with no parameters
	Body       BlockStatement
}

func (functionDeclaration *FunctionDeclaration) String() string {
	return functionDeclaration.Token.Value
}
