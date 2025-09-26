import gleeunit
import gleeunit/should

import lexer/lexer
import lexer/token

pub fn main() {
  gleeunit.main()
}

pub fn numbers_test() {
  let input = "123 45.67 0 9.0"
  let lx = lexer.new(input)

  let #(lx, t1) = lexer.next_token(lx)
  t1.token |> should.equal(token.IntLiteral(123))

  let #(lx, t2) = lexer.next_token(lx)
  t2.token |> should.equal(token.FloatLiteral(45.67))

  let #(lx, t4) = lexer.next_token(lx)
  t4.token |> should.equal(token.IntLiteral(0))

  let #(lx, t5) = lexer.next_token(lx)
  t5.token |> should.equal(token.FloatLiteral(9.0))

  let #(_lx, t6) = lexer.next_token(lx)
  t6.token |> should.equal(token.EOF)
}

pub fn identifiers_and_keywords_test() {
  let input = "let x func if return bool true false _id name1"
  let lx = lexer.new(input)

  let #(lx, t1) = lexer.next_token(lx)
  t1.token |> should.equal(token.Let)

  let #(lx, t2) = lexer.next_token(lx)
  t2.token |> should.equal(token.Identifier("x"))

  let #(lx, t3) = lexer.next_token(lx)
  t3.token |> should.equal(token.Func)

  let #(lx, t4) = lexer.next_token(lx)
  t4.token |> should.equal(token.If)

  let #(lx, t5) = lexer.next_token(lx)
  t5.token |> should.equal(token.Return)

  let #(lx, t6) = lexer.next_token(lx)
  t6.token |> should.equal(token.Bool)

  let #(lx, t7) = lexer.next_token(lx)
  t7.token |> should.equal(token.BoolLiteral(True))

  let #(lx, t8) = lexer.next_token(lx)
  t8.token |> should.equal(token.BoolLiteral(False))

  let #(lx, t9) = lexer.next_token(lx)
  t9.token |> should.equal(token.Identifier("_id"))

  let #(_lx, t10) = lexer.next_token(lx)
  t10.token |> should.equal(token.Identifier("name1"))
}

pub fn operators_and_delimiters_test() {
  let input = "+ - * / = == != < <= > >= ( ) { } [ ] , ; : ->"
  let lx = lexer.new(input)

  let #(lx, p1) = lexer.next_token(lx)
  p1.token |> should.equal(token.Plus)

  let #(lx, p2) = lexer.next_token(lx)
  p2.token |> should.equal(token.Minus)

  let #(lx, p3) = lexer.next_token(lx)
  p3.token |> should.equal(token.Asterisk)

  let #(lx, p4) = lexer.next_token(lx)
  p4.token |> should.equal(token.Slash)

  let #(lx, p5) = lexer.next_token(lx)
  p5.token |> should.equal(token.Assign)

  let #(lx, p6) = lexer.next_token(lx)
  p6.token |> should.equal(token.Equals)

  let #(lx, p7) = lexer.next_token(lx)
  p7.token |> should.equal(token.NotEquals)

  let #(lx, p8) = lexer.next_token(lx)
  p8.token |> should.equal(token.LessThan)

  let #(lx, p9) = lexer.next_token(lx)
  p9.token |> should.equal(token.LessThanOrEqual)

  let #(lx, p10) = lexer.next_token(lx)
  p10.token |> should.equal(token.GreaterThan)

  let #(lx, p11) = lexer.next_token(lx)
  p11.token |> should.equal(token.GreaterThanOrEqual)

  let #(lx, p12) = lexer.next_token(lx)
  p12.token |> should.equal(token.LeftParen)

  let #(lx, p13) = lexer.next_token(lx)
  p13.token |> should.equal(token.RightParen)

  let #(lx, p14) = lexer.next_token(lx)
  p14.token |> should.equal(token.LeftBrace)

  let #(lx, p15) = lexer.next_token(lx)
  p15.token |> should.equal(token.RightBrace)

  let #(lx, p16) = lexer.next_token(lx)
  p16.token |> should.equal(token.LeftBracket)

  let #(lx, p17) = lexer.next_token(lx)
  p17.token |> should.equal(token.RightBracket)

  let #(lx, p18) = lexer.next_token(lx)
  p18.token |> should.equal(token.Comma)

  let #(lx, p19) = lexer.next_token(lx)
  p19.token |> should.equal(token.Semicolon)

  let #(lx, p20) = lexer.next_token(lx)
  p20.token |> should.equal(token.Colon)

  let #(_lx, p21) = lexer.next_token(lx)
  p21.token |> should.equal(token.Arrow)
}

pub fn newline_and_position_test() {
  let input = "let\n123\n+"
  let lx = lexer.new(input)

  let #(lx, a1) = lexer.next_token(lx)
  a1.token |> should.equal(token.Let)
  a1.position |> should.equal(token.Position(1, 0))

  let #(lx, a2) = lexer.next_token(lx)
  a2.token |> should.equal(token.Newline)
  a2.position |> should.equal(token.Position(1, 3))

  let #(lx, a3) = lexer.next_token(lx)
  a3.token |> should.equal(token.IntLiteral(123))
  a3.position |> should.equal(token.Position(2, 0))

  let #(lx, a4) = lexer.next_token(lx)
  a4.token |> should.equal(token.Newline)
  a4.position |> should.equal(token.Position(2, 3))

  let #(_lx, a5) = lexer.next_token(lx)
  a5.token |> should.equal(token.Plus)
  a5.position |> should.equal(token.Position(3, 0))
}
