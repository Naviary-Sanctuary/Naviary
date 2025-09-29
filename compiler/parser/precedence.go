package parser

import "compiler/token"

// Operator precedence levels
// Higher value = higher precedence (binds tighter)
const (
	LOWEST      = iota // 0: Lowest precedence
	ASSIGNMENT         // 1: =, :=, +=, -=, *=, /=
	PIPELINE           // 2: |> (future)
	NILCOALESCE        // 3: ?? (future)
	LOGICAL_OR         // 4: || (future)
	LOGICAL_AND        // 5: && (future)
	TYPE_OP            // 6: is, as (future)
	BITWISE_OR         // 7: | (future)
	BITWISE_XOR        // 8: ^ (future)
	BITWISE_AND        // 9: & (future)
	EQUALITY           // 10: ==, != (future)
	COMPARISON         // 11: <, >, <=, >= (future)
	RANGE              // 12: .., ..= (future)
	SHIFT              // 13: <<, >>, >>> (future)
	SUM                // 14: +, -
	PRODUCT            // 15: *, /, %
	EXPONENT           // 16: ** (future)
	UNARY              // 17: !, ~, -, + prefix (future)
	CALL               // 18: function(), [], ., ?., :: (highest)
)

// precedenceMap maps token types to their precedence level
var precedenceMap = map[token.TokenType]int{
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,

	// Comparison operators
	// token.LESS_THAN:          COMPARISON,
	// token.GREATER_THAN:       COMPARISON,
	// token.LESS_THAN_EQUAL:    COMPARISON,
	// token.GREATER_THAN_EQUAL: COMPARISON,

	// Equality operators
	// token.EQUAL:     EQUALITY,
	// token.NOT_EQUAL: EQUALITY,

	// Logical operators
	// token.LOGICAL_AND: LOGICAL_AND,
	// token.LOGICAL_OR:  LOGICAL_OR,

	// Assignment operators
	// token.ASSIGN:          ASSIGNMENT,
	// token.COLON_ASSIGN:    ASSIGNMENT,
	// token.PLUS_ASSIGN:     ASSIGNMENT,
	// token.MINUS_ASSIGN:    ASSIGNMENT,
	// token.ASTERISK_ASSIGN: ASSIGNMENT,
	// token.SLASH_ASSIGN:    ASSIGNMENT,

	// Function call has highest precedence
	token.LEFT_PAREN: CALL,
}

// getPrecedence returns the precedence level for a given token type
// Returns LOWEST for non-operator tokens
func getPrecedence(tokenType token.TokenType) int {
	if precedence, ok := precedenceMap[tokenType]; ok {
		return precedence
	}
	return LOWEST
}
