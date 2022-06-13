package main

import (
	"fmt"
	"strconv"
	"strings"
)

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

func main() {
	err := ValidateOrderNumber2("122-321-111-563")
	if err != nil {
		fmt.Println(err)
	}
}
