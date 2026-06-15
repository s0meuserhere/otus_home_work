package hw02unpackstring

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "a4bc2d5e", expected: "aaaabccddddde"},
		{input: "abccd", expected: "abccd"},
		{input: "", expected: ""},
		{input: "aaa0b", expected: "aab"},
		{input: "🙃0", expected: ""},
		{input: "aaф0b", expected: "aab"},
		{input: `qwe\4\5`, expected: `qwe45`},
		{input: `qwe\45`, expected: `qwe44444`},
		{input: `qwe\\5`, expected: `qwe\\\\\`},
		{input: `qwe\\\3`, expected: `qwe\3`},
		{input: `a5b0c`, expected: `aaaaac`},
		{input: `a2b0c3`, expected: `aaccc`},
		{input: `a1`, expected: `a`},
		{input: `a9`, expected: `aaaaaaaaa`},
		{input: `x2y3z4`, expected: `xxyyyzzzz`},
		{input: `🙃3`, expected: `🙃🙃🙃`},
		{input: `ф2世2`, expected: `фф世世`},
		{input: `世2界3`, expected: `世世界界界`},
		{input: `\45`, expected: `44444`},
		{input: `\3`, expected: `3`},
		{input: `a`, expected: `a`},
		{input: `aa0`, expected: `a`},
		{input: `ab0cd0ef`, expected: `acef`},
		{input: `a8b9c`, expected: `aaaaaaaabbbbbbbbbc`},
		{input: `\9\8`, expected: `98`},
		{input: `a\1\2`, expected: `a12`},
		{input: `a\\b`, expected: `a\b`},
		{input: `\\2`, expected: `\\`},
		{input: `\\3`, expected: `\\\`},
		{input: `🙃\3`, expected: `🙃3`},
		{input: `\`, expected: `\`},
		{input: `abc\`, expected: `abc\`},
		{input: `a\`, expected: `a\`},
		{input: "d\n5abc", expected: "d\n\n\n\n\nabc"},
		{input: "abcd", expected: "abcd"},
		{input: `Привет, 0\0тус`, expected: `Привет,0тус`},
	}

	for _, tc := range tests {
		tc := tc //nolint:copyloopvar
		t.Run(tc.input, func(t *testing.T) {
			result, err := Unpack(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUnpackInvalidString(t *testing.T) {
	invalidStrings := []string{
		"3abc",
		"45",
		"aaa10b",
		`0`,
		`9`,
		`a12b`,
		`a99b`,
		`qw\ne`,
		`\n3`,
		`a\q`,
		`a\b`,
		`d\n5abc`,
		`10`,
		`1a`,
		`a10c`,
		`\a`,
		`\🙃`,
		`a\🙃`,
		`\\\a`,
		`qw\ne\`,
		`Привет, 0\Отус`,
	}
	for _, tc := range invalidStrings {
		tc := tc //nolint:copyloopvar
		t.Run(tc, func(t *testing.T) {
			_, err := Unpack(tc)
			require.Truef(t, errors.Is(err, ErrInvalidString), "actual error %q", err)
		})
	}
}
