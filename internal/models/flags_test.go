package models

import (
	"log/slog"
	"testing"
)

func TestGlobalFlags_AfterApply(t *testing.T) {
	flags := GlobalFlags{
		CoreFlags: CoreFlags{
			Simulate:  true,
			NoConfirm: true,
		},
		MediaFilterFlags: MediaFilterFlags{
			Ext: []string{"mp4", ".mkv"},
		},
		MergeFlags: MergeFlags{
			Ignore: true,
		},
	}
	err := flags.AfterApply()
	if err != nil {
		t.Fatalf("AfterApply failed: %v", err)
	}
	if flags.MediaFilterFlags.Ext[0] != ".mp4" {
		t.Errorf("Expected .mp4, got %s", flags.MediaFilterFlags.Ext[0])
	}
	if flags.MediaFilterFlags.Ext[1] != ".mkv" {
		t.Errorf("Expected .mkv, got %s", flags.MediaFilterFlags.Ext[1])
	}
	if !flags.CoreFlags.DryRun {
		t.Error("Expected DryRun to be true when Simulate is true")
	}
	if !flags.CoreFlags.Yes {
		t.Error("Expected Yes to be true when NoConfirm is true")
	}
	if !flags.MergeFlags.OnlyNewRows {
		t.Error("Expected OnlyNewRows to be true when Ignore is true")
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
