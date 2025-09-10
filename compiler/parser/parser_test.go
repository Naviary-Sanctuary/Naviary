package parser

import (
	"naviary/compiler/ast"
	"naviary/compiler/lexer"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLetStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected struct {
			identifier string
			mutable    bool
			value      interface{} // expected value
		}
	}{
		{
			name:  "immutable integer",
			input: "let x = 5",
			expected: struct {
				identifier string
				mutable    bool
				value      interface{}
			}{
				identifier: "x",
				mutable:    false,
				value:      "5",
			},
		},
		{
			name:  "mutable integer",
			input: "let mut y = 10",
			expected: struct {
				identifier string
				mutable    bool
				value      interface{}
			}{
				identifier: "y",
				mutable:    true,
				value:      "10",
			},
		},
		{
			name:  "mutable with colon assign",
			input: "let z := 15",
			expected: struct {
				identifier string
				mutable    bool
				value      interface{}
			}{
				identifier: "z",
				mutable:    true,
				value:      "15",
			},
		},
		{
			name:  "string literal",
			input: `let name = "Alice"`,
			expected: struct {
				identifier string
				mutable    bool
				value      interface{}
			}{
				identifier: "name",
				mutable:    false,
				value:      "Alice",
			},
		},
		{
			name:  "boolean true",
			input: "let flag = true",
			expected: struct {
				identifier string
				mutable    bool
				value      interface{}
			}{
				identifier: "flag",
				mutable:    false,
				value:      true,
			},
		},
		{
			name:  "boolean false",
			input: "let done = false",
			expected: struct {
				identifier string
				mutable    bool
				value      interface{}
			}{
				identifier: "done",
				mutable:    false,
				value:      false,
			},
		},
		{
			name:  "float literal",
			input: "let pi = 3.14",
			expected: struct {
				identifier string
				mutable    bool
				value      interface{}
			}{
				identifier: "pi",
				mutable:    false,
				value:      "3.14",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create lexer and parser
			lex := lexer.New(tt.input, "test.navi")
			parser := New(lex)

			// Parse program
			program := parser.ParseProgram()

			// Check for parse errors
			if lex.Errors().HasErrors() {
				t.Fatalf("lexer errors: %v", lex.Errors())
			}

			// Should have exactly one statement
			require.Len(t, program.Statements, 1, "program should have 1 statement")

			// Check it's a let statement
			stmt, ok := program.Statements[0].(*ast.LetStatement)
			require.True(t, ok, "statement should be LetStatement")

			// Check identifier name
			assert.Equal(t, tt.expected.identifier, stmt.Name.Value)

			// Check mutability
			assert.Equal(t, tt.expected.mutable, stmt.Mutable)

			// Check value based on type
			switch expected := tt.expected.value.(type) {
			case string:
				// Could be integer, float, or string literal
				switch expr := stmt.Value.(type) {
				case *ast.IntegerLiteral:
					assert.Equal(t, expected, expr.Value)
				case *ast.FloatLiteral:
					assert.Equal(t, expected, expr.Value)
				case *ast.StringLiteral:
					assert.Equal(t, expected, expr.Value)
				default:
					t.Fatalf("unexpected expression type: %T", expr)
				}
			case bool:
				boolExpr, ok := stmt.Value.(*ast.BooleanLiteral)
				require.True(t, ok, "value should be BooleanLiteral")
				assert.Equal(t, expected, boolExpr.Value)
			}
		})
	}
}

func TestParseMultipleStatements(t *testing.T) {
	input := `
let x = 5
let mut y = 10
let name := "Bob"
`

	lex := lexer.New(input, "test.navi")
	parser := New(lex)
	program := parser.ParseProgram()

	// Check no errors
	assert.False(t, lex.Errors().HasErrors())

	// Should have 3 statements
	assert.Len(t, program.Statements, 3)

	// Check each statement
	stmt1 := program.Statements[0].(*ast.LetStatement)
	assert.Equal(t, "x", stmt1.Name.Value)
	assert.False(t, stmt1.Mutable)

	stmt2 := program.Statements[1].(*ast.LetStatement)
	assert.Equal(t, "y", stmt2.Name.Value)
	assert.True(t, stmt2.Mutable)

	stmt3 := program.Statements[2].(*ast.LetStatement)
	assert.Equal(t, "name", stmt3.Name.Value)
	assert.True(t, stmt3.Mutable) // := means mutable
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "missing identifier",
			input:         "let = 5",
			expectedError: "expected 'IDENT', found '='",
		},
		{
			name:          "missing assignment",
			input:         "let x 5",
			expectedError: "expected '=' or ':=', found 'INT'",
		},
		{
			name:          "invalid expression",
			input:         "let x = +",
			expectedError: "unexpected token '+' in expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := lexer.New(tt.input, "test.navi")
			parser := New(lex)
			parser.ParseProgram()

			// Should have errors
			assert.True(t, lex.Errors().HasErrors())
		})
	}
}

