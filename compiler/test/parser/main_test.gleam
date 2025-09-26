import gleam/list
import gleam/string
import gleeunit
import gleeunit/should
import parser/parser

pub fn main() {
  gleeunit.main()
}

// Test: Program with main function should parse successfully
pub fn parse_with_main_function_test() {
  let source =
    "
    func main() {
      let a = 1 + 2
    }
  "

  let result = parser.parse(source)

  // Check that parsing succeeded
  case result {
    Ok(program) -> {
      // Verify we have exactly one function
      list.length(program.functions) |> should.equal(1)

      // Verify the function is named "main"
      case list.first(program.functions) {
        Ok(function) -> {
          function.name |> should.equal("main")
        }
        Error(_) -> panic as "Expected at least one function"
      }
    }
    Error(error) -> {
      let message = "Expected successful parse but got error: " <> error.message
      panic as message
    }
  }
}

// Test: Program without main function should fail validation
pub fn parse_without_main_function_test() {
  let source =
    "
    func add(x: int, y: int) -> int {
      return x + y
    }
  "

  let result = parser.parse(source)

  // Check that parsing failed with correct error
  case result {
    Ok(_) -> {
      panic as "Expected error for missing main function"
    }
    Error(error) -> {
      // Check error message contains "main"
      string.contains(error.message, "main") |> should.equal(True)
    }
  }
}
