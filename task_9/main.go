package main

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
)

func unpack(input string) (string, error) {
	if input == "" {
		return "", errors.New("empty input")
	}

	runes := []rune(input)
	var result []rune

	var lastChar rune
	escaped := false
	hasNonDigit := false

	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		if escaped {
			result = append(result, ch)
			lastChar = ch
			escaped = false
			hasNonDigit = true
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if unicode.IsDigit(ch) {
			if lastChar == 0 {
				return "", errors.New("некорректная строка: цифра в начале или после другой цифры без символа")
			}

			num, err := strconv.Atoi(string(ch))
			if err != nil {
				return "", err
			}

			if num > 1 {
				for j := 1; j < num; j++ {
					result = append(result, lastChar)
				}
			}
			continue
		}

		result = append(result, ch)
		lastChar = ch
		hasNonDigit = true
	}

	if !hasNonDigit {
		return "", errors.New("некорректная строка: нет символов для распаковки, только цифры")
	}

	if escaped {
		return "", errors.New("некорректная строка: заканчивается на '\\', нет символа после экранирования")
	}

	return string(result), nil
}

func main() {
	tests := []string{
		"a4bc2d5e",
		"abcd",
		"45",
		"a\\3",
		"",
		`qwe\4\5`,
		`qwe\45`,
	}

	for _, test := range tests {
		res, err := unpack(test)
		if err != nil {
			fmt.Printf("Input: %q --> Error: %v\n", test, err)
		} else {
			fmt.Printf("Input: %q --> Output: %q\n", test, res)
		}
	}
}
