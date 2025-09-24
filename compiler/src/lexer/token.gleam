pub type Token {
  Let
  Func
  If
  Return

  // Type Keywords
  Int
  Float
  String
  Bool
  Nil

  // Literals
  IntLiteral(value: Int)
  FloatLiteral(value: Float)
  StringLiteral(value: String)
  BoolLiteral(value: Bool)
  NilLiteral

  // Identifiers
  Identifier(name: String)

  // operators
  Plus
  Minus
  Asterisk
  Slash
  Assign
  Equals
  NotEquals
  LessThan
  GreaterThan
  LessThanOrEqual
  GreaterThanOrEqual

  // Delimiters
  LeftParen
  RightParen
  LeftBrace
  RightBrace
  LeftBracket
  RightBracket
  Comma
  Semicolon
  Colon
  Arrow

  //Special
  Newline
  EOF
}

pub type Position {
  Position(line: Int, column: Int)
}

pub type TokenWithPosition {
  TokenWithPosition(token: Token, position: Position)
}
