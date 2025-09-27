package parser

import (
	"compiler/errors"
	"compiler/lexer"
	"fmt"
	"strconv"
)

// Parser analyzes tokens and builds an AST
type Parser struct {
	lexer        *lexer.Lexer
	errors       *errors.ErrorCollector
	currentToken lexer.Token
	peekToken    lexer.Token
	fileName     string
}

// New creates a new Parser instance
func New(lexerInstance *lexer.Lexer, fileName string, errorCollector *errors.ErrorCollector) *Parser {
	parser := &Parser{
		lexer:    lexerInstance,
		errors:   errorCollector,
		fileName: fileName,
	}

	// Read two tokens to set currentToken and peekToken
	parser.nextToken()
	parser.nextToken()

	return parser
}

// ParseProgram parses the entire program
func (parser *Parser) ParseProgram() *Program {
	program := &Program{
		Functions: []FunctionDeclaration{},
	}

	// Parse all functions until EOF
	for !parser.currentTokenIs(lexer.EOF) {
		// Only function declarations are allowed at top level
		if parser.currentTokenIs(lexer.Func) {
			function := parser.parseFunctionDeclaration()
			if function != nil {
				program.Functions = append(program.Functions, *function)
			}
		} else {
			parser.errors.Add(
				errors.SyntaxError,
				fmt.Sprintf("Expected 'func' at top level, got %s", parser.currentToken.Literal),
				parser.currentToken.Line,
				parser.currentToken.Column,
				parser.fileName,
			)
			parser.nextToken() // Skip invalid token to continue parsing
		}
	}

	// Validate that we have exactly one main function for MVP
	if !parser.validateMainFunction(program) {
		return nil
	}

	return program
}

// parseFunctionDeclaration parses 'func name() { body }'
func (parser *Parser) parseFunctionDeclaration() *FunctionDeclaration {
	function := &FunctionDeclaration{
		Token:      parser.currentToken,
		Parameters: []string{}, // Empty for MVP
	}

	// Expect function name
	if !parser.expectPeek(lexer.Identifier) {
		parser.errors.Add(
			errors.SyntaxError,
			"Expected function name after 'func'",
			parser.peekToken.Line,
			parser.peekToken.Column,
			parser.fileName,
		)
		return nil
	}

	function.Name = parser.currentToken.Literal

	// Expect '(' for parameter list
	if !parser.expectPeek(lexer.LeftParen) {
		parser.errors.Add(
			errors.SyntaxError,
			fmt.Sprintf("Expected '(' after function name '%s'", function.Name),
			parser.peekToken.Line,
			parser.peekToken.Column,
			parser.fileName,
		)
		return nil
	}

	// Parse parameters (empty for MVP)
	if !parser.parseFunctionParameters(function) {
		return nil
	}

	// Expect '{' for function body
	if !parser.expectPeek(lexer.LeftBrace) {
		parser.errors.Add(
			errors.SyntaxError,
			fmt.Sprintf("Expected '{{' to start function body of '%s'", function.Name),
			parser.peekToken.Line,
			parser.peekToken.Column,
			parser.fileName,
		)
		return nil
	}

	// Parse function body
	function.Body = *parser.parseBlockStatement()

	// Move past the closing '}'
	parser.nextToken()

	return function
}

// parseFunctionParameters parses the parameter list (empty for MVP)
func (parser *Parser) parseFunctionParameters(function *FunctionDeclaration) bool {
	// For MVP, we only support empty parameter list
	if !parser.peekTokenIs(lexer.RightParen) {
		parser.errors.Add(
			errors.SyntaxError,
			"MVP only supports functions without parameters",
			parser.peekToken.Line,
			parser.peekToken.Column,
			parser.fileName,
		)
		return false
	}

	parser.nextToken() // consume ')'
	return true
}

// validateMainFunction checks if program has exactly one main function
func (parser *Parser) validateMainFunction(program *Program) bool {
	mainCount := 0

	for _, function := range program.Functions {
		if function.Name == "main" {
			mainCount++
		}
	}

	if mainCount == 0 {
		parser.errors.Add(
			errors.SyntaxError,
			"Program must have a 'main' function",
			1,
			1,
			parser.fileName,
		)
		return false
	}

	if mainCount > 1 {
		parser.errors.Add(
			errors.SyntaxError,
			"Program cannot have multiple 'main' functions",
			1,
			1,
			parser.fileName,
		)
		return false
	}

	return true
}

// nextToken advances to the next token
func (parser *Parser) nextToken() {
	parser.currentToken = parser.peekToken
	parser.peekToken = parser.lexer.NextToken()
}

// currentTokenIs checks if current token is of given type
func (parser *Parser) currentTokenIs(tokenType lexer.TokenType) bool {
	return parser.currentToken.Type == tokenType
}

// peekTokenIs checks if peek token is of given type
func (parser *Parser) peekTokenIs(tokenType lexer.TokenType) bool {
	return parser.peekToken.Type == tokenType
}

