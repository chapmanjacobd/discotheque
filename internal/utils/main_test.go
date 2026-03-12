package utils

import (
	"io"
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

	os.Exit(m.Run())
}
