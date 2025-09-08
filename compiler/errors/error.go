package errors

import (
	"fmt"
	"strings"
)

type ErrorType int

const (
	LexicalError ErrorType = iota
	SyntaxError
	TypeError
	RuntimeError
)

var errorTypeMap = map[ErrorType]string{
	LexicalError: "Lexical Error",
	SyntaxError:  "Syntax Error",
	TypeError:    "Type Error",
	RuntimeError: "Runtime Error",
}

func (e ErrorType) String() string {
	return errorTypeMap[e]
}

type CompileError struct {
	Type    ErrorType
	Message string
	File    string
	Line    int
	Column  int
	Length  int
	Source  string
}

func (e CompileError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s error: %s",
		e.File, e.Line, e.Column, e.Type, e.Message)
}

func (e CompileError) Display() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("\033[1;31merror\033[0m: %s\n", e.Message))
	builder.WriteString(fmt.Sprintf("  \033[1;34m-->\033[0m %s:%d:%d\n",
		e.File, e.Line, e.Column))

	if e.Source != "" {
		lineNum := fmt.Sprintf("%d", e.Line)
		builder.WriteString("   \033[1;34m|\033[0m\n")
		builder.WriteString(fmt.Sprintf(" \033[1;34m%s |\033[0m %s\n",
			lineNum, e.Source))

		spaces := e.Column - 1
		underline := strings.Repeat("^", e.Length)
		if e.Length == 0 {
			underline = "^"
		}
		builder.WriteString(fmt.Sprintf("   \033[1;34m|\033[0m %*s\033[1;31m%s\033[0m\n",
			spaces+len(lineNum)+1, "", underline))
	}

	return builder.String()
}
