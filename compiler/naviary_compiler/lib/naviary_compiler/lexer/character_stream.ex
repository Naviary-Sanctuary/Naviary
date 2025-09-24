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

  def get_position(%__MODULE__{line: line, column: column}) do
    {line, column}
  end

  @doc """
  Advances the stream by one character and return {char, new_stream}.
  Returns {nil, stream} if at end.
  Updates line and column appropriately for newlines.

  ## Examples

    iex> stream = CharacterStream.new("abc")
    iex> {char, stream2} = CharacterStream.advance(stream)
    iex> char
    "a"
    iex> stream2.column
    2
  """
  def advance(
        %__MODULE__{source: source, position: position, line: line, column: column} = stream
      ) do
    if at_end?(stream) do
      {nil, stream}
    else
      char = String.at(source, position)
      new_position = position + 1

      {new_line, new_column} =
        case char do
          "\n" ->
            {line + 1, 1}

          _ ->
            {line, column + 1}
        end

      new_stream = %__MODULE__{
        source: source,
        position: new_position,
        line: new_line,
        column: new_column
      }

      {char, new_stream}
    end
  end

  @doc """
  Advances the stream while the given predicate returns true.
  Returns the consumed string and the new stream.

  ## Examples

    iex> stream = CharacterStream.new("hello world")
    iex> {word, new_stream} = CharacterStream.advance_while(stream, &(&1 != " "))
    iex> word
    "hello"
  """
  def advance_while(%__MODULE__{} = stream, predicate) when is_function(predicate, 1) do
    do_advance_while(stream, predicate, "")
  end

  defp do_advance_while(stream, predicate, acc) do
    case peek(stream) do
      nil ->
        # End of stream
        {acc, stream}

      char ->
        if predicate.(char) do
          # Character matches predicate, continue collecting
          {_char, new_stream} = advance(stream)
          do_advance_while(new_stream, predicate, acc <> char)
        else
          # Character doesn't match, stop collecting
          {acc, stream}
        end
    end
  end

  @doc """
  Skips whitespace characters and return the new stream.
  Whitespace includes space, tab, newline, and carriage return.

  ## Examples

    iex> stream = CharacterStream.new("  \\n  hello")
    iex> new_stream = CharacterStream.skip_whitespace(stream)
    iex> CharacterStream.peek(new_stream)
    "h"
  """
  def skip_whitespace_and_comments(%__MODULE__{} = stream) do
    case peek(stream) do
      char when char in [" ", "\t", "\n", "\r"] ->
        {_char, new_stream} = advance(stream)
        skip_whitespace_and_comments(new_stream)

      "/" ->
        case peek_ahead(stream, 1) do
          "/" ->
            stream_after_comment = skip_line_comment(stream)
            skip_whitespace_and_comments(stream_after_comment)

          _ ->
            stream
        end

      _ ->
        stream
    end
  end

  @doc """
  Skip single line comment starting with //
  Assumes the stream is positioned at the first /
  Returns the stream after the comment
  """
  def skip_line_comment(%__MODULE__{} = stream) do
    case {peek(stream), peek_ahead(stream, 1)} do
      {"/", "/"} ->
        # Consume //
        {_, stream1} = advance(stream)
        {_, stream2} = advance(stream1)
        skip_until_newline(stream2)

      _ ->
        stream
    end
  end

  defp skip_until_newline(%__MODULE__{} = stream) do
    case peek(stream) do
      nil ->
        stream

      "\n" ->
        stream

      _ ->
        {_, next_stream} = advance(stream)
        skip_until_newline(next_stream)
    end
  end
end
