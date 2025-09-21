defmodule NaviaryCompiler.Lexer.CharacterUtils do
  @moduledoc """
  Utility functions for character classfication.
  """

  @spec is_digit(String.t() | nil) :: boolean()
  def is_digit(char) when is_binary(char) do
    char >= "0" and char <= "9"
  end

  @spec is_hex_digit(String.t() | nil) :: boolean()
  def is_hex_digit(char) when is_binary(char) do
    is_digit(char) or (char >= "a" and char <= "f") or (char >= "A" and char <= "F")
  end

  @spec is_binary_digit(String.t() | nil) :: boolean()
  def is_binary_digit(char) when is_binary(char) do
    char == "0" or char == "1"
  end

  @spec is_octal_digit(String.t() | nil) :: boolean()
  def is_octal_digit(char) when is_binary(char) do
    char >= "0" and char <= "7"
  end

  @doc """
  Checks if a character can start an identifier.
  Identifiers start with letters (a-z, A-Z) or underscore (_).
  """
  @spec is_identifier_start(String.t() | nil) :: boolean()
  def is_identifier_start(char) when is_binary(char) do
    (char >= "a" and char <= "z") or
      (char >= "A" and char <= "Z") or
      char == "_"
  end

  def is_identifier_start(_), do: false

  @doc """
  Checks if a character can continue an identifier.
  After the first character, identifiers can also contain digits.
  """
  @spec is_identifier_continue(String.t() | nil) :: boolean()
  def is_identifier_continue(char) when is_binary(char) do
    is_identifier_start(char) or is_digit(char)
  end

  def is_identifier_continue(_), do: false
end
