package mdcode

// Unfence parses a Markdown document and returns all fenced code blocks
// without modifying the source.
func Unfence(source []byte) (Blocks, error) {
	var blocks Blocks

	_, _, err := Walk(source, func(block *Block) error {
		blocks = append(blocks, block)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return blocks, nil
}
