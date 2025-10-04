package nir

import (
	"compiler/ast"
	"compiler/errors"
	"compiler/nir/instruction"
	"compiler/nir/value"
	"compiler/types"
	"fmt"
)

// Lowerer converts AST to NIR
// Lowering is the process of transforming high-level AST into low-level NIR (Naviary Intermediate Representation)
type Lowerer struct {
	builder         *Builder
	currentFunction *Function
	errorCollector  *errors.ErrorCollector
}

func NewLowerer(errorCollector *errors.ErrorCollector) *Lowerer {
	return &Lowerer{
		builder:         NewBuilder(),
		currentFunction: nil,
		errorCollector:  errorCollector,
	}
}

func (lowerer *Lowerer) Lower(program *ast.Program) *Module {
	module := NewModule("main")

	for _, statement := range program.Statements {
		switch stmt := statement.(type) {
		case *ast.FunctionStatement:
			function := lowerer.lowerFunction(stmt)
			if function != nil {
				module.AddFunction(function)
			}
		default:
			lowerer.errorCollector.Add(errors.SyntaxError,
				0, 0, 0,
				"Unknown statement type: %T",
				stmt,
			)
		}
	}

	return module
}

func (lowerer *Lowerer) lowerFunction(astFunc *ast.FunctionStatement) *Function {
	// Reset builder for new function
	lowerer.builder.Reset()

	// Convert parameters
	var parameters []Parameter
	for _, param := range astFunc.Parameters {
		// For now, assume all parameters are int type
		// TODO: Use type annotations when type system is implemented
		parameters = append(parameters, NewParameter(
			param.Name.Value,
			types.Int,
		))
	}

	// Determine return type
	// For now, default to nil
	// TODO: Use return type annotation when type system is implemented
	var returnType types.Type = types.Nil

	if astFunc.Name.Value == "main" {
		returnType = types.Int
	} else if astFunc.ReturnType != nil {
		returnType = lowerer.getType(astFunc.ReturnType)
	}

	// Create NIR function
	function := NewFunction(astFunc.Name.Value, parameters, returnType)
	lowerer.currentFunction = function

	// Create entry block
	entryBlock := NewBasicBlock("entry")
	lowerer.builder.SetInsertBlock(entryBlock)

	// Lower function body
	lowerer.lowerBlockStatement(astFunc.Body)

	// Add implicit return for void functions if missing
	if !entryBlock.IsComplete() {
		if astFunc.Name.Value == "main" {
			lowerer.builder.BuildReturn(lowerer.builder.CreateConstantInt(0))

		} else {
			lowerer.builder.BuildReturn(nil)
		}
	}

	function.AddBasicBlock(entryBlock)

	return function
}

// lowerBlockStatement lowers a block of statements
func (lowerer *Lowerer) lowerBlockStatement(block *ast.BlockStatement) {
	for _, statement := range block.Statements {
		lowerer.lowerStatement(statement)
	}
}

// lowerStatement lowers a single statement
func (lowerer *Lowerer) lowerStatement(statement ast.Statement) {
	switch stmt := statement.(type) {
	case *ast.LetStatement:
		lowerer.lowerLetStatement(stmt)
	case *ast.ReturnStatement:
		lowerer.lowerReturnStatement(stmt)
	case *ast.ExpressionStatement:
		lowerer.lowerExpressionStatement(stmt)
	default:
		lowerer.errorCollector.Add(
			errors.SyntaxError,
			0, 0, 0,
			"Unsupported statement type: %T",
			stmt,
		)
	}
}

// lowerLetStatement lowers a let statement
// Example: let x = 1 + 2
//
//	→ %0 = Constant(1)
//	  %1 = Constant(2)
//	  %2 = Add(%0, %1)
//	  %x = Alloc(int)
//	  Store(%x, %2)
func (lowerer *Lowerer) lowerLetStatement(letStmt *ast.LetStatement) {
	// Lower the initialization expression
	initValue := lowerer.lowerExpression(letStmt.Value)
	if initValue == nil {
		return
	}

	// Allocate variable
	variable := lowerer.builder.BuildAlloc(letStmt.Name.Value, initValue.Type())

	// Store initial value
	lowerer.builder.BuildStore(variable, initValue)
}

// lowerReturnStatement lowers a return statement
// Example: return x + 1
//
//	→ %0 = Load(%x)
//	  %1 = Constant(1)
//	  %2 = Add(%0, %1)
//	  Return(%2)
func (lowerer *Lowerer) lowerReturnStatement(returnStmt *ast.ReturnStatement) {
	if returnStmt.ReturnValue == nil {
		// Return void
		lowerer.builder.BuildReturn(nil)
		return
	}

	// Lower return expression
	returnValue := lowerer.lowerExpression(returnStmt.ReturnValue)
	if returnValue == nil {
		return
	}

	lowerer.builder.BuildReturn(returnValue)
}

