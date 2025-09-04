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

        let mut code_generator = CodeGenerator {
            context,
            module,
            builder,
            variables: HashMap::new(),
            functions: HashMap::new(),
            current_function: None,
        };

        code_generator.declare_external_functions();

        code_generator
    }

    fn get_native_int_type(&self) -> inkwell::types::IntType<'ctx> {
        #[cfg(target_pointer_width = "64")]
        return self.context.i64_type();

        #[cfg(target_pointer_width = "32")]
        return self.context.i32_type();
    }

    fn get_size_type(&self) -> inkwell::types::IntType<'ctx> {
        // 포인터와 같은 크기 (usize에 해당)
        #[cfg(target_pointer_width = "64")]
        return self.context.i64_type();

        #[cfg(target_pointer_width = "32")]
        return self.context.i32_type();
    }

    fn declare_external_functions(&mut self) {
        self.declare_c_standard_functions();
        self.declare_runtime_functions();
    }

    // 내장 함수 선언
    fn declare_c_standard_functions(&mut self) {
        let i32_type = self.context.i32_type();
        let i8_ptr_type = self.context.ptr_type(inkwell::AddressSpace::default());

        // printf - C 표준 출력 함수
        let printf_type = i32_type.fn_type(
            &[i8_ptr_type.into()],
            true, // variadic (가변 인자)
        );
        let printf_fn = self.module.add_function("printf", printf_type, None);
        self.functions.insert("printf".to_string(), printf_fn);
    }

    fn declare_runtime_functions(&mut self) {
        self.declare_runtime_memory_functions();
        self.declare_runtime_array_functions();
    }
    fn declare_runtime_memory_functions(&mut self) {
        let i8_ptr_type = self.context.ptr_type(inkwell::AddressSpace::default());
        let size_type = self.get_size_type(); // 플랫폼 의존적 (i32 또는 i64)

        // GC 초기화 - 인자 없음
        let gc_init_type = i8_ptr_type.fn_type(&[], false);
        self.module
            .add_function("naviary_gc_init", gc_init_type, None);

        // GC 실행 - GC 포인터만 받음
        let gc_collect_type = self
            .context
            .void_type()
            .fn_type(&[i8_ptr_type.into()], false);
        self.module
            .add_function("naviary_gc_collect", gc_collect_type, None);

        // 메모리 할당 - naviary_alloc(gc: *mut GC, size: usize) -> *mut u8
        let alloc_type = i8_ptr_type.fn_type(
            &[i8_ptr_type.into(), size_type.into()], // size는 플랫폼 의존적
            false,
        );
        self.module.add_function("naviary_alloc", alloc_type, None);

        // 루트 추가/제거 (GC 루트 관리)
        let add_root_type = self
            .context
            .void_type()
            .fn_type(&[i8_ptr_type.into(), i8_ptr_type.into()], false);
        self.module
            .add_function("naviary_gc_add_root", add_root_type, None);
        self.module
            .add_function("naviary_gc_remove_root", add_root_type, None);
    }

    fn declare_runtime_array_functions(&mut self) {
        let size_type = self.get_size_type();
        let native_int_type = self.get_native_int_type();
        let float_type = self.context.f64_type();
        let bool_type = self.context.bool_type();
        let i8_ptr_type = self.context.ptr_type(inkwell::AddressSpace::default());

        // Int 배열 함수들
        self.declare_array_functions_for_type("int", native_int_type.into(), size_type);

        // Float 배열 함수들
        self.declare_array_functions_for_type("float", float_type.into(), size_type);

        // Bool 배열 함수들
        self.declare_array_functions_for_type("bool", bool_type.into(), size_type);

        // String 배열 함수들
        self.declare_array_functions_for_type("string", i8_ptr_type.into(), size_type);
    }

    fn declare_array_functions_for_type(
        &mut self,
        type_name: &str,
        element_type: BasicTypeEnum<'ctx>,
        size_type: inkwell::types::IntType<'ctx>,
    ) {
        let i8_ptr_type = self.context.ptr_type(inkwell::AddressSpace::default());

        // naviary_allocate_{type}_array(gc: *GC, capacity: size_t) -> *Array
        let alloc_fn_name = format!("naviary_allocate_{}_array", type_name);
        let alloc_type = i8_ptr_type.fn_type(&[i8_ptr_type.into(), size_type.into()], false);
        self.module.add_function(&alloc_fn_name, alloc_type, None);

        // naviary_array_get_{type}(array: *Array, index: size_t) -> element
        let get_fn_name = format!("naviary_array_get_{}", type_name);
        let get_type = match element_type {
            BasicTypeEnum::IntType(t) => t.fn_type(&[i8_ptr_type.into(), size_type.into()], false),
            BasicTypeEnum::FloatType(t) => {
                t.fn_type(&[i8_ptr_type.into(), size_type.into()], false)
            }
            BasicTypeEnum::PointerType(t) => {
                t.fn_type(&[i8_ptr_type.into(), size_type.into()], false)
            }
            _ => panic!("Unsupported array element type"),
        };
        self.module.add_function(&get_fn_name, get_type, None);

        // naviary_array_set_{type}(array: *Array, index: size_t, value: element)
        let set_fn_name = format!("naviary_array_set_{}", type_name);
        let set_type = self.context.void_type().fn_type(
            &[i8_ptr_type.into(), size_type.into(), element_type.into()],
            false,
        );
        self.module.add_function(&set_fn_name, set_type, None);

        // naviary_array_len_{type}(array: *Array) -> size_t
        let len_fn_name = format!("naviary_array_len_{}", type_name);
        let len_type = size_type.fn_type(&[i8_ptr_type.into()], false);
        self.module.add_function(&len_fn_name, len_type, None);
    }

    // AST 타입을 LLVM 타입으로 변환
    fn get_llvm_type(&self, ty: &Type) -> BasicTypeEnum<'ctx> {
        match ty {
            Type::Int => self.get_native_int_type().into(),
            Type::Float => self.context.f64_type().into(),
            Type::Bool => self.context.bool_type().into(),
            Type::String => self
                .context
                .ptr_type(inkwell::AddressSpace::default())
                .into(),
            Type::IntArray | Type::FloatArray | Type::StringArray | Type::BoolArray => self
                .context
                .ptr_type(inkwell::AddressSpace::default())
                .into(),
        }
    }

    // 프로그램 전체 컴파일
    pub fn compile_program(&mut self, program: &Program) -> Result<()> {
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
                let var_type = if let Some(declared_type) = ty {
                    declared_type.clone()
                } else {
                    match value {
                        Expression::Array { elements } => {
                            if elements.is_empty() {
                                bail!("Cannot infer type of empty array without type annotation");
                            }
                            let elem_type = self.infer_expression_type(&elements[0])?;
                            match elem_type {
                                Type::Int => Type::IntArray,
                                Type::Float => Type::FloatArray,
                                Type::String => Type::StringArray,
                                Type::Bool => Type::BoolArray,
                                _ => bail!("Unsupported array element type: {:?}", elem_type),
                            }
                        }
                        _ => self.infer_expression_type(value)?,
                    }
                };

                let val = match value {
                    Expression::Array { elements } if elements.is_empty() => {
                        // 빈 배열인 경우 타입에 따라 처리
                        match &var_type {
                            Type::IntArray => self.compile_int_array(&[])?,
                            Type::FloatArray => self.compile_float_array(&[])?,
                            Type::StringArray => self.compile_string_array(&[])?,
                            Type::BoolArray => self.compile_bool_array(&[])?,
                            _ => {
                                bail!("Type mismatch: expected array type for empty array literal")
                            }
                        }
                    }
                    _ => self.compile_expression(value)?,
                };

                let alloca = self.create_entry_block_alloca(name, &var_type);
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
                let val = self.get_native_int_type().const_int(*n as u64, false);
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

            Expression::Array { elements } => {
                if elements.is_empty() {
                    bail!("Cannot compile empty array without type context");
                }

                let element_type = self.infer_expression_type(&elements[0])?;

                match element_type {
                    Type::Int => self.compile_int_array(elements),
                    Type::Float => self.compile_float_array(elements),
                    Type::String => self.compile_string_array(elements),
                    Type::Bool => self.compile_bool_array(elements),
                    _ => bail!("Unsupported array element type: {:?}", element_type),
                }
            }

            Expression::Index { object, index } => {
                // 인덱스 값 컴파일
                let index_value = self.compile_expression(index)?;

                // 인덱스를 size_type으로 변환
                let size_type = self.get_size_type();
                let index_converted = if index_value.is_int_value() {
                    let int_val = index_value.into_int_value();
                    if int_val.get_type().get_bit_width() < size_type.get_bit_width() {
                        self.builder
                            .build_int_s_extend(int_val, size_type, "index_extended")?
                            .into()
                    } else {
                        index_value
                    }
                } else {
                    index_value
                };

                let array_type = self.infer_expression_type(object)?;
                let array_ptr = self.compile_expression(object)?;

                match array_type {
                    Type::IntArray => {
                        let get_fn = self
                            .module
                            .get_function("naviary_array_get_int")
                            .ok_or_else(|| anyhow::anyhow!("Runtime function not found"))?;

                        let result = self.builder.build_call(
                            get_fn,
                            &[array_ptr.into(), index_converted.into()], // 변환된 인덱스
                            "array_get_int",
                        )?;

                        Ok(result
                            .try_as_basic_value()
                            .left()
                            .ok_or_else(|| anyhow::anyhow!("Expected return value"))?)
                    }
                    Type::FloatArray => {
                        let get_fn = self
                            .module
                            .get_function("naviary_array_get_float")
                            .ok_or_else(|| anyhow::anyhow!("Runtime function not found"))?;

                        let result = self.builder.build_call(
                            get_fn,
                            &[array_ptr.into(), index_converted.into()], // 변환된 인덱스
                            "array_get_float",
                        )?;

                        Ok(result
                            .try_as_basic_value()
                            .left()
                            .ok_or_else(|| anyhow::anyhow!("Expected return value"))?)
                    }
                    Type::BoolArray => {
                        let get_fn = self
                            .module
                            .get_function("naviary_array_get_bool")
                            .ok_or_else(|| anyhow::anyhow!("Runtime function not found"))?;

                        let result = self.builder.build_call(
                            get_fn,
                            &[array_ptr.into(), index_converted.into()], // 변환된 인덱스
                            "array_get_bool",
                        )?;

                        Ok(result
                            .try_as_basic_value()
                            .left()
                            .ok_or_else(|| anyhow::anyhow!("Expected return value"))?)
                    }
                    Type::StringArray => {
                        let get_fn = self
                            .module
                            .get_function("naviary_array_get_string")
                            .ok_or_else(|| anyhow::anyhow!("Runtime function not found"))?;

                        let result = self.builder.build_call(
                            get_fn,
                            &[array_ptr.into(), index_converted.into()], // 변환된 인덱스
                            "array_get_string",
                        )?;

                        Ok(result
                            .try_as_basic_value()
                            .left()
                            .ok_or_else(|| anyhow::anyhow!("Expected return value"))?)
                    }
                    _ => bail!("Cannot index non-array type: {:?}", array_type),
                }
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
            Expression::Call { name, args } if name == "print" => {
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
                                #[cfg(target_pointer_width = "64")]
                                let fmt_str = self
                                    .builder
                                    .build_global_string_ptr("%lld\n", "int_fmt_nl")?;
                                #[cfg(target_pointer_width = "32")]
                                let fmt_str =
                                    self.builder.build_global_string_ptr("%d\n", "int_fmt_nl")?;
                                fmt_str
                            } else {
                                #[cfg(target_pointer_width = "64")]
                                let fmt_str = self
                                    .builder
                                    .build_global_string_ptr("%lld ", "int_fmt_sp")?;
                                #[cfg(target_pointer_width = "32")]
                                let fmt_str =
                                    self.builder.build_global_string_ptr("%d ", "int_fmt_sp")?;
                                fmt_str
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

                            self.builder
                                .build_call(printf_fn, &[str_ptr.into()], "print_bool")?;
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
                        Type::IntArray | Type::FloatArray | Type::StringArray | Type::BoolArray => {
                            // 배열은 주소를 출력하거나 특별한 처리 필요
                            let fmt = if is_last {
                                self.builder
                                    .build_global_string_ptr("[array@%p]\n", "array_fmt_nl")?
                            } else {
                                self.builder
                                    .build_global_string_ptr("[array@%p] ", "array_fmt_sp")?
                            };
                            let val = self.compile_expression(arg)?;
                            self.builder.build_call(
                                printf_fn,
                                &[fmt.as_pointer_value().into(), val.into()],
                                "print_array_addr",
                            )?;
                        }
                    }
                }

                Ok(self.context.i32_type().const_int(0, false).into())
            }

            Expression::Call { name, args } => {
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

    fn compile_int_array(&mut self, elements: &[Expression]) -> Result<BasicValueEnum<'ctx>> {
        let capacity = elements.len();
        let size_type = self.get_size_type();

        // TODO: 실제로는 전역 GC 인스턴스 사용
        let gc_ptr = self
            .context
            .ptr_type(inkwell::AddressSpace::default())
            .const_null();

        let alloc_fn = self
            .module
            .get_function("naviary_allocate_int_array")
            .ok_or_else(|| anyhow::anyhow!("Runtime function not found"))?;

        let array_ptr = self.builder.build_call(
            alloc_fn,
            &[
                gc_ptr.into(),
                size_type.const_int(capacity as u64, false).into(),
            ],
            "new_int_array",
        )?;

        let array_ptr = array_ptr
            .try_as_basic_value()
            .left()
            .ok_or_else(|| anyhow::anyhow!("Expected return value"))?;

        // 각 요소 설정
        let set_fn = self
            .module
            .get_function("naviary_array_set_int")
            .ok_or_else(|| anyhow::anyhow!("Runtime function not found"))?;

        for (index, element) in elements.iter().enumerate() {
            let value = self.compile_expression(element)?;

            // value도 native_int_type으로 변환 필요
            let native_int_type = self.get_native_int_type();
            let value_converted = if value.is_int_value() {
                let int_val = value.into_int_value();
                if int_val.get_type().get_bit_width() < native_int_type.get_bit_width() {
                    self.builder
                        .build_int_s_extend(int_val, native_int_type, "value_extended")?
                        .into()
                } else {
                    value
                }
            } else {
                value
            };

            let index_val = size_type.const_int(index as u64, false);

            self.builder.build_call(
                set_fn,
                &[array_ptr.into(), index_val.into(), value_converted.into()],
                "array_set",
            )?;
        }

        Ok(array_ptr)
    }

    fn compile_float_array(&mut self, elements: &[Expression]) -> Result<BasicValueEnum<'ctx>> {
        let capacity = elements.len();
        let size_type = self.get_size_type();

        let gc_ptr = self
            .context
            .ptr_type(inkwell::AddressSpace::default())
            .const_null();

        let alloc_fn = self
            .module
            .get_function("naviary_allocate_float_array")
            .ok_or_else(|| anyhow::anyhow!("Runtime function not found"))?;

        let array_ptr = self.builder.build_call(
            alloc_fn,
            &[
                gc_ptr.into(),
                size_type.const_int(capacity as u64, false).into(),
            ],
            "new_float_array",
        )?;

        let array_ptr = array_ptr
            .try_as_basic_value()
            .left()
            .ok_or_else(|| anyhow::anyhow!("Expected return value"))?;

        let set_fn = self
            .module
            .get_function("naviary_array_set_float")
            .ok_or_else(|| anyhow::anyhow!("Runtime function not found"))?;

        for (index, element) in elements.iter().enumerate() {
            let value = self.compile_expression(element)?;
            let index_val = size_type.const_int(index as u64, false);

            self.builder.build_call(
                set_fn,
                &[array_ptr.into(), index_val.into(), value.into()],
                "array_set",
            )?;
        }

        Ok(array_ptr)
    }

    fn compile_bool_array(&mut self, elements: &[Expression]) -> Result<BasicValueEnum<'ctx>> {
        let capacity = elements.len();
        let size_type = self.get_size_type();

        // TODO: 실제로는 전역 GC 인스턴스 사용
        let gc_ptr = self
            .context
            .ptr_type(inkwell::AddressSpace::default())
            .const_null();

        // Bool 배열 할당
        let alloc_fn = self
            .module
            .get_function("naviary_allocate_bool_array")
            .ok_or_else(|| anyhow::anyhow!("Runtime function not found"))?;

        let array_ptr = self.builder.build_call(
            alloc_fn,
            &[
                gc_ptr.into(),
                size_type.const_int(capacity as u64, false).into(),
            ],
            "new_bool_array",
        )?;

        let array_ptr = array_ptr
            .try_as_basic_value()
            .left()
            .ok_or_else(|| anyhow::anyhow!("Expected return value"))?;

        // 각 요소 설정
        let set_fn = self
            .module
            .get_function("naviary_array_set_bool")
            .ok_or_else(|| anyhow::anyhow!("Runtime function not found"))?;

        for (index, element) in elements.iter().enumerate() {
            let value = self.compile_expression(element)?;
            let index_val = size_type.const_int(index as u64, false);

            self.builder.build_call(
                set_fn,
                &[array_ptr.into(), index_val.into(), value.into()],
                "array_set",
            )?;
        }

        Ok(array_ptr)
    }

    fn compile_string_array(&mut self, elements: &[Expression]) -> Result<BasicValueEnum<'ctx>> {
        let capacity = elements.len();
        let size_type = self.get_size_type();

        let gc_ptr = self
            .context
            .ptr_type(inkwell::AddressSpace::default())
            .const_null();

        // String 배열 할당
        let alloc_fn = self
            .module
            .get_function("naviary_allocate_string_array")
            .ok_or_else(|| anyhow::anyhow!("Runtime function not found"))?;

        let array_ptr = self.builder.build_call(
            alloc_fn,
            &[
                gc_ptr.into(),
                size_type.const_int(capacity as u64, false).into(),
            ],
            "new_string_array",
        )?;

        let array_ptr = array_ptr
            .try_as_basic_value()
            .left()
            .ok_or_else(|| anyhow::anyhow!("Expected return value"))?;

        // 각 문자열 요소 설정
        let set_fn = self
            .module
            .get_function("naviary_array_set_string")
            .ok_or_else(|| anyhow::anyhow!("Runtime function not found"))?;

        for (index, element) in elements.iter().enumerate() {
            // String expression을 컴파일하면 StringObject 포인터가 됨
            let value = self.compile_expression(element)?;
            let index_val = size_type.const_int(index as u64, false);

            self.builder.build_call(
                set_fn,
                &[array_ptr.into(), index_val.into(), value.into()],
                "array_set",
            )?;
        }

        Ok(array_ptr)
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

            Expression::Array { elements } => {
                if elements.is_empty() {
                    bail!("Cannot infer type of empty array");
                }
                let elem_type = self.infer_expression_type(&elements[0])?;
                match elem_type {
                    Type::Int => Ok(Type::IntArray),
                    Type::Float => Ok(Type::FloatArray),
                    Type::String => Ok(Type::StringArray),
                    Type::Bool => Ok(Type::BoolArray),
                    _ => bail!("Unsupported array element type"),
                }
            }

            Expression::Index { object, .. } => {
                let object_type = self.infer_expression_type(object)?;
                match object_type {
                    Type::IntArray => Ok(Type::Int),
                    Type::FloatArray => Ok(Type::Float),
                    Type::StringArray => Ok(Type::String),
                    Type::BoolArray => Ok(Type::Bool),
                    _ => bail!("Cannot index non-array type: {:?}", object_type),
                }
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
