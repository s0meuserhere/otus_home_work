package hw10programoptimization

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type User struct {
	Email string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	s, err := getStat(r, domain)
	if err != nil {
		return nil, fmt.Errorf("get stat error: %w", err)
	}

	return s, nil
}

func getStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	reader := bufio.NewReader(r)

	for {
		line, err := reader.ReadSlice('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("read line: %w", err)
		}

		line = bytes.TrimSpace(line)
		if len(line) > 0 {
			var u User
			if unmarshalErr := json.Unmarshal(line, &u); unmarshalErr != nil {
				return nil, fmt.Errorf("unmarshal user: %w", unmarshalErr)
			}

			if strings.Contains(u.Email, "."+domain) {
				parts := strings.SplitN(u.Email, "@", 2)
				if len(parts) == 2 {
					result[strings.ToLower(parts[1])]++
				}
			}
		}

		if errors.Is(err, io.EOF) {
			break
		}
	}

	return result, nil
}
