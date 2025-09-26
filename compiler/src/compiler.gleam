import codegen/codegen
import gleam/io
import gleam/list
import gleam/result
import gleam/string
import parser/parser
import simplifile
import typechecker/typechecker

// Main compiler pipeline
pub fn compile(source_file: String, output_file: String) -> Result(Nil, String) {
  // Read source file
  use source <- result.try(
    simplifile.read(source_file)
    |> result.map_error(fn(_) { "Failed to read file: " <> source_file }),
  )

  // Parse source code
  use ast <- result.try(
    parser.parse(source)
    |> result.map_error(fn(error) { "Parse error: " <> error.message }),
  )

  // Type check
  use _ <- result.try(
    typechecker.check_program(ast)
    |> result.map_error(fn(errors) {
      case errors {
        [first, ..] -> "Type error: " <> first.message
        [] -> "Unknown type error"
      }
    }),
  )

  // Generate Erlang code
  use erlang_code <- result.try(
    codegen.generate_module(ast)
    |> result.map_error(fn(error) { "Code generation error: " <> error }),
  )

  // Write Erlang file
  use _ <- result.try(
    simplifile.write(output_file, erlang_code)
    |> result.map_error(fn(_) { "Failed to write file: " <> output_file }),
  )

  // Compile with erlc
  compile_erlang(output_file)
}

// Compile Erlang source to BEAM
fn compile_erlang(erlang_file: String) -> Result(Nil, String) {
  // Use erlc to compile the Erlang file
  let command = "erlc " <> erlang_file

  case execute_command(command) {
    Ok(_) -> {
      io.println("✅ Compilation successful!")
      io.println("Generated: " <> erlang_file)
      let beam_file = string.replace(erlang_file, ".erl", ".beam")
      io.println("Generated: " <> beam_file)
      Ok(Nil)
    }
    Error(error) -> {
      Error("erlc compilation failed: " <> error)
    }
  }
}

@external(erlang, "os", "cmd")
fn os_cmd_erlang(command: List(Int)) -> List(Int)

// Wrapper to handle string conversion
fn os_cmd(command: String) -> String {
  // Convert string to character list (Erlang style)
  let char_list =
    string.to_utf_codepoints(command)
    |> list.map(string.utf_codepoint_to_int)

  // Call Erlang os:cmd
  let result = os_cmd_erlang(char_list)

  // Convert result back to string
  result
  |> list.map(fn(code) {
    case string.utf_codepoint(code) {
      Ok(s) -> string.from_utf_codepoints([s])
      Error(_) -> ""
    }
  })
  |> string.join("")
}

fn execute_command(command: String) -> Result(String, String) {
  let output = os_cmd(command)
  // Check if output contains error indicators
  case string.contains(output, "error") || string.contains(output, "Error") {
    True -> Error(output)
    False -> Ok(output)
  }
}

// Main entry point
pub fn main() -> Nil {
  // Example usage
  case compile("../examples/main.navi", "output.erl") {
    Ok(_) -> io.println("Compilation completed successfully!")
    Error(error) -> io.println("Compilation failed: " <> error)
  }
}

// CLI interface
pub fn run(args: List(String)) -> Result(Nil, String) {
  case args {
    [source_file] -> {
      let output_file = string.replace(source_file, ".navi", ".erl")
      compile(source_file, output_file)
    }
    [source_file, output_file] -> {
      compile(source_file, output_file)
    }
    _ -> {
      io.println("Usage: naviary <source.navi> [output.erl]")
      Error("Invalid arguments")
    }
  }
}
