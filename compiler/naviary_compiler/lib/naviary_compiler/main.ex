defmodule NaviaryCompiler.Main do
  @moduledoc """
  Main entry point for the Naviary compiler.
  Handles command-line arguments and orchestrates the compilation pieline
  """

  def main(args) do
    IO.puts("Naviary Compiler")

    case args do
      [] ->
        show_help()

      [filename] ->
        compile_file(filename)

      _ ->
        IO.puts("Error: Too many arguments")
        show_help()
        System.halt(1)
    end
  end

  defp show_help do
    IO.puts("""
    Usage: naviary_compiler <source_file>

    Arguments:
      <source_file> - Path to .navi source file to compile

    Examples:
      naviary_compiler hello.navi
    """)
  end

  defp compile_file(filename) do
    IO.puts("Compiling file: #{filename}")

    case File.read(filename) do
      {:ok, source_code} ->
        IO.puts("Source code length: #{String.length(source_code)} characters")
        IO.puts("Starting compilation pieline...")

      # TODO: Lexer, Parser, Type Checker, Code Generator

      {:error, reason} ->
        IO.puts("Error: #{reason}")
        System.halt(1)
    end
  end
end
