package main

import (
	"fmt"
	"naviary/compiler/ast"
	"naviary/compiler/lexer"
	"naviary/compiler/parser"
	typechecker "naviary/compiler/type-checker"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file.navi>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]

	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	sourceString := string(source)

	// 1. Lexical analysis
	lex := lexer.New(sourceString, filename)

	// 2. Parsing
	p := parser.New(lex)
	program := p.ParseProgram()

	// Check for parse errors
	if p.Errors().HasErrors() {
		p.Errors().Display()
		os.Exit(1)
	}

	// 3. Type checking
	checker := typechecker.New(sourceString, filename)
	checker.Check(program)

	// Check for type errors
	if checker.Errors().HasErrors() {
		checker.Errors().Display()
		os.Exit(1)
	}

	fmt.Printf("✓ Successfully compiled %s\n", filename)
	fmt.Println("=== AST Structure ===")
	for i, stmt := range program.Statements {
		fmt.Printf("Statement %d: %T\n", i, stmt)
		if funcStmt, ok := stmt.(*ast.FunctionStatement); ok {
			fmt.Printf("  Function: %s\n", funcStmt.Name.Value)
			fmt.Printf("  Body has %d statements:\n", len(funcStmt.Body.Statements))
			for j, bodyStmt := range funcStmt.Body.Statements {
				fmt.Printf("    %d: %T", j, bodyStmt)
				if letStmt, ok := bodyStmt.(*ast.LetStatement); ok {
					fmt.Printf(" (var: %s)", letStmt.Name.Value)
				}
				fmt.Println()
			}
		}
	}
	fmt.Println("==================")
}
