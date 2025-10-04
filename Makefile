# Compiler settings
ZIG := zig
GO := go
CC := clang

# Directories  
COMPILER_DIR := compiler
RUNTIME_DIR := runtime
BUILD_DIR := build
COMPILER_BUILD_DIR := $(BUILD_DIR)/compiler
RUNTIME_BUILD_DIR := $(BUILD_DIR)/runtime
BIN_DIR := $(BUILD_DIR)/bin

# Files
COMPILER_MAIN := $(COMPILER_DIR)/main.go
COMPILER_BIN := $(COMPILER_BUILD_DIR)/compiler
RUNTIME_LIB := $(RUNTIME_BUILD_DIR)/libnaviary_runtime.a
RUNTIME_SRC := $(RUNTIME_DIR)/src/lib.zig

# Create build directories
$(COMPILER_BUILD_DIR):
	@mkdir -p $(COMPILER_BUILD_DIR)

$(RUNTIME_BUILD_DIR):
	@mkdir -p $(RUNTIME_BUILD_DIR)

$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

# Build compiler executable
$(COMPILER_BIN): $(shell find $(COMPILER_DIR) -name '*.go') | $(COMPILER_BUILD_DIR)
	@echo "Building compiler..."
	@cd $(COMPILER_DIR) && $(GO) build -o ../$(COMPILER_BIN) .
	@echo "Compiler built: $(COMPILER_BIN)"

# Build runtime library
$(RUNTIME_LIB): $(RUNTIME_SRC) | $(RUNTIME_BUILD_DIR)
	@echo "Building runtime library..."
	@cd $(RUNTIME_DIR) && $(ZIG) build-lib src/lib.zig \
		-femit-bin=../$(RUNTIME_LIB) \
		-O ReleaseFast
	@echo "Runtime library built: $(RUNTIME_LIB)"

# Compiler target
.PHONY: compiler
compiler: $(COMPILER_BIN)

# Runtime target
.PHONY: runtime
runtime: $(RUNTIME_LIB)

# Build any .navi file to C
.PHONY: build
build: $(COMPILER_BIN)
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		echo "Usage: make build <file.navi>"; \
		exit 1; \
	fi
	@FILE=$(filter-out $@,$(MAKECMDGOALS)); \
	echo "Compiling $$FILE to LLVM IR..."; \
	$(COMPILER_BIN) $$FILE
	@echo "LLVM IR generated successfully"

# Run any .navi file (build C, compile to executable, and run)
.PHONY: run
run: $(COMPILER_BIN) $(RUNTIME_LIB) | $(BIN_DIR)
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		echo "Usage: make run <file.navi>"; \
		exit 1; \
	fi
	@FILE=$(filter-out $@,$(MAKECMDGOALS)); \
	BASENAME=$$(basename $$FILE .navi); \
	echo "Compiling $$FILE to C..."; \
	$(COMPILER_BIN) $$FILE; \
	C_FILE=$$(dirname $$FILE)/$$BASENAME.c; \
	echo "Building executable..."; \
	$(CC) $$C_FILE $(RUNTIME_LIB) -o $(BIN_DIR)/$$BASENAME; \
	echo "Running $$BASENAME..."; \
	echo "-------------------"; \
	$(BIN_DIR)/$$BASENAME

# Prevent make from interpreting .navi files as targets
%:
	@:

# Clean
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	find . -name "*.c" -not -path "./runtime/*" -delete