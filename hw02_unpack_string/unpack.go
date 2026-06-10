package main

import (
	"errors"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(input string) (string, error) {
	builder := strings.Builder{}
	runes := []rune(input)
	meta := make([]*runeMeta, len(runes))

	for i, r := range runes {
		meta[i] = newRuneMeta(r)
	}

	for i := 0; i < len(meta); i++ {
		current := meta[i]

		if current.Skipped {
			continue
		}

		if i < len(meta)-1 {
			next := meta[i+1]

			// Экранирование
			if current.IsBackslash {
				if next.IsDigit || next.IsBackslash {
					// Запишем следующий символ как есть
					builder.WriteRune(next.Rune)
					// И пометим его для пропуска, чтобы не обрабатывать как цифру/бэкслеш
					meta[i+1].Skipped = true
					continue
				}

				return "", ErrInvalidString
			}

			// Следующий 0 удаляет текущий символ
			if next.IsDigit && next.Repeat == 0 {
				continue
			}
		}

		if current.IsDigit {
			if i == 0 {
				return "", ErrInvalidString
			}

			prev := meta[i-1]

			if prev.IsDigit && !prev.Skipped {
				return "", ErrInvalidString
			}

			// Текущий 0 не пишется
			if current.Repeat == 0 {
				continue
			}

			builder.WriteString(strings.Repeat(string(prev.Rune), current.Repeat-1))
			continue
		}
		builder.WriteRune(current.Rune)
	}

	return builder.String(), nil
}
