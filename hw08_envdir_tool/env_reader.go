package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}

	env := make(Environment, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.Contains(name, "=") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", name, err)
		}

		if info.Size() == 0 {
			env[name] = EnvValue{NeedRemove: true}

			continue
		}

		value, err := readFirstLine(dir, name)
		if err != nil {
			return nil, err
		}

		env[name] = EnvValue{Value: value}
	}

	return env, nil
}

func readFirstLine(dir, name string) (string, error) {
	f, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		return "", fmt.Errorf("open %s: %w", name, err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	line, err := reader.ReadBytes('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read %s: %w", name, err)
	}

	line = bytes.TrimSuffix(line, []byte{'\n'})
	line = bytes.ReplaceAll(line, []byte{0}, []byte{'\n'})
	value := strings.TrimRight(string(line), " \t")

	return value, nil
}
