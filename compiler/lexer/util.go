package lexer

func isLetter(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_'
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

// keep new line(\n) for semicolon insertion
func isWhitespace(char byte) bool {
	return char == ' ' || char == '\t' || char == '\r'
}

func containsDot(s string) bool {
	for _, char := range s {
		if char == '.' {
			return true
		}
	}
	return false
}
