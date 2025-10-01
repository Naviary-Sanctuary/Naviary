package lexer

import (
	"compiler/errors"
	"compiler/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

type singleTokenTestCase struct {
	name          string
	input         string
	expectedType  token.TokenType
	expectedValue string
}

type multipleTokenTestCase struct {
	name           string
	input          string
	expectedTokens []struct {
		tokenType  token.TokenType
		tokenValue string
	}
}

func TestLexer(t *testing.T) {
	t.Run("Test single tokens", func(t *testing.T) {
		tests := []singleTokenTestCase{
			{
				name:          "Plus Operator",
				input:         "+",
				expectedType:  token.PLUS,
				expectedValue: "+",
			},
			{
				name:          "Minus Operator",
				input:         "-",
				expectedType:  token.MINUS,
				expectedValue: "-",
			},
			{
				name:          "Asterisk Operator",
				input:         "*",
				expectedType:  token.ASTERISK,
				expectedValue: "*",
			},
			{
				name:          "Slash Operator",
				input:         "/",
				expectedType:  token.SLASH,
				expectedValue: "/",
			},
			{
				name:          "Assign Operator",
				input:         "=",
				expectedType:  token.ASSIGN,
				expectedValue: "=",
			},
			{
				name:          "Left Parenthesis",
				input:         "(",
				expectedType:  token.LEFT_PAREN,
				expectedValue: "(",
			},
			{
				name:          "Right Parenthesis",
				input:         ")",
				expectedType:  token.RIGHT_PAREN,
				expectedValue: ")",
			},
			{
				name:          "Left Brace",
				input:         "{",
				expectedType:  token.LEFT_BRACE,
				expectedValue: "{",
			},
			{
				name:          "Right Brace",
				input:         "}",
				expectedType:  token.RIGHT_BRACE,
				expectedValue: "}",
			},
			{
				name:          "Comma",
				input:         ",",
				expectedType:  token.COMMA,
				expectedValue: ",",
			},
			{
				name:          "Semicolon",
				input:         ";",
				expectedType:  token.SEMICOLON,
				expectedValue: ";",
			},
			{
				name:          "Colon",
				input:         ":",
				expectedType:  token.COLON,
				expectedValue: ":",
			},
			{
				name:          "Arrow",
				input:         "->",
				expectedType:  token.ARROW,
				expectedValue: "->",
			},
			// INTEGER
			{
				name:          "Single digit integer",
				input:         "1",
				expectedType:  token.INT_LITERAL,
				expectedValue: "1",
			},
			{
				name:          "Multi digit integer",
				input:         "123",
				expectedType:  token.INT_LITERAL,
				expectedValue: "123",
			},
			// TODO:
			// {
			// 	name:          "Integer with underscore",
			// 	input:         "1_000",
			// 	expectedType:  token.INT_LITERAL,
			// 	expectedValue: "1_000",
			// },
			// IDENTIFIER
			{
				name:          "Simple identifier",
				input:         "x",
				expectedType:  token.IDENTIFIER,
				expectedValue: "x",
			},
			{
				name:          "Multi-character identifier",
				input:         "abc",
				expectedType:  token.IDENTIFIER,
				expectedValue: "abc",
			},
			{
				name:          "Identifier with underscore",
				input:         "x_y",
				expectedType:  token.IDENTIFIER,
				expectedValue: "x_y",
			},
			{
				name:          "Identifier with number",
				input:         "x1",
				expectedType:  token.IDENTIFIER,
				expectedValue: "x1",
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				errorCollector := errors.New(testCase.input, "test.navi")
				lexerInstance := New(testCase.input, "test.navi", errorCollector)

				tok := lexerInstance.NextToken()

				assert.Equal(t, testCase.expectedType, tok.Type,
					"Token type mismatch")
				assert.Equal(t, testCase.expectedValue, tok.Value,
					"Token value mismatch")

				// Verify no errors occurred
				assert.False(t, errorCollector.HasErrors(),
					"Lexer should not produce errors for valid input")
			})
		}
	})

	t.Run("Test keywords", func(t *testing.T) {
		tests := []singleTokenTestCase{
			{
				name:          "let keyword",
				input:         "let",
				expectedType:  token.LET,
				expectedValue: "let",
			},
			{
				name:          "func keyword",
				input:         "func",
				expectedType:  token.FUNC,
				expectedValue: "func",
			},
			{
				name:          "return keyword",
				input:         "return",
				expectedType:  token.RETURN,
				expectedValue: "return",
			},
			{
				name:          "mut keyword",
				input:         "mut",
				expectedType:  token.MUT,
				expectedValue: "mut",
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				errorCollector := errors.New(testCase.input, "test.navi")
				lexerInstance := New(testCase.input, "test.navi", errorCollector)

				tok := lexerInstance.NextToken()

				assert.Equal(t, testCase.expectedType, tok.Type,
					"Token type mismatch")
				assert.Equal(t, testCase.expectedValue, tok.Value,
					"Token value mismatch")

				// Verify no errors occurred
				assert.False(t, errorCollector.HasErrors(),
					"Lexer should not produce errors for valid input")
			})
		}
	})

	t.Run("Test Compound tokens", func(t *testing.T) {
		tests := []singleTokenTestCase{
			{
				name:          "Arrow token",
				input:         "->",
				expectedType:  token.ARROW,
				expectedValue: "->",
			},
			{
				name:          "Colon assign token",
				input:         ":=",
				expectedType:  token.COLON_ASSIGN,
				expectedValue: ":=",
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				errorCollector := errors.New(testCase.input, "test.navi")
				lexerInstance := New(testCase.input, "test.navi", errorCollector)

				tok := lexerInstance.NextToken()

				assert.Equal(t, testCase.expectedType, tok.Type,
					"Token type mismatch")
				assert.Equal(t, testCase.expectedValue, tok.Value,
					"Token value mismatch")
			})
		}
	})

	t.Run("Test complex  expressions", func(t *testing.T) {
		tests := []multipleTokenTestCase{
			{
				name:  "Variable declaration",
				input: "let x = 1 + 2",
				expectedTokens: []struct {
					tokenType  token.TokenType
					tokenValue string
				}{
					{tokenType: token.LET, tokenValue: "let"},
					{tokenType: token.IDENTIFIER, tokenValue: "x"},
					{tokenType: token.ASSIGN, tokenValue: "="},
					{tokenType: token.INT_LITERAL, tokenValue: "1"},
					{tokenType: token.PLUS, tokenValue: "+"},
					{tokenType: token.INT_LITERAL, tokenValue: "2"},
				},
			},
			{
				name:  "Arithmetic expression",
				input: "1 + 2 * 3",
				expectedTokens: []struct {
					tokenType  token.TokenType
					tokenValue string
				}{
					{tokenType: token.INT_LITERAL, tokenValue: "1"},
					{tokenType: token.PLUS, tokenValue: "+"},
					{tokenType: token.INT_LITERAL, tokenValue: "2"},
					{tokenType: token.ASTERISK, tokenValue: "*"},
					{tokenType: token.INT_LITERAL, tokenValue: "3"},
				},
			},
			{
				name:  "Function declaration",
				input: "func main() {}",
				expectedTokens: []struct {
					tokenType  token.TokenType
					tokenValue string
				}{
					{token.FUNC, "func"},
					{token.IDENTIFIER, "main"},
					{token.LEFT_PAREN, "("},
					{token.RIGHT_PAREN, ")"},
					{token.LEFT_BRACE, "{"},
					{token.RIGHT_BRACE, "}"},
					{token.EOF, ""},
				},
			},
			{
				name:  "Function with return type",
				input: "func add() -> int",
				expectedTokens: []struct {
					tokenType  token.TokenType
					tokenValue string
				}{
					{token.FUNC, "func"},
					{token.IDENTIFIER, "add"},
					{token.LEFT_PAREN, "("},
					{token.RIGHT_PAREN, ")"},
					{token.ARROW, "->"},
					{token.IDENTIFIER, "int"},
					{token.EOF, ""},
				},
			},
			{
				name:  "Mutable variable declaration",
				input: "let mut x := 10",
				expectedTokens: []struct {
					tokenType  token.TokenType
					tokenValue string
				}{
					{token.LET, "let"},
					{token.MUT, "mut"},
					{token.IDENTIFIER, "x"},
					{token.COLON_ASSIGN, ":="},
					{token.INT_LITERAL, "10"},
					{token.EOF, ""},
				},
			},
			{
				name:  "Function call",
				input: "print(42)",
				expectedTokens: []struct {
					tokenType  token.TokenType
					tokenValue string
				}{
					{token.IDENTIFIER, "print"},
					{token.LEFT_PAREN, "("},
					{token.INT_LITERAL, "42"},
					{token.RIGHT_PAREN, ")"},
					{token.EOF, ""},
				},
			},
			{
				name:  "Multiple operations",
				input: "a + b - c * d / e",
				expectedTokens: []struct {
					tokenType  token.TokenType
					tokenValue string
				}{
					{token.IDENTIFIER, "a"},
					{token.PLUS, "+"},
					{token.IDENTIFIER, "b"},
					{token.MINUS, "-"},
					{token.IDENTIFIER, "c"},
					{token.ASTERISK, "*"},
					{token.IDENTIFIER, "d"},
					{token.SLASH, "/"},
					{token.IDENTIFIER, "e"},
					{token.EOF, ""},
				},
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				errorCollector := errors.New(testCase.input, "test.navi")
				lexerInstance := New(testCase.input, "test.navi", errorCollector)

				for index, expected := range testCase.expectedTokens {
					tok := lexerInstance.NextToken()

					assert.Equal(t, expected.tokenType, tok.Type,
						"Token %d: type mismatch", index)
					assert.Equal(t, expected.tokenValue, tok.Value,
						"Token %d: value mismatch", index)
				}

				assert.False(t, errorCollector.HasErrors(),
					"Lexer should not produce errors for valid input")
			})
		}
	})

	t.Run("Test Lexer errors", func(t *testing.T) {
		tests := []struct {
			name               string
			input              string
			expectedErrorCount int
			shouldContainError string
		}{
			{
				name:               "Invalid character @",
				input:              "@",
				expectedErrorCount: 1,
				shouldContainError: "Unexpected character",
			},
			{
				name:               "Invalid character #",
				input:              "#",
				expectedErrorCount: 1,
				shouldContainError: "Unexpected character",
			},
			{
				name:               "Invalid character $",
				input:              "$",
				expectedErrorCount: 1,
				shouldContainError: "Unexpected character",
			},
			{
				name:               "Invalid number format",
				input:              "123abc",
				expectedErrorCount: 1,
				shouldContainError: "Invalid number format",
			},
			{
				name:               "Multiple invalid characters",
				input:              "let x = @ + #",
				expectedErrorCount: 2,
				shouldContainError: "Unexpected character",
			},
			{
				name:               "Invalid number in expression",
				input:              "let x = 123abc + 5",
				expectedErrorCount: 1,
				shouldContainError: "Invalid number format",
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				errorCollector := errors.New(testCase.input, "test.navi")
				lexerInstance := New(testCase.input, "test.navi", errorCollector)

				for {
					tok := lexerInstance.NextToken()
					if tok.Type == token.EOF {
						break
					}
				}

				assert.True(t, errorCollector.HasErrors(),
					"Lexer should produce errors for invalid input")
			})
		}
	})

	t.Run("Test whitespace handling", func(t *testing.T) {
		tests := []multipleTokenTestCase{
			{
				name:  "Multiple spaces between tokens",
				input: "let     x     =     5",
				expectedTokens: []struct {
					tokenType  token.TokenType
					tokenValue string
				}{
					{token.LET, "let"},
					{token.IDENTIFIER, "x"},
					{token.ASSIGN, "="},
					{token.INT_LITERAL, "5"},
					{token.EOF, ""},
				},
			},
			{
				name:  "Tabs between tokens",
				input: "let\tx\t=\t5",
				expectedTokens: []struct {
					tokenType  token.TokenType
					tokenValue string
				}{
					{token.LET, "let"},
					{token.IDENTIFIER, "x"},
					{token.ASSIGN, "="},
					{token.INT_LITERAL, "5"},
					{token.EOF, ""},
				},
			},
			{
				name:  "Newline between statements",
				input: "let x = 5\nlet y = 10",
				expectedTokens: []struct {
					tokenType  token.TokenType
					tokenValue string
				}{
					{token.LET, "let"},
					{token.IDENTIFIER, "x"},
					{token.ASSIGN, "="},
					{token.INT_LITERAL, "5"},
					{token.NEW_LINE, "\n"},
					{token.LET, "let"},
					{token.IDENTIFIER, "y"},
					{token.ASSIGN, "="},
					{token.INT_LITERAL, "10"},
					{token.EOF, ""},
				},
			},
			{
				name:  "Mixed whitespace",
				input: "  let  \t x \n = \t\t 5  ",
				expectedTokens: []struct {
					tokenType  token.TokenType
					tokenValue string
				}{
					{token.LET, "let"},
					{token.IDENTIFIER, "x"},
					{token.NEW_LINE, "\n"},
					{token.ASSIGN, "="},
					{token.INT_LITERAL, "5"},
					{token.EOF, ""},
				},
			},
			{
				name:  "Multiple newlines",
				input: "let x = 5\n\n\nlet y = 10",
				expectedTokens: []struct {
					tokenType  token.TokenType
					tokenValue string
				}{
					{token.LET, "let"},
					{token.IDENTIFIER, "x"},
					{token.ASSIGN, "="},
					{token.INT_LITERAL, "5"},
					{token.NEW_LINE, "\n"},
					{token.NEW_LINE, "\n"},
					{token.NEW_LINE, "\n"},
					{token.LET, "let"},
					{token.IDENTIFIER, "y"},
					{token.ASSIGN, "="},
					{token.INT_LITERAL, "10"},
					{token.EOF, ""},
				},
			},
			{
				name:  "No spaces (compact)",
				input: "let x=5",
				expectedTokens: []struct {
					tokenType  token.TokenType
					tokenValue string
				}{
					{token.LET, "let"},
					{token.IDENTIFIER, "x"},
					{token.ASSIGN, "="},
					{token.INT_LITERAL, "5"},
					{token.EOF, ""},
				},
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				errorCollector := errors.New(testCase.input, "test.navi")
				lexerInstance := New(testCase.input, "test.navi", errorCollector)

				for index, expected := range testCase.expectedTokens {
					tok := lexerInstance.NextToken()

					assert.Equal(t, expected.tokenType, tok.Type,
						"Token %d: type mismatch", index)
					assert.Equal(t, expected.tokenValue, tok.Value,
						"Token %d: value mismatch", index)
				}

				assert.False(t, errorCollector.HasErrors(),
					"Lexer should not produce errors for valid input with whitespace")
			})
		}
	})
	t.Run("Test token position", func(t *testing.T) {
		tests := []struct {
			name           string
			input          string
			expectedTokens []struct {
				tokenType token.TokenType
				value     string
				line      int
				column    int
			}
		}{
			{
				name:  "Single line positions",
				input: "let x = 5",
				expectedTokens: []struct {
					tokenType token.TokenType
					value     string
					line      int
					column    int
				}{
					{token.LET, "let", 1, 1},
					{token.IDENTIFIER, "x", 1, 5},
					{token.ASSIGN, "=", 1, 7},
					{token.INT_LITERAL, "5", 1, 9},
				},
			},
			{
				name:  "Multiple lines",
				input: "let x = 5\nlet y = 10",
				expectedTokens: []struct {
					tokenType token.TokenType
					value     string
					line      int
					column    int
				}{
					{token.LET, "let", 1, 1},
					{token.IDENTIFIER, "x", 1, 5},
					{token.ASSIGN, "=", 1, 7},
					{token.INT_LITERAL, "5", 1, 9},
					{token.NEW_LINE, "\n", 1, 10},
					{token.LET, "let", 2, 1},
					{token.IDENTIFIER, "y", 2, 5},
					{token.ASSIGN, "=", 2, 7},
					{token.INT_LITERAL, "10", 2, 9},
				},
			},
			{
				name:  "Function declaration with multiple lines",
				input: "func main() {\n  let x = 5\n}",
				expectedTokens: []struct {
					tokenType token.TokenType
					value     string
					line      int
					column    int
				}{
					{token.FUNC, "func", 1, 1},
					{token.IDENTIFIER, "main", 1, 6},
					{token.LEFT_PAREN, "(", 1, 10},
					{token.RIGHT_PAREN, ")", 1, 11},
					{token.LEFT_BRACE, "{", 1, 13},
					{token.NEW_LINE, "\n", 1, 14},
					{token.LET, "let", 2, 3},
					{token.IDENTIFIER, "x", 2, 7},
					{token.ASSIGN, "=", 2, 9},
					{token.INT_LITERAL, "5", 2, 11},
					{token.NEW_LINE, "\n", 2, 12},
					{token.RIGHT_BRACE, "}", 3, 1},
				},
			},
			{
				name:  "Compound tokens positions",
				input: "func add() -> int",
				expectedTokens: []struct {
					tokenType token.TokenType
					value     string
					line      int
					column    int
				}{
					{token.FUNC, "func", 1, 1},
					{token.IDENTIFIER, "add", 1, 6},
					{token.LEFT_PAREN, "(", 1, 9},
					{token.RIGHT_PAREN, ")", 1, 10},
					{token.ARROW, "->", 1, 12},
					{token.IDENTIFIER, "int", 1, 15},
				},
			},
			{
				name:  "Colon assign position",
				input: "let x := 5",
				expectedTokens: []struct {
					tokenType token.TokenType
					value     string
					line      int
					column    int
				}{
					{token.LET, "let", 1, 1},
					{token.IDENTIFIER, "x", 1, 5},
					{token.COLON_ASSIGN, ":=", 1, 7},
					{token.INT_LITERAL, "5", 1, 10},
				},
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				errorCollector := errors.New(testCase.input, "test.navi")
				lexerInstance := New(testCase.input, "test.navi", errorCollector)

				for index, expected := range testCase.expectedTokens {
					tok := lexerInstance.NextToken()

					assert.Equal(t, expected.tokenType, tok.Type,
						"Token %d: type mismatch", index)
					assert.Equal(t, expected.value, tok.Value,
						"Token %d: value mismatch", index)
					assert.Equal(t, expected.line, tok.Line,
						"Token %d (%s): line mismatch", index, expected.value)
					assert.Equal(t, expected.column, tok.Column,
						"Token %d (%s): column mismatch", index, expected.value)
				}

				// Verify no errors occurred
				assert.False(t, errorCollector.HasErrors(),
					"Lexer should not produce errors for valid input")
			})
		}
	})
}
