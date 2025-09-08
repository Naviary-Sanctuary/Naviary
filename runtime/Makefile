# OS
UNAME := $(shell uname)

ifeq ($(UNAME), Darwin)
    # macOS
    CXX = clang++
    AS = as
    LDFLAGS = -lSystem
else
    # Linux
    CXX = g++
    AS = as
    LDFLAGS = -lc
endif

# Compiler
CXXFLAGS = -std=c++2b ## c++2b -> c++23
CXXFLAGS += -O2 # optimization level 2
CXXFLAGS += -fPIC # Position Independent Code
CXXFLAGS += -Wall -Wextra  # activate all warnings
CXXFLAGS += -fno-exceptions # C++ exceptions are disabled
CXXFLAGS += -fno-rtti # deactivate RTTI 

# Assembler
ASFLAGS = #default falgs

# Linker
LD = ld #GNU linker

# targets
OBJECTS = io.o
LIBRARY = libnavi_runtime.a

# target: dependencies
all: $(LIBRARY)

# ar: archiving tool (pack objects into one file)
# rcs: r(replace) c(create) s(create index)
# $^: all dependencies (io.o)
$(LIBRARY): $(OBJECTS)
	ar rcs $@ $^
	@echo "Runtime library built: $(LIBRARY)"

# $<: first dependency (io.cpp)
# $@: target (io.o)
io.o: io.cpp runtime.h
	$(CXX) $(CXXFLAGS) -c $< -o $@
	@echo "Compiled: $<"

test: test_program
	./test_program
	@echo "Test Passed"

test_program: test.o $(LIBRARY)
ifeq ($(UNAME), Darwin)
	$(CXX) test.o -L. -lnavi_runtime -o test_program
else
	$(CXX) test.o -L. -lnavi_runtime -o test_program
endif

test.o: test.s
	$(AS) $(ASFLAGS) test.s -o test.o

clean:
	rm -f $(OBJECTS) $(LIBRARY) test_program test.o
	@echo "Cleand build files"

.PHONY: all clean test
