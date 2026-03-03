package mdcode

import (
	"testing"
)

func BenchmarkWalk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Walk(testdoc, func(block *Block) error {
			return nil
		})
	}
}
