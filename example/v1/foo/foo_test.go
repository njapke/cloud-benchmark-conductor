package foo

import (
	"strings"
	"testing"
)

func BenchmarkStringsCut(b *testing.B) {
	veryLongString := strings.Repeat("A", 10000) + "-B"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strings.Cut(veryLongString, "-")
	}
}
