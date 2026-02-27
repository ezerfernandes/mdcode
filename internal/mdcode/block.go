// Package mdcode extracts and manipulates fenced code blocks in Markdown documents.
package mdcode

// Block represents a single fenced code block parsed from a Markdown document.
type Block struct {
	Lang      string
	Meta      Meta
	Code      []byte
	StartLine int
	EndLine   int
}

// Blocks is a slice of code blocks extracted from a Markdown document.
type Blocks []*Block
