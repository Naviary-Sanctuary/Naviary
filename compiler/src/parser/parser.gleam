import ast/ast
import gleam/float
import gleam/int
import gleam/list
import lexer/lexer
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

pub fn parse(source: String) -> Result(ast.Program, ParseError) {
  let tokens = tokenize(source)
  let parser = new(tokens)

  case parse_program(parser) {
    Ok(#(_final_parser, program)) -> {
      case validate_main_function(program) {
        Ok(_) -> Ok(program)
        Error(error) -> Error(error)
      }
    }
    Error(error) -> Error(error)
  }
}

// Helper function to tokenize entire source
fn tokenize(source: String) -> List(TokenWithPosition) {
  let initial_lexer = lexer.new(source)
  tokenize_recursive(initial_lexer, [])
}

// Recursively tokenize until EOF
fn tokenize_recursive(
  lexer_state: lexer.Lexer,
  accumulated_tokens: List(TokenWithPosition),
) -> List(TokenWithPosition) {
  let #(next_lexer, token_with_position) = lexer_state |> lexer.next_token()

  case token_with_position.token {
    token.EOF -> {
      list.reverse([token_with_position, ..accumulated_tokens])
    }
    _ -> {
      let new_accumulated = [token_with_position, ..accumulated_tokens]
      tokenize_recursive(next_lexer, new_accumulated)
    }
  }
}

pub fn parse_program(
  parser: Parser,
) -> Result(#(Parser, ast.Program), ParseError) {
  parse_program_recursive(parser, [])
}

