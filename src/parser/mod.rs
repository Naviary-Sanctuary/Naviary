use crate::ast::*;
use crate::lexer::{Lexer, Token};
use anyhow::{Result, bail};

pub struct Parser<'a> {
    lexer: Lexer<'a>,
    current_token: Option<Token>,
}

impl<'a> Parser<'a> {
    pub fn new(mut lexer: Lexer<'a>) -> Self {
        let current_token = lexer.next_token();
        Parser {
            lexer,
            current_token,
        }
    }

    // 다음 토큰으로 이동
    fn advance(&mut self) {
        self.current_token = self.lexer.next_token();
    }

    // 현재 토큰 확인 (소비하지 않음)
    fn peek(&self) -> Option<&Token> {
        self.current_token.as_ref()
    }

    // 특정 토큰을 기대하고 소비
    fn expect(&mut self, expected: Token) -> Result<()> {
        match &self.current_token {
            Some(token) if *token == expected => {
                self.advance();
                Ok(())
            }
            Some(token) => bail!("Expected {:?}, found {:?}", expected, token),
            None => bail!("Expected {:?}, found EOF", expected),
        }
    }

    // 식별자를 기대하고 그 값을 반환
    fn expect_identifier(&mut self) -> Result<String> {
        match &self.current_token {
            Some(Token::Identifier(name)) => {
                let name = name.clone();
                self.advance();
                Ok(name)
            }
            _ => bail!("Expected identifier, found {:?}", self.current_token),
        }
    }

    // ===== 파싱 메서드들 =====

    // 프로그램 전체 파싱 (진입점)
    pub fn parse_program(&mut self) -> Result<Program> {
        let mut functions = Vec::new();

        while self.current_token.is_some() {
            functions.push(self.parse_function()?);
        }

        Ok(Program { functions })
    }

    // func name(params) -> type { body }
    fn parse_function(&mut self) -> Result<Function> {
        // "func" 키워드
        self.expect(Token::Func)?;

        // 함수 이름
        let name = self.expect_identifier()?;

        // 매개변수 리스트
        self.expect(Token::LeftParen)?;
        let params = self.parse_parameter_list()?;
        self.expect(Token::RightParen)?;

        // 반환 타입 (옵션)
        let return_type = if self.peek() == Some(&Token::Arrow) {
            self.advance(); // -> 소비
            Some(self.parse_type()?)
        } else {
            None
        };

        // 함수 본문
        self.expect(Token::LeftBrace)?;
        let body = self.parse_block()?;
        self.expect(Token::RightBrace)?;

        Ok(Function {
            name,
            params,
            return_type,
            body,
        })
    }

    // x: int, y: int
    fn parse_parameter_list(&mut self) -> Result<Vec<Parameter>> {
        let mut params = Vec::new();

        // 빈 매개변수 리스트
        if self.peek() == Some(&Token::RightParen) {
            return Ok(params);
        }

        loop {
            let name = self.expect_identifier()?;
            self.expect(Token::Colon)?;
            let ty = self.parse_type()?;

            params.push(Parameter { name, ty });

            // 더 있나?
            if self.peek() == Some(&Token::Comma) {
                self.advance();
            } else {
                break;
            }
        }

        Ok(params)
    }

    // 타입 파싱
    fn parse_type(&mut self) -> Result<Type> {
        let ty = match &self.current_token {
            Some(Token::Int) => Type::Int,
            Some(Token::Float) => Type::Float,
            Some(Token::String) => Type::String,
            Some(Token::Bool) => Type::Bool,
            _ => bail!("Expected type, found {:?}", self.current_token),
        };
        self.advance();
        Ok(ty)
    }

    // 블록 파싱 { statements }
    fn parse_block(&mut self) -> Result<Block> {
        let mut statements = Vec::new();

        while self.peek() != Some(&Token::RightBrace) && self.current_token.is_some() {
            statements.push(self.parse_statement()?);
        }

        Ok(Block { statements })
    }

    // 문장 파싱
    fn parse_statement(&mut self) -> Result<Statement> {
        match &self.current_token {
            Some(Token::Let) => self.parse_let_statement(),
            Some(Token::Return) => self.parse_return_statement(),
            Some(Token::If) => self.parse_if_statement(),
            _ => {
                // 표현식 문장 (함수 호출 등)
                let expr = self.parse_expression()?;
                self.expect(Token::Semicolon)?;
                Ok(Statement::Expression(expr))
            }
        }
    }

    // let name = value;
    // let name: type = value;
    fn parse_let_statement(&mut self) -> Result<Statement> {
        self.advance(); // 'let' 소비

        let name = self.expect_identifier()?;

        // 타입 명시 (옵션)
        let ty = if self.peek() == Some(&Token::Colon) {
            self.advance(); // ':' 소비
            Some(self.parse_type()?)
        } else {
            None
        };

        self.expect(Token::Equal)?;
        let value = self.parse_expression()?;
        self.expect(Token::Semicolon)?;

        Ok(Statement::Let { name, ty, value })
    }

    // return expr;
    fn parse_return_statement(&mut self) -> Result<Statement> {
        self.advance(); // 'return' 소비

        // return; (값 없음) 또는 return expr;
        let value = if self.peek() == Some(&Token::Semicolon) {
            None
        } else {
            Some(self.parse_expression()?)
        };

        self.expect(Token::Semicolon)?;
        Ok(Statement::Return(value))
    }

