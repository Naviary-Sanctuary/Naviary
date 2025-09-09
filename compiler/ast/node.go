package ast

type Node interface {
	TokenLiteral() string // token literal
	String() string       // for debugging
}

// Statements do not produce a value
type Statement interface {
	Node
	statementNode()
}

// Expressions produce a value
type Expression interface {
	Node
	expressionNode()
}
