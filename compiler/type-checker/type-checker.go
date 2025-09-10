package typechecker

import (
	"naviary/compiler/ast"
	"naviary/compiler/errors"
	"naviary/compiler/types"
)

type TypeChecker struct {
	errorCollector  *errors.ErrorCollector
	symbolTable     *types.SymbolTable
	currentFunction *types.FunctionType
}

func New(source string, filename string) *TypeChecker {
	typeChecker := &TypeChecker{
		errorCollector:  errors.New(source, filename),
		symbolTable:     types.NewSymbolTable(),
		currentFunction: nil,
	}

	typeChecker.registerBuiltins()

	return typeChecker
}

func (checker *TypeChecker) registerBuiltins() {
	// print function: print(value: any) -> void
	printType := &types.FunctionType{
		Parameters: []types.Type{types.Int}, // For now, only int
		ReturnType: nil,                     // void
	}
	checker.symbolTable.Define("print", printType, false)
}

func (checker *TypeChecker) Errors() *errors.ErrorCollector {
	return checker.errorCollector
}

// Check performs type checking on the entire program
func (checker *TypeChecker) Check(program *ast.Program) {
	// Check each statement in the program
	for _, statement := range program.Statements {
		checker.checkStatement(statement)
	}
}

func (checker *TypeChecker) checkExpression(expression ast.Expression) types.Type {
	switch expr := expression.(type) {
	case *ast.IntegerLiteral:
		return types.Int
	case *ast.FloatLiteral:
		return types.Float
	case *ast.StringLiteral:
		return types.String
	case *ast.BooleanLiteral:
		return types.Bool
	case *ast.Identifier:
		return checker.checkIdentifier(expr)
	case *ast.BinaryExpression:
		return checker.checkBinaryOperation(expr)
	case *ast.CallExpression:
		return checker.checkCallExpression(expr)
	default:
		return nil
	}
}

// checkStatement type checks a statement
func (checker *TypeChecker) checkStatement(statement ast.Statement) {
	switch stmt := statement.(type) {
	case *ast.LetStatement:
		checker.checkLetStatement(stmt)
	case *ast.FunctionStatement:
		checker.checkFunctionStatement(stmt)
	case *ast.ReturnStatement:
		checker.checkReturnStatement(stmt)
	case *ast.ExpressionStatement:
		checker.checkExpressionStatement(stmt)
	case *ast.BlockStatement:
		checker.checkBlockStatement(stmt)
	default:
		// Unknown statement type
	}
}

func (checker *TypeChecker) checkIdentifier(identifier *ast.Identifier) types.Type {
	symbol := checker.symbolTable.Lookup(identifier.Value)

	if symbol == nil {
		checker.errorCollector.Add(
			errors.TypeError,
			identifier.Token.Line,
			identifier.Token.Column,
			len(identifier.Value),
			"undefined variable %s",
			identifier.Value,
		)
		return nil
	}

	return symbol.Type
}

func (checker *TypeChecker) checkBinaryOperation(binary *ast.BinaryExpression) types.Type {
	leftType := checker.checkExpression(binary.Left)
	if leftType == nil {
		return nil // error already reported
	}

	rightType := checker.checkExpression(binary.Right)
	if rightType == nil {
		return nil // error already reported
	}

	if !leftType.Equals(rightType) {
		checker.errorCollector.Add(
			errors.TypeError,
			binary.Token.Line,
			binary.Token.Column,
			len(binary.Operator),
			"type mismatch: cannot apply '%s' to %s and %s",
			binary.Operator,
			leftType.String(),
			rightType.String(),
		)
		return nil
	}

	switch binary.Operator {
	case "+", "-", "*", "/":
		if leftType == types.Int || leftType == types.Float {
			return leftType
		}

		// string concat
		if binary.Operator == "+" && leftType == types.String {
			return types.String
		}
	case "==", "1=", ">", "<", ">=", "<=":
		return types.Bool
	}

	checker.errorCollector.Add(
		errors.TypeError,
		binary.Token.Line,
		binary.Token.Column,
		len(binary.Operator),
		"invalid operation: %s %s %s",
		leftType.String(),
		binary.Operator,
		rightType.String(),
	)
	return nil
}

