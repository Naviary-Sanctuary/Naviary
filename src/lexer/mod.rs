use logos::Logos;

#[derive(Logos, Debug, PartialEq, Clone)]
pub enum Token {
    // 키워드
    #[token("func")]
    Func,
    #[token("let")]
    Let,
    #[token("if")]
    If,
    #[token("else")]
    Else,
    #[token("return")]
    Return,
    #[token("mut")]
    Mut,
    #[token("for")]
    For,
    #[token("in")]
    In,

    // 타입
    #[token("int")]
    Int,
    #[token("float")]
    Float,
    #[token("string")]
    String,
    #[token("bool")]
    Bool,

    // 리터럴
    #[regex(r"-?[0-9]+", |lex| lex.slice().parse::<i64>().ok())]
    Number(i64),

    #[regex(r"-?[0-9]+\.[0-9]+", |lex| lex.slice().parse::<f64>().ok())]
    FloatNumber(f64),

    #[regex(r#""([^"\\]|\\.)*""#, |lex| {
        let s = lex.slice();
        Some(s[1..s.len()-1].to_string())
    })]
    StringLiteral(String),

    #[token("true", |_| true)]
    #[token("false", |_| false)]
    BoolLiteral(bool),

    // 식별자 (변수명, 함수명)
    #[regex("[a-zA-Z_][a-zA-Z0-9_]*", |lex| lex.slice().to_string())]
    Identifier(String),

    // 연산자
    #[token("+")]
    Plus,
    #[token("-")]
    Minus,
    #[token("*")]
    Star,
    #[token("/")]
    Slash,
    #[token("=")]
    Equal,

    // 비교 연산자
    #[token("==")]
    EqualEqual,
    #[token("!=")]
    NotEqual,
    #[token("<")]
    LessThan,
    #[token(">")]
    GreaterThan,
    #[token("<=")]
    LessThanEqual,
    #[token(">=")]
    GreaterThanEqual,

    // 범위 연산자
    #[token("..")] // 0..10 (exclusive)
    Range,
    #[token("..=")] // 0..=10 (inclusive)
    RangeEqual,

    // 구분자
    #[token("(")]
    LeftParen,
    #[token(")")]
    RightParen,
    #[token("{")]
    LeftBrace,
    #[token("}")]
    RightBrace,
    #[token(",")]
    Comma,
    #[token(";")]
    Semicolon,
    #[token(":")]
    Colon,
    #[token("->")]
    Arrow,

    // 공백과 주석 무시
    #[regex(r"[ \t\n\f]+", logos::skip)]
    #[regex(r"//[^\n]*", logos::skip)]
    Error,
}

// Lexer 래퍼
pub struct Lexer<'a> {
    inner: logos::Lexer<'a, Token>,
}

impl<'a> Lexer<'a> {
    pub fn new(input: &'a str) -> Self {
        Self {
            inner: Token::lexer(input),
        }
    }

    pub fn next_token(&mut self) -> Option<Token> {
        self.inner
            .next()
            .map(|result| result.unwrap_or(Token::Error))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_lexer() {
        let input = "func main() { let x = 10; }";
        let mut lexer = Lexer::new(input);

        assert_eq!(lexer.next_token(), Some(Token::Func));
        assert_eq!(
            lexer.next_token(),
            Some(Token::Identifier("main".to_string()))
        );
        assert_eq!(lexer.next_token(), Some(Token::LeftParen));
        assert_eq!(lexer.next_token(), Some(Token::RightParen));
        // ... 등등
    }
}
