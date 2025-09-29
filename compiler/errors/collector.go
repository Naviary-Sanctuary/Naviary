package errors

import (
	"fmt"
	"os"
	"strings"
)

type ErrorCollector struct {
	errors    []CompileError
	source    string
	filename  string
	maxErrors int // prevent spamming errors
}

func New(source, filename string) *ErrorCollector {
	return &ErrorCollector{
		errors:    make([]CompileError, 0),
		source:    source,
		filename:  filename,
		maxErrors: 100,
	}
}

func (collector *ErrorCollector) ReportAndExit() {
	if collector.HasErrors() {
		collector.Display()
		os.Exit(1)
	}
}

func (collector *ErrorCollector) HasErrors() bool {
	return len(collector.errors) > 0
}

func (collector *ErrorCollector) Display() {
	for _, err := range collector.errors {
		fmt.Fprint(os.Stderr, err.Display())
		fmt.Fprintln(os.Stderr)
	}

	if len(collector.errors) == 1 {
		fmt.Fprintf(os.Stderr, "\033[1;31merror\033[0m: aborting due to previous error\n")
	} else if len(collector.errors) > 1 {
		fmt.Fprintf(os.Stderr, "\033[1;31merror\033[0m: aborting due to %d previous errors\n",
			len(collector.errors))
	}
}

func (collector *ErrorCollector) Add(
	errorType ErrorType,
	line,
	column,
	length int,
	format string,
	args ...interface{},
) {
	sourceLine := collector.getSourceLine(line)

	err := CompileError{
		Type:    errorType,
		Message: fmt.Sprintf(format, args...), // Format message with args
		File:    collector.filename,
		Line:    line,
		Column:  column,
		Length:  length,
		Source:  sourceLine,
	}

	collector.errors = append(collector.errors, err)

	if len(collector.errors) >= collector.maxErrors {
		collector.ReportAndExit()
	}
}

func (collector *ErrorCollector) Clear() {
	collector.errors = collector.errors[:0]
}

func (collector *ErrorCollector) getSourceLine(lineNumber int) string {
	// Split source into lines
	lines := strings.Split(collector.source, "\n")

	// Check bounds (lineNumber is 1-based)
	if lineNumber > 0 && lineNumber <= len(lines) {
		return lines[lineNumber-1]
	}
	return ""
}
