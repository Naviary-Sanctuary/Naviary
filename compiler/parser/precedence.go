package parser

import "naviary/compiler/token"

const (
	LOWEST      = iota // 0: 가장 낮은 우선순위
	ASSIGNMENT         // 1: =, +=, -=, *=, /=
	PIPELINE           // 2: |>
	NILCOALESCE        // 3: ??
	LOGICAL_OR         // 4: ||
	LOGICAL_AND        // 5: &&
	TYPE_OP            // 6: is, as
	BITWISE_OR         // 7: |
	BITWISE_XOR        // 8: ^
	BITWISE_AND        // 9: &
	EQUALITY           // 10: ==, !=
	COMPARISON         // 11: <, >, <=, >=
	RANGE              // 12: .., ..=
	SHIFT              // 13: <<, >>, >>>
	SUM                // 14: +, -
	PRODUCT            // 15: *, /, %
	EXPONENT           // 16: **
	UNARY              // 17: !, ~, -, + (prefix)
	Grouping           // 18: (), [], ., ?., :: (가장 높은 우선순위)
)

var precedenceMap = map[token.TokenType]int{
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

	// Function call
	token.LEFT_PAREN: Grouping,
}

func getPrecedence(tokenType token.TokenType) int {
	if precedence, ok := precedenceMap[tokenType]; ok {
		return precedence
	}

	// For all non-operator token
	return LOWEST
}

func isRightAssociative(tokenType token.TokenType) bool {
	switch tokenType {
	case token.ASSIGN, // =
		token.COLON_ASSIGN,    // :=
		token.PLUS_ASSIGN,     // +=
		token.MINUS_ASSIGN,    // -=
		token.ASTERISK_ASSIGN, // *=
		token.SLASH_ASSIGN:    // /=
		return true
	// TODO: EXPONENT (**)
	default:
		return false
	}
}
