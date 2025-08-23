use crate::ast::*;
use anyhow::{Result, bail};
use std::collections::HashMap;

#[derive(Debug, Clone, PartialEq)]
pub enum FunctionKind {
    Regular {
        param_types: Vec<Type>,
        return_type: Option<Type>,
    },
    Builtin(BuiltinFunction),
}

#[derive(Debug, Clone, PartialEq)]
pub enum BuiltinFunction {
    Print, // 나중에 더 추가 가능
}
// 변수/함수의 타입 정보
#[derive(Debug, Clone, PartialEq)]
pub struct TypeInfo {
    pub ty: Type,
    pub is_mutable: bool,
    pub function_kind: Option<FunctionKind>,
}

// 타입 검사기
pub struct TypeChecker {
    // 전역 함수 테이블
    functions: HashMap<String, TypeInfo>,
    // 현재 스코프의 변수 테이블 (스택으로 관리)
    scopes: Vec<HashMap<String, TypeInfo>>,
    // 현재 함수의 반환 타입 (return 검사용)
    current_function_return_type: Option<Type>,
}

impl TypeChecker {
    pub fn new() -> Self {
        let mut checker = TypeChecker {
            functions: HashMap::new(),
            scopes: Vec::new(),
            current_function_return_type: None,
        };

        // 내장 함수 등록
        checker.register_builtin_functions();

        checker
    }

    // print 같은 내장 함수 등록
    fn register_builtin_functions(&mut self) {
        // print는 void 함수 (None 반환)
        self.functions.insert(
            "print".to_string(),
            TypeInfo {
                ty: Type::Int, // 무시됨
                is_mutable: false,
                function_kind: Some(FunctionKind::Builtin(BuiltinFunction::Print)),
            },
        );
    }

    // 새 스코프 시작 (함수, 블록 진입)
    fn push_scope(&mut self) {
        self.scopes.push(HashMap::new());
    }

    // 스코프 종료
    fn pop_scope(&mut self) {
        self.scopes.pop();
    }

    // 변수 등록
    fn declare_variable(&mut self, name: String, ty: Type, is_mutable: bool) -> Result<()> {
        let current_scope = self
            .scopes
            .last_mut()
            .ok_or_else(|| anyhow::anyhow!("No active scope"))?;

        if current_scope.contains_key(&name) {
            bail!("Variable '{}' already declared in this scope", name);
        }

        current_scope.insert(
            name,
            TypeInfo {
                ty,
                is_mutable,
                function_kind: None,
            },
        );

        Ok(())
    }

    // 변수/함수 조회
    fn lookup(&self, name: &str) -> Result<TypeInfo> {
        // 먼저 로컬 스코프에서 찾기 (가장 안쪽부터)
        for scope in self.scopes.iter().rev() {
            if let Some(info) = scope.get(name) {
                return Ok(info.clone());
            }
        }

        // 전역 함수에서 찾기
        if let Some(info) = self.functions.get(name) {
            return Ok(info.clone());
        }

        bail!("Undefined variable or function: '{}'", name)
    }

    // ===== 타입 검사 메서드들 =====

    pub fn check_program(&mut self, program: &Program) -> Result<()> {
        // 1단계: 모든 함수 시그니처 수집 (전방 선언 지원)
        for func in &program.functions {
            self.register_function(func)?;
        }

        // 2단계: 각 함수 본문 검사
        for func in &program.functions {
            self.check_function(func)?;
        }

        // main 함수 존재 확인
        if !self.functions.contains_key("main") {
            bail!("No main function found");
        }

        Ok(())
    }

    fn register_function(&mut self, func: &Function) -> Result<()> {
        let param_types: Vec<Type> = func.params.iter().map(|p| p.ty.clone()).collect();

        let info = TypeInfo {
            ty: func.return_type.clone().unwrap_or(Type::Int), // 이 필드는 함수에서 안 쓰임
            is_mutable: false,
            function_kind: Some(FunctionKind::Regular {
                param_types,
                return_type: func.return_type.clone(),
            }),
        };

        if self.functions.contains_key(&func.name) {
            bail!("Function '{}' already defined", func.name);
        }

        self.functions.insert(func.name.clone(), info);
        Ok(())
    }

    fn check_function(&mut self, func: &Function) -> Result<()> {
        // 새 스코프 시작
        self.push_scope();

        // 반환 타입 설정
        self.current_function_return_type = func.return_type.clone();

        // 매개변수를 스코프에 추가
        for param in &func.params {
            self.declare_variable(param.name.clone(), param.ty.clone(), false)?;
        }

        // 함수 본문 검사
        self.check_block(&func.body)?;

        // 반환 타입이 있는데 return이 없으면 에러 (간단한 검사)
        // TODO: 더 정교한 control flow 분석 필요

        // 스코프 종료
        self.pop_scope();

        Ok(())
    }

