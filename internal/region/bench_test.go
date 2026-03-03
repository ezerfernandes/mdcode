package region_test

import (
	"testing"

	"github.com/ezerfernandes/mdcode/internal/region"
)

func BenchmarkOutline(b *testing.B) {
	for i := 0; i < b.N; i++ {
		region.Outline(testdoc)
	}
}

func BenchmarkRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		region.Read(testdoc, "nonempty")
	}
}

func BenchmarkReplace(b *testing.B) {
	replacement := []byte("replaced content\n")
	for i := 0; i < b.N; i++ {
		region.Replace(testdoc, "nonempty", replacement)
	}
}
