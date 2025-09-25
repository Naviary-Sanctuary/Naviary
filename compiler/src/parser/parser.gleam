import ast/ast
import gleam/float
import gleam/int
import gleam/list
import lexer/token.{type Token, type TokenWithPosition}

pub type Parser {
  Parser(tokens: List(TokenWithPosition), consumed: List(TokenWithPosition))
}

pub type ParseError {
  ParseError(message: String, position: token.Position)
}

pub fn new(tokens: List(TokenWithPosition)) -> Parser {
  Parser(tokens, consumed: [])
}

pub fn parse_primary_expression(
  parser: Parser,
) -> Result(#(Parser, ast.Expression), ParseError) {
  case current_token(parser) {
    Ok(token_with_position) -> {
      case token_with_position.token {
        token.IntLiteral(value) -> {
          let new_parser = parser |> advance()
          Ok(#(new_parser, ast.IntegerLiteral(value)))
        }
        token.FloatLiteral(value) -> {
          let new_parser = parser |> advance()
          Ok(#(new_parser, ast.FloatLiteral(value)))
        }
        token.BoolLiteral(value) -> {
          let new_parser = parser |> advance()
          Ok(#(new_parser, ast.BoolLiteral(value)))
        }
        token.StringLiteral(value) -> {
          let new_parser = parser |> advance()
          Ok(#(new_parser, ast.StringLiteral(value)))
        }

        token.Identifier(name) -> {
          let new_parser = parser |> advance()
          case current_token(new_parser) {
            Ok(next_token) -> {
              case next_token.token {
                token.LeftParen -> {
                  parse_function_call(new_parser, name)
                }
                _ -> {
                  Ok(#(new_parser, ast.Identifier(name)))
                }
              }
            }
            Error(_) -> {
              Ok(#(new_parser, ast.Identifier(name)))
            }
          }
        }

        token.LeftParen -> {
          let parser_after_paren = parser |> advance()
          case parse_expression(parser_after_paren) {
            Ok(#(parser_with_expression, expression)) -> {
              case expect(parser_with_expression, token.RightParen) {
                Ok(final_parser) -> Ok(#(final_parser, expression))
                Error(parse_error) -> Error(parse_error)
              }
            }
            Error(parse_error) -> Error(parse_error)
          }
        }

        other -> {
          let message = "Expected expression but got " <> describe_token(other)
          Error(ParseError(message, token_with_position.position))
        }
      }
    }
    Error(parse_error) -> Error(parse_error)
  }
}

fn parse_function_call(
  parser: Parser,
  function_name: String,
) -> Result(#(Parser, ast.Expression), ParseError) {
  let parser_after_paren = parser |> advance()

  case parse_arguments(parser_after_paren, []) {
    Ok(#(parser_with_arguments, arguments)) -> {
      case expect(parser_with_arguments, token.RightParen) {
        Ok(final_parser) ->
          Ok(#(final_parser, ast.FunctionExpression(function_name, arguments)))
        Error(parse_error) -> Error(parse_error)
      }
    }
    Error(parse_error) -> Error(parse_error)
  }
}

fn parse_arguments(
  parser: Parser,
  accumulated: List(ast.Expression),
) -> Result(#(Parser, List(ast.Expression)), ParseError) {
  case current_token(parser) {
    Ok(token_with_position) -> {
      case token_with_position.token {
        token.RightParen -> {
          Ok(#(parser, list.reverse(accumulated)))
        }
        _ -> {
          case parse_expression(parser) {
            Ok(#(parser_after_expression, expression)) -> {
              let new_accumulated = [expression, ..accumulated]

              case current_token(parser_after_expression) {
                Ok(next_token) -> {
                  case next_token.token {
                    token.Comma -> {
                      let parser_after_comma =
                        parser_after_expression |> advance()
                      parse_arguments(parser_after_comma, new_accumulated)
                    }
                    token.RightParen -> {
                      Ok(#(
                        parser_after_expression,
                        list.reverse(new_accumulated),
                      ))
                    }
                    other -> {
                      let message =
                        "Expected ',' or ')' but got " <> describe_token(other)
                      Error(ParseError(message, next_token.position))
                    }
                  }
                }
                Error(parse_error) -> Error(parse_error)
              }
            }
            Error(parse_error) -> Error(parse_error)
          }
        }
      }
    }
    Error(parse_error) -> Error(parse_error)
  }
}

fn current_token(parser: Parser) -> Result(TokenWithPosition, ParseError) {
  case parser.tokens {
    [token, ..] -> Ok(token)
    [] -> {
      let position = token.Position(line: 0, column: 0)
      Error(ParseError(message: "Unexpected end of input", position: position))
    }
  }
}

fn peek_token(parser: Parser) -> Result(TokenWithPosition, ParseError) {
  case parser.tokens {
    [_, next, ..] -> Ok(next)
    [_] | [] -> {
      let position = token.Position(line: 0, column: 0)
      Error(ParseError(message: "Unexpected end of input", position: position))
    }
  }
}

fn advance(parser: Parser) -> Parser {
  case parser.tokens {
    [current, ..rest] ->
      Parser(tokens: rest, consumed: [current, ..parser.consumed])
    [] -> parser
  }
}

fn expect(parser: Parser, expected: Token) -> Result(Parser, ParseError) {
  case current_token(parser) {
    Ok(token_with_position) -> {
      case compare_token_type(token_with_position.token, expected) {
        True -> Ok(advance(parser))
        False -> {
          let message =
            "Expected "
            <> describe_token(expected)
            <> " but got "
            <> describe_token(token_with_position.token)
          Error(ParseError(message, token_with_position.position))
        }
      }
    }
    Error(parse_error) -> Error(parse_error)
  }
}

fn compare_token_type(actual: Token, expected: Token) -> Bool {
  case actual, expected {
    // For tokens with values, just check the variant type
    token.IntLiteral(_), token.IntLiteral(_) -> True
    token.FloatLiteral(_), token.FloatLiteral(_) -> True
    token.StringLiteral(_), token.StringLiteral(_) -> True
    token.BoolLiteral(_), token.BoolLiteral(_) -> True
    token.Identifier(_), token.Identifier(_) -> True
    // For simple tokens, they must be exactly the same
    a, b -> a == b
  }
}

fn describe_token(token: Token) -> String {
  case token {
    token.Let -> "'let'"
    token.Func -> "'func'"
    token.If -> "'if'"
    token.Return -> "'return'"
    token.Int -> "type 'int'"
    token.Float -> "type 'float'"
    token.String -> "type 'string'"
    token.Bool -> "type 'bool'"
    token.IntLiteral(value) -> "integer " <> int.to_string(value)
    token.FloatLiteral(value) -> "float " <> float.to_string(value)
    token.StringLiteral(value) -> "string \"" <> value <> "\""
    token.BoolLiteral(True) -> "true"
    token.BoolLiteral(False) -> "false"
    token.Identifier(name) -> "identifier '" <> name <> "'"
    token.Plus -> "'+'"
    token.Minus -> "'-'"
    token.Asterisk -> "'*'"
    token.Slash -> "'/'"
    token.Assign -> "'='"
    token.Equals -> "'=='"
    token.NotEquals -> "'!='"
    token.LessThan -> "'<'"
    token.GreaterThan -> "'>'"
    token.LeftParen -> "'('"
    token.RightParen -> "')'"
    token.LeftBrace -> "'{'"
    token.RightBrace -> "'}'"
    token.LeftBracket -> "'['"
    token.RightBracket -> "']'"
    token.Comma -> "','"
    token.Semicolon -> "';'"
    token.Colon -> "':'"
    token.Arrow -> "'->'"
    token.Newline -> "newline"
    token.EOF -> "end of file"
    token.Nil -> "'nil'"
    token.NilLiteral -> "nil"
    _ -> "unknown token"
  }
}

fn precedence(token: Token) -> Int {
  case token {
    token.Asterisk | token.Slash -> 6
    token.Plus | token.Minus -> 5
    token.LessThan
    | token.GreaterThan
    | token.LessThanOrEqual
    | token.GreaterThanOrEqual -> 4
    token.Equals | token.NotEquals -> 3
    _ -> 0
  }
}

fn token_to_binary_operator(token: Token) -> Result(ast.BinaryOperator, Nil) {
  case token {
    token.Plus -> Ok(ast.Add)
    token.Minus -> Ok(ast.Subtract)
    token.Asterisk -> Ok(ast.Multiply)
    token.Slash -> Ok(ast.Divide)
    token.Equals -> Ok(ast.Equal)
    token.NotEquals -> Ok(ast.NotEqual)
    token.LessThan -> Ok(ast.LessThan)
    token.GreaterThan -> Ok(ast.GreaterThan)
    _ -> Error(Nil)
  }
}

fn parse_expression(
  parser: Parser,
) -> Result(#(Parser, ast.Expression), ParseError) {
  parse_expression_pratt(parser, 0)
}

// Parse expression using Pratt parsing (operator precedence)
fn parse_expression_pratt(
  parser: Parser,
  min_precedence: Int,
) -> Result(#(Parser, ast.Expression), ParseError) {
  // Parse left side
  case parse_primary_expression(parser) {
    Error(err) -> Error(err)
    Ok(#(parser_after_left, left)) -> {
      // Try to extend with binary operators
      parse_binary_ops(parser_after_left, left, min_precedence)
    }
  }
}

// Parse binary operators with correct precedence
fn parse_binary_ops(
  parser: Parser,
  left: ast.Expression,
  min_precedence: Int,
) -> Result(#(Parser, ast.Expression), ParseError) {
  case current_token(parser) {
    Error(_) -> Ok(#(parser, left))
    // End of input
    Ok(token_with_pos) -> {
      let prec = precedence(token_with_pos.token)

      case prec <= min_precedence {
        True -> Ok(#(parser, left))
        // Precedence too low, stop here
        False -> {
          // Try to convert to binary operator
          case token_to_binary_operator(token_with_pos.token) {
            Error(_) -> Ok(#(parser, left))
            // Not an operator, stop here
            Ok(operator) -> {
              let parser_after_op = advance(parser)

              // Parse right side with higher precedence for left-associativity
              case parse_expression_pratt(parser_after_op, prec) {
                Error(err) -> Error(err)
                Ok(#(parser_after_right, right)) -> {
                  // Create new binary expression
                  let new_expr = ast.BinaryExpression(left, operator, right)
                  // Continue parsing with same min_precedence
                  parse_binary_ops(parser_after_right, new_expr, min_precedence)
                }
              }
            }
          }
        }
      }
    }
  }
}