func TestStatementTerminators(t *testing.T) {
	tests := []struct {
		name  string
		input string
		count int // expected statement count
	}{
		{
			name:  "semicolon terminated",
			input: "let x = 5; let y = 10;",
			count: 2,
		},
		{
			name:  "newline terminated",
			input: "let x = 5\nlet y = 10",
			count: 2,
		},
		{
			name:  "mixed terminators",
			input: "let x = 5;\nlet y = 10\n;let z = 15",
			count: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := lexer.New(tt.input, "test.navi")
			parser := New(lex)
			program := parser.ParseProgram()

			assert.False(t, lex.Errors().HasErrors())
			assert.Len(t, program.Statements, tt.count)
		})
	}
}

func TestParseInfixExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Expected string representation of AST
	}{
		{
			name:     "simple addition",
			input:    "let x = 5 + 3",
			expected: "let x = (5 + 3)",
		},
		{
			name:     "simple multiplication",
			input:    "let x = 5 * 3",
			expected: "let x = (5 * 3)",
		},
		{
			name:     "operator precedence - multiply first",
			input:    "let x = 2 + 3 * 4",
			expected: "let x = (2 + (3 * 4))",
		},
		{
			name:     "operator precedence - multiply first reversed",
			input:    "let x = 2 * 3 + 4",
			expected: "let x = ((2 * 3) + 4)",
		},
		{
			name:     "left associativity - subtraction",
			input:    "let x = 5 - 3 - 1",
			expected: "let x = ((5 - 3) - 1)",
		},
		{
			name:     "left associativity - division",
			input:    "let x = 20 / 4 / 2",
			expected: "let x = ((20 / 4) / 2)",
		},
		{
			name:     "complex expression",
			input:    "let x = 1 + 2 * 3 - 4 / 2",
			expected: "let x = ((1 + (2 * 3)) - (4 / 2))",
		},
		{
			name:     "all same precedence",
			input:    "let x = 1 + 2 + 3 + 4",
			expected: "let x = (((1 + 2) + 3) + 4)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := lexer.New(tt.input, "test.navi")
			parser := New(lex)
			program := parser.ParseProgram()

			// Check for errors
			assert.False(t, lex.Errors().HasErrors(),
				"parser errors: %v", lex.Errors())

			// Check we have one statement
			require.Len(t, program.Statements, 1)

			// Check the string representation matches expected
			stmt := program.Statements[0]
			assert.Equal(t, tt.expected, stmt.String())
		})
	}
}

func TestParseFunctionStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected struct {
			functionName string
			parameters   []struct {
				name     string
				typeName string
			}
			returnType     string
			bodyStatements int
		}
	}{
		{
			name:  "main function without parameters",
			input: "func main() { }",
			expected: struct {
				functionName string
				parameters   []struct {
					name     string
					typeName string
				}
				returnType     string
				bodyStatements int
			}{
				functionName:   "main",
				parameters:     []struct{ name, typeName string }{},
				returnType:     "",
				bodyStatements: 0,
			},
		},
		{
			name: "function with single parameter",
			input: `func greet(name: string) {
				let message = "Hello"
			}`,
			expected: struct {
				functionName string
				parameters   []struct {
					name     string
					typeName string
				}
				returnType     string
				bodyStatements int
			}{
				functionName: "greet",
				parameters: []struct{ name, typeName string }{
					{name: "name", typeName: "string"},
				},
				returnType:     "",
				bodyStatements: 1,
			},
		},
		{
			name:  "function with return type",
			input: "func add(x: int, y: int) -> int { }",
			expected: struct {
				functionName string
				parameters   []struct {
					name     string
					typeName string
				}
				returnType     string
				bodyStatements int
			}{
				functionName: "add",
				parameters: []struct{ name, typeName string }{
					{name: "x", typeName: "int"},
					{name: "y", typeName: "int"},
				},
				returnType:     "int",
				bodyStatements: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create lexer and parser
			lex := lexer.New(tt.input, "test.navi")
			parser := New(lex)

			// Parse program
			program := parser.ParseProgram()

			// Check for parse errors
			if parser.Errors().HasErrors() {
				t.Fatalf("parser errors: %v", parser.Errors())
			}

			// Should have exactly one statement
			require.Len(t, program.Statements, 1, "program should have 1 statement")

			// Check it's a function statement
			funcStmt, ok := program.Statements[0].(*ast.FunctionStatement)
			require.True(t, ok, "statement should be FunctionStatement")

			// Check function name
			assert.Equal(t, tt.expected.functionName, funcStmt.Name.Value)

			// Check parameters
			assert.Len(t, funcStmt.Parameters, len(tt.expected.parameters))
			for i, param := range funcStmt.Parameters {
				assert.Equal(t, tt.expected.parameters[i].name, param.Name.Value)
				assert.Equal(t, tt.expected.parameters[i].typeName, param.Type.Value)
			}

			// Check return type
			if tt.expected.returnType == "" {
				assert.Nil(t, funcStmt.ReturnType)
			} else {
				require.NotNil(t, funcStmt.ReturnType)
				assert.Equal(t, tt.expected.returnType, funcStmt.ReturnType.Value)
			}

			// Check body statements count
			assert.Len(t, funcStmt.Body.Statements, tt.expected.bodyStatements)
		})
	}
}

