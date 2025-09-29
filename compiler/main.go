package main

import (
	"compiler/codegen"
	"compiler/constants"
	"compiler/errors"
	"compiler/lexer"
	"compiler/parser"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CompileFile compiles a single Naviary source file to Erlang
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

	// Step 3: Code Generation
	generator := codegen.New(fileName, errorCollector)
	erlangCode, err := generator.GenerateToFile(program)
	if err != nil {
		if errorCollector.HasErrors() {
			errorCollector.Display()
		}
		return fmt.Errorf("code generation failed: %v", err)
	}

	// Step 4: Write output file
	outputPath := strings.TrimSuffix(inputPath, constants.NAVIARY_EXTENSION) + constants.ERLANG_EXTENSION
	err = os.WriteFile(outputPath, []byte(erlangCode), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file %s: %v", outputPath, err)
	}

	fmt.Printf("✓ Generated %s\n", outputPath)

	// Step 5: Compile to BEAM
	err = CompileToBeam(outputPath)
	if err != nil {
		return err
	}

	beamFile := strings.TrimSuffix(outputPath, constants.ERLANG_EXTENSION) + constants.BEAM_EXTENSION
	fmt.Printf("✓ Compiled to %s\n", beamFile)

	// Step 6: Run if requested
	if runAfterCompile {
		moduleName := strings.TrimSuffix(filepath.Base(inputPath), constants.NAVIARY_EXTENSION)
		moduleDir := filepath.Dir(outputPath)
		fmt.Println("\n--- Output ---")
		return RunBeam(moduleName, moduleDir)
	}

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

func CompileToBeam(erlangFile string) error {
	// Compile to BEAM and place output in the same directory as the .erl file
	outputDir := filepath.Dir(erlangFile)
	cmd := exec.Command("erlc", "-o", outputDir, erlangFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to compile to BEAM: %v\n%s", err, string(output))
	}

	return nil
}

// RunBeam runs the compiled BEAM file
func RunBeam(moduleName string, moduleDir string) error {
	// Run the Erlang module from its directory so the runtime can find the .beam file
	cmd := exec.Command("erl", "-pa", moduleDir, "-noshell", "-s", moduleName, "start", "-s", "init", "stop")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = moduleDir

	return cmd.Run()
}
