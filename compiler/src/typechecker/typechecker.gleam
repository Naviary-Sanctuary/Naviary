import ast/ast
import gleam/int
import gleam/list
import gleam/option
import typechecker/symbol_table.{type SymbolTable}

pub type TypeError {
  TypeError(message: String, line: Int, column: Int)
}

pub type TypeChecker {
  TypeChecker(
    symbol_table: SymbolTable,
    errors: List(TypeError),
    current_function_return_type: option.Option(ast.Type),
  )
}

// Create a new type checker
pub fn new() -> TypeChecker {
  TypeChecker(
    symbol_table: symbol_table.new() |> initialize_builtins(),
    errors: [],
    current_function_return_type: option.None,
  )
}

// Add an error to the type checker
fn add_error(
  checker: TypeChecker,
  message: String,
  line: Int,
  column: Int,
) -> TypeChecker {
  let error = TypeError(message, line, column)
  TypeChecker(..checker, errors: [error, ..checker.errors])
}

pub fn is_builtin_any_function(name: String) -> Bool {
  name == "print"
}

fn initialize_builtins(table: SymbolTable) -> SymbolTable {
  table
  |> symbol_table.add_function(
    "print",
    symbol_table.FunctionSignature(parameter_types: [], return_type: ast.Nil),
  )
}

pub fn check_expression(
  checker: TypeChecker,
  expression: ast.Expression,
) -> #(TypeChecker, Result(ast.Type, TypeError)) {
  case expression {
    ast.IntegerLiteral(_value) -> {
      #(checker, Ok(ast.Int))
    }
    ast.FloatLiteral(_value) -> {
      #(checker, Ok(ast.Float))
    }
    ast.BoolLiteral(_value) -> {
      #(checker, Ok(ast.Bool))
    }
    ast.StringLiteral(_value) -> {
      #(checker, Ok(ast.String))
    }
    ast.NilLiteral -> {
      #(checker, Ok(ast.Nil))
    }
    ast.Identifier(name) -> {
      case symbol_table.lookup_variable(checker.symbol_table, name) {
        option.Some(variable_type) -> {
          #(checker, Ok(variable_type))
        }
        option.None -> {
          let error = TypeError("Undefined variable '" <> name <> "'", 0, 0)
          #(
            add_error(checker, error.message, error.line, error.column),
            Error(error),
          )
        }
      }
    }

    ast.BinaryExpression(left, operator, right) -> {
      let #(checker_after_left, left_result) = check_expression(checker, left)
      let #(checker_after_right, right_result) =
        check_expression(checker_after_left, right)

      case left_result, right_result {
        Ok(left_type), Ok(right_type) -> {
          check_binary_operator(
            checker_after_right,
            operator,
            left_type,
            right_type,
          )
        }
        Error(_), _ -> {
          #(
            checker_after_right,
            Error(TypeError("Left operand has errors", 0, 0)),
          )
        }
        _, Error(_) -> {
          #(
            checker_after_right,
            Error(TypeError("Right operand has errors", 0, 0)),
          )
        }
      }
    }

    ast.FunctionExpression(name, arguments) -> {
      // Special handling for built-in functions that accept any type
      case is_builtin_any_function(name) {
        True -> {
          let checker_after_args =
            check_function_arguments_any(checker, arguments)
          #(checker_after_args, Ok(ast.Nil))
        }
        False -> {
          case symbol_table.lookup_function(checker.symbol_table, name) {
            option.Some(signature) -> {
              check_function_call(checker, name, arguments, signature)
            }
            option.None -> {
              let error = TypeError("Undefined function '" <> name <> "'", 0, 0)
              #(
                add_error(checker, error.message, error.line, error.column),
                Error(error),
              )
            }
          }
        }
      }
    }

    _ -> {
      let error =
        TypeError("Expression type checking not yet implemented", 0, 0)
      #(
        add_error(checker, error.message, error.line, error.column),
        Error(error),
      )
    }
  }
}

