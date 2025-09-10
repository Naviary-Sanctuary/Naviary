package lexer

import (
	"naviary/compiler/errors"
	"naviary/compiler/token"
)

type Lexer struct {
	input   string
	current int // current position in input
	next    int // next position to read
	char    byte
	line    int
	column  int
	errors  *errors.ErrorCollector
}

func New(input, filename string) *Lexer {
	lexer := &Lexer{input: input, line: 1, column: 0, errors: errors.New(input, filename)}
	lexer.advance()
	return lexer
}

func (lexer *Lexer) Errors() *errors.ErrorCollector {
	return lexer.errors
}

// Advance the lexer to the next character
func (lexer *Lexer) advance() {
	// Move to next position
	lexer.current = lexer.next
	lexer.next++

	// Update line and column based on CURRENT character
	if lexer.char == '\n' {
		lexer.line++
		lexer.column = 1
	} else {
		lexer.column++
	}

	// Read next character
	if lexer.next-1 < len(lexer.input) {
		lexer.char = lexer.input[lexer.next-1]
	} else {
		lexer.char = 0 // EOF
	}
}

// Look at the next character without advancing
func (lexer *Lexer) peek() byte {
	if lexer.next >= len(lexer.input) {
		return 0
	}

	return lexer.input[lexer.next]
}

// Read current char and advance
func (lexer *Lexer) consume() byte {
	char := lexer.char
	lexer.advance()
	return char
}

func (lexer *Lexer) skipWhitespace() {
	for isWhitespace(lexer.char) {
		lexer.advance()
	}
}

func (lexer *Lexer) skipLineComment() {

	// skip `//`
	lexer.advance()
	lexer.advance()

	for lexer.char != '\n' && lexer.char != 0 {
		lexer.advance()
	}
}

func (lexer *Lexer) readIdentifier() string {
	start := lexer.current

	for isLetter(lexer.char) || isDigit(lexer.char) {
		lexer.advance()
	}
	result := lexer.input[start:lexer.current]
	return result
}

func (lexer *Lexer) readNumber() string {
	var result []byte

	// Read integer part
	for isDigit(lexer.char) || lexer.char == '_' {
		if lexer.char == '_' {
			nextChar := lexer.peek()

			// Error: consecutive underscores
			if nextChar == '_' {
				lexer.errors.Add(
					errors.LexicalError,
					lexer.line,
					lexer.column,
					2,
					"consecutive underscores in number literal",
				)
				lexer.advance()
				continue
			}

			// Error: trailing underscore (not followed by digit or dot)
			if !isDigit(nextChar) && nextChar != '.' {
				lexer.errors.Add(
					errors.LexicalError,
					lexer.line,
					lexer.column,
					1,
					"number cannot end with underscore",
				)
				lexer.advance()
				break
			}

			// Valid underscore - skip it
			lexer.advance()

		} else {
			// Normal digit - add to result
			result = append(result, lexer.char)
			lexer.advance()
		}
	}

	// Check for decimal point (float)
	if lexer.char == '.' && isDigit(lexer.peek()) {
		// Add the dot
		result = append(result, '.')
		lexer.advance()

		// Read fractional part
		for isDigit(lexer.char) || lexer.char == '_' {
			if lexer.char == '_' {
				nextChar := lexer.peek()

				// Error: consecutive underscores in fractional part
				if nextChar == '_' {
					lexer.errors.Add(
						errors.LexicalError,
						lexer.line,
						lexer.column,
						2,
						"consecutive underscores in number literal",
					)
					lexer.advance()
					continue
				}

				// Error: trailing underscore in fractional part
				if !isDigit(nextChar) {
					lexer.errors.Add(
						errors.LexicalError,
						lexer.line,
						lexer.column,
						1,
						"number cannot end with underscore",
					)
					lexer.advance()
					break
				}

				// Valid underscore - skip it
				lexer.advance()

			} else {
				// Normal digit in fractional part
				result = append(result, lexer.char)
				lexer.advance()
			}
		}
	}

	return string(result)
}

