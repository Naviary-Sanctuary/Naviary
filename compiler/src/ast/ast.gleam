pub type Expression {
  // Literals
  IntegerLiteral(value: Int)
  FloatLiteral(value: Float)
  BoolLiteral(value: Bool)
  StringLiteral(value: String)
  NilLiteral

  // Variable reference
  Identifier(name: String)

  // Binary operations
  BinaryExpression(
    left: Expression,
    operator: BinaryOperator,
    right: Expression,
  )

  // Function calls
  FunctionExpression(name: String, arguments: List(Expression))
}

pub type BinaryOperator {
  Add
  Subtract
  Multiply
  Divide
  Equal
  NotEqual
  LessThan
  GreaterThan
}

pub type Statement {
  LetStatement(name: String, is_mutable: Bool, value: Expression)

  ExpressionStatement(expression: Expression)

  ReturnStatement(value: Expression)

  IfStatement(
    condition: Expression,
    then_branch: List(Statement),
    else_branch: List(Statement),
  )

  ForStatement(
    variable: String,
    start: Expression,
    end: Expression,
    body: List(Statement),
  )
}

pub type Function {
  Function(
    name: String,
    parameters: List(Parameter),
    return_type: Type,
    body: List(Statement),
  )
}

pub type Parameter {
  Parameter(name: String, parameter_type: Type)
}

pub type Type {
  Int
  Float
  String
  Bool
  Nil
}

pub type Program {
  Program(functions: List(Function))
}
