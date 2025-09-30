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
				expectedType:  token.INT,
				expectedValue: "1",
			},
			{
				name:          "Multi digit integer",
				input:         "123",
				expectedType:  token.INT,
				expectedValue: "123",
			},
			// TODO:
			// {
			// 	name:          "Integer with underscore",
			// 	input:         "1_000",
			// 	expectedType:  token.INT,
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
					{tokenType: token.INT, tokenValue: "1"},
					{tokenType: token.PLUS, tokenValue: "+"},
					{tokenType: token.INT, tokenValue: "2"},
				},
			},
			{
				name:  "Arithmetic expression",
				input: "1 + 2 * 3",
				expectedTokens: []struct {
					tokenType  token.TokenType
					tokenValue string
				}{
					{tokenType: token.INT, tokenValue: "1"},
					{tokenType: token.PLUS, tokenValue: "+"},
					{tokenType: token.INT, tokenValue: "2"},
					{tokenType: token.ASTERISK, tokenValue: "*"},
					{tokenType: token.INT, tokenValue: "3"},
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
					{token.INT, "10"},
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
					{token.INT, "42"},
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
					{token.INT, "5"},
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
					{token.INT, "5"},
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
					{token.INT, "5"},
					{token.NEW_LINE, "\n"},
					{token.LET, "let"},
					{token.IDENTIFIER, "y"},
					{token.ASSIGN, "="},
					{token.INT, "10"},
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
					{token.INT, "5"},
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
					{token.INT, "5"},
					{token.NEW_LINE, "\n"},
					{token.NEW_LINE, "\n"},
					{token.NEW_LINE, "\n"},
					{token.LET, "let"},
					{token.IDENTIFIER, "y"},
					{token.ASSIGN, "="},
					{token.INT, "10"},
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
					{token.INT, "5"},
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
}
