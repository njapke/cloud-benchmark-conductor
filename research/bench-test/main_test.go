package main

import "testing"

func BenchmarkValidateOrderNumber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ValidateOrderNumber("122-321-111-563")
	}
}

func BenchmarkValidateOrderNumber2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ValidateOrderNumber2("122-321-111-563")
	}
}
