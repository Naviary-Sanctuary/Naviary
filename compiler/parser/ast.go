package parser

import "compiler/lexer"

// Node is the base interface for all AST nodes
type Node interface {
	TokenLiteral() string
}

// Statement is a node that doesn't produce a value
type Statement interface {
	Node
	statementNode()
}

// Expression is a node that produces a value
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of the AST
type Program struct {
	Functions []FunctionDeclaration
}

func (program *Program) TokenLiteral() string {
	if len(program.Functions) > 0 {
		return program.Functions[0].TokenLiteral()
	}
	return ""
}

// FunctionDeclaration represents a function definition
type FunctionDeclaration struct {
	Token      lexer.Token // The 'func' token
	Name       string
	Parameters []string // For MVP, we only support main() with no parameters
	Body       BlockStatement
}

func (functionDeclaration *FunctionDeclaration) TokenLiteral() string {
	return functionDeclaration.Token.Literal
}

// BlockStatement represents a block of statements { ... }
type BlockStatement struct {
	Token      lexer.Token // The '{' token
	Statements []Statement
}

func (blockStatement *BlockStatement) TokenLiteral() string {
	return blockStatement.Token.Literal
}
func (blockStatement *BlockStatement) statementNode() {}

// LetStatement represents 'let x = expression'
type LetStatement struct {
	Token lexer.Token
	Name  string
	Value Expression
}

func (letStatement *LetStatement) TokenLiteral() string {
	return letStatement.Token.Literal
}
func (letStatement *LetStatement) statementNode() {}

// ExpressionStatement wraps an expression used as a statement
type ExpressionStatement struct {
	Token      lexer.Token
	Expression Expression
}

func (expressionStatement *ExpressionStatement) TokenLiteral() string {
	return expressionStatement.Token.Literal
}
func (expressionStatement *ExpressionStatement) statementNode() {}

// NumberLiteral represents an integer literal
type NumberLiteral struct {
	Token lexer.Token // The number token
	Value int64
}

func (numberLiteral *NumberLiteral) TokenLiteral() string {
	return numberLiteral.Token.Literal
}
func (numberLiteral *NumberLiteral) expressionNode() {}

// Identifier represents a variable name
type Identifier struct {
	Token lexer.Token // The identifier token
	Value string
}

func (identifier *Identifier) TokenLiteral() string {
	return identifier.Token.Literal
}
func (identifier *Identifier) expressionNode() {}

// BinaryExpression represents operations like 'a + b'
type BinaryExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (binaryExpression *BinaryExpression) TokenLiteral() string {
	return binaryExpression.Token.Literal
}
func (binaryExpression *BinaryExpression) expressionNode() {}

// CallExpression represents a function call like 'print(x)'
type CallExpression struct {
	Token     lexer.Token
	Function  string // For MVP, just the function name as string
	Arguments []Expression
}

func (callExpression *CallExpression) TokenLiteral() string {
	return callExpression.Token.Literal
}
func (callExpression *CallExpression) expressionNode() {}
