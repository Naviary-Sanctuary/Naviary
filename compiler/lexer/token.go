package lexer

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	EOF TokenType = iota
	Illegal

	// Literals
	Number     // 123
	Identifier // variable names

	// Keywords
	Let   // let
	Func  // func
	Print // print (built-in for MVP)

	// Operators
	Plus     // +
	Minus    // -
	Asterisk // *
	Slash    // /
	Assign   // =

	// Delimiters
	LeftParen  // (
	RightParen // )
	LeftBrace  // {
	RightBrace // }
)

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Keyword map for quick lookup
var keywords = map[string]TokenType{
	"let":   Let,
	"func":  Func,
	"print": Print,
}

// LookupIdentifier checks if an identifier is a keyword
func LookupIdentifier(identifier string) TokenType {
	if tokenType, ok := keywords[identifier]; ok {
		return tokenType
	}
	return Identifier
}
