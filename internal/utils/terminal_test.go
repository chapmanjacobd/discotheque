package utils

import "testing"

func TestTruncateMiddle(t *testing.T) {
	if TruncateMiddle("hello world", 10) != "hell…orld" {
		t.Errorf("got %s", TruncateMiddle("hello world", 10))
	}
	if TruncateMiddle("hello", 10) != "hello" {
		t.Errorf("got %s", TruncateMiddle("hello", 10))
	}
}

func TestGetTerminalWidth(t *testing.T) {
	w := GetTerminalWidth()
	if w <= 0 {
		t.Errorf("expected positive width")
	}
}

func TestCommandExists(t *testing.T) {
	if !CommandExists("go") {
		t.Errorf("expected go to exist")
	}
	if CommandExists("non-existent-command-12345") {
		t.Errorf("expected non-existent command not to exist")
	}
}
