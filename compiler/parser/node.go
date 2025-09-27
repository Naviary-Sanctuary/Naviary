package parser

import "compiler/lexer"

// Node is the base interface for all AST nodes
type Node interface {
	String() string
}

// Expression is a node that produces a value
type Expression interface {
	Node
	expressionNode()
}

// Statement is a node that doesn't produce a value
type Statement interface {
	Node
	statementNode()
}

// BlockStatement represents a block of statements { ... }
type BlockStatement struct {
	Token      lexer.Token // The '{' token
	Statements []Statement
}

func (blockStatement *BlockStatement) String() string {
	return blockStatement.Token.Value
}
func (blockStatement *BlockStatement) statementNode() {}