fn check_binary_operator(
  checker: TypeChecker,
  operator: ast.BinaryOperator,
  left_type: ast.Type,
  right_type: ast.Type,
) -> #(TypeChecker, Result(ast.Type, TypeError)) {
  case operator {
    // Arithmetic operators: only for numbers
    ast.Add | ast.Subtract | ast.Multiply | ast.Divide -> {
      case left_type, right_type {
        ast.Int, ast.Int -> #(checker, Ok(ast.Int))
        ast.Float, ast.Float -> #(checker, Ok(ast.Float))
        ast.Int, ast.Float | ast.Float, ast.Int -> #(checker, Ok(ast.Float))
        _, _ -> {
          let error =
            TypeError(
              "Arithmetic operator requires numeric types, got "
                <> type_to_string(left_type)
                <> " and "
                <> type_to_string(right_type),
              0,
              0,
            )
          #(
            add_error(checker, error.message, error.line, error.column),
            Error(error),
          )
        }
      }
    }

    // Comparison operators: return bool
    ast.Equal | ast.NotEqual -> {
      case left_type == right_type {
        True -> #(checker, Ok(ast.Bool))
        False -> {
          let error =
            TypeError(
              "Equality operator requires same types, got "
                <> type_to_string(left_type)
                <> " and "
                <> type_to_string(right_type),
              0,
              0,
            )
          #(
            add_error(checker, error.message, error.line, error.column),
            Error(error),
          )
        }
      }
    }

    ast.LessThan | ast.GreaterThan -> {
      case left_type, right_type {
        ast.Int, ast.Int
        | ast.Float, ast.Float
        | ast.Int, ast.Float
        | ast.Float, ast.Int
        -> {
          #(checker, Ok(ast.Bool))
        }
        _, _ -> {
          let error =
            TypeError("Comparison operator requires numeric types", 0, 0)
          #(
            add_error(checker, error.message, error.line, error.column),
            Error(error),
          )
        }
      }
    }
  }
}

fn type_to_string(type_: ast.Type) -> String {
  case type_ {
    ast.Int -> "int"
    ast.Float -> "float"
    ast.String -> "string"
    ast.Bool -> "bool"
    ast.Nil -> "nil"
  }
}

fn check_argument_types(
  checker: TypeChecker,
  arguments: List(ast.Expression),
  parameter_types: List(ast.Type),
  function_name: String,
) -> TypeChecker {
  case arguments, parameter_types {
    [], [] -> {
      // All arguments checked successfully
      checker
    }
    [arg, ..rest_args], [param_type, ..rest_params] -> {
      // Check current argument
      let #(checker_after_arg, arg_result) = check_expression(checker, arg)

      // Check if argument type matches parameter type
      let checker_with_check = case arg_result {
        Ok(arg_type) -> {
          case arg_type == param_type {
            True -> checker_after_arg
            False -> {
              let error =
                TypeError(
                  "Function '"
                    <> function_name
                    <> "' expects "
                    <> type_to_string(param_type)
                    <> " but got "
                    <> type_to_string(arg_type),
                  0,
                  0,
                )
              add_error(
                checker_after_arg,
                error.message,
                error.line,
                error.column,
              )
            }
          }
        }
        Error(_) -> {
          // Error already added during expression check
          checker_after_arg
        }
      }

      // Continue with remaining arguments
      check_argument_types(
        checker_with_check,
        rest_args,
        rest_params,
        function_name,
      )
    }
    _, _ -> {
      // This shouldn't happen if check_function_call checks count first
      checker
    }
  }
}

fn check_function_call(
  checker: TypeChecker,
  function_name: String,
  arguments: List(ast.Expression),
  signature: symbol_table.FunctionSignature,
) -> #(TypeChecker, Result(ast.Type, TypeError)) {
  // Check argument count first
  let arg_count = list.length(arguments)
  let param_count = list.length(signature.parameter_types)

  case arg_count == param_count {
    False -> {
      let error =
        TypeError(
          "Function '"
            <> function_name
            <> "' expects "
            <> int.to_string(param_count)
            <> " arguments, got "
            <> int.to_string(arg_count),
          0,
          0,
        )
      #(
        add_error(checker, error.message, error.line, error.column),
        Error(error),
      )
    }
    True -> {
      // Check argument types match parameter types
      let checker_after_args =
        check_argument_types(
          checker,
          arguments,
          signature.parameter_types,
          function_name,
        )
      #(checker_after_args, Ok(signature.return_type))
    }
  }
}

