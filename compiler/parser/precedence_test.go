package parser

import (
	"naviary/compiler/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPrecedence(t *testing.T) {
	tests := []struct {
		name      string
		tokenType token.TokenType
		expected  int
	}{
		// Arithmetic operators
		{"plus should have SUM precedence", token.PLUS, SUM},
		{"minus should have SUM precedence", token.MINUS, SUM},
		{"multiply should have PRODUCT precedence", token.ASTERISK, PRODUCT},
		{"divide should have PRODUCT precedence", token.SLASH, PRODUCT},

		// Comparison operators
		{"less than should have COMPARISON precedence", token.LESS_THAN, COMPARISON},
		{"greater than should have COMPARISON precedence", token.GREATER_THAN, COMPARISON},

		// Non-operators should have LOWEST precedence
		{"identifier should have LOWEST precedence", token.IDENT, LOWEST},
		{"integer should have LOWEST precedence", token.INT, LOWEST},
		{"EOF should have LOWEST precedence", token.EOF, LOWEST},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPrecedence(tt.tokenType)
			assert.Equal(t, tt.expected, result,
				"getPrecedence(%s) = %d, want %d",
				tt.tokenType, result, tt.expected)
		})
	}
}

func TestPrecedenceOrdering(t *testing.T) {
	// Verify that precedence values are in correct order (reversed)
	assert.True(t, PRODUCT > SUM, "PRODUCT should be lower precedence (higher number) than SUM")
	assert.True(t, SUM > COMPARISON, "SUM should be lower precedence than COMPARISON")
	assert.True(t, COMPARISON > EQUALITY, "COMPARISON should be lower precedence than EQUALITY")
}
