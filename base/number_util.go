package base

import (
	"errors"
	"fmt"
	"strconv"
)

// NumberToBinaryString 数字转二进制字符串
func NumberToBinaryString(number int64, bitSize int) (string, error) {
	if bitSize > 64 {
		return "", fmt.Errorf("bit size overflow, the max size is 64")
	}
	mask := int64(1<<bitSize - 1)
	endpoint := number & mask
	bitString := strconv.FormatInt(endpoint, 2)
	if len(bitString) < bitSize {
		bitString = fmt.Sprintf("%0"+strconv.Itoa(bitSize)+"s", bitString)
	}
	return bitString, nil
}

// BinaryStringToInt64 converts a binary string to an int64 value
func BinaryStringToInt64(binaryString string, bitSize int) (int64, error) {
	// Check if the string is empty
	if len(binaryString) == 0 {
		return 0, errors.New("binary string cannot be empty")
	}
	// Parse the binary string with base 2
	result, err := strconv.ParseInt(binaryString, 2, bitSize)
	if err != nil {
		return 0, fmt.Errorf("failed to parse binary string: %w", err)
	}
	return result, nil
}
