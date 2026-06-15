package hw02unpackstring

import (
	"strconv"
)

const backslash rune = 92

type runeMeta struct {
	Rune        rune
	IsDigit     bool
	Repeat      int
	IsBackslash bool
	Skipped     bool
}

func newRuneMeta(r rune) *runeMeta {
	meta := runeMeta{}
	meta.Rune = r
	meta.IsBackslash = r == backslash
	if value, err := strconv.Atoi(string(r)); err == nil {
		meta.Repeat = value
		meta.IsDigit = true
	}
	return &meta
}
