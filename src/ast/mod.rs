#[derive(Debug, Clone, PartialEq)]
pub struct Program {
    pub functions: Vec<Function>,
}

#[derive(Debug, Clone, Copy, PartialEq)]
pub enum Type {
    Int,
    Float,
    String,
    Bool,
}

#[derive(Debug, Clone, PartialEq)]
pub struct Function {
    pub name: String,
    pub params: Vec<Parameter>,
    pub return_type: Option<Type>,
    pub body: Block,
}

#[derive(Debug, Clone, PartialEq)]
pub struct Parameter {
    pub name: String,
    pub ty: Type,
}

#[derive(Debug, Clone, PartialEq)]
pub struct Block {
    pub statements: Vec<Statement>,
}

#[derive(Debug, Clone, PartialEq)]
pub enum Statement {
    // let x = 10;
    Let {
        name: String,
        ty: Option<Type>, // 타입 명시 옵션
        value: Expression,
        mutable: bool,
    },

    // 변수 할당: x = 10
    Assignment {
        name: String,
        value: Box<Expression>,
    },

    // 표현식 구문 (함수 호출 등)
    Expression(Expression),
    // return x;
    Return(Option<Expression>),

    If {
        condition: Expression,
        then_block: Block,
        else_block: Option<Block>,
    },

    For {
        variable: String,
        start: Expression,
        end: Expression,
        inclusive: bool,
        body: Block,
    },
}

#[derive(Debug, Clone, PartialEq)]
pub enum Expression {
    // 42, 3.14
    Number(i64),
    Float(f64),
    // "hello"
    String(String),
    // true, false
    Bool(bool),
    // 변수 참조: x
    Identifier(String),
    // 이항 연산: x + y
    Binary {
        left: Box<Expression>,
        op: BinaryOp,
        right: Box<Expression>,
    },
    // 함수 호출: print(x)
    Call {
        name: String,
        args: Vec<Expression>,
    },
}

#[derive(Debug, Clone, PartialEq)]
pub enum BinaryOp {
    Add,              // +
    Subtract,         // -
    Multiply,         // *
    Divide,           // /
    Equal,            // ==
    NotEqual,         // !=
    LessThan,         // <
    GreaterThan,      // >
    LessThanEqual,    // <=
    GreaterThanEqual, // >=
}