fn check_function_arguments_any(
  checker: TypeChecker,
  arguments: List(ast.Expression),
) -> TypeChecker {
  case arguments {
    [] -> checker
    [first, ..rest] -> {
      let #(checker_after_first, _result) = check_expression(checker, first)
      check_function_arguments_any(checker_after_first, rest)
    }
  }
}

// Check a single statement
pub fn check_statement(
  checker: TypeChecker,
  statement: ast.Statement,
) -> TypeChecker {
  case statement {
    ast.LetStatement(name, _is_mutable, value) -> {
      case
        symbol_table.has_variable_in_current_scope(checker.symbol_table, name)
      {
        True -> {
          let error =
            TypeError(
              "Variable '" <> name <> "' already declared in this scope",
              0,
              0,
            )
          add_error(checker, error.message, error.line, error.column)
        }
        False -> {
          // Check the value expression
          let #(checker_after_value, value_result) =
            check_expression(checker, value)

          case value_result {
            Ok(value_type) -> {
              // Add variable to symbol table
              let updated_table =
                symbol_table.add_variable(
                  checker_after_value.symbol_table,
                  name,
                  value_type,
                )
              TypeChecker(..checker_after_value, symbol_table: updated_table)
            }
            Error(_) -> {
              // Error already added during expression check
              checker_after_value
            }
          }
        }
      }
    }

    ast.ExpressionStatement(expression) -> {
      // Just check the expression, ignore the result type
      let #(checker_after_expr, _result) = check_expression(checker, expression)
      checker_after_expr
    }

    ast.ReturnStatement(value) -> {
      // Check the return value
      let #(checker_after_value, value_result) =
        check_expression(checker, value)

      // Check if return type matches function signature
      case checker_after_value.current_function_return_type {
        option.Some(expected_type) -> {
          case value_result {
            Ok(actual_type) -> {
              case actual_type == expected_type {
                True -> checker_after_value
                False -> {
                  let error =
                    TypeError(
                      "Return type mismatch: expected "
                        <> type_to_string(expected_type)
                        <> " but got "
                        <> type_to_string(actual_type),
                      0,
                      0,
                    )
                  add_error(
                    checker_after_value,
                    error.message,
                    error.line,
                    error.column,
                  )
                }
              }
            }
            Error(_) -> checker_after_value
          }
        }
        option.None -> {
          // Not inside a function (shouldn't happen in valid program)
          let error = TypeError("Return statement outside of function", 0, 0)
          add_error(
            checker_after_value,
            error.message,
            error.line,
            error.column,
          )
        }
      }
    }

    ast.IfStatement(condition, then_branch, else_branch) -> {
      // Check condition is boolean
      let #(checker_after_condition, condition_result) =
        check_expression(checker, condition)

      let checker_with_condition_check = case condition_result {
        Ok(condition_type) -> {
          case condition_type {
            ast.Bool -> checker_after_condition
            _ -> {
              let error =
                TypeError(
                  "If condition must be boolean, got "
                    <> type_to_string(condition_type),
                  0,
                  0,
                )
              add_error(
                checker_after_condition,
                error.message,
                error.line,
                error.column,
              )
            }
          }
        }
        Error(_) -> checker_after_condition
      }

      // Create new scope for then branch
      let then_table =
        symbol_table.new_with_parent(checker_with_condition_check.symbol_table)
      let checker_for_then =
        TypeChecker(..checker_with_condition_check, symbol_table: then_table)
      let checker_after_then = check_statements(checker_for_then, then_branch)

      // Create new scope for else branch  
      let else_table =
        symbol_table.new_with_parent(checker_with_condition_check.symbol_table)
      let checker_for_else =
        TypeChecker(..checker_with_condition_check, symbol_table: else_table)
      let checker_after_else = check_statements(checker_for_else, else_branch)

      // Merge errors from both branches
      TypeChecker(
        symbol_table: checker_with_condition_check.symbol_table,
        // Original scope
        errors: list.append(
          checker_after_else.errors,
          checker_after_then.errors,
        ),
        current_function_return_type: checker.current_function_return_type,
      )
    }

    ast.ForStatement(variable, start, end, body) -> {
      // Check start and end are integers
      let #(checker_after_start, start_result) =
        check_expression(checker, start)
      let #(checker_after_end, end_result) =
        check_expression(checker_after_start, end)

      let checker_with_range_check = case start_result, end_result {
        Ok(ast.Int), Ok(ast.Int) -> checker_after_end
        Ok(start_type), _ -> {
          let error =
            TypeError(
              "For loop start must be int, got " <> type_to_string(start_type),
              0,
              0,
            )
          add_error(checker_after_end, error.message, error.line, error.column)
        }
        _, Ok(end_type) -> {
          let error =
            TypeError(
              "For loop end must be int, got " <> type_to_string(end_type),
              0,
              0,
            )
          add_error(checker_after_end, error.message, error.line, error.column)
        }
        _, _ -> checker_after_end
      }

      // Create new scope with loop variable
      let loop_table =
        symbol_table.new_with_parent(checker_with_range_check.symbol_table)
      let loop_table_with_var =
        symbol_table.add_variable(loop_table, variable, ast.Int)
      let checker_for_body =
        TypeChecker(
          ..checker_with_range_check,
          symbol_table: loop_table_with_var,
        )

      // Check loop body
      let checker_after_body = check_statements(checker_for_body, body)

      // Return to original scope
      TypeChecker(
        symbol_table: checker_with_range_check.symbol_table,
        errors: checker_after_body.errors,
        current_function_return_type: checker.current_function_return_type,
      )
    }
  }
}

