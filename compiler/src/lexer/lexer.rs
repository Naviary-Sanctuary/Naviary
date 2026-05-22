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
    }
}
