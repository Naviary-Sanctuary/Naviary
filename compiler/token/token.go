package token

import "fmt"

type Token struct {
	Type    TokenType
	Literal string // actual text (e.g., "let", "x", "=", "42" )
	Line    int    // Line Number (for error messages)
	Column  int    // Column number (for error messages)
}

func (token Token) String() string {
	if token.Type.IsLiteral() || token.Type == IDENT {
		return fmt.Sprintf("<%s:%s at %d:%d>",
			token.Type, token.Literal, token.Line, token.Column)
	}

	return fmt.Sprintf("<%s at %d:%d>",
		token.Type, token.Line, token.Column)
}

var keywords = map[string]TokenType{
	"let":    LET,
	"mut":    MUT,
	"func":   FUNC,
	"if":     IF,
	"else":   ELSE,
	"for":    FOR,
	"return": RETURN,
	"class":  CLASS,
	"true":   TRUE,
	"false":  FALSE,
}

func LookupIdent(identifier string) TokenType {
	if tokenType, ok := keywords[identifier]; ok {
		return tokenType
	}
	return IDENT
}