// expectPeek advances if peek token matches, otherwise adds error
func (parser *Parser) expectPeek(tokenType lexer.TokenType) bool {
	if parser.peekTokenIs(tokenType) {
		parser.nextToken()
		return true
	}

	parser.peekError(tokenType)
	return false
}

// peekError adds an error for unexpected token
func (parser *Parser) peekError(expectedType lexer.TokenType) {
	message := fmt.Sprintf("Expected %v, got %v",
		expectedType,
		parser.peekToken.Type)

	parser.errors.Add(
		errors.SyntaxError,
		message,
		parser.peekToken.Line,
		parser.peekToken.Column,
		parser.fileName,
	)
}

// GetErrors returns accumulated parsing errors
func (parser *Parser) GetErrors() *errors.ErrorCollector {
	return parser.errors
}

// parsePrimaryExpression parses the simplest expressions: numbers, identifiers, grouped expressions

func (parser *Parser) parsePrimaryExpression() Expression {
	switch parser.currentToken.Type {
	case lexer.Number:
		return parser.parseNumberLiteral()
	case lexer.Identifier:
		return parser.parseIdentifier()
	case lexer.Print: // print도 identifier처럼 처리
		// print는 특별한 built-in 함수
		if parser.peekTokenIs(lexer.LeftParen) {
			return &CallExpression{
				Token:     parser.currentToken,
				Function:  "print",
				Arguments: parser.parseCallArgumentsAfterIdentifier(),
			}
		}
		// print 다음에 (가 없으면 에러
		parser.errors.Add(
			errors.SyntaxError,
			"Expected '(' after 'print'",
			parser.peekToken.Line,
			parser.peekToken.Column,
			parser.fileName,
		)
		return nil
	case lexer.LeftParen:
		return parser.parseGroupedExpression()
	default:
		parser.errors.Add(
			errors.SyntaxError,
			fmt.Sprintf("Unexpected token in expression: %s", parser.currentToken.Literal),
			parser.currentToken.Line,
			parser.currentToken.Column,
			parser.fileName,
		)
		return nil
	}
}

// parseNumberLiteral parses an integer literal
func (parser *Parser) parseNumberLiteral() Expression {
	literal := &NumberLiteral{Token: parser.currentToken}

	value, err := strconv.ParseInt(parser.currentToken.Literal, 10, 64)
	if err != nil {
		parser.errors.Add(
			errors.SyntaxError,
			fmt.Sprintf("Could not parse %q as integer", parser.currentToken.Literal),
			parser.currentToken.Line,
			parser.currentToken.Column,
			parser.fileName,
		)
		return nil
	}

	literal.Value = value
	return literal
}

// parseIdentifier parses a variable name or function call
func (parser *Parser) parseIdentifier() Expression {
	// Save current identifier token
	identToken := parser.currentToken

	// Check if this is a function call (including 'print')
	if parser.peekTokenIs(lexer.LeftParen) {
		// It's a function call
		return &CallExpression{
			Token:     identToken,
			Function:  identToken.Literal,
			Arguments: parser.parseCallArgumentsAfterIdentifier(),
		}
	}

	// Just a regular identifier
	return &Identifier{
		Token: identToken,
		Value: identToken.Literal,
	}
}

// parseCallArgumentsAfterIdentifier parses arguments after we've seen identifier(
func (parser *Parser) parseCallArgumentsAfterIdentifier() []Expression {
	parser.nextToken() // consume the identifier to get to '('

	if !parser.currentTokenIs(lexer.LeftParen) {
		return nil
	}

	arguments := []Expression{}

	// Empty argument list
	if parser.peekTokenIs(lexer.RightParen) {
		parser.nextToken() // consume ')'
		return arguments
	}

	parser.nextToken() // move to first argument
	arguments = append(arguments, parser.parseExpression())

	// For MVP, we only support single argument

	if !parser.expectPeek(lexer.RightParen) {
		return nil
	}

	return arguments
}

// parseGroupedExpression parses expressions in parentheses: (expression)
func (parser *Parser) parseGroupedExpression() Expression {
	parser.nextToken() // consume '('

	expression := parser.parseExpression()

	if !parser.expectPeek(lexer.RightParen) {
		return nil
	}

	return expression
}

// parseCallExpression parses function calls like print(x)
func (parser *Parser) parseCallExpression() Expression {
	call := &CallExpression{
		Token:    parser.currentToken,
		Function: parser.currentToken.Literal,
	}

	if !parser.expectPeek(lexer.LeftParen) {
		return nil
	}

	call.Arguments = parser.parseCallArguments()

	return call
}

// parseCallArguments parses the argument list of a function call
func (parser *Parser) parseCallArguments() []Expression {
	arguments := []Expression{}

	// Empty argument list
	if parser.peekTokenIs(lexer.RightParen) {
		parser.nextToken()
		return arguments
	}

	parser.nextToken() // move to first argument
	arguments = append(arguments, parser.parseExpression())

	// For MVP, we only support single argument
	// Later we'll add comma-separated arguments

	if !parser.expectPeek(lexer.RightParen) {
		return nil
	}

	return arguments
}

