import gleam/string

pub type Lexer {
  Lexer(
    input: String,
    position: Int,
    next_position: Int,
    current_char: String,
    line: Int,
    column: Int,
  )
}

pub fn new(input: String) -> Lexer {
  let first_char = case string.slice(input, 0, 1) {
    "" -> ""
    char -> char
  }

  Lexer(
    input,
    position: 0,
    next_position: 1,
    current_char: first_char,
    line: 1,
    column: 0,
  )
}

pub fn advance(lexer: Lexer) -> Lexer {
  let next_char = case string.slice(lexer.input, lexer.next_position, 1) {
    // End of input
    "" -> ""
    char -> char
  }

  Lexer(
    ..lexer,
    position: lexer.next_position,
    next_position: lexer.next_position + 1,
    current_char: next_char,
    column: lexer.column + 1,
  )
}

pub fn peek(lexer: Lexer) -> String {
  string.slice(lexer.input, lexer.next_position, 1)
}