// lowerExpressionStatement lowers an expression statement
// Example: print(42)
//
//	→ %0 = Constant(42)
//	  Call(print, [%0])
func (lowerer *Lowerer) lowerExpressionStatement(exprStmt *ast.ExpressionStatement) {
	lowerer.lowerExpression(exprStmt.Expression)
}

// lowerExpression lowers an expression to a NIR value
// This is where complex nested expressions get flattened
func (lowerer *Lowerer) lowerExpression(expr ast.Expression) value.Value {
	switch expression := expr.(type) {
	case *ast.IntegerLiteral:
		return lowerer.lowerIntegerLiteral(expression)
	case *ast.StringLiteral:
		return lowerer.lowerStringLiteral(expression)
	case *ast.Identifier:
		return lowerer.lowerIdentifier(expression)
	case *ast.BinaryExpression:
		return lowerer.lowerBinaryExpression(expression)
	case *ast.CallExpression:
		return lowerer.lowerCallExpression(expression)
	default:
		lowerer.errorCollector.Add(
			errors.SyntaxError,
			0, 0, 0,
			"Unsupported expression type: %T",
			expr,
		)
		return nil
	}
}

// lowerIntegerLiteral converts an integer literal to a constant
func (lowerer *Lowerer) lowerIntegerLiteral(literal *ast.IntegerLiteral) value.Value {
	// Parse integer value
	// For now, assume all integers are valid
	// TODO: Proper integer parsing with error handling
	var val int
	fmt.Sscanf(literal.Value, "%d", &val)

	return lowerer.builder.CreateConstantInt(val)
}

// lowerStringLiteral converts a string literal to a constant
func (lowerer *Lowerer) lowerStringLiteral(literal *ast.StringLiteral) value.Value {
	return lowerer.builder.CreateConstantString(literal.Value)
}

// lowerIdentifier converts an identifier to a load instruction
// Example: x  →  %0 = Load(%x)
func (lowerer *Lowerer) lowerIdentifier(identifier *ast.Identifier) value.Value {
	// Create variable reference
	// TODO: Look up actual variable from symbol table
	variable := lowerer.builder.CreateVariable(identifier.Value, types.Int)

	// Load the value
	return lowerer.builder.BuildLoad(variable)
}

// lowerBinaryExpression lowers a binary operation
// Example: 1 + 2
//
//	→ %0 = Constant(1)
//	  %1 = Constant(2)
//	  %2 = Add(%0, %1)
func (lowerer *Lowerer) lowerBinaryExpression(binary *ast.BinaryExpression) value.Value {
	// Lower left and right operands first
	left := lowerer.lowerExpression(binary.Left)
	if left == nil {
		return nil
	}

	right := lowerer.lowerExpression(binary.Right)
	if right == nil {
		return nil
	}

	// Generate appropriate instruction based on operator
	switch binary.Operator {
	case "+":
		return lowerer.builder.BuildBinary(left, right, instruction.BinaryAdd)
	case "-":
		return lowerer.builder.BuildBinary(left, right, instruction.BinarySubtract)
	case "*":
		return lowerer.builder.BuildBinary(left, right, instruction.BinaryMultiply)
	case "/":
		return lowerer.builder.BuildBinary(left, right, instruction.BinaryDivide)
	default:
		lowerer.errorCollector.Add(
			errors.SyntaxError,
			0, 0, 0,
			"Unsupported binary operator: %s",
			binary.Operator,
		)
		return nil
	}
}

// lowerCallExpression lowers a function call
// Example: print(42)
//
//	→ %0 = Constant(42)
//	  Call(print, [%0])
func (lowerer *Lowerer) lowerCallExpression(call *ast.CallExpression) value.Value {
	// Get function name
	functionName := ""
	if ident, ok := call.Function.(*ast.Identifier); ok {
		functionName = ident.Value
	} else {
		lowerer.errorCollector.Add(
			errors.SyntaxError,
			0, 0, 0,
			"Only simple function calls are supported",
		)
		return nil
	}

	// Lower arguments
	var arguments []value.Value
	for _, arg := range call.Arguments {
		argValue := lowerer.lowerExpression(arg)
		if argValue == nil {
			return nil
		}
		arguments = append(arguments, argValue)
	}

	// For now, assume all functions are void
	// TODO: Look up function signature from symbol table
	return lowerer.builder.BuildCall(functionName, arguments, nil)
}

// getType converts AST type annotation to NIR type
func (lowerer *Lowerer) getType(typeAnnotation *ast.TypeAnnotation) types.Type {
	switch typeAnnotation.Value {
	case "int":
		return types.Int
	case "float":
		return types.Float
	case "string":
		return types.String
	case "bool":
		return types.Bool
	case "nil":
		return types.Nil
	default:
		return types.Int // Default fallback
	}
}