// checkCallExpression checks function calls
func (checker *TypeChecker) checkCallExpression(call *ast.CallExpression) types.Type {
	// Get function identifier
	funcIdent, ok := call.Function.(*ast.Identifier)
	if !ok {
		checker.errorCollector.Add(
			errors.TypeError,
			call.Token.Line,
			call.Token.Column,
			1,
			"invalid function call: not an identifier",
		)
		return nil
	}

	// Look up function in symbol table
	symbol := checker.symbolTable.Lookup(funcIdent.Value)
	if symbol == nil {
		checker.errorCollector.Add(
			errors.TypeError,
			funcIdent.Token.Line,
			funcIdent.Token.Column,
			len(funcIdent.Value),
			"undefined function: %s",
			funcIdent.Value,
		)
		return nil
	}

	// Check if it's actually a function type
	funcType, ok := symbol.Type.(*types.FunctionType)
	if !ok {
		checker.errorCollector.Add(
			errors.TypeError,
			funcIdent.Token.Line,
			funcIdent.Token.Column,
			len(funcIdent.Value),
			"'%s' is not a function",
			funcIdent.Value,
		)
		return nil
	}

	// Check argument count
	if len(call.Arguments) != len(funcType.Parameters) {
		checker.errorCollector.Add(
			errors.TypeError,
			call.Token.Line,
			call.Token.Column,
			1,
			"wrong number of arguments: expected %d, got %d",
			len(funcType.Parameters),
			len(call.Arguments),
		)
		return nil
	}

	// Check each argument type
	for i, arg := range call.Arguments {
		argType := checker.checkExpression(arg)
		if argType == nil {
			continue // Error already reported
		}

		expectedType := funcType.Parameters[i]
		if !argType.Equals(expectedType) {
			checker.errorCollector.Add(
				errors.TypeError,
				call.Token.Line,
				call.Token.Column,
				1,
				"argument %d: expected %s, got %s",
				i+1,
				expectedType.String(),
				argType.String(),
			)
		}
	}

	return funcType.ReturnType // Can be nil for void functions
}

// checkLetStatement checks variable declarations
func (checker *TypeChecker) checkLetStatement(letStmt *ast.LetStatement) {
	// First, check the value expression to get its type
	valueType := checker.checkExpression(letStmt.Value)
	if valueType == nil {
		return // Error already reported
	}

	// Now define the variable with its type
	if !checker.symbolTable.Define(letStmt.Name.Value, valueType, letStmt.Mutable) {
		checker.errorCollector.Add(
			errors.TypeError,
			letStmt.Name.Token.Line,
			letStmt.Name.Token.Column,
			len(letStmt.Name.Value),
			"variable '%s' already defined in this scope",
			letStmt.Name.Value,
		)
		return
	}
}

