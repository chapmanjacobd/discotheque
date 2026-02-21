package models

import (
	"log/slog"
	"testing"
)

func TestGlobalFlags_AfterApply(t *testing.T) {
	flags := GlobalFlags{
		FilterFlags: FilterFlags{
			Ext: []string{"mp4", ".mkv"},
		},
	}
	err := flags.AfterApply()
	if err != nil {
		t.Fatalf("AfterApply failed: %v", err)
	}
	if flags.Ext[0] != ".mp4" {
		t.Errorf("Expected .mp4, got %s", flags.Ext[0])
	}
	if flags.Ext[1] != ".mkv" {
		t.Errorf("Expected .mkv, got %s", flags.Ext[1])
	}
}

func TestSetupLogging(t *testing.T) {
	SetupLogging(true)
	if LogLevel.Level() != slog.LevelDebug {
		t.Errorf("Expected debug level, got %v", LogLevel.Level())
	}

	SetupLogging(false)
	if LogLevel.Level() != slog.LevelInfo {
		t.Errorf("Expected info level, got %v", LogLevel.Level())
	}
}
