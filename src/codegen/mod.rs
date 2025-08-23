use crate::ast::*;
use anyhow::{Result, bail};
use inkwell::IntPredicate;
use inkwell::builder::Builder;
use inkwell::context::Context;
use inkwell::module::Module;
use inkwell::types::{BasicMetadataTypeEnum, BasicTypeEnum};
use inkwell::values::{BasicValueEnum, FunctionValue, PointerValue};
use std::collections::HashMap;

pub struct CodeGenerator<'ctx> {
    context: &'ctx Context,
    module: Module<'ctx>,
    builder: Builder<'ctx>,
    // 변수 심볼 테이블 (변수명 -> LLVM 값)
    variables: HashMap<String, (PointerValue<'ctx>, Type, bool)>,
    // 함수 심볼 테이블
    functions: HashMap<String, FunctionValue<'ctx>>,
    // 현재 함수
    current_function: Option<FunctionValue<'ctx>>,
}

impl<'ctx> CodeGenerator<'ctx> {
    pub fn new(context: &'ctx Context, module_name: &str) -> Self {
        let module = context.create_module(module_name);
        let builder = context.create_builder();

        CodeGenerator {
            context,
            module,
            builder,
            variables: HashMap::new(),
            functions: HashMap::new(),
            current_function: None,
        }
    }

    // 내장 함수 선언
    fn declare_builtin_functions(&mut self) {
        let i32_type = self.context.i32_type();
        let i8_ptr_type = self
            .context
            .i8_type()
            .ptr_type(inkwell::AddressSpace::default());

        let printf_type = i32_type.fn_type(
            &[i8_ptr_type.into()],
            true, // variadic (가변 인자)
        );

        let printf_fn = self.module.add_function("printf", printf_type, None);
        self.functions.insert("printf".to_string(), printf_fn);
    }

