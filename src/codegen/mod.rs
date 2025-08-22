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
    variables: HashMap<String, PointerValue<'ctx>>,
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
        // print 함수 선언 (외부 C 함수로)
        // void print(int value)
        let i32_type = self.context.i32_type();
        let print_type = self
            .context
            .void_type()
            .fn_type(&[BasicMetadataTypeEnum::from(i32_type)], false);
        let print_fn = self.module.add_function("print", print_type, None);
        self.functions.insert("print".to_string(), print_fn);

        // printBool 함수 선언 (외부 C 함수로)
        let bool_type = self.context.bool_type();
        let print_bool_type = self
            .context
            .void_type()
            .fn_type(&[BasicMetadataTypeEnum::from(bool_type)], false);
        let print_bool_fn = self.module.add_function("printBool", print_bool_type, None);
        self.functions
            .insert("printBool".to_string(), print_bool_fn);

        // printString 함수 선언 (외부 C 함수로)
        let string_type = self.context.ptr_type(inkwell::AddressSpace::default());
        let print_string_type = self
            .context
            .void_type()
            .fn_type(&[BasicMetadataTypeEnum::from(string_type)], false);
        let print_string_fn = self
            .module
            .add_function("printString", print_string_type, None);
        self.functions
            .insert("printString".to_string(), print_string_fn);
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
            self.variables.insert(param.name.clone(), alloca);
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
            Statement::Let { name, ty, value } => {
                // 값 계산
                let val = self.compile_expression(value)?;

                // 변수를 위한 스택 공간 할당
                let var_type = ty.as_ref().unwrap_or(&Type::Int); // 타입 추론된 경우 기본값 (실제로는 type checker가 처리)
                let alloca = self.create_entry_block_alloca(name, var_type);

                // 값 저장
                self.builder.build_store(alloca, val)?;
                self.variables.insert(name.clone(), alloca);
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
                let condition_value = self.compile_expression(condition)?;

                let function = self.current_function.unwrap();

                // 블록 생성
                let then_bb = self.context.append_basic_block(function, "then");
                let else_bb = self.context.append_basic_block(function, "else");
                let merge_bb = self.context.append_basic_block(function, "merge");

                // 조건 분기
                self.builder.build_conditional_branch(
                    condition_value.into_int_value(),
                    then_bb,
                    if else_block.is_some() {
                        else_bb
                    } else {
                        merge_bb
                    },
                )?;

                // then 블록 컴파일
                self.builder.position_at_end(then_bb);
                self.compile_block(then_block)?;
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
                // 변수 값 로드
                let ptr = self
                    .variables
                    .get(name)
                    .ok_or_else(|| anyhow::anyhow!("Undefined variable: {}", name))?;
                let val = self
                    .builder
                    .build_load(self.context.i32_type(), *ptr, name)?;
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
                // 함수 핸들을 복사하여 `self`의 불변 대여를 즉시 종료
                let function = *self
                    .functions
                    .get(name)
                    .ok_or_else(|| anyhow::anyhow!("Undefined function: {}", name))?;

                // 인자 컴파일
                let mut arg_values = Vec::new();
                for arg in args {
                    arg_values.push(self.compile_expression(arg)?.into());
                }

                // 함수 호출
                let call_site = self.builder.build_call(function, &arg_values, name)?;

                // 반환값이 있으면 반환, 없으면 임의 값 반환
                // inkwell의 Either 타입 사용
                match call_site.try_as_basic_value() {
                    inkwell::Either::Left(val) => Ok(val),
                    inkwell::Either::Right(_) => {
                        // void 함수 - 임의 값 반환 (사용되지 않을 것)
                        Ok(self.context.i32_type().const_int(0, false).into())
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

    // LLVM IR을 문자열로 반환
    pub fn get_ir_string(&self) -> String {
        self.module.print_to_string().to_string()
    }
}