// checkFunctionStatement checks function declarations
func (checker *TypeChecker) checkFunctionStatement(funcStmt *ast.FunctionStatement) {
	// Create function type from AST
	paramTypes := make([]types.Type, len(funcStmt.Parameters))
	for i, param := range funcStmt.Parameters {
		paramType := types.GetPrimitiveType(param.Type.Value)
		if paramType == nil {
			checker.errorCollector.Add(
				errors.TypeError,
				param.Type.Token.Line,
				param.Type.Token.Column,
				len(param.Type.Value),
				"unknown type: %s",
				param.Type.Value,
			)
			return
		}
		paramTypes[i] = paramType
	}

	// Get return type (nil if no return type specified)
	var returnType types.Type
	if funcStmt.ReturnType != nil {
		returnType = types.GetPrimitiveType(funcStmt.ReturnType.Value)
		if returnType == nil {
			checker.errorCollector.Add(
				errors.TypeError,
				funcStmt.ReturnType.Token.Line,
				funcStmt.ReturnType.Token.Column,
				len(funcStmt.ReturnType.Value),
				"unknown return type: %s",
				funcStmt.ReturnType.Value,
			)
			return
		}
	}

	// Create function type
	funcType := &types.FunctionType{
		Parameters: paramTypes,
		ReturnType: returnType,
	}

	// Define function in symbol table
	if !checker.symbolTable.Define(funcStmt.Name.Value, funcType, false) {
		checker.errorCollector.Add(
			errors.TypeError,
			funcStmt.Name.Token.Line,
			funcStmt.Name.Token.Column,
			len(funcStmt.Name.Value),
			"function '%s' already defined",
			funcStmt.Name.Value,
		)
		return
	}

	// Create new scope for function body
	checker.symbolTable = checker.symbolTable.NewChildScope()

	// Add parameters to the new scope
	for i, param := range funcStmt.Parameters {
		if !checker.symbolTable.Define(param.Name.Value, paramTypes[i], false) {
			checker.errorCollector.Add(
				errors.TypeError,
				param.Name.Token.Line,
				param.Name.Token.Column,
				len(param.Name.Value),
				"duplicate parameter: %s",
				param.Name.Value,
			)
		}
	}

	// Save current function for return type checking
	previousFunction := checker.currentFunction
	checker.currentFunction = funcType

	// Check function body
	checker.checkBlockStatement(funcStmt.Body)

	// Restore previous function and scope
	checker.currentFunction = previousFunction
	checker.symbolTable = checker.symbolTable.Parent()
}

// checkBlockStatement checks a block of statements
func (checker *TypeChecker) checkBlockStatement(block *ast.BlockStatement) {
	// Create new scope for the block
	checker.symbolTable = checker.symbolTable.NewChildScope()

	// Check each statement in the block
	for _, statement := range block.Statements {
		checker.checkStatement(statement)
	}

	// Restore parent scope
	checker.symbolTable = checker.symbolTable.Parent()
}

// checkReturnStatement checks return statements
func (checker *TypeChecker) checkReturnStatement(returnStmt *ast.ReturnStatement) {
	// Check if we're inside a function
	if checker.currentFunction == nil {
		checker.errorCollector.Add(
			errors.TypeError,
			returnStmt.Token.Line,
			returnStmt.Token.Column,
			len(returnStmt.Token.Literal),
			"return statement outside function",
		)
		return
	}

	// Check return value
	if returnStmt.ReturnValue == nil {
		// No return value
		if checker.currentFunction.ReturnType != nil {
			checker.errorCollector.Add(
				errors.TypeError,
				returnStmt.Token.Line,
				returnStmt.Token.Column,
				len(returnStmt.Token.Literal),
				"missing return value: expected %s",
				checker.currentFunction.ReturnType.String(),
			)
		}
		return
	}

	// Has return value - check its type
	returnType := checker.checkExpression(returnStmt.ReturnValue)
	if returnType == nil {
		return // Error already reported
	}

	// Check if return type matches function signature
	if checker.currentFunction.ReturnType == nil {
		checker.errorCollector.Add(
			errors.TypeError,
			returnStmt.Token.Line,
			returnStmt.Token.Column,
			len(returnStmt.Token.Literal),
			"unexpected return value in void function",
		)
	} else if !returnType.Equals(checker.currentFunction.ReturnType) {
		checker.errorCollector.Add(
			errors.TypeError,
			returnStmt.Token.Line,
			returnStmt.Token.Column,
			len(returnStmt.Token.Literal),
			"return type mismatch: expected %s, got %s",
			checker.currentFunction.ReturnType.String(),
			returnType.String(),
		)
	}
}

// checkExpressionStatement checks expression statements
func (checker *TypeChecker) checkExpressionStatement(exprStmt *ast.ExpressionStatement) {
	// Just check the expression, ignore the return type
	checker.checkExpression(exprStmt.Expression)
}
