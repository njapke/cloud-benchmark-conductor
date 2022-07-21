package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func BenchmarkSplit(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strings.Split("A,B,C,D", ",")
	}
}

func ValidateOrderNumber(orderNumber string) error {
	parts := strings.Split(orderNumber, "-")
	if len(parts) != 4 {
		return fmt.Errorf("invalid order number: %s", orderNumber)
	}
	checksums := make([]int, 3)
	for i := 0; i < 3; i++ {
		sum := 0
		for _, c := range strings.Split(parts[i], "") {
			val, err := strconv.ParseInt(c, 10, 8)
			if err != nil {
				return err
			}
			sum += int(val)
		}
		checksums[i] = sum % 10
	}
	validChecksums := strings.Split(parts[3], "")
	if len(validChecksums) != 3 {
		return fmt.Errorf("invalid checksum: %s", parts[3])
	}
	for i, c := range validChecksums {
		val, err := strconv.ParseInt(c, 10, 8)
		if err != nil {
			return err
		}
		if checksums[i] != int(val) {
			return fmt.Errorf("invalid checksum: %s", parts[3])
		}
	}
	return nil
}

func BenchmarkValidateOrderNumber(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateOrderNumber("122-321-111-563")
	}
}