    // AST 타입을 LLVM 타입으로 변환
    fn get_llvm_type(&self, ty: &Type) -> BasicTypeEnum<'ctx> {
        match ty {
            Type::Int => self.context.i32_type().into(),
            Type::Float => self.context.f64_type().into(),
            Type::Bool => self.context.bool_type().into(),
            Type::String => self
                .context
                .ptr_type(inkwell::AddressSpace::default())
                .into(),
        }
    }

    // 프로그램 전체 컴파일
    pub fn compile_program(&mut self, program: &Program) -> Result<()> {
        // 내장 함수 선언
        self.declare_builtin_functions();

        // 모든 함수 선언 (전방 선언 지원)
        for func in &program.functions {
            self.declare_function(func)?;
        }

        // 각 함수 본문 컴파일
        for func in &program.functions {
            self.compile_function(func)?;
        }

        // LLVM IR 검증
        if let Err(e) = self.module.verify() {
            bail!("Module verification failed: {}", e.to_string());
        }

        Ok(())
    }

    // 함수 선언
    fn declare_function(&mut self, func: &Function) -> Result<()> {
        // 매개변수 타입들 - BasicMetadataTypeEnum으로 변환
        let param_types: Vec<BasicMetadataTypeEnum> = func
            .params
            .iter()
            .map(|p| BasicMetadataTypeEnum::from(self.get_llvm_type(&p.ty)))
            .collect();

        // 함수 타입
        let fn_type = if let Some(ref return_type) = func.return_type {
            let ret_type = self.get_llvm_type(return_type);
            match ret_type {
                BasicTypeEnum::IntType(t) => t.fn_type(&param_types, false),
                BasicTypeEnum::FloatType(t) => t.fn_type(&param_types, false),
                _ => bail!("Unsupported return type"),
            }
        } else {
            // void 반환
            self.context.void_type().fn_type(&param_types, false)
        };

        // 함수 추가
        let function = self.module.add_function(&func.name, fn_type, None);
        self.functions.insert(func.name.clone(), function);

        Ok(())
    }

    // 함수 본문 컴파일
    fn compile_function(&mut self, func: &Function) -> Result<()> {
        let function = *self
            .functions
            .get(&func.name)
            .ok_or_else(|| anyhow::anyhow!("Function not found: {}", func.name))?;

        self.current_function = Some(function);

        // entry 블록 생성
        let entry = self.context.append_basic_block(function, "entry");
        self.builder.position_at_end(entry);

        // 새 변수 스코프
        self.variables.clear();

        // 매개변수를 변수로 저장
        for (i, param) in func.params.iter().enumerate() {
            let arg = function
                .get_nth_param(i as u32)
                .ok_or_else(|| anyhow::anyhow!("Missing parameter"))?;

            // 매개변수를 위한 alloca (스택 할당)
            let alloca = self.create_entry_block_alloca(&param.name, &param.ty);
            self.builder.build_store(alloca, arg)?;
            self.variables
                .insert(param.name.clone(), (alloca, param.ty.clone(), false));
        }

        // 함수 본문 컴파일
        self.compile_block(&func.body)?;

        // 명시적 return이 없는 경우 처리
        if self.current_block_has_no_terminator() {
            if func.return_type.is_none() {
                // void 함수
                self.builder.build_return(None)?;
            } else if func.name == "main" {
                // main 함수는 기본적으로 0 반환
                let zero = self.context.i32_type().const_int(0, false);
                self.builder.build_return(Some(&zero))?;
            } else {
                // 다른 함수는 에러 (return이 필요함)
                bail!("Function '{}' must return a value", func.name);
            }
        }

        Ok(())
    }

    // 현재 블록이 terminator(return, br 등)를 가지지 않는지 확인
    fn current_block_has_no_terminator(&self) -> bool {
        let block = self.builder.get_insert_block().unwrap();
        block.get_terminator().is_none()
    }

    // 함수 진입점에 alloca 생성 (스택 변수 할당)
    fn create_entry_block_alloca(&self, name: &str, ty: &Type) -> PointerValue<'ctx> {
        let builder = self.context.create_builder();
        let entry = self
            .current_function
            .unwrap()
            .get_first_basic_block()
            .unwrap();

        match entry.get_first_instruction() {
            Some(first_instr) => builder.position_before(&first_instr),
            None => builder.position_at_end(entry),
        }

        let llvm_type = self.get_llvm_type(ty);
        builder.build_alloca(llvm_type, name).unwrap()
    }

    // 블록 컴파일
    fn compile_block(&mut self, block: &Block) -> Result<()> {
        for stmt in &block.statements {
            self.compile_statement(stmt)?;
        }
        Ok(())
    }

    // 문장 컴파일
    fn compile_statement(&mut self, stmt: &Statement) -> Result<()> {
        match stmt {
            Statement::Let {
                name,
                ty,
                value,
                mutable,
            } => {
                // 값 계산
                let val = self.compile_expression(value)?;

                // 변수를 위한 스택 공간 할당
                let var_type = ty.as_ref().unwrap_or(&Type::Int); // 타입 추론된 경우 기본값 (실제로는 type checker가 처리)
                let alloca = self.create_entry_block_alloca(name, var_type);

                // 값 저장
                self.builder.build_store(alloca, val)?;
                self.variables
                    .insert(name.clone(), (alloca, var_type.clone(), *mutable));
            }

            Statement::Assignment { name, value } => {
                let ptr = match self.variables.get(name) {
                    Some(&(ptr, _, _)) => ptr,
                    None => bail!("Undefined variable: {}", name),
                };

                let new_value = self.compile_expression(value)?;
                self.builder.build_store(ptr, new_value)?;
            }

            Statement::Return(expr) => {
                if let Some(expr) = expr {
                    let val = self.compile_expression(expr)?;
                    self.builder.build_return(Some(&val))?;
                } else {
                    self.builder.build_return(None)?;
                }
            }

            Statement::Expression(expr) => {
                // 표현식 실행 (결과 무시)
                self.compile_expression(expr)?;
            }

            Statement::If {
                condition,
                then_block,
                else_block,
            } => {
                // 1. 조건식 계산 (x > 5) -> true/false 값 생성
                let condition_value = self.compile_expression(condition)?;

                // 현재 함수 가져오기
                let function = self.current_function.unwrap();

                // 블록 생성
                // then_bb: if가 true일 때 실행할 블록
                // else_bb: if가 false일 때 실행할 블록
                // merge_bb: 둘다 끝나고 만나는 합류점
                let then_bb = self.context.append_basic_block(function, "then");
                let else_bb = self.context.append_basic_block(function, "else");
                let merge_bb = self.context.append_basic_block(function, "merge");

                // 조건 분기
                // condition_value가 true면 then_bb로 이동, 아니면 else_bb로 이동하지만 else_bb가 없으면 merge_bb로 이동해라
                self.builder.build_conditional_branch(
                    condition_value.into_int_value(),
                    then_bb,
                    if else_block.is_some() {
                        else_bb
                    } else {
                        merge_bb
                    },
                )?;

                // then 블록 컴파일, if가 false면 어차피 실행되지 않는다.
                self.builder.position_at_end(then_bb);
                self.compile_block(then_block)?;
                // then 블록에서 명시적 return이 없으면 merge_bb로 이동해라
                if self.current_block_has_no_terminator() {
                    self.builder.build_unconditional_branch(merge_bb)?;
                }

                // else 블록 컴파일
                if let Some(else_block) = else_block {
                    self.builder.position_at_end(else_bb);
                    self.compile_block(else_block)?;
                    if self.current_block_has_no_terminator() {
                        self.builder.build_unconditional_branch(merge_bb)?;
                    }
                } else {
                    self.builder.position_at_end(else_bb);
                    self.builder.build_unconditional_branch(merge_bb)?;
                }

                self.builder.position_at_end(merge_bb);
            }
            Statement::For {
                variable,
                start,
                end,
                inclusive,
                body,
            } => {
                // 1. start와 end 값 계산
                let start_val = self.compile_expression(start)?;
                let end_val = self.compile_expression(end)?;

                let function = self.current_function.unwrap();

                // 2. 필요한 블록들 생성
                let loop_header = self.context.append_basic_block(function, "loop_header");
                let loop_body = self.context.append_basic_block(function, "loop_body");
                let loop_exit = self.context.append_basic_block(function, "loop_exit");

                // 3. loop 변수를 위한 alloca (함수 entry에)
                let loop_var = self.create_entry_block_alloca(variable, &Type::Int);

                // 4. 초기값 저장
                self.builder.build_store(loop_var, start_val)?;

                // 5. loop_header로 점프
                self.builder.build_unconditional_branch(loop_header)?;

                // 6. loop_header: 조건 체크
                self.builder.position_at_end(loop_header);
                let current_val =
                    self.builder
                        .build_load(self.context.i32_type(), loop_var, variable)?;

                // 비교 (inclusive에 따라 < 또는 <=)
                let op = if *inclusive {
                    IntPredicate::SLE
                } else {
                    IntPredicate::SLT
                };

                let condition = self.builder.build_int_compare(
                    op,
                    current_val.into_int_value(),
                    end_val.into_int_value(),
                    "loop_cond",
                )?;

                self.builder
                    .build_conditional_branch(condition, loop_body, loop_exit)?;

                // 7. loop_body: 본문 실행
                self.builder.position_at_end(loop_body);

                // loop 변수를 스코프에 추가 (immutable)
                let old_vars = self.variables.clone(); // 기존 변수 백업
                self.variables
                    .insert(variable.clone(), (loop_var, Type::Int, false));

                // body 컴파일
                self.compile_block(body)?;

                // i++ (증가)
                let current = self
                    .builder
                    .build_load(self.context.i32_type(), loop_var, "i")?;
                let next = self.builder.build_int_add(
                    current.into_int_value(),
                    self.context.i32_type().const_int(1, false),
                    "next_i",
                )?;
                self.builder.build_store(loop_var, next)?;

                // loop_header로 다시
                self.builder.build_unconditional_branch(loop_header)?;

                // 8. loop_exit: 루프 종료 후
                self.builder.position_at_end(loop_exit);

                // 변수 스코프 복원
                self.variables = old_vars;
            }
        }

        Ok(())
    }

    // 표현식 컴파일
    fn compile_expression(&mut self, expr: &Expression) -> Result<BasicValueEnum<'ctx>> {
        match expr {
            Expression::Number(n) => {
                let val = self.context.i32_type().const_int(*n as u64, false);
                Ok(val.into())
            }

            Expression::Float(f) => {
                let val = self.context.f64_type().const_float(*f);
                Ok(val.into())
            }

            Expression::Bool(b) => {
                let val = self.context.bool_type().const_int(*b as u64, false);
                Ok(val.into())
            }

            Expression::String(s) => {
                // 문자열 리터럴 (전역 상수로)
                let val = self.builder.build_global_string_ptr(s, "str")?;
                Ok(val.as_pointer_value().into())
            }

            Expression::Identifier(name) => {
                let (ptr, ty) = match self.variables.get(name) {
                    Some(&(ptr, ty, _)) => (ptr, ty), // 둘 다 Copy!
                    None => bail!("Undefined variable: {}", name),
                };

                let llvm_type = self.get_llvm_type(&ty);
                let val = self.builder.build_load(llvm_type, ptr, name)?;
                Ok(val)
            }
            Expression::Binary { left, op, right } => {
                let lhs = self.compile_expression(left)?;
                let rhs = self.compile_expression(right)?;

                match op {
                    BinaryOp::Add => {
                        if lhs.is_int_value() {
                            let result = self.builder.build_int_add(
                                lhs.into_int_value(),
                                rhs.into_int_value(),
                                "add",
                            )?;
                            Ok(result.into())
                        } else {
                            let result = self.builder.build_float_add(
                                lhs.into_float_value(),
                                rhs.into_float_value(),
                                "fadd",
                            )?;
                            Ok(result.into())
                        }
                    }

                    BinaryOp::Subtract => {
                        if lhs.is_int_value() {
                            let result = self.builder.build_int_sub(
                                lhs.into_int_value(),
                                rhs.into_int_value(),
                                "sub",
                            )?;
                            Ok(result.into())
                        } else {
                            let result = self.builder.build_float_sub(
                                lhs.into_float_value(),
                                rhs.into_float_value(),
                                "fsub",
                            )?;
                            Ok(result.into())
                        }
                    }

                    BinaryOp::Multiply => {
                        if lhs.is_int_value() {
                            let result = self.builder.build_int_mul(
                                lhs.into_int_value(),
                                rhs.into_int_value(),
                                "mul",
                            )?;
                            Ok(result.into())
                        } else {
                            let result = self.builder.build_float_mul(
                                lhs.into_float_value(),
                                rhs.into_float_value(),
                                "fmul",
                            )?;
                            Ok(result.into())
                        }
                    }

                    BinaryOp::Divide => {
                        if lhs.is_int_value() {
                            let result = self.builder.build_int_signed_div(
                                lhs.into_int_value(),
                                rhs.into_int_value(),
                                "div",
                            )?;
                            Ok(result.into())
                        } else {
                            let result = self.builder.build_float_div(
                                lhs.into_float_value(),
                                rhs.into_float_value(),
                                "fdiv",
                            )?;
                            Ok(result.into())
                        }
                    }

                    BinaryOp::Equal => {
                        let result = if lhs.is_int_value() {
                            self.builder.build_int_compare(
                                IntPredicate::EQ,
                                lhs.into_int_value(),
                                rhs.into_int_value(),
                                "eq",
                            )?
                        } else {
                            self.builder.build_float_compare(
                                inkwell::FloatPredicate::OEQ,
                                lhs.into_float_value(),
                                rhs.into_float_value(),
                                "feq",
                            )?
                        };
                        Ok(result.into())
                    }

                    BinaryOp::NotEqual => {
                        let result = if lhs.is_int_value() {
                            self.builder.build_int_compare(
                                IntPredicate::NE,
                                lhs.into_int_value(),
                                rhs.into_int_value(),
                                "ne",
                            )?
                        } else {
                            self.builder.build_float_compare(
                                inkwell::FloatPredicate::ONE,
                                lhs.into_float_value(),
                                rhs.into_float_value(),
                                "fne",
                            )?
                        };
                        Ok(result.into())
                    }

                    BinaryOp::LessThan => {
                        let result = if lhs.is_int_value() {
                            self.builder.build_int_compare(
                                IntPredicate::SLT,
                                lhs.into_int_value(),
                                rhs.into_int_value(),
                                "lt",
                            )?
                        } else {
                            self.builder.build_float_compare(
                                inkwell::FloatPredicate::OLT,
                                lhs.into_float_value(),
                                rhs.into_float_value(),
                                "olt",
                            )?
                        };
                        Ok(result.into())
                    }

                    BinaryOp::GreaterThan => {
                        let result = if lhs.is_int_value() {
                            self.builder.build_int_compare(
                                IntPredicate::SGT,
                                lhs.into_int_value(),
                                rhs.into_int_value(),
                                "gt",
                            )?
                        } else {
                            self.builder.build_float_compare(
                                inkwell::FloatPredicate::OGT,
                                lhs.into_float_value(),
                                rhs.into_float_value(),
                                "ogt",
                            )?
                        };
                        Ok(result.into())
                    }

                    BinaryOp::LessThanEqual => {
                        let result = if lhs.is_int_value() {
                            self.builder.build_int_compare(
                                IntPredicate::SLE,
                                lhs.into_int_value(),
                                rhs.into_int_value(),
                                "le",
                            )?
                        } else {
                            self.builder.build_float_compare(
                                inkwell::FloatPredicate::OLE,
                                lhs.into_float_value(),
                                rhs.into_float_value(),
                                "ole",
                            )?
                        };
                        Ok(result.into())
                    }

                    BinaryOp::GreaterThanEqual => {
                        let result = if lhs.is_int_value() {
                            self.builder.build_int_compare(
                                IntPredicate::SGE,
                                lhs.into_int_value(),
                                rhs.into_int_value(),
                                "ge",
                            )?
                        } else {
                            self.builder.build_float_compare(
                                inkwell::FloatPredicate::OGE,
                                lhs.into_float_value(),
                                rhs.into_float_value(),
                                "oge",
                            )?
                        };
                        Ok(result.into())
                    }
                }
            }

            Expression::Call { name, args } => {
                // print 특별 처리
                if name == "print" {
                    let printf_fn = *self
                        .functions
                        .get("printf")
                        .ok_or_else(|| anyhow::anyhow!("printf not found"))?;

                    if args.is_empty() {
                        bail!("print() expects at least 1 argument");
                    }

                    // 모든 인자를 순서대로 출력
                    for (i, arg) in args.iter().enumerate() {
                        let arg_type = self.infer_expression_type(arg)?;

                        // 마지막 인자만 줄바꿈, 나머지는 공백
                        let is_last = i == args.len() - 1;

                        match arg_type {
                            Type::Int => {
                                let fmt = if is_last {
                                    self.builder.build_global_string_ptr("%d\n", "int_fmt_nl")?
                                } else {
                                    self.builder.build_global_string_ptr("%d ", "int_fmt_sp")?
                                };
                                let val = self.compile_expression(arg)?;
                                self.builder.build_call(
                                    printf_fn,
                                    &[fmt.as_pointer_value().into(), val.into()],
                                    "print_int",
                                )?;
                            }
                            Type::String => {
                                let fmt = if is_last {
                                    self.builder.build_global_string_ptr("%s\n", "str_fmt_nl")?
                                } else {
                                    self.builder.build_global_string_ptr("%s", "str_fmt")? // 공백 없음
                                };
                                let val = self.compile_expression(arg)?;
                                self.builder.build_call(
                                    printf_fn,
                                    &[fmt.as_pointer_value().into(), val.into()],
                                    "print_string",
                                )?;
                            }
                            Type::Bool => {
                                let val = self.compile_expression(arg)?;
                                let true_str = if is_last {
                                    self.builder.build_global_string_ptr("true\n", "true_nl")?
                                } else {
                                    self.builder.build_global_string_ptr("true ", "true_sp")?
                                };
                                let false_str = if is_last {
                                    self.builder
                                        .build_global_string_ptr("false\n", "false_nl")?
                                } else {
                                    self.builder.build_global_string_ptr("false ", "false_sp")?
                                };

                                let str_ptr = self.builder.build_select(
                                    val.into_int_value(),
                                    true_str.as_pointer_value(),
                                    false_str.as_pointer_value(),
                                    "bool_str",
                                )?;

                                self.builder.build_call(
                                    printf_fn,
                                    &[str_ptr.into()],
                                    "print_bool",
                                )?;
                            }
                            Type::Float => {
                                let fmt = if is_last {
                                    self.builder
                                        .build_global_string_ptr("%f\n", "float_fmt_nl")?
                                } else {
                                    self.builder
                                        .build_global_string_ptr("%f ", "float_fmt_sp")?
                                };
                                let val = self.compile_expression(arg)?;
                                self.builder.build_call(
                                    printf_fn,
                                    &[fmt.as_pointer_value().into(), val.into()],
                                    "print_float",
                                )?;
                            }
                        }
                    }

                    Ok(self.context.i32_type().const_int(0, false).into())
                } else {
                    // 일반 함수 호출 (기존 코드)
                    let function = *self
                        .functions
                        .get(name)
                        .ok_or_else(|| anyhow::anyhow!("Undefined function: {}", name))?;

                    let mut arg_values = Vec::new();
                    for arg in args {
                        arg_values.push(self.compile_expression(arg)?.into());
                    }

                    let call_site = self.builder.build_call(function, &arg_values, name)?;

                    match call_site.try_as_basic_value() {
                        inkwell::Either::Left(val) => Ok(val),
                        inkwell::Either::Right(_) => {
                            Ok(self.context.i32_type().const_int(0, false).into())
                        }
                    }
                }
            }
        }
    }

    // LLVM IR을 파일로 저장
    pub fn write_to_file(&self, filename: &str) -> Result<()> {
        self.module
            .print_to_file(filename)
            .map_err(|e| anyhow::anyhow!("Failed to write LLVM IR: {}", e.to_string()))
    }

    fn infer_expression_type(&self, expr: &Expression) -> Result<Type> {
        match expr {
            Expression::Number(_) => Ok(Type::Int),
            Expression::Float(_) => Ok(Type::Float),
            Expression::String(_) => Ok(Type::String),
            Expression::Bool(_) => Ok(Type::Bool),
            Expression::Identifier(name) => {
                let (_, ty, _) = self
                    .variables
                    .get(name)
                    .ok_or_else(|| anyhow::anyhow!("Unknown variable: {}", name))?;
                Ok(*ty)
            }
            Expression::Binary { left, op, .. } => {
                match op {
                    BinaryOp::Equal
                    | BinaryOp::NotEqual
                    | BinaryOp::LessThan
                    | BinaryOp::GreaterThan
                    | BinaryOp::LessThanEqual
                    | BinaryOp::GreaterThanEqual => Ok(Type::Bool),
                    _ => self.infer_expression_type(left), // 산술 연산은 왼쪽 타입 반환
                }
            }
            Expression::Call { .. } => {
                bail!("Cannot infer type of function call in codegen");
            }
        }
    }
}
