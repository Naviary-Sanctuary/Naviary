package token

import "fmt"

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	EOF TokenType = iota
	ILLEGAL

	// Literals
	INT        // 123
	IDENTIFIER // variable names

	// Keywords
	LET    // let
	MUT    // mut
	RETURN // return
	FUNC   // func

	operatorBegin
	// Operators
	PLUS         // +
	MINUS        // -
	ASTERISK     // *
	SLASH        // /
	ASSIGN       // =
	COLON_ASSIGN // :=

	operatorEnd

	// Delimiters
	LEFT_PAREN  // (
	RIGHT_PAREN // )
	LEFT_BRACE  // {
	RIGHT_BRACE // }
	COMMA       // ,
	SEMICOLON   // ;
	COLON       // :
	ARROW       // ->

	NEW_LINE // \n
)

func (tokenType TokenType) String() string {
	if int(tokenType) < len(tokenMap) {
		return tokenMap[tokenType]
	}

	return fmt.Sprintf("TOKEN(%d)", int(tokenType))
}

func (tokenType TokenType) IsOperator() bool {
	return tokenType >= operatorBegin && tokenType <= operatorEnd
}
