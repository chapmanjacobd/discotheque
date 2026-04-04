package utils_test

import (
	"strings"
	"testing"

	"github.com/chapmanjacobd/discoteca/internal/utils"
)

func TestGetMpvListenSocket(t *testing.T) {
	got := utils.GetMpvListenSocket()
	if got == "" {
		t.Errorf("GetMpvListenSocket returned empty string")
	}
	if !strings.Contains(got, "mpv_socket") {
		t.Errorf("GetMpvListenSocket mismatch: %s", got)
	}
}

func TestGetMpvWatchSocket(t *testing.T) {
	got := utils.GetMpvWatchSocket()
	if got == "" {
		t.Errorf("GetMpvWatchSocket returned empty string")
	}
}

func TestGetMpvWatchLaterDir(t *testing.T) {
	got := utils.GetMpvWatchLaterDir()
	if got == "" {
		t.Errorf("GetMpvWatchLaterDir returned empty string")
	}
}

func TestGetDirs(t *testing.T) {
	tests := []struct {
		name string
		fn   func() string
	}{
		{"GetTempDir", utils.GetTempDir},
		{"GetCattNowPlayingFile", utils.GetCattNowPlayingFile},
		{"GetConfigDir", utils.GetConfigDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(); got == "" {
				t.Errorf("%s returned empty string", tt.name)
			}
		})
	}
}
