package codegen

import (
	"compiler/ast"
	"compiler/errors"
)

type CGenerator struct {
	emitter        *Emitter
	errorCollector *errors.ErrorCollector
}

func NewCGenerator(errorCollector *errors.ErrorCollector) *CGenerator {
	return &CGenerator{
		emitter:        NewEmitter(),
		errorCollector: errorCollector,
	}
}

func (generator *CGenerator) Generate(program *ast.Program) string {
	generator.EmitHeaders()

	for _, statement := range program.Statements {
		generator.generateStatement(statement)
	}

	return generator.emitter.GetOutput()
}

func (generator *CGenerator) EmitHeaders() {
	// TODO: dynamic header include
	generator.emitter.EmitLine("#include <stdio.h>")
	generator.emitter.EmitNewLine()
	generator.emitter.EmitLine("extern void print(int value);")
	generator.emitter.EmitNewLine()
}

func (generator *CGenerator) generateStatement(statement ast.Statement) {
	switch stmt := statement.(type) {
	case *ast.FunctionStatement:
		generator.generateFunction(stmt)
	case *ast.LetStatement:
		generator.generateLet(stmt)
	case *ast.ReturnStatement:
		generator.generateReturnStatement(stmt)
	case *ast.ExpressionStatement:
		generator.generateExpressionStatement(stmt)
	default:
		generator.errorCollector.Add(errors.SyntaxError,
			0, 0, 0,
			"Unknown statement type: %T",
			stmt,
		)
	}
	generator.emitter.EmitNewLine()
}

func (generator *CGenerator) generateFunction(function *ast.FunctionStatement) {
	// TODO: return type is only int for now
	generator.emitter.Emit("int ")
	generator.emitter.Emit(function.Name.Value)
	generator.emitter.Emit("(")
	// TODO: parameters
	generator.emitter.Emit(")")
	generator.emitter.EmitLine(" {")
	generator.emitter.IncreaseIndent()

	for _, stmt := range function.Body.Statements {
		generator.generateStatement(stmt)
	}

	if function.Name.Value == "main" {
		generator.emitter.EmitLine("return 0;")
	}

	generator.emitter.DecreaseIndent()
	generator.emitter.EmitLine("}")
}

func (generator *CGenerator) generateReturnStatement(returnStmt *ast.ReturnStatement) {
	generator.emitter.Emit("return")

	// Check if there's a return value
	if returnStmt.ReturnValue != nil {
		generator.emitter.Emit(" ")
		generator.generateExpression(returnStmt.ReturnValue)
	}

	generator.emitter.EmitLine(";")
}

func (generator *CGenerator) generateLet(let *ast.LetStatement) {
	// TODO: type is only int for now
	generator.emitter.Emit("int ")

	generator.emitter.Emit(let.Name.Value)

	generator.emitter.Emit(" = ")

	generator.generateExpression(let.Value)

	generator.emitter.Emit(";")
}

func (generator *CGenerator) generateExpression(expr ast.Expression) {
	switch expression := expr.(type) {
	case *ast.IntegerLiteral:
		generator.emitter.Emit(expression.Value)
	case *ast.Identifier:
		generator.emitter.Emit(expression.Value)
	case *ast.BinaryExpression:
		generator.generateExpression(expression.Left)
		generator.emitter.Emit(" ")
		generator.emitter.Emit(expression.Operator)
		generator.emitter.Emit(" ")
		generator.generateExpression(expression.Right)
	case *ast.CallExpression:
		generator.generateExpression(expression.Function)
		generator.emitter.Emit("(")
		for i, arg := range expression.Arguments {
			generator.generateExpression(arg)
			if i < len(expression.Arguments)-1 {
				generator.emitter.Emit(", ")
			}
		}
		generator.emitter.Emit(")")
	default:
		generator.errorCollector.Add(
			errors.SyntaxError,
			0, 0, 0,
			"Unsupported expression type: %T", expr,
		)
	}
}

func (generator *CGenerator) generateExpressionStatement(exprStmt *ast.ExpressionStatement) {
	generator.generateExpression(exprStmt.Expression)
	generator.emitter.Emit(";")
}
