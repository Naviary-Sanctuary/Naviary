defmodule NaviaryCompiler.Lexer.CharacterStream do
  @moduledoc """
  Manages character-by-character reading of source code.
  Tracks current position (line, column) for error reporting.
  """

  @enforce_keys [:source, :position, :line, :column]
  defstruct [:source, :position, :line, :column]

  @type t :: %__MODULE__{
          source: String.t(),
          position: non_neg_integer(),
          line: non_neg_integer(),
          column: non_neg_integer()
        }

  @doc """
  Create a new CharacterStream from source code string.
  Starts at position 0, line 1, column 1.
  """
  def new(source) when is_binary(source) do
    %__MODULE__{
      source: source,
      position: 0,
      line: 1,
      column: 1
    }
  end

  @doc """
  Returns the current character without advancing.
  Returns nil if at end of stream.

  ## Examples

    iex> stream = CharacterStream.new("abc")
    iex> CharacterStream.peek(stream)
    "a"
  """
  def peek(%__MODULE__{source: source, position: position}) do
    if position < byte_size(source) do
      String.at(source, position)
    else
      nil
    end
  end

  @doc """
  Returns the character at position + offset without advancing.

    iex> stream = CharacterStream.new("abc")
    iex> CharacterStream.peek_ahead(stream, 2)
    "c"
  """
  def peek_ahead(%__MODULE__{source: source, position: position}, offset)
      when is_integer(offset) and offset >= 0 do
    target_position = position + offset

    if target_position < byte_size(source) do
      String.at(source, target_position)
    else
      nil
    end
  end

  @doc """
  Checks if the stream is at the end.
  """
  def at_end?(%__MODULE__{position: position, source: source}) do
    position >= byte_size(source)
  end

  @doc """
  Returns the current position info for error reporting.
  """
  def get_position(%__MODULE__{position: position}) do
    {line, column}
  end
end
