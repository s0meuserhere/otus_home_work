package logger

import "testing"

func TestLogger(t *testing.T) {
	t.Parallel()

	log := New("debug", "local")
	if log == nil {
		t.Fatal("expected logger")
	}
	if log.Slog() == nil {
		t.Fatal("expected slog logger")
	}
}
