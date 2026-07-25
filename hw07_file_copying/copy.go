package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cheggaaa/pb/v3" //nolint:depguard
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
	ErrPathFromEmpty         = errors.New("path from is empty")
	ErrPathToEmpty           = errors.New("path to is empty")
	ErrPathFromToSameFile    = errors.New("path from is same file")
	ErrSourceFileNotFound    = errors.New("source file not found")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	if len(fromPath) == 0 {
		return ErrPathFromEmpty
	}
	if len(toPath) == 0 {
		return ErrPathToEmpty
	}
	if fromPath == toPath {
		return ErrPathFromToSameFile
	}
	fStat, err := os.Stat(fromPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrSourceFileNotFound
		}
		return fmt.Errorf("stat %s: %w", fromPath, err)
	}
	if !fStat.Mode().IsRegular() {
		return ErrUnsupportedFile
	}
	if offset > fStat.Size() {
		return ErrOffsetExceedsFileSize
	}

	fTo, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", toPath, err)
	}
	defer func() {
		if err := fTo.Close(); err != nil {
			log.Printf("error closing 'TO' file: %v", err)
			return
		}
	}()

	fFrom, err := os.Open(fromPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", fromPath, err)
	}
	defer func() {
		if err := fFrom.Close(); err != nil {
			log.Printf("error closing 'FROM' file: %v", err)
			return
		}
	}()

	total := fStat.Size() - offset
	if limit > 0 && limit < total {
		total = limit
	}

	bar := pb.Full.Start64(total)
	defer bar.Finish()

	_, err = fFrom.Seek(offset, io.SeekStart)
	if err != nil {
		return fmt.Errorf("seek %s: %w", fromPath, err)
	}

	_, err = io.CopyN(fTo, bar.NewProxyReader(fFrom), total)
	if err != nil {
		return fmt.Errorf("copy %s: %w", toPath, err)
	}

	return nil
}