// Check a list of statements
pub fn check_statements(
  checker: TypeChecker,
  statements: List(ast.Statement),
) -> TypeChecker {
  case statements {
    [] -> checker
    [first, ..rest] -> {
      let checker_after_first = check_statement(checker, first)
      check_statements(checker_after_first, rest)
    }
  }
}

// Check a function definition
pub fn check_function(
  checker: TypeChecker,
  function: ast.Function,
) -> TypeChecker {
  // Create function signature
  let parameter_types =
    list.map(function.parameters, fn(param) { param.parameter_type })
  let signature =
    symbol_table.FunctionSignature(parameter_types, function.return_type)

  // Add function to global symbol table
  let global_table_with_function =
    symbol_table.add_function(checker.symbol_table, function.name, signature)

  // Create new scope for function body
  let function_table = symbol_table.new_with_parent(global_table_with_function)

  // Add parameters to function scope
  let function_table_with_params =
    list.fold(function.parameters, function_table, fn(table, param) {
      symbol_table.add_variable(table, param.name, param.parameter_type)
    })

  // Check function body with return type context
  let checker_for_body =
    TypeChecker(
      symbol_table: function_table_with_params,
      errors: checker.errors,
      current_function_return_type: option.Some(function.return_type),
    )

  let checker_after_body = check_statements(checker_for_body, function.body)

  // Return to global scope but keep errors
  TypeChecker(
    symbol_table: global_table_with_function,
    errors: checker_after_body.errors,
    current_function_return_type: option.None,
  )
}

// Check entire program
pub fn check_program(program: ast.Program) -> Result(Nil, List(TypeError)) {
  let initial_checker = new()

  // Check all functions
  let final_checker =
    list.fold(program.functions, initial_checker, fn(checker, function) {
      check_function(checker, function)
    })

  // Return errors if any
  case final_checker.errors {
    [] -> Ok(Nil)
    errors -> Error(list.reverse(errors))
    // Reverse to show errors in order
  }
}
