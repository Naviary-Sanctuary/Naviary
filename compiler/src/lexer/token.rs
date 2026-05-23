use crate::source::span::Span;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum TokenKind {
    Func,
    Let,
    Identifier,

    Integer,

    LeftParenthesis,
    RightParenthesis,
    LeftBrace,
    RightBrace,

    Plus,
    Minus,
    Equal,

    Semicolon,

    EndOfFile,
    Unknown,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Token {
    kind: TokenKind,
    span: Span,
}

impl Token {
    pub fn kind(&self) -> TokenKind {
        self.kind
    }
    pub fn span(&self) -> Span {
        self.span
    }
    pub fn new(kind: TokenKind, span: Span) -> Self {
        Self { kind, span }
    }
}

#[cfg(test)]
mod tests {
    use super::Token;
    use super::TokenKind;

    mod new_test {
        use super::Token;
        use super::TokenKind;
        use crate::source::span::Span;

        #[test]
        fn create_token_with_args() {
            let token = Token::new(TokenKind::Let, Span::new(0, 2));

            assert_eq!(token.kind(), TokenKind::Let);
            assert_eq!(token.span(), Span::new(0, 2));
        }
    }
}