    fn check_block(&mut self, block: &Block) -> Result<()> {
        for stmt in &block.statements {
            self.check_statement(stmt)?;
        }
        Ok(())
    }

    fn check_statement(&mut self, stmt: &Statement) -> Result<()> {
        match stmt {
            Statement::Let {
                name,
                ty,
                value,
                mutable,
            } => {
                // 값의 타입 추론
                let value_type = self.infer_expression_type(value)?;

                // 명시된 타입이 있으면 일치하는지 확인
                let var_type = if let Some(declared_type) = ty {
                    if *declared_type != value_type {
                        bail!(
                            "Type mismatch: variable '{}' declared as {:?} but initialized with {:?}",
                            name,
                            declared_type,
                            value_type
                        );
                    }
                    declared_type.clone()
                } else {
                    // 타입 추론
                    value_type
                };

                // 변수 등록
                self.declare_variable(name.clone(), var_type, *mutable)?;
            }

            Statement::Assignment { name, value } => {
                let info = self.lookup(name)?;

                if info.function_kind.is_some() {
                    bail!("Cannot assign to function '{}'", name);
                }

                if !info.is_mutable {
                    bail!("Cannot assign to immutable variable '{}'", name);
                }

                let value_type = self.infer_expression_type(value)?;
                if value_type != info.ty {
                    bail!(
                        "Type mismatch in assignment: expected {:?}, found {:?}",
                        info.ty,
                        value_type
                    );
                }
            }

            Statement::Return(expr) => {
                let return_type = if let Some(expr) = expr {
                    Some(self.infer_expression_type(expr)?)
                } else {
                    None
                };

                // 함수의 반환 타입과 일치하는지 확인
                if return_type != self.current_function_return_type {
                    bail!(
                        "Return type mismatch: expected {:?}, found {:?}",
                        self.current_function_return_type,
                        return_type
                    );
                }
            }

            Statement::Expression(expr) => {
                // 표현식의 타입 검사
                // void 함수 호출도 허용
                let _ = self.check_expression_statement(expr);
            }

            Statement::If {
                condition,
                then_block,
                else_block,
            } => {
                let condition_type = self.infer_expression_type(condition)?;
                if condition_type != Type::Bool {
                    bail!("If condition must be bool, found {:?}", condition_type)
                }

                self.check_block(then_block)?;

                if let Some(else_block) = else_block {
                    self.check_block(else_block)?;
                }
            }

            Statement::For {
                variable,
                start,
                end,
                inclusive: _,
                body,
            } => {
                let start_type = self.infer_expression_type(start)?;
                let end_type = self.infer_expression_type(end)?;

                if start_type != end_type {
                    bail!(
                        "Start and end of for loop must have the same type, found {:?} and {:?}",
                        start_type,
                        end_type
                    );
                }

                match start_type {
                    Type::Int => {
                        // 새 스코프 시작 (for 블록)
                        self.push_scope();
                        // loop 변수를 immutable int로 등록
                        self.declare_variable(variable.clone(), Type::Int, false)?;
                        // body 체크
                        self.check_block(body)?;
                        // 스코프 종료
                        self.pop_scope();
                    }
                    _ => bail!(
                        "For loop range must be numeric type, found {:?}",
                        start_type
                    ),
                }
            }
        }

        Ok(())
    }

    // Expression statement를 위한 별도 메서드 (void 함수 호출 허용)
    fn check_expression_statement(&self, expr: &Expression) -> Result<()> {
        match expr {
            Expression::Call { name, args } => {
                let info = self.lookup(name)?;

                match &info.function_kind {
                    Some(FunctionKind::Builtin(BuiltinFunction::Print)) => {
                        // 최소 1개 이상
                        if args.is_empty() {
                            bail!("print() expects at least 1 argument");
                        }

                        // 모든 인자가 출력 가능한 타입인지 확인
                        for arg in args {
                            let arg_type = self.infer_expression_type(arg)?;
                            match arg_type {
                                Type::Int | Type::Float | Type::String | Type::Bool => {}
                                _ => bail!("Cannot print type {:?}", arg_type),
                            }
                        }
                        Ok(())
                    }
                    Some(FunctionKind::Regular {
                        param_types,
                        return_type: _,
                    }) => {
                        if args.len() != param_types.len() {
                            bail!(
                                "Function '{}' expects {} arguments, but {} provided",
                                name,
                                param_types.len(),
                                args.len()
                            );
                        }

                        for (i, (arg, expected_type)) in args.iter().zip(param_types).enumerate() {
                            let arg_type = self.infer_expression_type(arg)?;
                            if arg_type != *expected_type {
                                bail!(
                                    "Type mismatch in argument {} of function '{}': expected {:?}, found {:?}",
                                    i + 1,
                                    name,
                                    expected_type,
                                    arg_type
                                );
                            }
                        }
                        Ok(())
                    }
                    None => bail!("'{}' is not a function", name),
                }
            }
            _ => {
                self.infer_expression_type(expr)?;
                Ok(())
            }
        }
    }

