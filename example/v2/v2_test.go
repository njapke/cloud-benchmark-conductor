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

func parseUint(s string) int {
	ret := 0
	pos := 1
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] > '9' || s[i] < '0' {
			return -1
		}
		ret += int(s[i]-'0') * pos
		pos *= 10
	}
	return ret
}

func TestParseUint(t *testing.T) {
	if parseUint("999") != 999 {
		t.Fatal("parseUint failed")
	}
}

func BenchmarkParseUint(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseUint("999")
	}
}
