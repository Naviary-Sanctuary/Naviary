use crate::lexer::token::{Token, TokenKind};
use crate::source::span::Span;

pub struct Lexer<'source> {
    source: &'source str,
    position: usize,
}

impl<'source> Lexer<'source> {
    pub fn source(&self) -> &'source str {
        self.source
    }
    pub fn position(&self) -> usize {
        self.position
    }
    pub fn new(source: &'source str) -> Self {
        Self {
            source,
            position: 0,
        }
    }

    pub fn is_at_end(&self) -> bool {
        self.position() >= self.source().len()
    }

    pub fn peek(&self) -> Option<char> {
        self.source()[self.position()..].chars().next()
    }

    pub fn advance(&mut self) -> Option<char> {
        let character = self.peek()?;
        self.position += character.len_utf8();

        Some(character)
    }

    pub fn skip_whitespace(&mut self) {
        while matches!(self.peek(), Some(' ' | '\t' | '\n' | '\r')) {
            self.advance();
        }
    }

    pub fn next_token(&mut self) -> Token {
        self.skip_whitespace();

        let start = self.position();

        let Some(character) = self.advance() else {
            return Token::new(TokenKind::EndOfFile, Span::new(start, start));
        };

        let kind = self.char_to_token_kind(character);

        Token::new(kind, Span::new(start, self.position()))
    }

    fn char_to_token_kind(&mut self, character: char) -> TokenKind {
        match character {
            '(' => TokenKind::LeftParenthesis,
            ')' => TokenKind::RightParenthesis,
            '{' => TokenKind::LeftBrace,
            '}' => TokenKind::RightBrace,
            '+' => TokenKind::Plus,
            '-' => TokenKind::Minus,
            '=' => TokenKind::Equal,
            ';' => TokenKind::Semicolon,
            _ => TokenKind::Unknown,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::Lexer;

    mod new_tests {
        use super::Lexer;

        #[test]
        fn create_lexer_with_args() {
            let lexer = Lexer::new("func main");

            assert_eq!(lexer.source(), "func main");
            assert_eq!(lexer.position(), 0);
        }
    }

    mod is_at_end_tests {
        use super::Lexer;
        #[test]
        fn returns_false_at_initial_position_when_source_is_not_empty() {
            let lexer = Lexer::new("func");
            assert!(!lexer.is_at_end());
        }
        #[test]
        fn returns_true_when_source_is_empty() {
            let lexer = Lexer::new("");
            assert!(lexer.is_at_end());
        }
    }

    mod peek_tests {
        use super::Lexer;

        #[test]
        fn returns_char_on_current_position() {
            let lexer = Lexer::new("func");

            assert_eq!(lexer.peek(), Some('f'));
        }

        #[test]
        fn returns_none_when_source_is_empty() {
            let lexer = Lexer::new("");

            assert_eq!(lexer.peek(), None);
        }

        #[test]
        fn does_not_advance_position() {
            let lexer = Lexer::new("func");

            lexer.peek();

            assert_eq!(lexer.position(), 0);
        }
    }

    mod advance_tests {
        use super::Lexer;

        #[test]
        fn return_current_character_and_advance_position() {
            let mut lexer = Lexer::new("func");
            let result = lexer.advance();

            assert_eq!(result, Some('f'));
            assert_eq!(lexer.position(), 1);
        }

        #[test]
        fn returns_none_when_source_is_empty() {
            let mut lexer = Lexer::new("");

            assert_eq!(lexer.advance(), None);
        }
    }

    mod skip_whitespace_tests {
        use super::Lexer;

        #[test]
        fn skip_whitespaces() {
            let mut lexer = Lexer::new("     func");

            lexer.skip_whitespace();

            assert_eq!(lexer.position(), 5);
            assert_eq!(lexer.peek(), Some('f'));
        }

        #[test]
        fn skip_common_whitespace_characters() {
            let mut lexer = Lexer::new(" \t\n\rfunc");

            lexer.skip_whitespace();

            assert_eq!(lexer.position(), 4);
            assert_eq!(lexer.peek(), Some('f'));
        }

        #[test]
        fn does_nothing_when_current_char_is_not_whitespace() {
            let mut lexer = Lexer::new("func");
            lexer.skip_whitespace();

            assert_eq!(lexer.position(), 0);
            assert_eq!(lexer.peek(), Some('f'));
        }
    }
    mod next_token_tests {
        use super::Lexer;
        use crate::lexer::token::TokenKind;
        use crate::source::span::Span;
        #[test]
        fn returns_plus_token() {
            let mut lexer = Lexer::new("+");
            let token = lexer.next_token();
            assert_eq!(token.kind(), TokenKind::Plus);
            assert_eq!(token.span(), Span::new(0, 1));
        }
        #[test]
        fn returns_end_of_file_for_empty_source() {
            let mut lexer = Lexer::new("");
            let token = lexer.next_token();
            assert_eq!(token.kind(), TokenKind::EndOfFile);
            assert_eq!(token.span(), Span::new(0, 0));
        }
        #[test]
        fn skips_whitespace_before_token() {
            let mut lexer = Lexer::new("   +");
            let token = lexer.next_token();
            assert_eq!(token.kind(), TokenKind::Plus);
            assert_eq!(token.span(), Span::new(3, 4));
        }
        #[test]
        fn returns_unknown_for_unrecognized_character() {
            let mut lexer = Lexer::new("@");
            let token = lexer.next_token();
            assert_eq!(token.kind(), TokenKind::Unknown);
            assert_eq!(token.span(), Span::new(0, 1));
        }

        #[test]
        fn returns_single_character_tokens() {
            let cases = [
                ("(", TokenKind::LeftParenthesis),
                (")", TokenKind::RightParenthesis),
                ("{", TokenKind::LeftBrace),
                ("}", TokenKind::RightBrace),
                ("+", TokenKind::Plus),
                ("-", TokenKind::Minus),
                ("=", TokenKind::Equal),
                (";", TokenKind::Semicolon),
            ];
            for (source, expected_kind) in cases {
                let mut lexer = Lexer::new(source);
                let token = lexer.next_token();
                assert_eq!(token.kind(), expected_kind);
                assert_eq!(token.span(), Span::new(0, 1));
            }
        }
    }
}
