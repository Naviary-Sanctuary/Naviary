package codegen

import (
	"naviary/compiler/ast"
	"naviary/compiler/token" // token import 추가
	"strings"
	"testing"
)

func TestGenerateSimpleMain(t *testing.T) {
	// Create AST for: func main() { print(42) }
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.FunctionStatement{
				Token: token.Token{Type: token.FUNC, Literal: "func"},
				Name: &ast.Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "main"},
					Value: "main",
				},
				Parameters: []*ast.FunctionParameter{},
				ReturnType: nil,
				Body: &ast.BlockStatement{
					Statements: []ast.Statement{
						&ast.ExpressionStatement{
							Expression: &ast.CallExpression{
								Function: &ast.Identifier{
									Value: "navi_print_int",
								},
								Arguments: []ast.Expression{
									&ast.IntegerLiteral{
										Value: "42",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Generate assembly
	emitter := NewDarwinARM64Emitter()
	generator := New(emitter)
	generator.Generate(program)

	assembly := generator.GenerateAssembly()

	// Check key parts
	if !strings.Contains(assembly, "_main:") {
		t.Errorf("Missing _main label")
	}

	if !strings.Contains(assembly, "mov x0, #42") {
		t.Errorf("Missing mov x0, #42")
	}

	if !strings.Contains(assembly, "bl _navi_print_int") {
		t.Errorf("Missing function call")
	}

	if !strings.Contains(assembly, "ret") {
		t.Errorf("Missing return")
	}

	// Print for manual inspection
	t.Logf("Generated assembly:\n%s", assembly)
}
