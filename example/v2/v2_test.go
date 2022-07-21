package main

import (
	"fmt"
	"strings"
	"testing"
)

func split(input string, c byte) []string {
	res := make([]string, 0)
	for {
		pos := strings.IndexByte(input, c)
		if pos == -1 {
			return append(res, input)
		}
		res = append(res, input[:pos])
		input = input[pos+1:]
	}
}

func BenchmarkSplit(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		split("A,B,C,D", ',')
	}
}

func ValidateOrderNumber2(orderNumber string) error {
	checksums := make([]int32, 3)
	onRest := orderNumber
	for i := 0; i < 3; i++ {
		foundDash := strings.IndexByte(onRest, '-')
		if foundDash == -1 {
			return fmt.Errorf("invalid order number: %s", orderNumber)
		}
		for _, c := range onRest[:foundDash] {
			checksums[i] += c - '0'
		}
		checksums[i] %= 10
		onRest = onRest[foundDash+1:]
	}
	if len(onRest) != 3 {
		return fmt.Errorf("invalid checksum length: %s", onRest)
	}
	for i, c := range onRest {
		if checksums[i] != c-'0' {
			return fmt.Errorf("invalid checksum: %s", onRest)
		}
	}
	return nil
}

func BenchmarkValidateOrderNumber(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateOrderNumber2("122-321-111-563")
	}
}
