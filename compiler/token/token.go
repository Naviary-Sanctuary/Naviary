package token

// Token represents a lexical token
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

func New(tokenType TokenType, value string, line int, column int) Token {
	return Token{
		Type:   tokenType,
		Value:  value,
		Line:   line,
		Column: column,
	}
}

// Keyword map for quick lookup
var keywords = map[string]TokenType{
	"let":    LET,
	"func":   FUNC,
	"return": RETURN,
	"mut":    MUT,
	"class":  CLASS,
	"this":   THIS,

	"int":    INT,
	"float":  FLOAT,
	"string": STRING,
	"bool":   BOOL,
}

// LookupIdentifier checks if an identifier is a keyword
func LookupIdentifier(identifier string) TokenType {
	if tokenType, ok := keywords[identifier]; ok {
		return tokenType
	}
	return IDENTIFIER
}
