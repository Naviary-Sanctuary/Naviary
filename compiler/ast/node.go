package ast

type Node interface {
	TokenLiteral() string
	String() string // for debugging
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}
