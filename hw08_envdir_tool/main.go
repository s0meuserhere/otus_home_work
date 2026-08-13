package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: %s /path/to/env/dir command [args...]\n", filepath.Base(os.Args[0]))
		os.Exit(envdirFailureCode)
	}

	dir := os.Args[1]
	cmd := os.Args[2:]

	env, err := ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read env dir: %v\n", err)
		os.Exit(envdirFailureCode)
	}

	os.Exit(RunCmd(cmd, env))
}
