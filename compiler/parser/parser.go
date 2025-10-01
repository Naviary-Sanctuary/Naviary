package parser

import (
	"compiler/ast"
	"compiler/errors"
	"compiler/lexer"
	"compiler/token"
)

// Parser analyzes tokens and builds an AST
type Parser struct {
	lexer          *lexer.Lexer
	currentToken   token.Token
	peekToken      token.Token
	errorCollector *errors.ErrorCollector
}

func New(lexer *lexer.Lexer, errorCollector *errors.ErrorCollector) *Parser {
	parser := &Parser{
		lexer:          lexer,
		errorCollector: errorCollector,
	}

	parser.advance()
	parser.advance()

	return parser
}

func (parser *Parser) advance() {
	parser.currentToken = parser.peekToken
	parser.peekToken = parser.lexer.NextToken()
}

func (parser *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{
		Statements: []ast.Statement{},
	}

	for parser.currentToken.Type != token.EOF {
		if parser.currentToken.Type == token.NEW_LINE {
			parser.advance()
			continue
		}

		statement := parser.parseStatement()

		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}

		if parser.currentToken.Type != token.EOF && parser.currentToken.Type != token.NEW_LINE {
			parser.advance()
		}
	}

	return program
}

func (parser *Parser) parseStatement() ast.Statement {
	switch parser.currentToken.Type {
	case token.LET:
		return parser.parseLetStatement()
	case token.FUNC:
		return parser.parseFunctionStatement()
	case token.RETURN:
		return parser.parseReturnStatement()
	case token.IDENTIFIER:
		return parser.parseExpressionStatement()
	default:
		return nil
	}
}

func (parser *Parser) parseLetStatement() ast.Statement {
	letToken := parser.currentToken

	isMutable := false

	if parser.peekToken.Type == token.MUT {
		parser.advance() // advance to mut
		isMutable = true
	}

	parser.advance() // advance to identifier

	name := &ast.Identifier{
		Token: parser.currentToken,
		Value: parser.currentToken.Value,
	}

	parser.advance() // consume identifier

	var typeAnnotation *ast.TypeAnnotation
	if parser.currentToken.Type == token.COLON {

		typeAnnotation = parser.parseTypeAnnotation()

		if typeAnnotation == nil {
			return nil
		}
	}

	switch parser.currentToken.Type {
	case token.COLON_ASSIGN:
		isMutable = true
		parser.advance()
	case token.ASSIGN:
		parser.advance()
	default:
		parser.errorCollector.Add(errors.SyntaxError, parser.peekToken.Line, parser.peekToken.Column, len(parser.peekToken.Value), "Expected := or =, got %s", parser.peekToken.Type.String())
		return nil
	}

	value := parser.parseExpression(LOWEST)

	statement := &ast.LetStatement{
		Token:          letToken,
		Name:           name,
		Value:          value,
		TypeAnnotation: typeAnnotation,
		Mutable:        isMutable,
	}

	parser.skipEndOfStatement()

	return statement
}

func (parser *Parser) parseFunctionStatement() ast.Statement {
	function := &ast.FunctionStatement{
		Token: parser.currentToken,
	}

	if !parser.expectPeek(token.IDENTIFIER) {
		return nil
	}

	parser.advance() // consume func

	function.Name = &ast.Identifier{
		Token: parser.currentToken,
		Value: parser.currentToken.Value,
	}

	parser.advance() // consume function name
	if !parser.expect(token.LEFT_PAREN) {
		return nil
	}

	function.Parameters = parser.parseFunctionParameters()

	if parser.peekToken.Type == token.ARROW {
		parser.advance()
		parser.advance() // consume '->'

		if !parser.expect(token.IDENTIFIER) {
			return nil
		}

		function.ReturnType = &ast.TypeAnnotation{
			Token: parser.currentToken,
			Value: parser.currentToken.Value,
		}

		parser.advance() // consume return type
	}

	if !parser.expect(token.LEFT_BRACE) {
		return nil
	}

	function.Body = parser.parseBlockStatement()

	if function.Body == nil {
		return nil
	}

	return function
}

func (parser *Parser) parseFunctionParameters() []*ast.FunctionParameter {
	parameters := []*ast.FunctionParameter{}

	if parser.peekToken.Type == token.RIGHT_PAREN {
		parser.advance()
		parser.advance() // consume '()'
		return parameters
	}

	parser.advance() // consume '('

	for {
		if !parser.expect(token.IDENTIFIER) {
			return nil
		}

		parameter := &ast.FunctionParameter{
			Name: &ast.Identifier{
				Token: parser.currentToken,
				Value: parser.currentToken.Value,
			},
		}

		parser.advance() // consume parameter name

		parameterType := parser.parseTypeAnnotation()
		if parameterType == nil {
			return nil
		}

		parameter.Type = *parameterType
		parameters = append(parameters, parameter)

		if parser.currentToken.Type == token.COMMA {
			parser.advance() // consume comma
			continue
		}
		if !parser.expect(token.RIGHT_PAREN) {
			return nil
		}

		break
	}

	parser.advance() // consume ')'

	return parameters
}

