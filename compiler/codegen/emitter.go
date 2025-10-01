package codegen

import (
	"bytes"
	"fmt"
)

type Emitter struct {
	buffer             bytes.Buffer
	indentLevel        int
	indentString       string
	isCurrentLineEmpty bool
}

func NewEmitter() *Emitter {
	return &Emitter{
		indentString:       "    ", // 4 spaces
		indentLevel:        0,      // start with no indentation
		isCurrentLineEmpty: true,   // start with empty line
	}
}

// Add text to buffer without newline
func (emitter *Emitter) Emit(text string, args ...any) {
	if emitter.isCurrentLineEmpty && emitter.indentLevel > 0 {
		for i := 0; i < emitter.indentLevel; i++ {
			emitter.buffer.WriteString(emitter.indentString)
		}
	}

	if len(args) > 0 {
		emitter.buffer.WriteString(fmt.Sprintf(text, args...))
	} else {
		emitter.buffer.WriteString(text)
	}

	emitter.isCurrentLineEmpty = false
}

func (emitter *Emitter) EmitNewLine() {
	emitter.buffer.WriteString("\n")

	emitter.isCurrentLineEmpty = true
}

func (emitter *Emitter) EmitLine(text string, args ...any) {
	emitter.Emit(text, args...)
	emitter.EmitNewLine()
}

func (emitter *Emitter) IncreaseIndent() {
	emitter.indentLevel++
}

func (emitter *Emitter) DecreaseIndent() {
	emitter.indentLevel--
}

func (emitter *Emitter) GetOutput() string {
	return emitter.buffer.String()
}
