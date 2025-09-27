package lexer

// isLetter checks if a character can start an identifier
func isLetter(char byte) bool {
	return ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '_'
}

// isDigit checks if a character is a digit
func isDigit(char byte) bool {
	return '0' <= char && char <= '9'
}
