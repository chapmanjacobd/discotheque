package utils

import (
	"io"
	"os"
	"testing"

	"github.com/chapmanjacobd/discoteca/internal/models"
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

	// Set verbose logging (go test will only show output on test failure)
	models.SetupLogging(2) // Debug level

	os.Exit(m.Run())
}
