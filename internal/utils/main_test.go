package utils

import (
	"io"
	"log/slog"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Silence Stdout and Stderr during tests
	origStdout := Stdout
	origStderr := Stderr
	Stdout = io.Discard
	Stderr = io.Discard
	defer func() {
		Stdout = origStdout
		Stderr = origStderr
	}()

	// Silence slog during tests
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	os.Exit(m.Run())
}
