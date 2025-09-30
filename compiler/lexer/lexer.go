package lexer

import (
	"compiler/errors"
	"compiler/token"
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
func (lexer *Lexer) NextToken() token.Token {
	var t token.Token

	lexer.skipWhitespace()

	// Save current position for token
	t.Line = lexer.line
	t.Column = lexer.column

	switch lexer.currentChar {
	case '=':
		t = token.New(token.ASSIGN, string(lexer.currentChar), lexer.line, lexer.column)
		lexer.advance()
	case '+':
		t = token.New(token.PLUS, string(lexer.currentChar), lexer.line, lexer.column)
		lexer.advance()
	case '-':
		if lexer.peek() == '>' {
			startColumn := lexer.column
			lexer.advance() // consume '-'
			lexer.advance() // consume '>'
			t = token.New(token.ARROW, "->", lexer.line, startColumn)
		} else {
			t = token.New(token.MINUS, string(lexer.currentChar), lexer.line, lexer.column)
			lexer.advance()
		}
	case '*':
		t = token.New(token.ASTERISK, string(lexer.currentChar), lexer.line, lexer.column)
		lexer.advance()
	case '/':
		t = token.New(token.SLASH, string(lexer.currentChar), lexer.line, lexer.column)
		lexer.advance()
	case '(':
		t = token.New(token.LEFT_PAREN, string(lexer.currentChar), lexer.line, lexer.column)
		lexer.advance()
	case ')':
		t = token.New(token.RIGHT_PAREN, string(lexer.currentChar), lexer.line, lexer.column)
		lexer.advance()
	case '{':
		t = token.New(token.LEFT_BRACE, string(lexer.currentChar), lexer.line, lexer.column)
		lexer.advance()
	case '}':
		t = token.New(token.RIGHT_BRACE, string(lexer.currentChar), lexer.line, lexer.column)
		lexer.advance()
	case ',':
		t = token.New(token.COMMA, string(lexer.currentChar), lexer.line, lexer.column)
		lexer.advance()
	case ';':
		t = token.New(token.SEMICOLON, string(lexer.currentChar), lexer.line, lexer.column)
		lexer.advance()
	case ':':
		// Check for := (colon assign)
		if lexer.peek() == '=' {
			startColumn := lexer.column
			lexer.advance() // consume ':'
			lexer.advance() // consume '='
			t = token.New(token.COLON_ASSIGN, ":=", lexer.line, startColumn)
		} else {
			t = token.New(token.COLON, string(lexer.currentChar), lexer.line, lexer.column)
			lexer.advance()
		}
	case '\n':
		t = token.New(token.NEW_LINE, string(lexer.currentChar), lexer.line, lexer.column)
		lexer.advance()
	case 0:
		t.Type = token.EOF
		t.Value = ""
	default:
		if isLetter(lexer.currentChar) {
			t.Value = lexer.readIdentifier()
			t.Type = token.LookupIdentifier(t.Value)
			return t // readIdentifier already advanced position
		} else if isDigit(lexer.currentChar) {
			t.Value = lexer.readNumber()
			t.Type = token.INT
			return t // readNumber already advanced position
		} else {
			t = token.New(token.ILLEGAL, string(lexer.currentChar), lexer.line, lexer.column)
			lexer.errors.Add(
				errors.LexicalError,
				lexer.line,
				lexer.column,
				len(string(lexer.currentChar)),
				"Unexpected character: %s",
				string(lexer.currentChar),
			)
			lexer.advance()
		}
	}

	return t
}

// Tokenize processes the entire input and returns all tokens
func (lexer *Lexer) Tokenize() []token.Token {
	var tokens []token.Token

	for {
		t := lexer.NextToken()
		tokens = append(tokens, t)

		if t.Type == token.EOF {
			break
		}

		// Stop if too many errors
		if lexer.errors.HasErrors() {
			tokens = append(tokens, token.Token{Type: token.EOF, Line: lexer.line, Column: lexer.column})
			break
		}
	}

	return tokens
}

// advances the lexer to the next character
func (lexer *Lexer) advance() {
	// Handle newline based on the character being consumed (previous currentChar)
	if lexer.currentChar == '\n' {
		lexer.line++
		lexer.column = 0
	}

	if lexer.readPosition >= len(lexer.input) {
		lexer.currentChar = 0 // ASCII code for NUL, signifies EOF
	} else {
		lexer.currentChar = lexer.input[lexer.readPosition]
	}

	lexer.position = lexer.readPosition
	lexer.readPosition++
	lexer.column++
}

// skipWhitespace skips spaces, tabs, and newlines
func (lexer *Lexer) skipWhitespace() {
	for lexer.currentChar == ' ' || lexer.currentChar == '\t' || lexer.currentChar == '\r' {
		lexer.advance()
	}
}

// readNumber reads a number from the input
func (lexer *Lexer) readNumber() string {
	startPosition := lexer.position

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
			lexer.line,
			lexer.column,
			len(invalidToken),
			"Invalid number format: %s",
			invalidToken,
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

func (lexer *Lexer) peek() byte {
	if lexer.readPosition >= len(lexer.input) {
		return 0
	}

	return lexer.input[lexer.readPosition]
}
