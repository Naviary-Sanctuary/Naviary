package ast

import (
	"bytes"
	token "naviary/compiler/token"
)

type AssignmentStatement struct {
	Token    token.Token
	Name     *Identifier
	Value    Expression
	Operator string
}

func (assignment *AssignmentStatement) statementNode() {}

func (assignment *AssignmentStatement) TokenLiteral() string {
	return assignment.Token.Literal
}

func (assignment *AssignmentStatement) String() string {
	var out bytes.Buffer

	out.WriteString(assignment.Name.String())
	out.WriteString(" ")
	out.WriteString(assignment.Operator)
	out.WriteString(" ")
	out.WriteString(assignment.Value.String())

	return out.String()

}
