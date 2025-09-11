package main

import (
	"fmt"
	"naviary/compiler/codegen"
	"naviary/compiler/lexer"
	"naviary/compiler/parser"
	typechecker "naviary/compiler/type-checker"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	// 4. Code generation
	emitter := codegen.NewDarwinARM64Emitter()
	generator := codegen.New(emitter)
	generator.Generate(program)

	assembly := generator.GenerateAssembly()

	// 5. Write assembly to file
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
	asmFile := baseName + ".s"

	err = os.WriteFile(asmFile, []byte(assembly), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing assembly file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated assembly: %s\n", asmFile)

	// 6. Assemble to object file
	objFile := baseName + ".o"
	cmd := exec.Command("as", "-o", objFile, asmFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Assembly failed: %v\n%s\n", err, output)
		os.Exit(1)
	}

	fmt.Printf("✓ Created object file: %s\n", objFile)

	// 7. Link with runtime
	execFile := baseName
	cmd = exec.Command("ld",
		"-o", execFile,
		objFile,
		"-L", "runtime", // Runtime library directory
		"-lnavi_runtime", // Link libnavi_runtime.a
		"-lSystem",       // macOS system libraries
		"-syslibroot", "/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk",
		"-e", "_main", // Entry point
		"-arch", "arm64", // Architecture
	)
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Linking failed: %v\n%s\n", err, output)
		os.Exit(1)
	}

	fmt.Printf("✓ Created executable: %s\n", execFile)

	// 8. Optionally run the program
	fmt.Println("\nRun the program:")
	fmt.Printf("  ./%s\n", execFile)
}
