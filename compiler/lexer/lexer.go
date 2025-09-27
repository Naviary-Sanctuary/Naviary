package lexer

import (
	"compiler/errors"
)

// Lexer tokenizes the input source code
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	currentChar  byte // current char under examination
	line         int
	column       int
	errors       *errors.ErrorCollector
	fileName     string
}

// New creates a new Lexer instance
func New(input string, fileName string, errorCollector *errors.ErrorCollector) *Lexer {
	lexer := &Lexer{
		input:    input,
		line:     1,
		column:   0,
		errors:   errorCollector, // 외부에서 받음
		fileName: fileName,
	}
	lexer.advance()
	return lexer
}

// NextToken returns the next token from the input
func (lexer *Lexer) NextToken() Token {
	var token Token

	lexer.skipWhitespace()

	// Save current position for token
	token.Line = lexer.line
	token.Column = lexer.column

	switch lexer.currentChar {
	case '=':
		token = lexer.newToken(Assign, lexer.currentChar)
		lexer.advance()
	case '+':
		token = lexer.newToken(Plus, lexer.currentChar)
		lexer.advance()
	case '-':
		token = lexer.newToken(Minus, lexer.currentChar)
		lexer.advance()
	case '*':
		token = lexer.newToken(Asterisk, lexer.currentChar)
		lexer.advance()
	case '/':
		token = lexer.newToken(Slash, lexer.currentChar)
		lexer.advance()
	case '(':
		token = lexer.newToken(LeftParen, lexer.currentChar)
		lexer.advance()
	case ')':
		token = lexer.newToken(RightParen, lexer.currentChar)
		lexer.advance()
	case '{':
		token = lexer.newToken(LeftBrace, lexer.currentChar)
		lexer.advance()
	case '}':
		token = lexer.newToken(RightBrace, lexer.currentChar)
		lexer.advance()
	case 0:
		token.Type = EOF
		token.Literal = ""
	default:
		if isLetter(lexer.currentChar) {
			token.Literal = lexer.readIdentifier()
			token.Type = LookupIdentifier(token.Literal)
			return token // readIdentifier already advanced position
		} else if isDigit(lexer.currentChar) {
			token.Literal = lexer.readNumber()
			token.Type = Number
			return token // readNumber already advanced position
		} else {
			token = lexer.newToken(Illegal, lexer.currentChar)
			lexer.errors.Add(
				errors.LexicalError,
				"Unexpected character: "+string(lexer.currentChar),
				lexer.line,
				lexer.column,
				lexer.fileName,
			)
			lexer.advance()
		}
	}

	return token
}

// Tokenize processes the entire input and returns all tokens
func (lexer *Lexer) Tokenize() []Token {
	var tokens []Token

	for {
		token := lexer.NextToken()
		tokens = append(tokens, token)

		if token.Type == EOF {
			break
		}

		// Stop if too many errors
		if lexer.errors.Count() > 10 {
			tokens = append(tokens, Token{Type: EOF, Line: lexer.line, Column: lexer.column})
			break
		}
	}

	return tokens
}

// newToken creates a new token with the given type and character
func (lexer *Lexer) newToken(tokenType TokenType, char byte) Token {
	return Token{
		Type:    tokenType,
		Literal: string(char),
		Line:    lexer.line,
		Column:  lexer.column,
	}
}

// GetErrors returns the accumulated errors
func (lexer *Lexer) GetErrors() *errors.ErrorCollector {
	return lexer.errors
}

// advances the lexer to the next character
func (lexer *Lexer) advance() {
	if lexer.readPosition >= len(lexer.input) {
		lexer.currentChar = 0 // ASCII code for NUL, signifies EOF
	} else {
		lexer.currentChar = lexer.input[lexer.readPosition]
	}

	lexer.position = lexer.readPosition
	lexer.readPosition++
	lexer.column++

	// Handle newline
	if lexer.currentChar == '\n' {
		lexer.line++
		lexer.column = 0
	}
}

// peekChar looks at the next character without advancing
func (lexer *Lexer) peekChar() byte {
	if lexer.readPosition >= len(lexer.input) {
		return 0
	}
	return lexer.input[lexer.readPosition]
}

// skipWhitespace skips spaces, tabs, and newlines
func (lexer *Lexer) skipWhitespace() {
	for lexer.currentChar == ' ' || lexer.currentChar == '\t' ||
		lexer.currentChar == '\n' || lexer.currentChar == '\r' {
		lexer.advance()
	}
}

// isLetter checks if a character can start an identifier
func isLetter(char byte) bool {
	return ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '_'
}

// isDigit checks if a character is a digit
func isDigit(char byte) bool {
	return '0' <= char && char <= '9'
}

// readNumber reads a number from the input
func (lexer *Lexer) readNumber() string {
	startPosition := lexer.position
	startColumn := lexer.column

	// Read all consecutive digits
	for isDigit(lexer.currentChar) {
		lexer.advance()
	}

	// Check for invalid number format (e.g., 123abc)
	if isLetter(lexer.currentChar) {
		// Continue reading to capture the full invalid token
		for isLetter(lexer.currentChar) || isDigit(lexer.currentChar) {
			lexer.advance()
		}
		invalidToken := lexer.input[startPosition:lexer.position]
		lexer.errors.Add(
			errors.LexicalError,
			"Invalid number format: "+invalidToken,
			lexer.line,
			startColumn,
			lexer.fileName,
		)
		return invalidToken
	}

	return lexer.input[startPosition:lexer.position]
}

// readIdentifier reads an identifier or keyword from the input
func (lexer *Lexer) readIdentifier() string {
	startPosition := lexer.position

	// First character must be a letter or underscore
	if !isLetter(lexer.currentChar) {
		return ""
	}

	// Read first character
	lexer.advance()

	// Continue reading letters, digits, and underscores
	for isLetter(lexer.currentChar) || isDigit(lexer.currentChar) {
		lexer.advance()
	}

	return lexer.input[startPosition:lexer.position]
}