func (parser *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{
		Token: parser.currentToken,
	}

	parser.advance() // consume '{'

	for parser.currentToken.Type != token.RIGHT_BRACE && parser.currentToken.Type != token.EOF {
		if parser.currentToken.Type == token.NEW_LINE {
			parser.advance()
			continue
		}

		statement := parser.parseStatement()
		if statement != nil {
			block.Statements = append(block.Statements, statement)
		}

		if parser.currentToken.Type != token.RIGHT_BRACE && parser.currentToken.Type != token.EOF {
			parser.advance()
		}
	}

	if !parser.expect(token.RIGHT_BRACE) {
		return nil
	}

	parser.advance() // consume '}'
	return block
}

func (parser *Parser) parseExpressionStatement() ast.Statement {
	statement := &ast.ExpressionStatement{
		Token:      parser.currentToken,
		Expression: parser.parseExpression(LOWEST),
	}

	parser.skipEndOfStatement()

	return statement
}

func (parser *Parser) parseReturnStatement() ast.Statement {
	returnStatement := &ast.ReturnStatement{
		Token: parser.currentToken,
	}

	parser.advance()

	returnStatement.ReturnValue = parser.parseExpression(LOWEST)

	parser.skipEndOfStatement()

	return returnStatement

}

func (parser *Parser) parseExpression(precedence int) ast.Expression {
	left := parser.parseAtom()
	if left == nil {
		return nil
	}

	for !parser.isStatementEnd() && precedence < getPrecedence(parser.peekToken.Type) {

		if parser.peekToken.Type == token.LEFT_PAREN {
			parser.advance() // advance to '('
			left = parser.parseCallExpression(left)
			continue
		}

		if !parser.peekToken.Type.IsOperator() {
			break
		}

		operatorToken := parser.peekToken
		operatorPrecedence := getPrecedence(operatorToken.Type)

		parser.advance()
		parser.advance() // advance to right operand

		right := parser.parseExpression(operatorPrecedence)
		if right == nil {
			return nil
		}

		left = &ast.BinaryExpression{
			Token:    operatorToken,
			Left:     left,
			Operator: operatorToken.Value,
			Right:    right,
		}
	}

	return left
}

func (parser *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	call := &ast.CallExpression{
		Token:     parser.currentToken,
		Function:  function,
		Arguments: []ast.Expression{},
	}

	call.Arguments = parser.parseCallArguments()

	return call
}

func (parser *Parser) parseCallArguments() []ast.Expression {
	arguments := []ast.Expression{}

	if parser.peekToken.Type == token.RIGHT_PAREN {
		parser.advance() // consume ')'
		return arguments
	}

	parser.advance() // consume '('
	arguments = append(arguments, parser.parseExpression(LOWEST))

	for parser.peekToken.Type == token.COMMA {
		parser.advance() // consume argument
		parser.advance() // consume comma

		arguments = append(arguments, parser.parseExpression(LOWEST))
	}
	if !parser.expectPeek(token.RIGHT_PAREN) {
		return nil
	}

	parser.advance() // consume ')'

	return arguments
}

func (parser *Parser) skipEndOfStatement() {
	if parser.peekToken.Type == token.SEMICOLON || parser.peekToken.Type == token.NEW_LINE {
		parser.advance()
	}
}

// parseAtom parses an literals and identifiers
func (parser *Parser) parseAtom() ast.Expression {
	switch parser.currentToken.Type {
	case token.INT_LITERAL:
		return &ast.IntegerLiteral{
			Token: parser.currentToken,
			Value: parser.currentToken.Value,
		}
	case token.IDENTIFIER:
		return &ast.Identifier{
			Token: parser.currentToken,
			Value: parser.currentToken.Value,
		}
	default:
		parser.errorCollector.Add(errors.SyntaxError,
			parser.currentToken.Line,
			parser.currentToken.Column,
			len(parser.currentToken.Value),
			"Unexpected token '%s' in expression",
			parser.currentToken.Type.String(),
		)
		return nil
	}
}

func (parser *Parser) isStatementEnd() bool {
	if parser.currentToken.Type == token.SEMICOLON || parser.currentToken.Type == token.NEW_LINE || parser.currentToken.Type == token.EOF {
		return true
	}

	return false
}

func (parser *Parser) expectPeek(tokenType token.TokenType) bool {
	if parser.peekToken.Type == tokenType {
		return true
	}

	parser.errorCollector.Add(
		errors.SyntaxError,
		parser.peekToken.Line,
		parser.peekToken.Column,
		len(parser.peekToken.Value),
		"expected %s, got %s",
		tokenType.String(),
		parser.peekToken.Type.String(),
	)

	return false
}

func (parser *Parser) expect(tokenType token.TokenType) bool {
	if parser.currentToken.Type == tokenType {
		return true
	}

	parser.errorCollector.Add(errors.SyntaxError,
		parser.currentToken.Line,
		parser.currentToken.Column,
		len(parser.currentToken.Value),
		"expected %s, got %s",
		tokenType.String(),
		parser.currentToken.Type.String(),
	)
	return false
}

func (parser *Parser) parseTypeAnnotation() *ast.TypeAnnotation {
	if !parser.expect(token.COLON) {
		return nil
	}

	parser.advance() // consume ':'

	switch parser.currentToken.Type {
	case token.INT, token.FLOAT, token.STRING, token.BOOL, token.IDENTIFIER:

		typeAnnotation := &ast.TypeAnnotation{
			Token: parser.currentToken,
			Value: parser.currentToken.Value,
		}

		parser.advance() // consume type

		return typeAnnotation
	default:
		return nil
	}

}
