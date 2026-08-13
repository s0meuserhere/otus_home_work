package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

const (
	// envdirFailureCode возвращается при ошибке самой утилиты.
	// См. https://www.unix.com/man-page/debian/8/envdir/
	envdirFailureCode = 111
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return envdirFailureCode
	}

	command := exec.Command(cmd[0], cmd[1:]...) //nolint:gosec
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Env = prepareEnv(os.Environ(), env)

	err := command.Run()
	if err == nil {
		return 0
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}

	return envdirFailureCode
}

func prepareEnv(current []string, env Environment) []string {
	envMap := make(map[string]string, len(current))

	for _, item := range current {
		key, value, ok := strings.Cut(item, "=")
		if !ok {
			continue
		}

		envMap[key] = value
	}

	for key, envValue := range env {
		if envValue.NeedRemove {
			delete(envMap, key)

			continue
		}

		envMap[key] = envValue.Value
	}

	result := make([]string, 0, len(envMap))
	for key, value := range envMap {
		result = append(result, key+"="+value)
	}

	return result
}