func (lexer *Lexer) readString() string {
	// Skip opening quote
	lexer.advance()

	start := lexer.current
	startLine := lexer.line
	startColumn := lexer.column - 1 // -1 for the opening quote

	for {
		if lexer.char == '"' {
			// Found closing quote - success
			result := lexer.input[start:lexer.current]
			lexer.advance() // consume closing quote
			return result
		}

		if lexer.char == '\\' {
			// Handle escape sequences
			lexer.advance() // consume backslash

			if lexer.char == 0 {
				// EOF after backslash
				lexer.errors.Add(
					errors.LexicalError,
					startLine,
					startColumn,
					lexer.current-start+1,
					"unterminated string literal",
				)
				return lexer.input[start:lexer.current]
			}

			// Check for valid escape sequences
			switch lexer.char {
			case 'n', 't', 'r', '\\', '"':
				// Valid escape sequences
				lexer.advance()
			default:
				// Invalid escape sequence
				lexer.errors.Add(
					errors.LexicalError,
					lexer.line,
					lexer.column-1, // -1 for backslash position
					2,              // backslash + character
					"invalid escape sequence '\\%c'",
					lexer.char,
				)
				lexer.advance()
			}

		} else if lexer.char == 0 {
			// EOF without closing quote
			lexer.errors.Add(
				errors.LexicalError,
				startLine,
				startColumn,
				lexer.current-start+1,
				"unterminated string literal",
			)
			return lexer.input[start:lexer.current]

		} else if lexer.char == '\n' {
			// Newline in string (not allowed in Naviary)
			lexer.errors.Add(
				errors.LexicalError,
				lexer.line,
				lexer.column,
				1,
				"unexpected newline in string literal",
			)
			// Continue reading to find more errors
			lexer.advance()

		} else {
			// Normal character
			lexer.advance()
		}
	}
}

// create Token
func (lexer *Lexer) newToken(tokenType token.TokenType, char byte) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(char),
		Line:    lexer.line,
		Column:  lexer.column,
	}
}

