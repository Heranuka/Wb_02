package main

import (
	"errors"
	"testing"

	"github.com/go-playground/assert/v2"
)

func Test_unpack(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		want        string
		expectError error
	}{
		{
			name:        "OK",
			input:       "a4bc2d5e",
			want:        "aaaabccddddde",
			expectError: nil,
		},
		{
			name:        "Without nums",
			input:       "abcd",
			want:        "abcd",
			expectError: nil,
		},
		{
			name:        "Digits only",
			input:       "45",
			want:        "",
			expectError: errors.New("некорректная строка: цифра в начале или после другой цифры без символа"),
		},
		{
			name:        "empty",
			input:       "",
			want:        "",
			expectError: errors.New("empty input"), // В вашем коде для пустой строки нет ошибки, возвращается пустая строка без ошибки
		},
		{
			name:        "with slashes",
			input:       `qwe\4\5`,
			want:        "qwe45",
			expectError: nil,
		},
		{
			name:        "something",
			input:       `qwe\45`,
			want:        "qwe44444",
			expectError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := unpack(testCase.input)

			if testCase.expectError == nil && err != nil {
				t.Fatalf("unexpected error for input %q: %v", testCase.input, err)
			}
			if testCase.expectError != nil && err == nil {
				t.Fatalf("expected error %v but got nil for input %q", testCase.expectError, testCase.input)
			}
			if testCase.expectError != nil && err != nil {
				// сравниваем тексты ошибок
				if err.Error() != testCase.expectError.Error() {
					t.Fatalf("expected error %v but got %v for input %q", testCase.expectError, err, testCase.input)
				}
			}

			assert.Equal(t, testCase.want, got)
		})
	}
}
