package lexer

import (
	"fmt"
	"naviary/compiler/token"
)

type Lexer struct {
	input   string
	current int // current position in input
	next    int // next position to read
	char    byte
	line    int
	column  int
}

func New(input string) *Lexer {
	lexer := &Lexer{input: input, line: 1, column: 0}
	lexer.advance()
	return lexer
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

	for isDigit(lexer.char) || lexer.char == '_' {
		if lexer.char != '_' {
			result = append(result, lexer.char)
		} else {
			// start with underscore
			if len(result) == 0 {
				break
			}

			if !isDigit(lexer.peek()) && lexer.peek() != '.' {
				lexer.advance()
				return string(result)
			}
		}

		lexer.advance()
	}

	if lexer.char == '.' && isDigit(lexer.peek()) {
		result = append(result, '.')
		lexer.advance()

		for isDigit(lexer.char) || lexer.char == '_' {
			if lexer.char != '_' {
				result = append(result, lexer.char)
			} else {
				if !isDigit(lexer.peek()) {
					lexer.advance()
					return string(result)
				}
			}
			lexer.advance()
		}
	}

	return string(result)
}

func (lexer *Lexer) readString() string {
	lexer.advance()

	start := lexer.current

	for {
		if lexer.char == '"' {
			// Found closing quote
			result := lexer.input[start:lexer.current]
			lexer.advance() // consume closing quote
			return result
		}

		if lexer.char == '\\' {
			// Handle escape sequences
			lexer.advance() // consume backslash
			if lexer.char != 0 {
				lexer.advance() // consume escaped character
			}
		} else if lexer.char == 0 {
			// Unterminated string (EOF reached)
			return lexer.input[start:lexer.current]
		} else {
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
			char := lexer.char
			lexer.advance()
			lexer.advance()

			t = token.Token{
				Type:    token.PLUS_ASSIGN,
				Literal: string(char) + "=",
				Line:    t.Line,
				Column:  t.Column,
			}
		} else {
			t = lexer.newToken(token.PLUS, lexer.char)
			lexer.advance()
		}
	case '-':
		if lexer.peek() == '=' {
			char := lexer.char
			lexer.advance()
			lexer.advance()

			t = token.Token{
				Type:    token.MINUS_ASSIGN,
				Literal: string(char) + "=",
				Line:    t.Line,
				Column:  t.Column,
			}
		} else {
			t = lexer.newToken(token.MINUS, lexer.char)
			lexer.advance()
		}
	case '*':
		if lexer.peek() == '=' {
			char := lexer.char
			lexer.advance()
			lexer.advance()

			t = token.Token{
				Type:    token.ASTERISK_ASSIGN,
				Literal: string(char) + "=",
				Line:    t.Line,
				Column:  t.Column,
			}
		} else {
			t = lexer.newToken(token.ASTERISK, lexer.char)
			lexer.advance()
		}
	case '/':
		if lexer.peek() == '=' {
			char := lexer.char
			lexer.advance()
			lexer.advance()

			t = token.Token{
				Type:    token.SLASH_ASSIGN,
				Literal: string(char) + "=",
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
			char := lexer.char
			lexer.advance()
			lexer.advance()

			t = token.Token{
				Type:    token.EQUAL,
				Literal: string(char) + "=",
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
			t = lexer.newToken(token.ILLEGAL, lexer.char)
			lexer.advance()
		}
	}

	return t
}
