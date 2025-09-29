package ast

import (
	"bytes"
	"compiler/token"
)

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (block *BlockStatement) statementNode() {}

func (block *BlockStatement) TokenLiteral() string {
	return block.Token.Value
}

func (block *BlockStatement) String() string {
	var out bytes.Buffer

	out.WriteString("{\n")
	for _, stmt := range block.Statements {
		out.WriteString("  ")
		out.WriteString(stmt.String())
		out.WriteString("\n")
	}
	out.WriteString("}")
	return out.String()
}