// Operator precedence levels
const (
	LOWEST         = 1
	ADDITIVE       = 2 // + -
	MULTIPLICATIVE = 3 // * /
)

// precedenceMap defines operator precedence
var precedenceMap = map[lexer.TokenType]int{
	lexer.Plus:     ADDITIVE,
	lexer.Minus:    ADDITIVE,
	lexer.Asterisk: MULTIPLICATIVE,
	lexer.Slash:    MULTIPLICATIVE,
}

// peekPrecedence returns the precedence of the peek token
func (parser *Parser) peekPrecedence() int {
	if precedence, ok := precedenceMap[parser.peekToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

// currentPrecedence returns the precedence of the current token
func (parser *Parser) currentPrecedence() int {
	if precedence, ok := precedenceMap[parser.currentToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

// parseExpression parses expressions with operator precedence (Pratt parsing)
func (parser *Parser) parseExpression() Expression {
	return parser.parseExpressionWithPrecedence(LOWEST)
}

// parseExpressionWithPrecedence implements Pratt parser algorithm
func (parser *Parser) parseExpressionWithPrecedence(precedence int) Expression {
	// Parse left side (prefix/primary expression)
	left := parser.parsePrimaryExpression()
	if left == nil {
		return nil
	}

	// Keep parsing while we have higher precedence operators
	for !parser.peekTokenIs(lexer.EOF) && precedence < parser.peekPrecedence() {
		// Check if next token is a binary operator
		if !parser.isBinaryOperator(parser.peekToken.Type) {
			return left
		}

		parser.nextToken()
		left = parser.parseBinaryExpression(left)
		if left == nil {
			return nil
		}
	}

	return left
}

// isBinaryOperator checks if token is a binary operator
func (parser *Parser) isBinaryOperator(tokenType lexer.TokenType) bool {
	return tokenType == lexer.Plus ||
		tokenType == lexer.Minus ||
		tokenType == lexer.Asterisk ||
		tokenType == lexer.Slash
}

// parseBinaryExpression parses binary operations like a + b
func (parser *Parser) parseBinaryExpression(left Expression) Expression {
	expression := &BinaryExpression{
		Token:    parser.currentToken,
		Operator: parser.currentToken.Literal,
		Left:     left,
	}

	precedence := parser.currentPrecedence()
	parser.nextToken()

	// Parse right side with higher precedence
	// This ensures correct associativity
	expression.Right = parser.parseExpressionWithPrecedence(precedence)

	if expression.Right == nil {
		return nil
	}

	return expression
}

// parseStatement parses different types of statements
func (parser *Parser) parseStatement() Statement {
	switch parser.currentToken.Type {
	case lexer.Let:
		return parser.parseLetStatement()
	case lexer.LeftBrace:
		return parser.parseBlockStatement()
	default:
		return parser.parseExpressionStatement()
	}
}

// parseLetStatement parses 'let identifier = expression'
func (parser *Parser) parseLetStatement() Statement {
	statement := &LetStatement{Token: parser.currentToken}

	// Expect identifier after 'let'
	if !parser.expectPeek(lexer.Identifier) {
		parser.errors.Add(
			errors.SyntaxError,
			"Expected identifier after 'let'",
			parser.peekToken.Line,
			parser.peekToken.Column,
			parser.fileName,
		)
		return nil
	}

	statement.Name = parser.currentToken.Literal

	// Expect '=' after identifier
	if !parser.expectPeek(lexer.Assign) {
		parser.errors.Add(
			errors.SyntaxError,
			fmt.Sprintf("Expected '=' after identifier '%s'", statement.Name),
			parser.peekToken.Line,
			parser.peekToken.Column,
			parser.fileName,
		)
		return nil
	}

	// Move to the expression
	parser.nextToken()

	// Parse the value expression
	statement.Value = parser.parseExpression()
	if statement.Value == nil {
		return nil
	}

	return statement
}

// parseBlockStatement parses { statements... }
func (parser *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{
		Token:      parser.currentToken,
		Statements: []Statement{},
	}

	parser.nextToken() // consume '{'

	// Parse statements until we hit '}' or EOF
	for !parser.currentTokenIs(lexer.RightBrace) && !parser.currentTokenIs(lexer.EOF) {
		statement := parser.parseStatement()
		if statement != nil {
			block.Statements = append(block.Statements, statement)
		}
		parser.nextToken()
	}

	// Check if we properly closed the block
	if !parser.currentTokenIs(lexer.RightBrace) {
		parser.errors.Add(
			errors.SyntaxError,
			"Expected '}' to close block",
			parser.currentToken.Line,
			parser.currentToken.Column,
			parser.fileName,
		)
		return nil
	}

	return block
}

// parseExpressionStatement parses an expression used as a statement
func (parser *Parser) parseExpressionStatement() Statement {
	statement := &ExpressionStatement{
		Token: parser.currentToken,
	}

	statement.Expression = parser.parseExpression()
	if statement.Expression == nil {
		return nil
	}

	return statement
}
