package codegen

import (
	"compiler/errors"
	"compiler/parser"
	"fmt"
	"strings"
)

// ErlangGenerator generates Erlang source code from AST
type ErlangGenerator struct {
	errors     *errors.ErrorCollector
	fileName   string
	moduleName string
	indent     int // Current indentation level
}

// New creates a new Erlang code generator
func New(fileName string, errorCollector *errors.ErrorCollector) *ErlangGenerator {
	// Convert filename to valid Erlang module name
	// Remove .navi extension and convert to lowercase
	moduleName := strings.TrimSuffix(fileName, ".navi")
	moduleName = strings.ToLower(moduleName)
	moduleName = strings.ReplaceAll(moduleName, "-", "_")

	return &ErlangGenerator{
		errors:     errorCollector,
		fileName:   fileName,
		moduleName: moduleName,
		indent:     0,
	}
}

// indentString returns the current indentation as spaces
func (generator *ErlangGenerator) indentString() string {
	return strings.Repeat("    ", generator.indent)
}

// GetErrors returns accumulated errors
func (generator *ErlangGenerator) GetErrors() *errors.ErrorCollector {
	return generator.errors
}

// generateModuleHeader generates Erlang module declaration and exports
func (generator *ErlangGenerator) generateModuleHeader() string {
	var builder strings.Builder

	// Module declaration
	builder.WriteString(fmt.Sprintf("-module(%s).\n", generator.moduleName))

	// Export main function as start/0 for OTP compatibility
	builder.WriteString("-export([start/0]).\n\n")

	return builder.String()
}

// generateModuleFooter generates any necessary module footer code
func (generator *ErlangGenerator) generateModuleFooter() string {
	// For MVP, no footer needed
	return ""
}

// generateExpression converts an AST expression to Erlang code
func (generator *ErlangGenerator) generateExpression(expression parser.Expression) string {
	if expression == nil {
		return "undefined"
	}

	switch node := expression.(type) {
	case *parser.IntegerLiteral:
		return generator.generateIntegerLiteral(node)
	case *parser.Identifier:
		return generator.generateIdentifier(node)
	case *parser.BinaryExpression:
		return generator.generateBinaryExpression(node)
	case *parser.CallExpression:
		return generator.generateCallExpression(node)
	default:
		generator.errors.Add(
			errors.CodeGenerationError,
			fmt.Sprintf("Unknown expression type: %T", expression),
			0, 0, generator.fileName,
		)
		return "undefined"
	}
}

// generateIntegerLiteral converts a integer literal to Erlang
func (generator *ErlangGenerator) generateIntegerLiteral(integer *parser.IntegerLiteral) string {
	return fmt.Sprintf("%d", integer.Value)
}

// generateIdentifier converts an identifier to Erlang (uppercase for variables)
func (generator *ErlangGenerator) generateIdentifier(identifier *parser.Identifier) string {
	// Erlang variables must start with uppercase
	return generator.naviaryToErlangVariable(identifier.Value)
}

// naviaryToErlangVariable converts Naviary variable names to Erlang format
func (generator *ErlangGenerator) naviaryToErlangVariable(name string) string {
	if len(name) == 0 {
		return "Undefined"
	}

	// Capitalize first letter for Erlang
	firstChar := strings.ToUpper(string(name[0]))
	if len(name) == 1 {
		return firstChar
	}

	return firstChar + name[1:]
}

// generateBinaryExpression converts binary operations to Erlang
func (generator *ErlangGenerator) generateBinaryExpression(binary *parser.BinaryExpression) string {
	left := generator.generateExpression(binary.Left)
	right := generator.generateExpression(binary.Right)

	// Map Naviary operators to Erlang operators
	var operator string
	switch binary.Operator {
	case "+":
		operator = "+"
	case "-":
		operator = "-"
	case "*":
		operator = "*"
	case "/":
		operator = "div" // Integer division in Erlang
	default:
		generator.errors.Add(
			errors.CodeGenerationError,
			fmt.Sprintf("Unknown operator: %s", binary.Operator),
			0, 0, generator.fileName,
		)
		operator = "+"
	}

	// Special handling for division
	if binary.Operator == "/" {
		return fmt.Sprintf("(%s %s %s)", left, operator, right)
	}

	return fmt.Sprintf("(%s %s %s)", left, operator, right)
}

// generateCallExpression converts function calls to Erlang
func (generator *ErlangGenerator) generateCallExpression(call *parser.CallExpression) string {
	// Special handling for built-in functions
	if call.Function == "print" {
		return generator.generatePrintCall(call)
	}

	// Regular function calls (not used in MVP)
	arguments := []string{}
	for _, arg := range call.Arguments {
		arguments = append(arguments, generator.generateExpression(arg))
	}

	if len(arguments) == 0 {
		return fmt.Sprintf("%s()", call.Function)
	}

	return fmt.Sprintf("%s(%s)", call.Function, strings.Join(arguments, ", "))
}

// generatePrintCall handles the special 'print' built-in
func (generator *ErlangGenerator) generatePrintCall(call *parser.CallExpression) string {
	if len(call.Arguments) == 0 {
		return "io:format(\"~n\")"
	}

	// For MVP, assume single argument
	argument := generator.generateExpression(call.Arguments[0])
	return fmt.Sprintf("io:format(\"~p~n\", [%s])", argument)
}

