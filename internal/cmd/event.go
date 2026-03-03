package cmd

import (
	"fmt"
	"io"
)

// Event kind constants using dot-notation taxonomy.
const (
	OpStart      = "op.start"
	FileProcess  = "file.process"
	BlockHeader  = "block.header"
	BlockCommand = "block.command"
	BlockDone    = "block.done"
	BatchHeader  = "batch.header"
	WarnExit     = "warn.exit"
	WarnIO       = "warn.io"
)

// emitter classifies and emits structured status events.
type emitter interface {
	Emit(kind string, format string, args ...any)
}

// textEmitter writes human-readable formatted text to w.
type textEmitter struct {
	w io.Writer
}

func (e *textEmitter) Emit(_ string, format string, args ...any) {
	fmt.Fprintf(e.w, format, args...)
}

// nopEmitter discards all events.
type nopEmitter struct{}

func (e *nopEmitter) Emit(_ string, _ string, _ ...any) {}
