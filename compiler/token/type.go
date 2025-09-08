package token

import "fmt"

type TokenType int

const (
	// Special Tokens
	ILLEGAL TokenType = iota // Unrecognized Character
	EOF                      // End of File

	literalBegin
	//Literals
	INT    // 123, 456
	FLOAT  // 3.14
	STRING // "hello"

	literalEnd

	// Identifier
	IDENT // Variable names, function names, etc

	keywordBigin
	// Keywords
	LET    // let
	MUT    // mut
	FUNC   // func
	IF     // if
	ELSE   // else
	FOR    // for
	RETURN // return
	CLASS  // class

	// Boolean Literals
	TRUE  // true
	FALSE // false

	keywordEnd

	operatorBegin
	// Operators
	PLUS               // +
	MINUS              // -
	ASTERISK           // *
	SLASH              // /
	ASSIGN             // =
	COLON_ASSIGN       // :=
	EQUAL              // ==
	NOT_EQUAL          // !=
	GREATER_THAN       // >
	LESS_THAN          // <
	GREATER_THAN_EQUAL // >=
	LESS_THAN_EQUAL    // <=

	// Compound Assignment Operators
	PLUS_ASSIGN     // +=
	MINUS_ASSIGN    // -=
	ASTERISK_ASSIGN // *=
	SLASH_ASSIGN    // /=

	operatorEnd

	// Delimiters
	LEFT_PAREN    // (
	RIGHT_PAREN   // )
	LEFT_BRACE    // {
	RIGHT_BRACE   // }
	LEFT_BRACKET  // [
	RIGHT_BRACKET // ]
	COMMA         // ,
	COLON         // :
	SEMICOLON     // ;
	ARROW         // ->
	DOT           // .

	// Line Break
	NEWLINE // \n
)

func (tokenType TokenType) String() string {
	if int(tokenType) < len(tokenMap) {
		return tokenMap[tokenType]
	}
	return fmt.Sprintf("TOKEN(%d)", int(tokenType))
}

func (tokenType TokenType) IsKeyword() bool {
	return tokenType >= keywordBigin && tokenType <= keywordEnd
}

func (tokenType TokenType) IsLiteral() bool {
	return tokenType >= literalBegin && tokenType <= literalEnd
}

func (tokenType TokenType) IsOperator() bool {
	return tokenType >= operatorBegin && tokenType <= operatorEnd
}
