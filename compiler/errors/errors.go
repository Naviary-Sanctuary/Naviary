package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents different kinds of compilation errors
type ErrorType int

const (
	LexicalError ErrorType = iota
	SyntaxError
	TypeError
	CodeGenerationError
)

// CompilerError represents an error during compilation
type CompilerError struct {
	Type    ErrorType
	Message string
	Line    int
	Column  int
	File    string
}

// ErrorCollector collects and reports compilation errors
type ErrorCollector struct {
	Errors     []CompilerError
	sourceCode string
	lines      []string
	fileName   string
}

// NewErrorCollector creates a new error collector
func NewErrorCollector(sourceCode string, fileName string) *ErrorCollector {
	return &ErrorCollector{
		Errors:     []CompilerError{},
		sourceCode: sourceCode,
		lines:      strings.Split(sourceCode, "\n"),
		fileName:   fileName,
	}
}

// Add appends a new error to the collector
func (collector *ErrorCollector) Add(errorType ErrorType, message string, line, column int, file string) {
	collector.Errors = append(collector.Errors, CompilerError{
		Type:    errorType,
		Message: message,
		Line:    line,
		Column:  column,
		File:    file,
	})
}

// HasErrors checks if there are any errors
func (collector *ErrorCollector) HasErrors() bool {
	return len(collector.Errors) > 0
}

// Count returns the number of errors
func (collector *ErrorCollector) Count() int {
	return len(collector.Errors)
}

// Print displays all errors with source context
func (collector *ErrorCollector) Print() {
	for i, err := range collector.Errors {
		collector.printError(err)
		if i < len(collector.Errors)-1 {
			fmt.Println()
		}
	}

	if collector.HasErrors() {
		errorWord := "error"
		if len(collector.Errors) > 1 {
			errorWord = "errors"
		}
		fmt.Printf("\n%serror: could not compile due to %d %s%s\n",
			collector.colorCode("red"),
			len(collector.Errors),
			errorWord,
			collector.colorCode("reset"))
	}
}

// printError displays a single error with source context
func (collector *ErrorCollector) printError(err CompilerError) {
	// Error header
	fmt.Printf("%serror[%s]:%s %s\n",
		collector.colorCode("red"),
		collector.errorTypeName(err.Type),
		collector.colorCode("reset"),
		err.Message)

	// File location
	fmt.Printf("%s  --> %s%s:%d:%d\n",
		collector.colorCode("blue"),
		collector.colorCode("reset"),
		err.File,
		err.Line,
		err.Column)

	// Source context
	collector.printSourceContext(err.Line, err.Column)
}

// printSourceContext shows the code around the error
func (collector *ErrorCollector) printSourceContext(line int, column int) {
	lineIndex := line - 1
	if lineIndex < 0 || lineIndex >= len(collector.lines) {
		return
	}

	gutterWidth := len(fmt.Sprintf("%d", line+1))

	// Show line before
	if lineIndex > 0 {
		fmt.Printf("   %*d | %s\n", gutterWidth, line-1, collector.lines[lineIndex-1])
	}

	// Show error line
	fmt.Printf("   %s%*d |%s %s\n",
		collector.colorCode("blue"),
		gutterWidth,
		line,
		collector.colorCode("reset"),
		collector.lines[lineIndex])

	// Show caret
	fmt.Printf("   %s | ", strings.Repeat(" ", gutterWidth))
	if column > 0 {
		fmt.Print(strings.Repeat(" ", column-1))
	}
	fmt.Printf("%s^", collector.colorCode("red"))

	// Wavy underline
	tokenLength := collector.estimateTokenLength(collector.lines[lineIndex], column-1)
	if tokenLength > 1 {
		fmt.Print(strings.Repeat("~", tokenLength-1))
	}
	fmt.Printf("%s\n", collector.colorCode("reset"))

	// Show line after
	if lineIndex < len(collector.lines)-1 {
		fmt.Printf("   %*d | %s\n", gutterWidth, line+1, collector.lines[lineIndex+1])
	}
}

// estimateTokenLength estimates token length at position
func (collector *ErrorCollector) estimateTokenLength(line string, startPos int) int {
	if startPos < 0 || startPos >= len(line) {
		return 1
	}

	length := 0
	for i := startPos; i < len(line); i++ {
		char := line[i]
		if char == ' ' || char == '\t' || char == '(' || char == ')' ||
			char == '{' || char == '}' || char == ',' || char == ';' {
			break
		}
		length++
	}

	if length == 0 {
		return 1
	}
	return length
}

// errorTypeName returns human-readable error type
func (collector *ErrorCollector) errorTypeName(errorType ErrorType) string {
	switch errorType {
	case LexicalError:
		return "Lexical Error"
	case SyntaxError:
		return "Syntax Error"
	case TypeError:
		return "Type Error"
	case CodeGenerationError:
		return "Code Generation Error"
	default:
		return "Error"
	}
}

// colorCode returns ANSI color code
func (collector *ErrorCollector) colorCode(color string) string {
	codes := map[string]string{
		"red":   "\033[31;1m",
		"blue":  "\033[34;1m",
		"reset": "\033[0m",
	}

	if code, ok := codes[color]; ok {
		return code
	}
	return ""
}
