package mdcode

type Block struct {
	Lang      string
	Meta      Meta
	Code      []byte
	StartLine int
	EndLine   int
}

type Blocks []*Block
