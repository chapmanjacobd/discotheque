package utils

import (
	"strings"
	"testing"
)

func TestGetMpvListenSocket(t *testing.T) {
	got := GetMpvListenSocket()
	if got == "" {
		t.Error("GetMpvListenSocket returned empty string")
	}
	if !strings.Contains(got, "mpv_socket") {
		t.Errorf("GetMpvListenSocket mismatch: %s", got)
	}
}

func TestGetMpvWatchSocket(t *testing.T) {
	got := GetMpvWatchSocket()
	if got == "" {
		t.Error("GetMpvWatchSocket returned empty string")
	}
}

func TestGetMpvWatchLaterDir(t *testing.T) {
	got := GetMpvWatchLaterDir()
	if got == "" {
		t.Error("GetMpvWatchLaterDir returned empty string")
	}
}

func TestGetDirs(t *testing.T) {
	if GetTempDir() == "" {
		t.Error("GetTempDir empty")
	}
	if GetCattNowPlayingFile() == "" {
		t.Error("GetCattNowPlayingFile empty")
	}
	if GetConfigDir() == "" {
		t.Error("GetConfigDir empty")
	}
}