    fn infer_expression_type(&self, expr: &Expression) -> Result<Type> {
        match expr {
            Expression::Number(_) => Ok(Type::Int),
            Expression::Float(_) => Ok(Type::Float),
            Expression::String(_) => Ok(Type::String),
            Expression::Bool(_) => Ok(Type::Bool),

            Expression::Identifier(name) => {
                let info = self.lookup(name)?;
                if info.function_kind.is_some() {
                    bail!("Cannot use function '{}' as a value", name);
                }
                Ok(info.ty)
            }

            Expression::Binary { left, op, right } => {
                let left_type = self.infer_expression_type(left)?;
                let right_type = self.infer_expression_type(right)?;

                // 타입 호환성 검사
                match op {
                    BinaryOp::Add | BinaryOp::Subtract | BinaryOp::Multiply | BinaryOp::Divide => {
                        // 숫자 연산
                        if left_type != right_type {
                            bail!(
                                "Type mismatch in binary operation: {:?} {:?} {:?}",
                                left_type,
                                op,
                                right_type
                            );
                        }

                        match left_type {
                            Type::Int | Type::Float => Ok(left_type),
                            _ => bail!(
                                "Invalid types for arithmetic operation: {:?} {:?} {:?}",
                                left_type,
                                op,
                                right_type
                            ),
                        }
                    }

                    BinaryOp::Equal | BinaryOp::NotEqual => {
                        // 비교 연산
                        if left_type != right_type {
                            bail!(
                                "Cannot compare different types: {:?} and {:?}",
                                left_type,
                                right_type
                            );
                        }
                        Ok(Type::Bool)
                    }

                    BinaryOp::LessThan
                    | BinaryOp::GreaterThan
                    | BinaryOp::LessThanEqual
                    | BinaryOp::GreaterThanEqual => {
                        if left_type != right_type {
                            bail!(
                                "Cannot compare different types: {:?} and {:?}",
                                left_type,
                                right_type
                            );
                        }

                        match left_type {
                            Type::Int | Type::Float => Ok(Type::Bool),
                            _ => bail!(
                                "Invalid types for comparison operation: {:?} {:?} {:?}",
                                left_type,
                                op,
                                right_type
                            ),
                        }
                    }
                }
            }

            Expression::Call { name, args } => {
                let info = self.lookup(name)?;

                match &info.function_kind {
                    Some(FunctionKind::Builtin(BuiltinFunction::Print)) => {
                        bail!("Void function 'print' cannot be used as a value");
                    }
                    Some(FunctionKind::Regular {
                        param_types,
                        return_type,
                    }) => {
                        if args.len() != param_types.len() {
                            bail!(
                                "Function '{}' expects {} arguments, but {} provided",
                                name,
                                param_types.len(),
                                args.len()
                            );
                        }

                        for (i, (arg, expected_type)) in args.iter().zip(param_types).enumerate() {
                            let arg_type = self.infer_expression_type(arg)?;
                            if arg_type != *expected_type {
                                bail!(
                                    "Type mismatch in argument {} of function '{}': expected {:?}, found {:?}",
                                    i + 1,
                                    name,
                                    expected_type,
                                    arg_type
                                );
                            }
                        }

                        return_type.clone().ok_or_else(|| {
                            anyhow::anyhow!("Void function '{}' cannot be used as a value", name)
                        })
                    }
                    None => bail!("'{}' is not a function", name),
                }
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::lexer::Lexer;
    use crate::parser::Parser;

    #[test]
    fn test_type_check_valid() {
        let input = r#"
            func add(x: int, y: int) -> int {
                return x + y;
            }
            
            func main() {
                let a = 10;
                let b = 20;
                let result = add(a, b);
                print(result);
            }
        "#;

        let lexer = Lexer::new(input);
        let mut parser = Parser::new(lexer);
        let program = parser.parse_program().unwrap();

        let mut checker = TypeChecker::new();
        assert!(checker.check_program(&program).is_ok());
    }

    #[test]
    fn test_type_mismatch() {
        let input = r#"
            func main() {
                let x: int = "hello";
            }
        "#;

        let lexer = Lexer::new(input);
        let mut parser = Parser::new(lexer);
        let program = parser.parse_program().unwrap();

        let mut checker = TypeChecker::new();
        assert!(checker.check_program(&program).is_err());
    }
}
