package main

import (
	"testing"
)

func TestReverse(t *testing.T) {
	type testCase struct {
		name  string
		input string
		want  string
	}

	tests := []testCase{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "simple string",
			input: "Hello, World!",
			want:  "!dlroW ,olleH",
		},
		{
			name:  "1 whitespace string",
			input: " ",
			want:  " ",
		},
		{
			name:  "3 whitespace string",
			input: "   ",
			want:  "   ",
		},
		{
			name:  "palindrome string v1",
			input: "12345qwertyytrewq54321",
			want:  "12345qwertyytrewq54321",
		},
		{
			name:  "palindrome string v2",
			input: "12345qwerty1ytrewq54321",
			want:  "12345qwerty1ytrewq54321",
		},
		{
			name:  "digits only string",
			input: "12345",
			want:  "54321",
		},
		{
			name:  "special characters string",
			input: "!@#`~",
			want:  "~`#@!",
		},
		{
			name:  "newline and tab string",
			input: "a\n\tb",
			want:  "b\t\na",
		},
		{
			name:  "string with emoji",
			input: "Hello, OTUS! 👋",
			want:  "👋 !SUTO ,olleH",
		},
		{
			name:  "cyrillic",
			input: "Привет, ОТУС!",
			want:  "!СУТО ,тевирП",
		},
		{
			name:  "cyrillic with emoji",
			input: "👋 Привет, ОТУС!",
			want:  "!СУТО ,тевирП 👋",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Reverse(tc.input)
			if got != tc.want {
				t.Fatalf("got string (%v); want string (%v)", got, tc.want)
			}
			t.Logf("Reverse(%q) = %q", tc.input, got)
		})
	}
}
