package parser

import "naviary/compiler/token"

const (
	Grouping    = iota // 0: (), [], ., ?., ::
	UNARY              // 1: !, ~, -, + (prefix)
	EXPONENT           // 2: **
	PRODUCT            // 3: *, /, %
	SUM                // 4: +, -
	SHIFT              // 5: <<, >>, >>>
	RANGE              // 6: .., ..=
	COMPARISON         // 7: <, >, <=, >=
	EQUALITY           // 8: ==, !=
	BITWISE_AND        // 9: &
	BITWISE_XOR        // 10: ^
	BITWISE_OR         // 11: |
	TYPE_OP            // 12: is, as
	LOGICAL_AND        // 13: &&
	LOGICAL_OR         // 14: ||
	NILCOALESCE        // 15: ??
	PIPELINE           // 16: |>
	ASSIGNMENT         // 17: =, +=, -=, *=, /=
	LOWEST
)

var procedenceMap = map[token.TokenType]int{
	// Arithmetic operators
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
	// token.PERCENT:  PRODUCT,
	token.PLUS:  SUM,
	token.MINUS: SUM,

	// Comparison operators
	token.GREATER_THAN:       COMPARISON,
	token.GREATER_THAN_EQUAL: COMPARISON,
	token.LESS_THAN:          COMPARISON,
	token.LESS_THAN_EQUAL:    COMPARISON,

	// Equality operators
	token.EQUAL:     EQUALITY,
	token.NOT_EQUAL: EQUALITY,

	// Assignment operators
	token.ASSIGN:          ASSIGNMENT,
	token.COLON_ASSIGN:    ASSIGNMENT,
	token.PLUS_ASSIGN:     ASSIGNMENT,
	token.MINUS_ASSIGN:    ASSIGNMENT,
	token.ASTERISK_ASSIGN: ASSIGNMENT,
	token.SLASH_ASSIGN:    ASSIGNMENT,
}

func getPrecedence(tokenType token.TokenType) int {
	if precedence, ok := procedenceMap[tokenType]; ok {
		return precedence
	}

	// For all non-operator token
	return LOWEST
}
