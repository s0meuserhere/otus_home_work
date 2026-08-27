package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseArgs(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		timeout, host, port, err := parseArgs([]string{"localhost", "4242"})
		require.NoError(t, err)
		require.Equal(t, 10*time.Second, timeout)
		require.Equal(t, "localhost", host)
		require.Equal(t, "4242", port)
	})

	t.Run("custom timeout", func(t *testing.T) {
		timeout, host, port, err := parseArgs([]string{"--timeout=3s", "1.1.1.1", "123"})
		require.NoError(t, err)
		require.Equal(t, 3*time.Second, timeout)
		require.Equal(t, "1.1.1.1", host)
		require.Equal(t, "123", port)
	})

	t.Run("missing args", func(t *testing.T) {
		_, _, _, err := parseArgs([]string{"localhost"})
		require.Error(t, err)
	})

	t.Run("invalid flag", func(t *testing.T) {
		_, _, _, err := parseArgs([]string{"--unknown", "localhost", "80"})
		require.Error(t, err)
	})
}
