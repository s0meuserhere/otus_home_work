package hw03frequencyanalysis

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const top10 = 10

var punctuation = regexp.MustCompile(`^[^\p{L}\p{N}-]+|[^\p{L}\p{N}-]+$`)

func Top10(input string) []string {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	words := strings.Fields(input)
	frequency := make(map[string]int)

	for _, word := range words {
		if word == `-` {
			continue
		}
		w := strings.ToLower(word)
		w = punctuation.ReplaceAllString(w, "")
		frequency[w]++
	}

	prepared := make([]string, 0)
	for k := range frequency {
		prepared = append(prepared, k)
	}

	sort.Slice(prepared, func(i, j int) bool {
		if frequency[prepared[i]] == frequency[prepared[j]] {
			return prepared[i] < prepared[j]
		}
		return frequency[prepared[i]] > frequency[prepared[j]]
	})

	if len(prepared) >= top10 {
		return prepared[:top10]
	}

	return prepared
}
