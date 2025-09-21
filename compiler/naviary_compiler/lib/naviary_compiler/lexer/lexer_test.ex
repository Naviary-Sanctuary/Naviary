defmodule NaviaryCompiler.Lexer.LexerTest do
  use ExUnit.Case

  alias NaviaryCompiler.Lexer.Lexer

  test "tokenize simple integer" do
    {:ok, tokens} = Lexer.tokenize("123")
    assert length(tokens) == 2

    [number_token, eof_token] = tokens

    assert number_token.type == :integer_literal
    assert number_token.value == "123"
    assert eof_token.type == :eof
  end
end
