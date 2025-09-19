defmodule NaviaryCompiler.Token do
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
  Converts a token to a human-readable string for debugging.
  """
  def to_string(%__MODULE__{type: type, value: value, line: line, column: column}) do
    "#{type}: #{value} at line #{line}, column #{column}"
  end
end
