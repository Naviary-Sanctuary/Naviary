defmodule NaviaryCompiler.Token do
  alias NaviaryCompiler.Token.TokenType
  @moduledoc """
  Represents a single token form the lexical analysis phase.
  Contains its type, value, and position information for error reporting.
  """

  @enforce_keys [:type, :value, :line, :column]
  defstruct [:type, :value, :line, :column]

  @type t :: %__MODULE__{
          type: atom(),
          value: String.t(),
          line: non_neg_integer(),
          column: non_neg_integer()
        }

  @doc """
  Create a new token with the given type, value, and position.

  ## Examples

    iex> Token.new(:let, "let", 1, 5)
    %Token{type: :let, value: "let", line: 1, column: 5}
  """
  def new(type, value, line, column) do
    %__MODULE__{type: type, value: value, line: line, column: column}
  end

  @doc """
  Create a new keyword token.
  """
  def keyword(type, line, column) when is_atom(type) do
    value = Atom.to_string(type)
    new(type, value, line, column)
  end

  @doc """
  Create an identifier token with the given name
  """
  def identifier(name, line, column) when is_binary(name) do
    new(:identifier, name, line, column)
  end

  @doc """
  Create an integer literal token.
  """
  def integer(value, line, column) when is_binary(value) do
    new(:integer_literal, value, line, column)
  end

  @doc """
  Create a float literal token.
  """
  def float(value, line, column) when is_binary(value) do
    new(:float_literal, value, line, column)
  end

  @doc """
  Create a string literal token (without quotes in the value).
  """
  def string(value, line, column) when is_binary(value) do
    new(:string_literal, value, line, column)
  end

  @doc """
  Create an operator token.
  """
  def operator(op_string, line, column) when is_binary(op_string) do
    case TokenType.operator_type(op_string) do
      nil -> {:error, "Unknown operator: #{op_string}"}
      type -> {:ok, new(type, op_string, line, column)}
    end
  end

  @doc """
  Create a delimiter token.
  """
  def delimiter(delim_string, line, column) when is_binary(delim_string) do
    case TokenType.delimiter_type(delim_string) do
      nil -> {:error, "Unknown delimiter: #{delim_string}"}
      type -> {:ok, new(type, delim_string, line, column)}
    end
  end

  @doc """
  Create an EOF (end of file) token.
  """
  def eof(line, column) do
    new(:eof, "", line, column)
  end

  @doc """
  Converts a token to a human-readable string for debugging.
  """
  def to_string(%__MODULE__{type: type, value: value, line: line, column: column}) do
    case type do
      :string_literal ->
        "#{type}: \"#{value}\" at line #{line}, column #{column}"

      :eof ->
        "EOF at line #{line}, column #{column}"

      _ ->
        "#{type}: #{value} at line #{line}, column #{column}"
    end
  end

  @doc """
  Check if a token is of a specific type.
  """
  def is_type?(%__MODULE__{type: type}, expected_type) when is_atom(expected_type) do
    type == expected_type
  end

  @doc """
  Check if a token is any of the given types.
  """
  def is_any_type?(%__MODULE__{type: type}, types) when is_list(types) do
    type in types
  end
end