// generateStatement converts an AST statement to Erlang code
func (generator *ErlangGenerator) generateStatement(statement parser.Statement) string {
	if statement == nil {
		return ""
	}

	switch node := statement.(type) {
	case *parser.LetStatement:
		return generator.generateLetStatement(node)
	case *parser.BlockStatement:
		return generator.generateBlockStatement(node)
	case *parser.ExpressionStatement:
		return generator.generateExpressionStatement(node)
	default:
		generator.errors.Add(
			errors.CodeGenerationError,
			fmt.Sprintf("Unknown statement type: %T", statement),
			0, 0, generator.fileName,
		)
		return ""
	}
}

// generateLetStatement converts 'let x = expr' to Erlang variable binding
func (generator *ErlangGenerator) generateLetStatement(letStmt *parser.LetStatement) string {
	variable := generator.naviaryToErlangVariable(letStmt.Name)
	value := generator.generateExpression(letStmt.Value)

	return fmt.Sprintf("%s%s = %s",
		generator.indentString(),
		variable,
		value)
}

// generateBlockStatement converts a block of statements to Erlang
func (generator *ErlangGenerator) generateBlockStatement(block *parser.BlockStatement) string {
	var builder strings.Builder

	// Generate each statement
	for i, stmt := range block.Statements {
		code := generator.generateStatement(stmt)
		if code != "" {
			builder.WriteString(code)

			// Add comma between statements (Erlang sequence)
			if i < len(block.Statements)-1 {
				builder.WriteString(",\n")
			}
		}
	}

	return builder.String()
}

// generateExpressionStatement converts an expression used as a statement
func (generator *ErlangGenerator) generateExpressionStatement(exprStmt *parser.ExpressionStatement) string {
	expression := generator.generateExpression(exprStmt.Expression)
	return fmt.Sprintf("%s%s", generator.indentString(), expression)
}

// generateStatementSequence generates a sequence of statements with proper Erlang syntax
func (generator *ErlangGenerator) generateStatementSequence(statements []parser.Statement) string {
	if len(statements) == 0 {
		return "ok" // Erlang convention for empty function body
	}

	var builder strings.Builder
	generator.indent++

	for i, stmt := range statements {
		code := generator.generateStatement(stmt)
		if code != "" {
			builder.WriteString(code)

			// In Erlang, statements in a sequence are separated by commas
			// The last statement ends with nothing (will be followed by period or semicolon)
			if i < len(statements)-1 {
				builder.WriteString(",\n")
			}
		}
	}

	generator.indent--
	return builder.String()
}

// Generate produces complete Erlang source code from AST
func (generator *ErlangGenerator) Generate(program *parser.Program) string {
	var builder strings.Builder

	// Generate module header
	builder.WriteString(generator.generateModuleHeader())

	// Generate all functions
	for _, function := range program.Functions {
		functionCode := generator.generateFunction(&function)
		builder.WriteString(functionCode)
		builder.WriteString("\n")
	}

	// Generate module footer (if any)
	builder.WriteString(generator.generateModuleFooter())

	return builder.String()
}

// generateFunction converts a function declaration to Erlang
func (generator *ErlangGenerator) generateFunction(function *parser.FunctionDeclaration) string {
	var builder strings.Builder

	// Map 'main' to 'start' for Erlang/OTP convention
	functionName := function.Name
	if functionName == "main" {
		functionName = "start"
	}

	// Function signature
	// For MVP, all functions have 0 parameters
	builder.WriteString(fmt.Sprintf("%s() ->\n", functionName))

	// Generate function body
	bodyCode := generator.generateStatementSequence(function.Body.Statements)

	// If body is empty, use 'ok'
	if bodyCode == "" {
		bodyCode = generator.indentString() + "ok"
	}

	builder.WriteString(bodyCode)
	builder.WriteString(".\n") // Erlang function ends with period

	return builder.String()
}

// GenerateToFile generates Erlang code and returns it as string for file writing
func (generator *ErlangGenerator) GenerateToFile(program *parser.Program) (string, error) {
	if program == nil {
		generator.errors.Add(
			errors.CodeGenerationError,
			"Cannot generate code from nil program",
			0, 0, generator.fileName,
		)
		return "", fmt.Errorf("nil program")
	}

	// Check for any accumulated errors
	if generator.errors.HasErrors() {
		return "", fmt.Errorf("code generation failed with %d errors", generator.errors.Count())
	}

	code := generator.Generate(program)

	// Final validation of generated code
	if !generator.validateErlangCode(code) {
		return "", fmt.Errorf("invalid Erlang code generated")
	}

	return code, nil
}

// validateErlangCode performs basic validation on generated Erlang code
func (generator *ErlangGenerator) validateErlangCode(code string) bool {
	// Basic checks
	if code == "" {
		generator.errors.Add(
			errors.CodeGenerationError,
			"Generated code is empty",
			0, 0, generator.fileName,
		)
		return false
	}

	// Check for module declaration
	if !strings.Contains(code, "-module(") {
		generator.errors.Add(
			errors.CodeGenerationError,
			"Missing module declaration",
			0, 0, generator.fileName,
		)
		return false
	}

	// Check for export declaration
	if !strings.Contains(code, "-export([") {
		generator.errors.Add(
			errors.CodeGenerationError,
			"Missing export declaration",
			0, 0, generator.fileName,
		)
		return false
	}

	// Check for start/0 function
	if !strings.Contains(code, "start() ->") {
		generator.errors.Add(
			errors.CodeGenerationError,
			"Missing start/0 function (main)",
			0, 0, generator.fileName,
		)
		return false
	}

	return true
}
