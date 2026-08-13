package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDir(t *testing.T) {
	env, err := ReadDir("testdata/env")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	tests := []struct {
		name       string
		wantValue  string
		wantRemove bool
	}{
		{name: "HELLO", wantValue: `"hello"`},
		{name: "BAR", wantValue: "bar"},
		{name: "FOO", wantValue: "   foo\nwith new line"},
		{name: "EMPTY", wantValue: ""},
		{name: "UNSET", wantRemove: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := env[tt.name]
			if !ok {
				t.Fatalf("variable %q not found", tt.name)
			}

			if got.NeedRemove != tt.wantRemove {
				t.Fatalf("NeedRemove = %v, want %v", got.NeedRemove, tt.wantRemove)
			}

			if got.Value != tt.wantValue {
				t.Fatalf("Value = %q, want %q", got.Value, tt.wantValue)
			}
		})
	}
}

func TestReadDirSkipEqualsInName(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "VALID"), []byte("ok\n"), 0o644); err != nil {
		t.Fatalf("write VALID: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "BAD=NAME"), []byte("skip\n"), 0o644); err != nil {
		t.Fatalf("write BAD=NAME: %v", err)
	}

	env, err := ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	if _, ok := env["BAD=NAME"]; ok {
		t.Fatal("variable with '=' in name must be skipped")
	}

	got, ok := env["VALID"]
	if !ok {
		t.Fatal("VALID not found")
	}

	if got.Value != "ok" || got.NeedRemove {
		t.Fatalf("VALID = %+v, want Value=ok NeedRemove=false", got)
	}
}

func TestReadDirNotFound(t *testing.T) {
	_, err := ReadDir("testdata/no_such_dir")
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}