func TestParseFunctionCalls(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected struct {
			functionName   string
			argumentCount  int
			argumentValues []string // String representation of arguments
		}
	}{
		{
			name:  "function call without arguments",
			input: "let x = getValue()",
			expected: struct {
				functionName   string
				argumentCount  int
				argumentValues []string
			}{
				functionName:   "getValue",
				argumentCount:  0,
				argumentValues: []string{},
			},
		},
		{
			name:  "function call with single argument",
			input: "let x = print(42)",
			expected: struct {
				functionName   string
				argumentCount  int
				argumentValues []string
			}{
				functionName:   "print",
				argumentCount:  1,
				argumentValues: []string{"42"},
			},
		},
		{
			name:  "function call with multiple arguments",
			input: "let result = add(10, 20)",
			expected: struct {
				functionName   string
				argumentCount  int
				argumentValues []string
			}{
				functionName:   "add",
				argumentCount:  2,
				argumentValues: []string{"10", "20"},
			},
		},
		{
			name:  "function call with expression arguments",
			input: "let result = multiply(2 + 3, 4 * 5)",
			expected: struct {
				functionName   string
				argumentCount  int
				argumentValues []string
			}{
				functionName:   "multiply",
				argumentCount:  2,
				argumentValues: []string{"(2 + 3)", "(4 * 5)"},
			},
		},
		{
			name:  "nested function calls",
			input: "let x = add(multiply(2, 3), 4)",
			expected: struct {
				functionName   string
				argumentCount  int
				argumentValues []string
			}{
				functionName:   "add",
				argumentCount:  2,
				argumentValues: []string{"multiply(2, 3)", "4"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create lexer and parser
			lex := lexer.New(tt.input, "test.navi")
			parser := New(lex)

			// Parse program
			program := parser.ParseProgram()

			// Check for parse errors
			if parser.Errors().HasErrors() {
				t.Fatalf("parser errors: %v", parser.Errors())
			}

			// Should have exactly one statement
			require.Len(t, program.Statements, 1, "program should have 1 statement")

			// Check it's a let statement
			letStmt, ok := program.Statements[0].(*ast.LetStatement)
			require.True(t, ok, "statement should be LetStatement")

			// Check the value is a call expression
			callExpr, ok := letStmt.Value.(*ast.CallExpression)
			require.True(t, ok, "value should be CallExpression, got %T", letStmt.Value)

			// Check function name
			funcIdent, ok := callExpr.Function.(*ast.Identifier)
			require.True(t, ok, "function should be Identifier")
			assert.Equal(t, tt.expected.functionName, funcIdent.Value)

			// Check arguments
			assert.Len(t, callExpr.Arguments, tt.expected.argumentCount)

			for i, arg := range callExpr.Arguments {
				assert.Equal(t, tt.expected.argumentValues[i], arg.String())
			}
		})
	}
}

// Test function calls as statements (not just in let statements)
func TestParseFunctionCallStatements(t *testing.T) {
	input := `
func main() {
	print(42)
	doSomething()
	calculate(1 + 2, 3 * 4)
}
`

	lex := lexer.New(input, "test.navi")
	parser := New(lex)
	program := parser.ParseProgram()

	// Check for parse errors
	assert.False(t, parser.Errors().HasErrors(), "parser should have no errors")

	// Should have one function statement
	require.Len(t, program.Statements, 1)

	funcStmt, ok := program.Statements[0].(*ast.FunctionStatement)
	require.True(t, ok, "should be FunctionStatement")

	// Check function has 3 statements in body
	assert.Len(t, funcStmt.Body.Statements, 3)

	// Verify each statement is an ExpressionStatement containing a CallExpression
	for i, stmt := range funcStmt.Body.Statements {
		_, ok := stmt.(*ast.ExpressionStatement)
		assert.True(t, ok, "statement %d should be ExpressionStatement", i)
	}
}
