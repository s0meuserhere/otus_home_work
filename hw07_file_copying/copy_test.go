package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {
	tests := []struct {
		name     string
		offset   int64
		limit    int64
		expected string
	}{
		{
			name:     "offset0_limit0",
			offset:   0,
			limit:    0,
			expected: "testdata/out_offset0_limit0.txt",
		},
		{
			name:     "offset0_limit10",
			offset:   0,
			limit:    10,
			expected: "testdata/out_offset0_limit10.txt",
		},
		{
			name:     "offset0_limit1000",
			offset:   0,
			limit:    1000,
			expected: "testdata/out_offset0_limit1000.txt",
		},
		{
			name:     "offset0_limit10000",
			offset:   0,
			limit:    10000,
			expected: "testdata/out_offset0_limit10000.txt",
		},
		{
			name:     "offset100_limit1000",
			offset:   100,
			limit:    1000,
			expected: "testdata/out_offset100_limit1000.txt",
		},
		{
			name:     "offset6000_limit1000",
			offset:   6000,
			limit:    1000,
			expected: "testdata/out_offset6000_limit1000.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toPath := filepath.Join(t.TempDir(), "out.txt")

			err := Copy("testdata/input.txt", toPath, tt.offset, tt.limit)
			if err != nil {
				t.Fatalf("Copy() error = %v", err)
			}

			got, err := os.ReadFile(toPath)
			if err != nil {
				t.Fatalf("read result: %v", err)
			}

			want, err := os.ReadFile(tt.expected)
			if err != nil {
				t.Fatalf("read expected: %v", err)
			}

			if !bytes.Equal(got, want) {
				t.Errorf("result mismatch: got %d bytes, want %d bytes", len(got), len(want))
			}
		})
	}
}

func TestCopyErrors(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "src.txt")
	dstPath := filepath.Join(tmpDir, "dst.txt")

	if err := os.WriteFile(srcPath, []byte("hello OTUS"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	tests := []struct {
		name    string
		from    string
		to      string
		offset  int64
		limit   int64
		wantErr error
	}{
		{
			name:    "empty from",
			from:    "",
			to:      dstPath,
			wantErr: ErrPathFromEmpty,
		},
		{
			name:    "empty to",
			from:    srcPath,
			to:      "",
			wantErr: ErrPathToEmpty,
		},
		{
			name:    "same path",
			from:    srcPath,
			to:      srcPath,
			wantErr: ErrPathFromToSameFile,
		},
		{
			name:    "offset exceeds file size",
			from:    srcPath,
			to:      dstPath,
			offset:  100,
			wantErr: ErrOffsetExceedsFileSize,
		},
		{
			name:    "unsupported directory",
			from:    tmpDir,
			to:      dstPath,
			wantErr: ErrUnsupportedFile,
		},
		{
			name:    "source not found",
			from:    filepath.Join(tmpDir, "missing.txt"),
			to:      dstPath,
			wantErr: ErrSourceFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Copy(tt.from, tt.to, tt.offset, tt.limit)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCopyLimitExceedsFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "src.txt")
	dstPath := filepath.Join(tmpDir, "dst.txt")
	content := []byte("hello OTUS")

	if err := os.WriteFile(srcPath, content, 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	err := Copy(srcPath, dstPath, 0, 1000)
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}

	if !bytes.Equal(got, content) {
		t.Errorf("got %q, want %q", got, content)
	}
}
