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

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	// TODO: Will implement later
	return ""
}
