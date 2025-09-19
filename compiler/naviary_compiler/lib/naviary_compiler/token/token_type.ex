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
    "class" => :class
  }

  # Type Keywords - primitive types
  @type_keywords %{
    "int" => :int,
    "float" => :float,
    "string" => :string,
    "bool" => :bool
  }

  @all_keywords Map.merge(@keywords, @type_keywords)

  @doc """
    Return the token type atom for a given keyword string.

    ## Examples

      iex> TokenType.get_type("let")
      :let

      iex> TokenType.get_type("int")
      :int

      iex> TokenType.get_type("variable_name")
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
end
