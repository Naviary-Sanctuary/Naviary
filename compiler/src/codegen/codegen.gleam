import ast/ast
import gleam/float
import gleam/int
import gleam/list
import gleam/string

pub fn generate(program: ast.Program) -> String {
  let _main_function = find_main_function(program.functions)
  let header = generate_module_header()
  let functions = generate_functions(program.functions)

  header <> "\n" <> functions
}

fn find_main_function(
  functions: List(ast.Function),
) -> Result(ast.Function, Nil) {
  list.find(functions, fn(f) { f.name == "main" })
}

fn generate_module_header() -> String {
  "-module(main).\n" <> "-export([start/0]).\n"
}

fn generate_functions(functions: List(ast.Function)) -> String {
  functions
  |> list.map(generate_function)
  |> string.join("\n")
}

// Generate a single function
fn generate_function(function: ast.Function) -> String {
  case function.name {
    // Special handling for main function - rename to start
    "main" -> {
      "start() ->\n" <> generate_function_body(function.body) <> "."
    }
    // Regular functions
    _ -> {
      let params = generate_parameters(function.parameters)
      function.name
      <> "("
      <> params
      <> ") ->\n"
      <> generate_function_body(function.body)
      <> "."
    }
  }
}

// Generate parameter list
fn generate_parameters(parameters: List(ast.Parameter)) -> String {
  parameters
  |> list.map(fn(param) { erlang_variable_name(param.name) })
  |> string.join(", ")
}

// Convert variable name to Erlang format (capitalize first letter)
fn erlang_variable_name(name: String) -> String {
  case string.length(name) {
    0 -> "_"
    _ -> {
      let first = string.slice(name, 0, 1) |> string.uppercase
      let rest = string.slice(name, 1, string.length(name) - 1)
      first <> rest
    }
  }
}

// Generate function body from statements
fn generate_function_body(statements: List(ast.Statement)) -> String {
  case statements {
    [] -> "    ok\n"
    [single] -> {
      // Single statement - no comma needed
      "    " <> generate_statement(single) <> "\n"
    }
    _ -> {
      // Multiple statements - add commas between them
      let #(all_but_last, last) = split_last(statements)

      let body_with_commas =
        all_but_last
        |> list.map(fn(stmt) { "    " <> generate_statement(stmt) })
        |> string.join(",\n")

      let last_statement = "    " <> generate_statement(last)

      body_with_commas <> ",\n" <> last_statement <> "\n"
    }
  }
}

// Helper function to split list into all-but-last and last element
fn split_last(items: List(a)) -> #(List(a), a) {
  case list.reverse(items) {
    [last, ..rest] -> #(list.reverse(rest), last)
    [] -> panic("split_last called on empty list")
  }
}

// Generate a single statement
fn generate_statement(statement: ast.Statement) -> String {
  case statement {
    ast.LetStatement(name, _is_mutable, value) -> {
      let var_name = erlang_variable_name(name)
      let expr = generate_expression(value)
      var_name <> " = " <> expr
    }

    ast.ExpressionStatement(expression) -> {
      generate_expression(expression)
    }

    ast.ReturnStatement(value) -> {
      // In Erlang, last expression is automatically returned
      generate_expression(value)
    }

    ast.IfStatement(_condition, _then_branch, _else_branch) -> {
      // Not implemented in MVP
      "ok"
    }

    ast.ForStatement(_variable, _start, _end, _body) -> {
      // Not implemented in MVP
      "ok"
    }
  }
}

// Generate expression
fn generate_expression(expression: ast.Expression) -> String {
  case expression {
    // Literals
    ast.IntegerLiteral(value) -> {
      int.to_string(value)
    }

    ast.FloatLiteral(value) -> {
      float.to_string(value)
    }

    ast.BoolLiteral(True) -> "true"
    ast.BoolLiteral(False) -> "false"

    ast.StringLiteral(value) -> {
      "\"" <> escape_string(value) <> "\""
    }

    ast.NilLiteral -> "nil"

    // Variable reference
    ast.Identifier(name) -> {
      erlang_variable_name(name)
    }

    // Binary operations
    ast.BinaryExpression(left, operator, right) -> {
      let left_code = generate_expression(left)
      let op_code = generate_operator(operator)
      let right_code = generate_expression(right)

      // Parentheses for safety
      "(" <> left_code <> " " <> op_code <> " " <> right_code <> ")"
    }

    // Function calls
    ast.FunctionExpression(name, arguments) -> {
      generate_function_call(name, arguments)
    }
  }
}

// Escape special characters in strings
fn escape_string(s: String) -> String {
  s
  |> string.replace("\\", "\\\\")
  |> string.replace("\"", "\\\"")
  |> string.replace("\n", "\\n")
  |> string.replace("\r", "\\r")
  |> string.replace("\t", "\\t")
}

// Generate operator
fn generate_operator(operator: ast.BinaryOperator) -> String {
  case operator {
    ast.Add -> "+"
    ast.Subtract -> "-"
    ast.Multiply -> "*"
    ast.Divide -> "div"
    // Integer division in Erlang
    ast.Equal -> "=="
    ast.NotEqual -> "/="
    // Not equal in Erlang is /=
    ast.LessThan -> "<"
    ast.GreaterThan -> ">"
  }
}

// Generate function call
fn generate_function_call(
  name: String,
  arguments: List(ast.Expression),
) -> String {
  case name {
    // Special case for print - map to io:format
    "print" -> {
      case arguments {
        [arg] -> {
          let arg_code = generate_expression(arg)
          "io:format(\"~p~n\", [" <> arg_code <> "])"
        }
        _ -> {
          // Multiple arguments or no arguments
          let args =
            arguments
            |> list.map(generate_expression)
            |> string.join(", ")
          "io:format(\"~p~n\", [" <> args <> "])"
        }
      }
    }

    // Regular function calls
    _ -> {
      let args =
        arguments
        |> list.map(generate_expression)
        |> string.join(", ")
      name <> "(" <> args <> ")"
    }
  }
}

// Generate complete Erlang module file
pub fn generate_module(program: ast.Program) -> Result(String, String) {
  // Check for main function
  case find_main_function(program.functions) {
    Ok(_main) -> {
      let code = generate(program)
      Ok(code)
    }
    Error(_) -> {
      Error("No main function found")
    }
  }
}

// Helper to compile and save to file
pub fn compile_to_file(
  program: ast.Program,
  _filename: String,
) -> Result(String, String) {
  case generate_module(program) {
    Ok(erlang_code) -> {
      // Return the code and filename for external file writing
      Ok(erlang_code)
    }
    Error(error) -> Error(error)
  }
}
