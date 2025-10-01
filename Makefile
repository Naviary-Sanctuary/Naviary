.PHONY: all build compile run clean

ROOT_DIR := $(CURDIR)
COMPILER_DIR := $(ROOT_DIR)/compiler
BINARY_NAME := naviary
BINARY_PATH := $(COMPILER_DIR)/$(BINARY_NAME)
DEFAULT_SRC := $(ROOT_DIR)/examples/main.navi
RUNTIME_SRC := $(ROOT_DIR)/runtime/print.zig

all: build

build:
	cd $(COMPILER_DIR) && go build -o $(BINARY_PATH)

compile: build
	@ARGS="$(filter-out compile,$(MAKECMDGOALS))"; \
	if [ -z "$$ARGS" ]; then \
		echo "No source provided. Using default: $(DEFAULT_SRC)"; \
		$(BINARY_PATH) $(DEFAULT_SRC); \
	else \
		$(BINARY_PATH) $$ARGS; \
	fi

clean:
	rm -f $(BINARY_PATH)

# Swallow extra goals like file paths so they can be forwarded to recipes without errors
%:
	@:

