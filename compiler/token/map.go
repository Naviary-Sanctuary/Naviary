package token

var tokenMap = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	// Literals show their type
	INT:    "INT",
	FLOAT:  "FLOAT",
	STRING: "STRING",

	// Identifier
	IDENT: "IDENT",

	// Keywords show actual keyword
	LET:    "let",
	MUT:    "mut",
	FUNC:   "func",
	IF:     "if",
	ELSE:   "else",
	FOR:    "for",
	RETURN: "return",
	CLASS:  "class",
	TRUE:   "true",
	FALSE:  "false",

	// Operators show actual symbol
	PLUS:         "+",
	MINUS:        "-",
	ASTERISK:     "*",
	SLASH:        "/",
	ASSIGN:       "=",
	COLON_ASSIGN: ":=",
	EQUAL:        "==",
	NOT_EQUAL:    "!=",
	LESS_THAN:    "<",
	GREATER_THAN: ">",

	PLUS_ASSIGN:     "+=",
	MINUS_ASSIGN:    "-=",
	ASTERISK_ASSIGN: "*=",
	SLASH_ASSIGN:    "/=",

	// Delimiters show actual symbol
	LEFT_PAREN:    "(",
	RIGHT_PAREN:   ")",
	LEFT_BRACE:    "{",
	RIGHT_BRACE:   "}",
	LEFT_BRACKET:  "[",
	RIGHT_BRACKET: "]",
	COMMA:         ",",
	COLON:         ":",
	SEMICOLON:     ";",
	ARROW:         "->",
	DOT:           ".",

	NEWLINE: "\\n",
}
