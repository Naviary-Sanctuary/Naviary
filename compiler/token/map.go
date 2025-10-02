package token

var tokenMap = [...]string{
	// Special tokens
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	// Literals
	INT_LITERAL:    "INT_LITERAL",
	STRING_LITERAL: "STRING_LITERAL",

	// Identifier
	IDENTIFIER: "IDENTIFIER",

	// Keywords
	LET:    "let",
	MUT:    "mut",
	FUNC:   "func",
	RETURN: "return",
	CLASS:  "class",
	THIS:   "this",

	// Type keywords
	INT:    "int",
	FLOAT:  "float",
	STRING: "string",
	BOOL:   "bool",

	// Operators
	PLUS:         "+",
	MINUS:        "-",
	ASTERISK:     "*",
	SLASH:        "/",
	ASSIGN:       "=",
	COLON_ASSIGN: ":=",
	DOT:          ".",

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
