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

  @spec scan_identifier(CharacterStream.t()) :: scan_result
  defp scan_identifier(stream) do
    start_line = stream.line
    start_column = stream.column

    {identifier_text, new_stream} =
      CharacterStream.advance_while(
        stream,
        &CharacterUtils.is_identifier_continue/1
      )

    if identifier_text == "" do
      {nil, stream}
    else
      token =
        case TokenType.keyword_type(identifier_text) do
          nil ->
            Token.identifier(identifier_text, start_line, start_column)

          keyword_atom ->
            Token.keyword(keyword_atom, start_line, start_column)
        end

      {token, new_stream}
    end
  end

  @spec scan_operator_or_delimiter(CharacterStream.t()) :: scan_result
  defp scan_operator_or_delimiter(stream) do
    start_line = stream.line
    start_column = stream.column

    current = CharacterStream.peek(stream)
    next = CharacterStream.peek_ahead(stream, 1)

    case check_two_char_operator(current, next) do
      {token_type, 2} ->
        {_, stream1} = CharacterStream.advance(stream)
        {_, stream2} = CharacterStream.advance(stream1)

        token = Token.new(token_type, current <> next, start_line, start_column)
        {token, stream2}

      nil ->
        scan_single_char_operator_or_delimiter(stream, current, start_line, start_column)
    end
  end

  @spec scan_single_char_operator_or_delimiter(
          CharacterStream.t(),
          String.t() | nil,
          non_neg_integer(),
          non_neg_integer()
        ) :: scan_result
  defp scan_single_char_operator_or_delimiter(stream, current, start_line, start_column) do
    # Get token type from TokenType module
    token_type = TokenType.operator_type(current) || TokenType.delimiter_type(current)

    case token_type do
      nil ->
        # Not an operator or delimiter
        {nil, stream}

      type ->
        # Consume one character and create token
        {_, new_stream} = CharacterStream.advance(stream)
        token = Token.new(type, current, start_line, start_column)
        {token, new_stream}
    end
  end

  @spec check_two_char_operator(String.t() | nil, String.t() | nil) :: {atom(), 2} | nil
  defp check_two_char_operator(first, second) when is_binary(first) and is_binary(second) do
    two_char = first <> second

    case two_char do
      "==" -> {:equal, 2}
      "!=" -> {:not_equal, 2}
      "->" -> {:arrow, 2}
      _ -> nil
    end
  end

  defp check_two_char_operator(_, _), do: nil
end
