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
