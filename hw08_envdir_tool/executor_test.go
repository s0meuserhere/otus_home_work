package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestRunCmd(t *testing.T) {
	t.Setenv("EXISTING", "old")
	t.Setenv("TO_REMOVE", "present")

	env := Environment{
		"EXISTING":  {Value: "new"},
		"TO_REMOVE": {NeedRemove: true},
		"EMPTY":     {Value: ""},
		"ADDED":     {Value: "added"},
	}

	code, out := runCmdCapture(t, []string{
		"/bin/sh", "-c",
		`printf '%s|%s|%s|%s' "$EXISTING" "${TO_REMOVE-unset}" "$EMPTY" "$ADDED"`,
	}, env)
	if code != 0 {
		t.Fatalf("return code = %d, want 0", code)
	}

	want := "new|unset||added"
	if out != want {
		t.Fatalf("stdout = %q, want %q", out, want)
	}
}

func TestRunCmdExitCode(t *testing.T) {
	code := RunCmd([]string{"/bin/sh", "-c", "exit 42"}, nil)
	if code != 42 {
		t.Fatalf("return code = %d, want 42", code)
	}
}

func TestRunCmdEmptyCommand(t *testing.T) {
	code := RunCmd(nil, nil)
	if code != envdirFailureCode {
		t.Fatalf("return code = %d, want %d", code, envdirFailureCode)
	}
}

func TestRunCmdNotFound(t *testing.T) {
	code := RunCmd([]string{"/path/to/missing/binary"}, nil)
	if code != envdirFailureCode {
		t.Fatalf("return code = %d, want %d", code, envdirFailureCode)
	}
}

func runCmdCapture(t *testing.T, cmd []string, env Environment) (int, string) {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	oldStdout := os.Stdout
	os.Stdout = w

	code := RunCmd(cmd, env)

	os.Stdout = oldStdout
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	_ = r.Close()

	return code, buf.String()
}
