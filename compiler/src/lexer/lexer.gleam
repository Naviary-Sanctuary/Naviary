import gleam/float
import gleam/int
import gleam/string
import lexer/character
import lexer/token.{type Token, type TokenWithPosition}

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

  let is_newline = case lexer.current_char {
    "\n" -> True
    _ -> False
  }

  let new_line = case is_newline {
    True -> lexer.line + 1
    False -> lexer.line
  }
  let new_column = case is_newline {
    True -> 0
    False -> lexer.column + 1
  }

  Lexer(
    ..lexer,
    position: lexer.next_position,
    next_position: lexer.next_position + 1,
    current_char: next_char,
    line: new_line,
    column: new_column,
  )
}

/// Peek at the next character without advancing the lexer
pub fn peek(lexer: Lexer) -> String {
  string.slice(lexer.input, lexer.next_position, 1)
}

/// Skip whitespace characters (spaces, tabs, carriage returns)
/// 
/// Return a new lexer positioned at the first non-whitespace character
pub fn skip_whitespace(lexer: Lexer) -> Lexer {
  case character.is_whitespace(lexer.current_char) {
    True -> {
      lexer |> advance() |> skip_whitespace()
    }
    False -> lexer
  }
}

fn read_digits(lexer: Lexer, accumulated: String) -> #(Lexer, String) {
  case character.is_digit(lexer.current_char) {
    True -> {
      let new_accumulated = accumulated <> lexer.current_char
      let new_lexer = lexer |> advance()
      read_digits(new_lexer, new_accumulated)
    }
    False -> #(lexer, accumulated)
  }
}

/// Read a number (integer or float)
///
/// Return the updated lexer and the token
pub fn read_number(lexer: Lexer) -> #(Lexer, Token) {
  let #(lexer_after_int, integer_part) = read_digits(lexer, "")

  case lexer_after_int.current_char {
    "." -> {
      let next_char = peek(lexer_after_int)
      case character.is_digit(next_char) {
        True -> {
          let lexer_after_dot = lexer_after_int |> advance()
          let #(final_lexer, fractional_part) = read_digits(lexer_after_dot, "")

          let float_string = integer_part <> "." <> fractional_part
          case float.parse(float_string) {
            Ok(value) -> #(final_lexer, token.FloatLiteral(value))
            Error(_) -> #(final_lexer, token.EOF)
          }
        }
        False -> {
          case int.parse(integer_part) {
            Ok(value) -> #(lexer_after_int, token.IntLiteral(value))
            Error(_) -> #(lexer_after_int, token.EOF)
          }
        }
      }
    }
    _ -> {
      case int.parse(integer_part) {
        Ok(value) -> #(lexer_after_int, token.IntLiteral(value))
        Error(_) -> #(lexer_after_int, token.EOF)
      }
    }
  }
}

fn read_identifier(lexer: Lexer, accumulated: String) -> #(Lexer, String) {
  case character.is_identifier_continue(lexer.current_char) {
    True -> {
      let new_accumulated = accumulated <> lexer.current_char
      let new_lexer = lexer |> advance()
      read_identifier(new_lexer, new_accumulated)
    }
    False -> #(lexer, accumulated)
  }
}

fn identifier_to_token(identifier: String) -> Token {
  case identifier {
    "let" -> token.Let
    "func" -> token.Func
    "if" -> token.If
    "return" -> token.Return
    "int" -> token.Int
    "float" -> token.Float
    "string" -> token.String
    "bool" -> token.Bool
    "nil" -> token.Nil
    "true" -> token.BoolLiteral(True)
    "false" -> token.BoolLiteral(False)
    _ -> token.Identifier(identifier)
  }
}

pub fn read_identifier_or_keyword(lexer: Lexer) -> #(Lexer, Token) {
  let #(new_lexer, identifier) = read_identifier(lexer, "")
  let token = identifier_to_token(identifier)
  #(new_lexer, token)
}

pub fn read_operator(lexer: Lexer) -> #(Lexer, Token) {
  case lexer.current_char {
    "+" -> #(lexer |> advance(), token.Plus)
    "-" -> {
      case peek(lexer) {
        ">" -> {
          let lexer_after_arrow = lexer |> advance() |> advance()
          #(lexer_after_arrow, token.Arrow)
        }
        _ -> #(lexer |> advance(), token.Minus)
      }
    }
    "*" -> #(lexer |> advance(), token.Asterisk)
    "/" -> #(lexer |> advance(), token.Slash)
    "=" -> {
      case peek(lexer) {
        "=" -> #(lexer |> advance() |> advance(), token.Equals)
        _ -> #(lexer |> advance(), token.Assign)
      }
    }
    "!" -> {
      case peek(lexer) {
        "=" -> #(lexer |> advance() |> advance(), token.NotEquals)
        // TODO: handle this case
        _ -> #(lexer |> advance(), token.EOF)
      }
    }
    "<" -> {
      case peek(lexer) {
        "=" -> #(lexer |> advance() |> advance(), token.LessThanOrEqual)
        _ -> #(lexer |> advance(), token.LessThan)
      }
    }
    ">" -> {
      case peek(lexer) {
        "=" -> #(lexer |> advance() |> advance(), token.GreaterThanOrEqual)
        _ -> #(lexer |> advance(), token.GreaterThan)
      }
    }
    "(" -> #(lexer |> advance(), token.LeftParen)
    ")" -> #(lexer |> advance(), token.RightParen)
    "{" -> #(lexer |> advance(), token.LeftBrace)
    "}" -> #(lexer |> advance(), token.RightBrace)
    "[" -> #(lexer |> advance(), token.LeftBracket)
    "]" -> #(lexer |> advance(), token.RightBracket)
    "," -> #(lexer |> advance(), token.Comma)
    ";" -> #(lexer |> advance(), token.Semicolon)
    ":" -> #(lexer |> advance(), token.Colon)
    _ -> #(lexer |> advance(), token.EOF)
  }
}

pub fn next_token(lexer: Lexer) -> #(Lexer, TokenWithPosition) {
  let lexer_after_whitespace = lexer |> skip_whitespace()

  let position =
    token.Position(
      line: lexer_after_whitespace.line,
      column: lexer_after_whitespace.column,
    )

  case lexer_after_whitespace.current_char {
    "" -> {
      let token_with_position = token.TokenWithPosition(token.EOF, position)
      #(lexer_after_whitespace, token_with_position)
    }
    "\n" -> {
      let new_lexer = advance(lexer_after_whitespace)
      let token_with_position = token.TokenWithPosition(token.Newline, position)
      #(new_lexer, token_with_position)
    }
    char -> {
      let #(new_lexer, token_type) = case character.is_digit(char) {
        True -> read_number(lexer_after_whitespace)
        False -> {
          case character.is_identifier_start(char) {
            True -> read_identifier_or_keyword(lexer_after_whitespace)
            False -> read_operator(lexer_after_whitespace)
          }
        }
      }

      #(new_lexer, token.TokenWithPosition(token_type, position))
    }
  }
}
