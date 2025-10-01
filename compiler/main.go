package main

import (
	"compiler/codegen"
	"compiler/constants"
	"compiler/errors"
	"compiler/lexer"
	"compiler/parser"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CompileFile compiles a single Naviary source file
func CompileFile(inputPath string, runAfterCompile bool) error {
	sourceCode, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", inputPath, err)
	}

	fileName := filepath.Base(inputPath)
	sourceStr := string(sourceCode)

	// Create global error collector with source code
	errorCollector := errors.New(sourceStr, fileName)

	// Step 1: Lexical Analysis
	fmt.Printf("Compiling %s...\n", fileName)
	lexerInstance := lexer.New(sourceStr, fileName, errorCollector)

	// Transfer lexer errors to main collector
	if errorCollector.HasErrors() {
		errorCollector.Display()
		return fmt.Errorf("compilation failed")
	}

	// Step 2: Parsing
	parserInstance := parser.New(lexerInstance, errorCollector)
	program := parserInstance.ParseProgram()

	// Transfer parser errors to main collector
	if errorCollector.HasErrors() {
		errorCollector.Display()
		return fmt.Errorf("compilation failed")
	}

	//Step 3: Code Generation
	generator := codegen.NewCGenerator(errorCollector)
	cCode := generator.Generate(program)

	if errorCollector.HasErrors() {
		errorCollector.Display()
		return fmt.Errorf("compilation failed")
	}

	//Step 4: Write C file
	outputPath := strings.TrimSuffix(inputPath, constants.NAVIARY_EXTENSION) + ".c"
	err = os.WriteFile(outputPath, []byte(cCode), 0644)
	if err != nil {
		return fmt.Errorf("failed to write C file: %v", err)
	}

	fmt.Printf("C code generated successfully: %s\n", outputPath)

	return nil
}

func main() {
	// Parse command line arguments
	runFlag := false
	args := os.Args[1:]

	if len(args) > 0 && args[0] == "run" {
		runFlag = true
		args = args[1:]
	}

	if len(args) < 1 {
		fmt.Printf("Usage: naviary [run] <source_file%s>\n", constants.NAVIARY_EXTENSION)
		fmt.Printf("  naviary hello%s       # Compile only\n", constants.NAVIARY_EXTENSION)
		fmt.Printf("  naviary run hello%s   # Compile and run\n", constants.NAVIARY_EXTENSION)
		os.Exit(1)
	}

	inputFile := args[0]

	// Validate file extension
	if !strings.HasSuffix(inputFile, constants.NAVIARY_EXTENSION) {
		fmt.Printf("Error: Input file must have %s extension\n", constants.NAVIARY_EXTENSION)
		os.Exit(1)
	}

	// Check if file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Printf("Error: File '%s' not found\n", inputFile)
		os.Exit(1)
	}

	// Compile the file
	if err := CompileFile(inputFile, runFlag); err != nil {
		fmt.Printf("Compilation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Compilation successful!")
}
