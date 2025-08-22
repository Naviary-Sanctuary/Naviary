mod ast;
mod codegen;
mod lexer;
mod parser;
mod typechecker;

use codegen::CodeGenerator;
use colored::*;
use inkwell::context::Context;
use lexer::Lexer;
use parser::Parser;
use std::fs;
use std::path::Path;
use std::process::Command;
use typechecker::TypeChecker;

fn main() {
    let input = r#"
        func add(x: int, y: int) -> int {
            return x + y;
        }
        
        func main() {
            let mut a = 10;
            a = a + 10;
            print(a);

            a = 10;
            let b = 20;
            let result = add(a, b);
            if result > 10 {
                printString("result is greater than 10");
                print(result);
            } else {
                printString("result is less than or equal to 10");
                print(result);
            }
        }
    "#;

    println!("{}", "=== Naviary Compiler v0.0.1 ===".blue().bold());
    println!("{}", "Input:".green());
    println!("{}", input);

    // generate Lexer
    println!("\n{}", "Step 1: Generating Lexer...".yellow());
    let lexer = Lexer::new(input);

    // Parsing with Lexer
    println!("{}", "Step 2: Parsing...".yellow());
    let mut parser = Parser::new(lexer);
    let program_ast = match parser.parse_program() {
        Ok(program) => {
            println!("{}", "✓ Parsing successful".green());
            println!("{:#?}", program);
            program
        }
        Err(e) => {
            println!("{} {}", "✗ Parsing failed:".red(), e);
            return;
        }
    };

    // Type Checking
    println!("{}", "Step 3: Type Checking...".yellow());
    let mut type_checker = TypeChecker::new();
    match type_checker.check_program(&program_ast) {
        Ok(_) => {
            println!("{}", "✓ Type check passed".green());
        }
        Err(e) => {
            println!("{} {}", "✗ Type check failed:".red(), e);
            return;
        }
    }

    // Code Generation
    println!("{}", "Step 4: Code Generation...".yellow());
    let context = Context::create();
    let mut codegen = CodeGenerator::new(&context, "naviary_module");

    match codegen.compile_program(&program_ast) {
        Ok(_) => {
            println!("{}", "✓ Code generation successful".green());
        }
        Err(e) => {
            println!("{} {}", "✗ Code generation failed:".red(), e);
            return;
        }
    }

    // LLVM IR 출력
    let _ir = codegen.get_ir_string();
    println!("\n{}", "=== Generated LLVM IR ===".cyan());

    // IR을 파일로 저장
    if let Err(e) = codegen.write_to_file("output.ll") {
        println!("{} {}", "✗ Failed to write LLVM IR:".red(), e);
        return;
    }
    println!("\n{}", "✓ LLVM IR saved to output.ll".green());

    // 실행 파일 생성
    println!("\n{}", "Step 5: Creating executable...".yellow());
    if let Err(e) = compile_and_run() {
        println!("{} {}", "✗ Compilation failed:".red(), e);
    }
}

fn compile_and_run() -> Result<(), String> {
    // runtime.c 파일이 없으면 생성
    if !Path::new("runtime.c").exists() {
        println!("{}", "Creating runtime.c...".yellow());
        fs::write(
            "runtime.c",
            r#"#include <stdio.h>

void print(int value) {
    printf("%d\n", value);
}"#,
        )
        .map_err(|e| format!("Failed to create runtime.c: {}", e))?;
    }

    // runtime.c 컴파일
    println!("Compiling runtime.c...");
    let output = Command::new("clang")
        .args(&["-c", "runtime.c", "-o", "runtime.o"])
        .output()
        .map_err(|e| format!("Failed to execute clang: {}", e))?;

    if !output.status.success() {
        return Err(format!(
            "Failed to compile runtime.c:\n{}",
            String::from_utf8_lossy(&output.stderr)
        ));
    }

    // LLVM IR을 오브젝트 파일로 컴파일
    println!("Compiling LLVM IR...");
    let output = Command::new("clang")
        .args(&["-c", "output.ll", "-o", "output.o"])
        .output()
        .map_err(|e| format!("Failed to compile LLVM IR: {}", e))?;

    if !output.status.success() {
        return Err(format!(
            "Failed to compile LLVM IR:\n{}",
            String::from_utf8_lossy(&output.stderr)
        ));
    }

    // 링킹
    println!("Linking...");
    let output = Command::new("clang")
        .args(&["output.o", "runtime.o", "-o", "program"])
        .output()
        .map_err(|e| format!("Failed to link: {}", e))?;

    if !output.status.success() {
        return Err(format!(
            "Failed to link:\n{}",
            String::from_utf8_lossy(&output.stderr)
        ));
    }

    println!("{}", "✓ Executable created: ./program".green());

    // 실행
    println!("\n{}", "=== Running Program ===".magenta().bold());
    let output = Command::new("./program")
        .output()
        .map_err(|e| format!("Failed to run program: {}", e))?;

    // stdout 출력
    if !output.stdout.is_empty() {
        print!("Output: {}", String::from_utf8_lossy(&output.stdout));
    }

    // stderr 출력 (있을 경우)
    if !output.stderr.is_empty() {
        print!("Debug: {}", String::from_utf8_lossy(&output.stderr));
    }

    // exit code 확인
    if output.status.success() {
        println!("{}", "✓ Program executed successfully".green());
    } else {
        if let Some(code) = output.status.code() {
            println!("{}", format!("Program exited with code: {}", code).yellow());
        }
    }

    Ok(())
}