fn parse_program_recursive(
  parser: Parser,
  accumulated_functions: List(ast.Function),
) -> Result(#(Parser, ast.Program), ParseError) {
  let parser_cleaned = skip_newlines(parser)

  case current_token(parser_cleaned) {
    Error(_) -> {
      let functions = list.reverse(accumulated_functions)
      let program = ast.Program(functions: functions)
      Ok(#(parser_cleaned, program))
    }
    Ok(token_with_position) -> {
      case token_with_position.token {
        token.EOF -> {
          let functions = list.reverse(accumulated_functions)
          let program = ast.Program(functions: functions)
          Ok(#(parser_cleaned, program))
        }
        token.Func -> {
          case parse_function(parser_cleaned) {
            Error(error) -> Error(error)
            Ok(#(parser_after_function, function)) -> {
              let new_accumulated = [function, ..accumulated_functions]
              parse_program_recursive(parser_after_function, new_accumulated)
            }
          }
        }
        other -> {
          let message =
            "Expected 'func' or end of file but got " <> describe_token(other)
          Error(ParseError(message, token_with_position.position))
        }
      }
    }
  }
}

fn skip_newlines(parser: Parser) -> Parser {
  case current_token(parser) {
    Ok(token_with_position) -> {
      case token_with_position.token {
        token.Newline -> {
          parser |> advance() |> skip_newlines()
        }
        _ -> parser
      }
    }
    Error(_) -> parser
  }
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
    token.IntLiteral(_), token.IntLiteral(_) -> True
    token.FloatLiteral(_), token.FloatLiteral(_) -> True
    token.StringLiteral(_), token.StringLiteral(_) -> True
    token.BoolLiteral(_), token.BoolLiteral(_) -> True
    token.Identifier(_), token.Identifier(_) -> True
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

pub fn parse_statement(
  parser: Parser,
) -> Result(#(Parser, ast.Statement), ParseError) {
  case current_token(parser) {
    Error(error) -> Error(error)
    Ok(token_with_position) -> {
      case token_with_position.token {
        token.Let -> parse_let_statement(parser)
        token.Return -> parse_return_statement(parser)
        // token.If -> parse_if_statement(parser)
        // token.For -> parse_for_statement(parser)
        _ -> {
          parse_expression_statement(parser)
        }
      }
    }
  }
}

fn parse_let_statement(
  parser: Parser,
) -> Result(#(Parser, ast.Statement), ParseError) {
  case expect(parser, token.Let) {
    Error(parse_error) -> Error(parse_error)
    Ok(parser_after_let) -> {
      let #(is_mutable, parser_after_mut) = case
        current_token(parser_after_let)
      {
        Ok(token_with_position) -> {
          case token_with_position.token {
            token.Identifier("mut") -> #(True, parser_after_let |> advance())
            _ -> #(False, parser_after_let)
          }
        }
        Error(_) -> #(False, parser_after_let)
      }

      case current_token(parser_after_mut) {
        Error(parse_error) -> Error(parse_error)
        Ok(token_with_position) -> {
          case token_with_position.token {
            token.Identifier(variable_name) -> {
              let parser_after_name = parser_after_mut |> advance()

              case expect(parser_after_name, token.Assign) {
                Error(parse_error) -> Error(parse_error)
                Ok(parser_after_assign) -> {
                  case parse_expression(parser_after_assign) {
                    Error(parse_error) -> Error(parse_error)
                    Ok(#(final_parser, value_expression)) -> {
                      let statement =
                        ast.LetStatement(
                          variable_name,
                          is_mutable,
                          value_expression,
                        )
                      Ok(#(final_parser, statement))
                    }
                  }
                }
              }
            }

            other -> {
              let message =
                "Expected variable name after 'let' but got "
                <> describe_token(other)
              Error(ParseError(message, token_with_position.position))
            }
          }
        }
      }
    }
  }
}

fn parse_return_statement(
  parser: Parser,
) -> Result(#(Parser, ast.Statement), ParseError) {
  case expect(parser, token.Return) {
    Error(parse_error) -> Error(parse_error)
    Ok(parser_after_return) -> {
      case current_token(parser_after_return) {
        Error(_) -> {
          // End of input, return with nil value
          let statement = ast.ReturnStatement(value: ast.NilLiteral)
          Ok(#(parser_after_return, statement))
        }
        Ok(token_with_position) -> {
          case token_with_position.token {
            token.Semicolon | token.RightBrace | token.Newline -> {
              let statement = ast.ReturnStatement(value: ast.NilLiteral)
              Ok(#(parser_after_return, statement))
            }
            _ -> {
              case parse_expression(parser_after_return) {
                Error(parse_error) -> Error(parse_error)
                Ok(#(final_parser, return_expression)) -> {
                  let statement = ast.ReturnStatement(value: return_expression)
                  Ok(#(final_parser, statement))
                }
              }
            }
          }
        }
      }
    }
  }
}

fn parse_expression_statement(
  parser: Parser,
) -> Result(#(Parser, ast.Statement), ParseError) {
  case parse_expression(parser) {
    Error(error) -> Error(error)
    Ok(#(parser_after_expression, expression)) -> {
      let statement = ast.ExpressionStatement(expression: expression)
      Ok(#(parser_after_expression, statement))
    }
  }
}

fn parse_block(
  parser: Parser,
) -> Result(#(Parser, List(ast.Statement)), ParseError) {
  case expect(parser, token.LeftBrace) {
    Error(error) -> Error(error)
    Ok(parser_after_brace) -> {
      parse_block_statements(parser_after_brace, [])
    }
  }
}

fn parse_block_statements(
  parser: Parser,
  accumulated_statements: List(ast.Statement),
) -> Result(#(Parser, List(ast.Statement)), ParseError) {
  let parser_cleaned = skip_statement_terminators(parser)

  case current_token(parser_cleaned) {
    Error(error) -> Error(error)
    Ok(token_with_position) -> {
      case token_with_position.token {
        token.RightBrace -> {
          let parser_after_brace = advance(parser_cleaned)
          let statements = list.reverse(accumulated_statements)
          Ok(#(parser_after_brace, statements))
        }
        _ -> {
          case parse_statement(parser_cleaned) {
            Error(error) -> Error(error)
            Ok(#(parser_after_statement, statement)) -> {
              let new_accumulated = [statement, ..accumulated_statements]

              let parser_after_terminator =
                consume_optional_terminator(parser_after_statement)

              parse_block_statements(parser_after_terminator, new_accumulated)
            }
          }
        }
      }
    }
  }
}

fn skip_statement_terminators(parser: Parser) -> Parser {
  case current_token(parser) {
    Ok(token_with_position) -> {
      case token_with_position.token {
        token.Newline | token.Semicolon -> {
          parser |> advance() |> skip_statement_terminators()
        }
        _ -> parser
      }
    }
    Error(_) -> parser
  }
}

fn consume_optional_terminator(parser: Parser) -> Parser {
  case current_token(parser) {
    Ok(token_with_position) -> {
      case token_with_position.token {
        token.Newline | token.Semicolon -> advance(parser)
        _ -> parser
      }
    }
    Error(_) -> parser
  }
}

// Parse type annotation: int, float, string, bool, nil
fn parse_type(parser: Parser) -> Result(#(Parser, ast.Type), ParseError) {
  case current_token(parser) {
    Error(error) -> Error(error)
    Ok(token_with_position) -> {
      case token_with_position.token {
        token.Int -> Ok(#(advance(parser), ast.Int))
        token.Float -> Ok(#(advance(parser), ast.Float))
        token.String -> Ok(#(advance(parser), ast.String))
        token.Bool -> Ok(#(advance(parser), ast.Bool))
        token.Nil -> Ok(#(advance(parser), ast.Nil))
        other -> {
          let message = "Expected type but got " <> describe_token(other)
          Error(ParseError(message, token_with_position.position))
        }
      }
    }
  }
}

// Parse function parameter: name: Type
fn parse_parameter(
  parser: Parser,
) -> Result(#(Parser, ast.Parameter), ParseError) {
  case current_token(parser) {
    Error(error) -> Error(error)
    Ok(token_with_position) -> {
      case token_with_position.token {
        token.Identifier(param_name) -> {
          let parser_after_name = advance(parser)

          // Expect colon
          case expect(parser_after_name, token.Colon) {
            Error(error) -> Error(error)
            Ok(parser_after_colon) -> {
              // Parse the type
              case parse_type(parser_after_colon) {
                Error(error) -> Error(error)
                Ok(#(parser_after_type, param_type)) -> {
                  let parameter =
                    ast.Parameter(name: param_name, parameter_type: param_type)
                  Ok(#(parser_after_type, parameter))
                }
              }
            }
          }
        }
        other -> {
          let message =
            "Expected parameter name but got " <> describe_token(other)
          Error(ParseError(message, token_with_position.position))
        }
      }
    }
  }
}

// Parse parameter list: (param1: Type1, param2: Type2, ...)
fn parse_parameter_list(
  parser: Parser,
) -> Result(#(Parser, List(ast.Parameter)), ParseError) {
  // Expect opening parenthesis
  case expect(parser, token.LeftParen) {
    Error(error) -> Error(error)
    Ok(parser_after_paren) -> {
      // Check if parameters exist or empty list
      case current_token(parser_after_paren) {
        Error(error) -> Error(error)
        Ok(token_with_position) -> {
          case token_with_position.token {
            // Empty parameter list
            token.RightParen -> {
              let parser_after_close = advance(parser_after_paren)
              Ok(#(parser_after_close, []))
            }
            // Parse parameters
            _ -> parse_parameters_recursive(parser_after_paren, [])
          }
        }
      }
    }
  }
}

// Helper to parse multiple parameters recursively
fn parse_parameters_recursive(
  parser: Parser,
  accumulated_params: List(ast.Parameter),
) -> Result(#(Parser, List(ast.Parameter)), ParseError) {
  // Parse one parameter
  case parse_parameter(parser) {
    Error(error) -> Error(error)
    Ok(#(parser_after_param, parameter)) -> {
      let new_accumulated = [parameter, ..accumulated_params]

      // Check what comes next
      case current_token(parser_after_param) {
        Error(error) -> Error(error)
        Ok(token_with_position) -> {
          case token_with_position.token {
            // More parameters
            token.Comma -> {
              let parser_after_comma = advance(parser_after_param)
              parse_parameters_recursive(parser_after_comma, new_accumulated)
            }
            // End of parameter list
            token.RightParen -> {
              let parser_after_paren = advance(parser_after_param)
              let parameters = list.reverse(new_accumulated)
              Ok(#(parser_after_paren, parameters))
            }
            other -> {
              let message =
                "Expected ',' or ')' in parameter list but got "
                <> describe_token(other)
              Error(ParseError(message, token_with_position.position))
            }
          }
        }
      }
    }
  }
}

// Parse function: func name(params) -> ReturnType { body }
pub fn parse_function(
  parser: Parser,
) -> Result(#(Parser, ast.Function), ParseError) {
  // Expect 'func' keyword
  case expect(parser, token.Func) {
    Error(error) -> Error(error)
    Ok(parser_after_func) -> {
      // Parse function name
      case current_token(parser_after_func) {
        Error(error) -> Error(error)
        Ok(token_with_position) -> {
          case token_with_position.token {
            token.Identifier(function_name) -> {
              let parser_after_name = advance(parser_after_func)

              // Parse parameter list
              case parse_parameter_list(parser_after_name) {
                Error(error) -> Error(error)
                Ok(#(parser_after_params, parameters)) -> {
                  // Parse return type (optional, defaults to Nil)
                  let #(parser_after_return, return_type) =
                    parse_optional_return_type(parser_after_params)

                  // Parse function body
                  case parse_block(parser_after_return) {
                    Error(error) -> Error(error)
                    Ok(#(final_parser, body_statements)) -> {
                      let function =
                        ast.Function(
                          name: function_name,
                          parameters: parameters,
                          return_type: return_type,
                          body: body_statements,
                        )
                      Ok(#(final_parser, function))
                    }
                  }
                }
              }
            }
            other -> {
              let message =
                "Expected function name but got " <> describe_token(other)
              Error(ParseError(message, token_with_position.position))
            }
          }
        }
      }
    }
  }
}

// Parse optional return type: -> Type or nothing (defaults to Nil)
fn parse_optional_return_type(parser: Parser) -> #(Parser, ast.Type) {
  case current_token(parser) {
    Ok(token_with_position) -> {
      case token_with_position.token {
        token.Arrow -> {
          let parser_after_arrow = advance(parser)
          // Try to parse the return type
          case parse_type(parser_after_arrow) {
            Ok(#(parser_after_type, return_type)) -> {
              #(parser_after_type, return_type)
            }
            Error(_) -> {
              // If type parsing fails, default to Nil
              #(parser_after_arrow, ast.Nil)
            }
          }
        }
        _ -> {
          // No arrow, default return type is Nil
          #(parser, ast.Nil)
        }
      }
    }
    Error(_) -> {
      // End of input, default to Nil
      #(parser, ast.Nil)
    }
  }
}

pub fn validate_main_function(program: ast.Program) -> Result(Nil, ParseError) {
  let has_main =
    list.any(program.functions, fn(function) { function.name == "main" })

  case has_main {
    True -> Ok(Nil)
    False -> {
      let position = token.Position(line: 0, column: 0)
      Error(ParseError(
        message: "No 'main' function found. Every program must have a main function as entry point.",
        position: position,
      ))
    }
  }
}
