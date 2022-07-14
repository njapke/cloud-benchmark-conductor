package main

import "testing"

func BenchmarkSplit(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		split("A,B,C,D", ',')
	}
}
