defmodule NaviaryCompiler.Token.TokenType do
  @moduledoc """
  Defines all token types used in the Naviary language lexer.
  """

  # Keywords - core language structs
  @keywords %{
    "let" => :let,
    "func" => :func,
    "if" => :if,
    "for" => :for,
    "return" => :return,
    "class" => :class,
    # Added for mutable variables
    "mut" => :mut,
    # Added for if-else
    "else" => :else
  }

  # Type Keywords - primitive types
  @type_keywords %{
    "int" => :int,
    "float" => :float,
    "string" => :string,
    "bool" => :bool
  }

  # Boolean literals
  @bool_literals %{
    "true" => :true_literal,
    "false" => :false_literal
  }

  # Operators - arithmetic and comparison
  @operators %{
    # Single character operators
    # Addition
    "+" => :plus,
    # Subtraction
    "-" => :minus,
    # Multiplication
    "*" => :asterisk,
    # Division
    "/" => :slash,
    # Assignment
    "=" => :assign,
    # Logical NOT
    "!" => :not,
    # Less than
    "<" => :less,
    # Greater than
    ">" => :greater,

    # Two character operators
    # Equality comparison
    "==" => :equal,
    # Not equal
    "!=" => :not_equal,
    # Function return type
    "->" => :arrow
  }

  # Delimiters - for grouping and separation
  @delimiters %{
    "(" => :left_paren,
    ")" => :right_paren,
    "{" => :left_brace,
    "}" => :right_brace,
    "[" => :left_bracket,
    "]" => :right_bracket,
    "," => :comma,
    ":" => :colon,
    ";" => :semicolon
  }

  # Combine all keyword-like tokens
  @all_keywords Map.merge(@keywords, @type_keywords) |> Map.merge(@bool_literals)

  # Define all possible token types as a type spec
  @type token_type ::
          :let
          # Keywords
          | :func
          | :if
          | :for
          | :return
          | :class
          | :mut
          | :else
          # Type keywords
          | :int
          | :float
          | :string
          | :bool
          # Boolean literals
          | :true_literal
          | :false_literal
          # Operators
          | :plus
          | :minus
          | :asterisk
          | :slash
          | :assign
          | :not
          | :less
          | :greater
          | :equal
          | :not_equal
          | :arrow
          # Delimiters
          | :left_paren
          | :right_paren
          | :left_brace
          | :right_brace
          | :left_bracket
          | :right_bracket
          | :comma
          | :colon
          | :semicolon
          # Dynamic tokens (values vary)
          # Variable/function names
          | :identifier
          # 42, 100, 0xFF
          | :integer_literal
          # 3.14, 0.5
          | :float_literal
          # "hello world"
          | :string_literal
          # Special

          # Line breaks (if significant)
          | :newline
          # End of file marker
          | :eof

  # Export the maps for lexer use
  def operators, do: @operators
  def delimiters, do: @delimiters
  def bool_literals, do: @bool_literals

  @doc """
  Return the token type atom for a given keyword string.
  Returns nil if not a keyword.

  ## Examples

    iex> TokenType.keyword_type("let")
    :let

    iex> TokenType.keyword_type("int")
    :int

    iex> TokenType.keyword_type("variable_name")
    nil
  """
  def keyword_type(string) when is_binary(string) do
    Map.get(@all_keywords, string)
  end

  @doc """
  Check if the given string is a keyword.
  """
  def keyword?(string) when is_binary(string) do
    Map.has_key?(@all_keywords, string)
  end

  @doc """
  Check if the given string is an operator.
  """
  def operator_type(string) when is_binary(string) do
    Map.get(@operators, string)
  end

  @doc """
  Check if the given string is a delimiter.
  """
  def delimiter_type(string) when is_binary(string) do
    Map.get(@delimiters, string)
  end

  @doc """
  Returns the list of all token types that carry dynamic values.
  These tokens need their actual value stored, not just the type.
  """
  def literal_types do
    [:identifier, :integer_literal, :float_literal, :string_literal]
  end

  @doc """
  Checks if a token type is a literal (carries a value).
  """
  def is_literal?(type) when is_atom(type) do
    type in literal_types()
  end

  @doc """
  Checks if a token type is a comparison operator.
  """
  def is_comparison?(type) when is_atom(type) do
    type in [:equal, :not_equal, :less, :greater]
  end

  @doc """
  Checks if a token type is an arithmetic operator.
  """
  def is_arithmetic?(type) when is_atom(type) do
    type in [:plus, :minus, :asterisk, :slash]
  end
end
