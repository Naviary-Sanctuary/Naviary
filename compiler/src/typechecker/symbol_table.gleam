import ast/ast
import gleam/dict.{type Dict}
import gleam/option.{type Option, None, Some}

pub type FunctionSignature {
  FunctionSignature(parameter_types: List(ast.Type), return_type: ast.Type)
}

pub type SymbolTable {
  SymbolTable(
    variables: Dict(String, ast.Type),
    functions: Dict(String, FunctionSignature),
    parent: Option(SymbolTable),
  )
}

pub fn new() -> SymbolTable {
  SymbolTable(variables: dict.new(), functions: dict.new(), parent: None)
}

pub fn new_with_parent(parent: SymbolTable) -> SymbolTable {
  SymbolTable(
    variables: dict.new(),
    functions: dict.new(),
    parent: Some(parent),
  )
}

pub fn add_variable(
  table: SymbolTable,
  name: String,
  variable_type: ast.Type,
) -> SymbolTable {
  SymbolTable(
    ..table,
    variables: dict.insert(table.variables, name, variable_type),
  )
}

pub fn lookup_variable(table: SymbolTable, name: String) -> Option(ast.Type) {
  case dict.get(table.variables, name) {
    Ok(variable_type) -> Some(variable_type)
    Error(_) -> {
      case table.parent {
        Some(parent) -> lookup_variable(parent, name)
        None -> None
      }
    }
  }
}

pub fn add_function(
  table: SymbolTable,
  name: String,
  signature: FunctionSignature,
) -> SymbolTable {
  SymbolTable(..table, functions: dict.insert(table.functions, name, signature))
}

pub fn lookup_function(
  table: SymbolTable,
  name: String,
) -> Option(FunctionSignature) {
  case dict.get(table.functions, name) {
    Ok(function_signature) -> Some(function_signature)
    Error(_) -> {
      case table.parent {
        Some(parent) -> lookup_function(parent, name)
        None -> None
      }
    }
  }
}

pub fn has_variable_in_current_scope(table: SymbolTable, name: String) -> Bool {
  case dict.get(table.variables, name) {
    Ok(_) -> True
    Error(_) -> False
  }
}
