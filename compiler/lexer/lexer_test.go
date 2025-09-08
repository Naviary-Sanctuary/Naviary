package lexer

import (
	"naviary/compiler/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNextToken(t *testing.T) {
	t.Run("parse simple code", func(t *testing.T) {
		input := `let five = 5
let ten = 10

let add = func(x, y) {
    x + y
}

let result = add(five, ten)
// This is a comment
counter := 0
counter += 1
`

		lexer := New(input)

		expectedTokens := []token.Token{
			{Type: token.LET, Literal: "let", Line: 1, Column: 1},
			{Type: token.IDENT, Literal: "five", Line: 1, Column: 5},
			{Type: token.ASSIGN, Literal: "=", Line: 1, Column: 10},
			{Type: token.INT, Literal: "5", Line: 1, Column: 12},
			{Type: token.NEWLINE, Literal: "\n", Line: 1, Column: 13},
			{Type: token.LET, Literal: "let", Line: 2, Column: 1},
			{Type: token.IDENT, Literal: "ten", Line: 2, Column: 5},
			{Type: token.ASSIGN, Literal: "=", Line: 2, Column: 9},
			{Type: token.INT, Literal: "10", Line: 2, Column: 11},
			{Type: token.NEWLINE, Literal: "\n", Line: 2, Column: 13},
			{Type: token.NEWLINE, Literal: "\n", Line: 3, Column: 1},
			{Type: token.LET, Literal: "let", Line: 4, Column: 1},
			{Type: token.IDENT, Literal: "add", Line: 4, Column: 5},
			{Type: token.ASSIGN, Literal: "=", Line: 4, Column: 9},
			{Type: token.FUNC, Literal: "func", Line: 4, Column: 11},
			{Type: token.LEFT_PAREN, Literal: "(", Line: 4, Column: 15},
			{Type: token.IDENT, Literal: "x", Line: 4, Column: 16},
			{Type: token.COMMA, Literal: ",", Line: 4, Column: 17},
			{Type: token.IDENT, Literal: "y", Line: 4, Column: 19},
			{Type: token.RIGHT_PAREN, Literal: ")", Line: 4, Column: 20},
			{Type: token.LEFT_BRACE, Literal: "{", Line: 4, Column: 22},
			{Type: token.NEWLINE, Literal: "\n", Line: 4, Column: 23},
			{Type: token.IDENT, Literal: "x", Line: 5, Column: 5},
			{Type: token.PLUS, Literal: "+", Line: 5, Column: 7},
			{Type: token.IDENT, Literal: "y", Line: 5, Column: 9},
			{Type: token.NEWLINE, Literal: "\n", Line: 5, Column: 10},
			{Type: token.RIGHT_BRACE, Literal: "}", Line: 6, Column: 1},
			{Type: token.NEWLINE, Literal: "\n", Line: 6, Column: 2},
			{Type: token.NEWLINE, Literal: "\n", Line: 7, Column: 1},
			{Type: token.LET, Literal: "let", Line: 8, Column: 1},
			{Type: token.IDENT, Literal: "result", Line: 8, Column: 5},
			{Type: token.ASSIGN, Literal: "=", Line: 8, Column: 12},
			{Type: token.IDENT, Literal: "add", Line: 8, Column: 14},
			{Type: token.LEFT_PAREN, Literal: "(", Line: 8, Column: 17},
			{Type: token.IDENT, Literal: "five", Line: 8, Column: 18},
			{Type: token.COMMA, Literal: ",", Line: 8, Column: 22},
			{Type: token.IDENT, Literal: "ten", Line: 8, Column: 24},
			{Type: token.RIGHT_PAREN, Literal: ")", Line: 8, Column: 27},
			{Type: token.NEWLINE, Literal: "\n", Line: 8, Column: 28},
			{Type: token.NEWLINE, Literal: "\n", Line: 9, Column: 21},
			{Type: token.IDENT, Literal: "counter", Line: 10, Column: 1},
			{Type: token.COLON_ASSIGN, Literal: ":=", Line: 10, Column: 9},
			{Type: token.INT, Literal: "0", Line: 10, Column: 12},
			{Type: token.NEWLINE, Literal: "\n", Line: 10, Column: 13},
			{Type: token.IDENT, Literal: "counter", Line: 11, Column: 1},
			{Type: token.PLUS_ASSIGN, Literal: "+=", Line: 11, Column: 9},
			{Type: token.INT, Literal: "1", Line: 11, Column: 12},
			{Type: token.NEWLINE, Literal: "\n", Line: 11, Column: 13},
			{Type: token.EOF, Literal: "", Line: 12, Column: 1},
		}

		for i, expected := range expectedTokens {
			actual := lexer.NextToken()
			assert.Equal(t, expected, actual, "Token %d mismatch", i)
		}
	})

}
