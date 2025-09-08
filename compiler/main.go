package main

import (
	"fmt"
	"naviary/compiler/lexer"
	"naviary/compiler/token"
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

	// Create lexer
	lex := lexer.New(string(source), filename)

	// Tokenize
	var tokens []token.Token
	for {
		t := lex.NextToken()
		tokens = append(tokens, t)

		if t.Type == token.EOF {
			break
		}
	}

	// Check for errors
	if lex.Errors().HasErrors() {
		lex.Errors().Display()
		os.Exit(1)
	}

	fmt.Printf("Successfully tokenized %d tokens\n", len(tokens))
}
