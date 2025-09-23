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

  describe "single character operators" do
    test "arithmetic operators" do
      {:ok, tokens} = Lexer.tokenize("+ - * /")

      assert length(tokens) == 5
      assert Enum.at(tokens, 0).type == :plus
      assert Enum.at(tokens, 1).type == :minus
      assert Enum.at(tokens, 2).type == :asterisk
      assert Enum.at(tokens, 3).type == :slash
      assert Enum.at(tokens, 4).type == :eof
    end

    test "assignment operator" do
      {:ok, tokens} = Lexer.tokenize("=")

      assert Enum.at(tokens, 0).type == :assign
      assert Enum.at(tokens, 0).value == "="
    end
  end

  describe "two character operators" do
    test "equality operators" do
      {:ok, tokens} = Lexer.tokenize("== !=")

      assert length(tokens) == 3
      assert Enum.at(tokens, 0).type == :equal
      assert Enum.at(tokens, 0).value == "=="
      assert Enum.at(tokens, 1).type == :not_equal
      assert Enum.at(tokens, 1).value == "!="
      assert Enum.at(tokens, 2).type == :eof
    end

    test "arrow operator" do
      {:ok, tokens} = Lexer.tokenize("->")

      assert Enum.at(tokens, 0).type == :arrow
      assert Enum.at(tokens, 0).value == "->"
    end
  end

  describe "delimiters" do
    test "parentheses and braces" do
      {:ok, tokens} = Lexer.tokenize("(){}[]")

      assert length(tokens) == 7
      assert Enum.at(tokens, 0).type == :left_paren
      assert Enum.at(tokens, 1).type == :right_paren
      assert Enum.at(tokens, 2).type == :left_brace
      assert Enum.at(tokens, 3).type == :right_brace
      assert Enum.at(tokens, 4).type == :left_bracket
      assert Enum.at(tokens, 5).type == :right_bracket
      assert Enum.at(tokens, 6).type == :eof
    end

    test "comma and semicolon" do
      {:ok, tokens} = Lexer.tokenize(", ;")

      assert Enum.at(tokens, 0).type == :comma
      assert Enum.at(tokens, 1).type == :semicolon
      assert Enum.at(tokens, 2).type == :eof
    end
  end

  describe "mixed expressions" do
    test "arithmetic expression" do
      {:ok, tokens} = Lexer.tokenize("1 + 2")

      assert length(tokens) == 4
      assert Enum.at(tokens, 0).type == :integer_literal
      assert Enum.at(tokens, 0).value == "1"
      assert Enum.at(tokens, 1).type == :plus
      assert Enum.at(tokens, 2).type == :integer_literal
      assert Enum.at(tokens, 2).value == "2"
      assert Enum.at(tokens, 3).type == :eof
    end
  end

  describe "keywords" do
    test "language keywords" do
      {:ok, tokens} = Lexer.tokenize("let func if for return class mut else")

      assert length(tokens) == 9
      assert Enum.at(tokens, 0).type == :let
      assert Enum.at(tokens, 1).type == :func
      assert Enum.at(tokens, 2).type == :if
      assert Enum.at(tokens, 3).type == :for
      assert Enum.at(tokens, 4).type == :return
      assert Enum.at(tokens, 5).type == :class
      assert Enum.at(tokens, 6).type == :mut
      assert Enum.at(tokens, 7).type == :else
      assert Enum.at(tokens, 8).type == :eof
    end

    test "boolean literals" do
      {:ok, tokens} = Lexer.tokenize("true false")

      assert Enum.at(tokens, 0).type == :true_literal
      assert Enum.at(tokens, 1).type == :false_literal
      assert Enum.at(tokens, 2).type == :eof
    end
  end

  describe "identifiers" do
    test "simple identifiers" do
      {:ok, tokens} = Lexer.tokenize("myVariable userName")

      assert Enum.at(tokens, 0).type == :identifier
      assert Enum.at(tokens, 0).value == "myVariable"
      assert Enum.at(tokens, 1).type == :identifier
      assert Enum.at(tokens, 1).value == "userName"
      assert Enum.at(tokens, 2).type == :eof
    end
  end

  describe "mixed keywords and identifiers" do
    test "let x = 10" do
      {:ok, tokens} = Lexer.tokenize("let x = 10")

      assert Enum.at(tokens, 0).type == :let
      assert Enum.at(tokens, 1).type == :identifier
      assert Enum.at(tokens, 1).value == "x"
      assert Enum.at(tokens, 2).type == :assign
      assert Enum.at(tokens, 3).type == :integer_literal
      assert Enum.at(tokens, 3).value == "10"
      assert Enum.at(tokens, 4).type == :eof
    end
  end

  describe "basic strings" do
    test "simple string" do
      {:ok, tokens} = Lexer.tokenize("\"Hello World\"")

      assert Enum.at(tokens, 0).type == :string_literal
      assert Enum.at(tokens, 0).value == "Hello World"
    end

    test "empty string" do
      {:ok, tokens} = Lexer.tokenize("\"\"")

      assert Enum.at(tokens, 0).type == :string_literal
      assert Enum.at(tokens, 0).value == ""
    end
  end

  describe "escape sequences" do
    test "escaped quote" do
      {:ok, tokens} = Lexer.tokenize("\"Say \\\"Hi\\\"\"")

      assert Enum.at(tokens, 0).type == :string_literal
      # We store escape as-is
      assert Enum.at(tokens, 0).value == "Say \\\"Hi\\\""
    end
  end
end
