package parser

import (
	"naviary/compiler/ast"
	"naviary/compiler/errors"
	"naviary/compiler/lexer"
	"naviary/compiler/token"
)

type Parser struct {
	lexer          *lexer.Lexer
	currentToken   token.Token
	peekToken      token.Token
	errorCollector *errors.ErrorCollector
}

func New(lex *lexer.Lexer) *Parser {
	parser := &Parser{lexer: lex, errorCollector: lex.Errors()}

	// Read two tokens to initialize currentToken and peekToken
	parser.advance()
	parser.advance()

	return parser
}

func (parser *Parser) Errors() *errors.ErrorCollector {
	return parser.errorCollector
}

func (parser *Parser) advance() {
	parser.currentToken = parser.peekToken
	parser.peekToken = parser.lexer.NextToken()
}

func (parser *Parser) isCurrentToken(tokenType token.TokenType) bool {
	return parser.currentToken.Type == tokenType
}

func (parser *Parser) isPeekToken(tokenType token.TokenType) bool {
	return parser.peekToken.Type == tokenType
}

func (parser *Parser) expect(tokenType token.TokenType) bool {
	if parser.isCurrentToken(tokenType) {
		return true
	}

	parser.errorCollector.Add(
		errors.SyntaxError,
		parser.currentToken.Line,
		parser.currentToken.Column,
		len(parser.currentToken.Literal),
		"expected %s, got %s",
		tokenType.String(),
		parser.currentToken.Type.String(),
	)

	return false
}

func (parser *Parser) expectPeek(tokenType token.TokenType) bool {
	if parser.isPeekToken(tokenType) {
		return true
	}

	parser.errorCollector.Add(
		errors.SyntaxError,
		parser.peekToken.Line,
		parser.peekToken.Column,
		len(parser.peekToken.Literal),
		"expected %s, got %s",
		tokenType.String(),
		parser.peekToken.Type.String(),
	)

	return false
}

func (parser *Parser) consume(tokenType token.TokenType) bool {
	if !parser.expect(tokenType) {
		return false
	}

	parser.advance()
	return true
}

// Entry point for parsing
// @return root node of the AST
func (parser *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{
		Statements: []ast.Statement{},
	}

	for !parser.isCurrentToken(token.EOF) {
		if parser.isCurrentToken(token.NEWLINE) {
			parser.advance()
			continue
		}

		statement := parser.parseStatement()

		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}

		if !parser.isCurrentToken(token.EOF) && !parser.isCurrentToken(token.NEWLINE) {
			parser.advance()
		}
	}

	return program
}

func (parser *Parser) parseStatement() ast.Statement {
	switch parser.currentToken.Type {
	case token.LET:
		return parser.parseLetStatement()
	default:
		return nil
	}
}

func (parser *Parser) parseLetStatement() ast.Statement {
	letToken := parser.currentToken

	isMutable := false

	if parser.isPeekToken(token.MUT) {
		parser.advance() // consume mut
		isMutable = true
	}

	if !parser.expectPeek(token.IDENT) {
		return nil
	}
	parser.advance()

	name := &ast.Identifier{
		Token: parser.currentToken,
		Value: parser.currentToken.Literal,
	}

	if parser.isPeekToken(token.COLON_ASSIGN) {
		isMutable = true
		parser.advance()
	} else if parser.isPeekToken(token.ASSIGN) {
		parser.advance()
	} else {
		parser.errorCollector.Add(
			errors.SyntaxError,
			parser.peekToken.Line,
			parser.peekToken.Column,
			len(parser.peekToken.Literal),
			"expected := or =, got %s",
			parser.peekToken.Type.String(),
		)
		return nil
	}

	parser.advance()

	value := parser.parseExpression()

	statement := &ast.LetStatement{
		Token:   letToken,
		Name:    name,
		Value:   value,
		Mutable: isMutable,
	}

	if parser.isCurrentToken(token.SEMICOLON) || parser.isCurrentToken(token.NEWLINE) {
		parser.advance()
	}

	return statement
}

// parseExpression is the main entry for expression parsing
func (parser *Parser) parseExpression() ast.Expression {
	return parser.parsePrimary()
}

// parsePrimary handles literals and identifiers
func (parser *Parser) parsePrimary() ast.Expression {
	switch parser.currentToken.Type {
	case token.INT:
		return parser.parseIntegerLiteral()
	case token.FLOAT:
		return parser.parseFloatLiteral()
	case token.STRING:
		return parser.parseStringLiteral()
	case token.TRUE, token.FALSE:
		return parser.parseBooleanLiteral()
	case token.IDENT:
		return parser.parseIdentifier()
	default:
		parser.errorCollector.Add(
			errors.SyntaxError,
			parser.currentToken.Line,
			parser.currentToken.Column,
			len(parser.currentToken.Literal),
			"unexpected token '%s' in expression",
			parser.currentToken.Type.String(),
		)
		return nil
	}
}

func (parser *Parser) parseIntegerLiteral() ast.Expression {
	return &ast.IntegerLiteral{
		Token: parser.currentToken,
		Value: parser.currentToken.Literal,
	}
}

func (parser *Parser) parseFloatLiteral() ast.Expression {
	return &ast.FloatLiteral{
		Token: parser.currentToken,
		Value: parser.currentToken.Literal,
	}
}

func (parser *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{
		Token: parser.currentToken,
		Value: parser.currentToken.Literal,
	}
}

func (parser *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{
		Token: parser.currentToken,
		Value: parser.isCurrentToken(token.TRUE),
	}
}

func (parser *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{
		Token: parser.currentToken,
		Value: parser.currentToken.Literal,
	}
}