    fn parse_if_statement(&mut self) -> Result<Statement> {
        self.expect(Token::If)?;
        let condition = self.parse_expression()?;
        self.expect(Token::LeftBrace)?;
        let then_block = self.parse_block()?;
        self.expect(Token::RightBrace)?;

        let else_block = if self.peek() == Some(&Token::Else) {
            self.advance();

            if self.peek() == Some(&Token::If) {
                // else if 케이스 - 재귀적으로 if문 파싱
                Some(Block {
                    statements: vec![self.parse_if_statement()?],
                })
            } else {
                // 일반 else 케이스
                self.expect(Token::LeftBrace)?;
                let block = self.parse_block()?;
                self.expect(Token::RightBrace)?;
                Some(block)
            }
        } else {
            None
        };

        Ok(Statement::If {
            condition,
            then_block,
            else_block,
        })
    }

    // 표현식 파싱 (우선순위 처리)
    fn parse_expression(&mut self) -> Result<Expression> {
        self.parse_comparison()
    }

    fn parse_comparison(&mut self) -> Result<Expression> {
        let mut left = self.parse_additive()?;

        while let Some(token) = self.peek() {
            let op = match token {
                Token::EqualEqual => BinaryOp::Equal,
                Token::NotEqual => BinaryOp::NotEqual,
                Token::LessThan => BinaryOp::LessThan,
                Token::GreaterThan => BinaryOp::GreaterThan,
                Token::LessThanEqual => BinaryOp::LessThanEqual,
                Token::GreaterThanEqual => BinaryOp::GreaterThanEqual,
                _ => break,
            };

            self.advance();
            let right = self.parse_additive()?;
            left = Expression::Binary {
                left: Box::new(left),
                op,
                right: Box::new(right),
            };
        }

        Ok(left)
    }

    // 덧셈/뺄셈 (낮은 우선순위)
    fn parse_additive(&mut self) -> Result<Expression> {
        let mut left = self.parse_multiplicative()?;

        while let Some(token) = self.peek() {
            let op = match token {
                Token::Plus => BinaryOp::Add,
                Token::Minus => BinaryOp::Subtract,
                _ => break,
            };

            self.advance();
            let right = self.parse_multiplicative()?;
            left = Expression::Binary {
                left: Box::new(left),
                op,
                right: Box::new(right),
            };
        }

        Ok(left)
    }

    // 곱셈/나눗셈 (높은 우선순위)
    fn parse_multiplicative(&mut self) -> Result<Expression> {
        let mut left = self.parse_primary()?;

        while let Some(token) = self.peek() {
            let op = match token {
                Token::Star => BinaryOp::Multiply,
                Token::Slash => BinaryOp::Divide,
                _ => break,
            };

            self.advance();
            let right = self.parse_primary()?;
            left = Expression::Binary {
                left: Box::new(left),
                op,
                right: Box::new(right),
            };
        }

        Ok(left)
    }

    // 기본 표현식 (리터럴, 변수, 함수 호출, 괄호)
    fn parse_primary(&mut self) -> Result<Expression> {
        match &self.current_token.clone() {
            Some(Token::Number(n)) => {
                let n = *n;
                self.advance();
                Ok(Expression::Number(n))
            }
            Some(Token::FloatNumber(f)) => {
                let f = *f;
                self.advance();
                Ok(Expression::Float(f))
            }
            Some(Token::StringLiteral(s)) => {
                let s = s.clone();
                self.advance();
                Ok(Expression::String(s))
            }
            Some(Token::BoolLiteral(b)) => {
                let b = *b;
                self.advance();
                Ok(Expression::Bool(b))
            }
            Some(Token::Identifier(name)) => {
                let name = name.clone();
                self.advance();

                // 함수 호출인가?
                if self.peek() == Some(&Token::LeftParen) {
                    self.advance(); // '(' 소비
                    let args = self.parse_argument_list()?;
                    self.expect(Token::RightParen)?;
                    Ok(Expression::Call { name, args })
                } else {
                    // 단순 변수 참조
                    Ok(Expression::Identifier(name))
                }
            }
            Some(Token::LeftParen) => {
                // 괄호 표현식
                self.advance();
                let expr = self.parse_expression()?;
                self.expect(Token::RightParen)?;
                Ok(expr)
            }
            _ => bail!("Unexpected token in expression: {:?}", self.current_token),
        }
    }

    // 함수 호출 인자 리스트
    fn parse_argument_list(&mut self) -> Result<Vec<Expression>> {
        let mut args = Vec::new();

        if self.peek() == Some(&Token::RightParen) {
            return Ok(args);
        }

        loop {
            args.push(self.parse_expression()?);

            if self.peek() == Some(&Token::Comma) {
                self.advance();
            } else {
                break;
            }
        }

        Ok(args)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_simple_function() {
        let input = r#"
            func main() {
                let x = 10;
                let y = 20;
                return x + y;
            }
        "#;

        let lexer = Lexer::new(input);
        let mut parser = Parser::new(lexer);
        let program = parser.parse_program().unwrap();

        assert_eq!(program.functions.len(), 1);
        assert_eq!(program.functions[0].name, "main");
    }

    #[test]
    fn test_operator_precedence() {
        let input = "func test() { let x = 10 + 20 * 3; }";

        let lexer = Lexer::new(input);
        let mut parser = Parser::new(lexer);
        parser.parse_program().unwrap();

        // 10 + (20 * 3) 로 파싱되어야 함
        // AST 구조 확인 가능
    }
}
