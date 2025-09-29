package token

var tokenMap = [...]string{
	// Special tokens
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	// Literals
	INT: "INT",

	// Identifier
	IDENTIFIER: "IDENTIFIER",

	// Keywords
	LET:    "let",
	MUT:    "mut",
	FUNC:   "func",
	RETURN: "return",

	// Operators
	PLUS:         "+",
	MINUS:        "-",
	ASTERISK:     "*",
	SLASH:        "/",
	ASSIGN:       "=",
	COLON_ASSIGN: ":=",

	// Delimiters
	LEFT_PAREN:  "(",
	RIGHT_PAREN: ")",
	LEFT_BRACE:  "{",
	RIGHT_BRACE: "}",
	COMMA:       ",",
	SEMICOLON:   ";",
	COLON:       ":",
	ARROW:       "->",

	NEW_LINE: "\\n",
}