func (lexer *Lexer) NextToken() token.Token {
	var t token.Token

	lexer.skipWhitespace()

	t.Line = lexer.line
	t.Column = lexer.column

	switch lexer.char {
	case '(':
		t = lexer.newToken(token.LEFT_PAREN, lexer.char)
		lexer.advance()
	case ')':
		t = lexer.newToken(token.RIGHT_PAREN, lexer.char)
		lexer.advance()
	case '{':
		t = lexer.newToken(token.LEFT_BRACE, lexer.char)
		lexer.advance()
	case '}':
		t = lexer.newToken(token.RIGHT_BRACE, lexer.char)
		lexer.advance()
	case '[':
		t = lexer.newToken(token.LEFT_BRACKET, lexer.char)
		lexer.advance()
	case ']':
		t = lexer.newToken(token.RIGHT_BRACKET, lexer.char)
		lexer.advance()
	case ',':
		t = lexer.newToken(token.COMMA, lexer.char)
		lexer.advance()
	case ';':
		t = lexer.newToken(token.SEMICOLON, lexer.char)
		lexer.advance()
	case '.':
		t = lexer.newToken(token.DOT, lexer.char)
		lexer.advance()
	case '\n':
		t = lexer.newToken(token.NEWLINE, lexer.char)
		lexer.advance()

	case '+':
		if lexer.peek() == '=' {
			lexer.advance()
			lexer.advance()

			t = token.Token{
				Type:    token.PLUS_ASSIGN,
				Literal: "+=",
				Line:    t.Line,
				Column:  t.Column,
			}
		} else {
			t = lexer.newToken(token.PLUS, lexer.char)
			lexer.advance()
		}
	case '-':
		if lexer.peek() == '=' {
			lexer.advance()
			lexer.advance()

			t = token.Token{
				Type:    token.MINUS_ASSIGN,
				Literal: "-=",
				Line:    t.Line,
				Column:  t.Column,
			}
		} else if lexer.peek() == '>' {
			lexer.advance()
			lexer.advance()

			t = token.Token{
				Type:    token.ARROW,
				Literal: "->",
				Line:    t.Line,
				Column:  t.Column,
			}
		} else {
			t = lexer.newToken(token.MINUS, lexer.char)
			lexer.advance()
		}
	case '*':
		if lexer.peek() == '=' {
			lexer.advance()
			lexer.advance()

			t = token.Token{
				Type:    token.ASTERISK_ASSIGN,
				Literal: "*=",
				Line:    t.Line,
				Column:  t.Column,
			}
		} else {
			t = lexer.newToken(token.ASTERISK, lexer.char)
			lexer.advance()
		}
	case '/':
		if lexer.peek() == '=' {
			lexer.advance()
			lexer.advance()

			t = token.Token{
				Type:    token.SLASH_ASSIGN,
				Literal: "/=",
				Line:    t.Line,
				Column:  t.Column,
			}
		} else if lexer.peek() == '/' {
			lexer.skipLineComment()
			return lexer.NextToken()
		} else {
			t = lexer.newToken(token.SLASH, lexer.char)
			lexer.advance()
		}
	case '=':
		if lexer.peek() == '=' {
			lexer.advance()
			lexer.advance()

			t = token.Token{
				Type:    token.EQUAL,
				Literal: "==",
				Line:    t.Line,
				Column:  t.Column,
			}
		} else {
			t = lexer.newToken(token.ASSIGN, lexer.char)
			lexer.advance()
		}

	case '!':
		if lexer.peek() == '=' {
			lexer.advance()
			lexer.advance()
			t = token.Token{
				Type:    token.NOT_EQUAL,
				Literal: "!=",
				Line:    t.Line,
				Column:  t.Column,
			}
		} else {
			t = lexer.newToken(token.ILLEGAL, lexer.char)
			lexer.advance()
		}

	case '<':
		if lexer.peek() == '=' {
			lexer.advance()
			lexer.advance()
			t = token.Token{
				Type:    token.LESS_THAN_EQUAL,
				Literal: "<=",
				Line:    t.Line,
				Column:  t.Column,
			}
		} else {
			t = lexer.newToken(token.LESS_THAN, lexer.char)
			lexer.advance()
		}

	case '>':
		if lexer.peek() == '=' {
			lexer.advance()
			lexer.advance()
			t = token.Token{
				Type:    token.GREATER_THAN_EQUAL,
				Literal: ">=",
				Line:    t.Line,
				Column:  t.Column,
			}
		} else {
			t = lexer.newToken(token.GREATER_THAN, lexer.char)
			lexer.advance()
		}
	case ':':
		if lexer.peek() == '=' {
			lexer.advance()
			lexer.advance()
			t = token.Token{
				Type:    token.COLON_ASSIGN,
				Literal: ":=",
				Line:    t.Line,
				Column:  t.Column,
			}
		} else {
			t = lexer.newToken(token.COLON, lexer.char)
			lexer.advance()
		}
	case '"':
		t.Type = token.STRING
		t.Literal = lexer.readString()
	case 0:
		// End of file
		t.Type = token.EOF
		t.Literal = ""

	default:
		// Complex cases
		if isLetter(lexer.char) {
			// Identifier or keyword
			t.Literal = lexer.readIdentifier()
			t.Type = token.LookupIdent(t.Literal)
			return t

		} else if isDigit(lexer.char) {
			// Number literal
			t.Literal = lexer.readNumber()
			if containsDot(t.Literal) {
				t.Type = token.FLOAT
			} else {
				t.Type = token.INT
			}
			return t // readNumber already advanced

		} else {
			// Unknown character
			lexer.errors.Add(errors.LexicalError, lexer.line, lexer.column, 1, "unknown character: %s", string(lexer.char))

			t = lexer.newToken(token.ILLEGAL, lexer.char)
			lexer.advance()
		}
	}

	return t
}
