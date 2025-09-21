defmodule NaviaryCompiler.Lexer.Lexer do
  @moduledoc """
  Lexical analyzer (tokenizer) for the Naviary.

  Converts source code strings into streams of tokens for parsing.
  Uses a character-by-character scanning approach with lookahead.
  """

  alias NaviaryCompiler.Lexer.CharacterStream
  alias NaviaryCompiler.Token
  alias NaviaryCompiler.Token.TokenType
  alias NaviaryCompiler.Lexer.CharacterUtils

  @type lexer_result :: {:ok, [Token.t()]} | {:error, String.t()}
  @type scan_result :: {Token.t() | nil, CharacterStream.t()}

  @doc """
  Main entry point for tokenization

  Takes source code as a string and return either:
  - {:ok, [Token.t()]} if successful
  - {:error, String.t()} if there was an error
  """
  @spec tokenize(String.t()) :: lexer_result()
  def tokenize(source_code) when is_binary(source_code) do
    try do
      stream = CharacterStream.new(source_code)

      {tokens, _final_stream} = tokenize_stream(stream, [])
      {:ok, tokens}
    rescue
      error -> {:error, "Lexer Error: #{Exception.message(error)}"}
    end
  end

  @spec tokenize_stream(CharacterStream.t(), [Token.t()]) :: {[Token.t()], CharacterStream.t()}
  defp tokenize_stream(stream, tokens_accumulator) do
    stream = CharacterStream.skip_whitespace(stream)

    case CharacterStream.at_end?(stream) do
      true ->
        eof_token = Token.eof(stream.line, stream.column)
        final_tokens = [eof_token | tokens_accumulator] |> Enum.reverse()
        {final_tokens, stream}

      false ->
        {token, new_stream} = scan_next_token(stream)

        case token do
          nil ->
            {char, skip_stream} = CharacterStream.advance(stream)
            error_token = Token.new(:error, char, stream.line, stream.column)
            tokenize_stream(skip_stream, [error_token | tokens_accumulator])

          _ ->
            tokenize_stream(new_stream, [token | tokens_accumulator])
        end
    end
  end

  @spec scan_next_token(CharacterStream.t()) :: scan_result
  defp scan_next_token(stream) do
    current_char = CharacterStream.peek(stream)

    cond do
      # Numbers (0-9)
      current_char != nil and current_char >= "0" and current_char <= "9" ->
        scan_number(stream)

      # String literals (")
      current_char == "\"" ->
        scan_string(stream)

      # Identifiers and keywords (a-z, A-Z, _)
      is_identifier_start(current_char) ->
        scan_identifier(stream)

      # Operators and delimiters (+, -, =, (, ), etc.)
      is_operator_or_delimiter_start(current_char) ->
        scan_operator_or_delimiter(stream)

      # Unrecognized character
      true ->
        {nil, stream}
    end
  end

  # Helper functions for character classification

  defp is_identifier_start(char) when is_binary(char) do
    (char >= "a" and char <= "z") or
      (char >= "A" and char <= "Z") or
      char == "_"
  end

  defp is_identifier_start(_), do: false

  defp is_operator_or_delimiter_start(char) when is_binary(char) do
    # Check if character is in our operators or delimiters
    TokenType.operator_type(char) != nil or TokenType.delimiter_type(char) != nil
  end

  defp is_operator_or_delimiter_start(_), do: false

  @spec scan_number(CharacterStream.t()) :: scan_result
  defp scan_number(stream) do
    start_line = stream.line
    start_column = stream.column

    scan_decimal_integer(stream, start_line, start_column)

    # TODO: Implement hex, octal, and binary integer scanning
    # case CharacterStream.peek(stream) do
    #   "0" ->
    #     {_zero, stream_after_zero} = CharacterStream.advance(stream)
    # end
  end

  defp scan_decimal_integer(stream, start_line, start_column) do
    {digits, new_stream} = CharacterStream.advance_while(stream, &CharacterUtils.is_digit/1)

    if digits == "" do
      {nil, stream}
    else
      token = Token.integer(digits, start_line, start_column)
      {token, new_stream}
    end
  end

  defp scan_string(_stream) do
    raise "scan_string not implemented yet"
  end

  defp scan_identifier(_stream) do
    raise "scan_identifier not implemented yet"
  end

  defp scan_operator_or_delimiter(_stream) do
    raise "scan_operator_or_delimiter not implemented yet"
  end
end
