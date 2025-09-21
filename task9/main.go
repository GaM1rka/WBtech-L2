package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func unpack(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	var result strings.Builder
	var prev rune
	escaped := false

	for i, char := range s {
		// Если символ "\" - берем следующий
		if char == '\\' && !escaped {
			escaped = true
			continue
		}

		if escaped {
			result.WriteRune(char)
			prev = char
			escaped = false
			continue
		}

		// Если символ число - повторяем предыдущий символ столько раз
		if unicode.IsDigit(char) {
			if i == 0 {
				return "", errors.New("string starts with digit")
			}
			if prev == 0 {
				return "", errors.New("invalid string format")
			}

			numStr := string(char)
			j := i + 1
			for j < len(s) && unicode.IsDigit(rune(s[j])) {
				numStr += string(s[j])
				j++
			}

			count, err := strconv.Atoi(numStr)
			if err != nil {
				return "", fmt.Errorf("invalid number: %v", err)
			}

			result.WriteString(strings.Repeat(string(prev), count-1))
			prev = 0
			continue
		}

		// Обычный символ
		result.WriteRune(char)
		prev = char
	}

	return result.String(), nil
}

func main() {
	// Входные данные из условия
	testCases := []string{
		"a4bc2d5e",
		"abcd",
		"45",
		"",
		"qwe\\4\\5",
		"qwe\\45",
	}
	// Проходимся по данным из условия
	for _, test := range testCases {
		fmt.Printf("Input: %q\n", test)
		result, err := unpack(test)
		if err != nil {
			fmt.Printf("Error: %v\n\n", err)
		} else {
			fmt.Printf("Output: %q\n\n", result)
		}
	}
}
