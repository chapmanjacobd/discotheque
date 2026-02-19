package utils

import (
	"testing"
)

func TestCleanPath(t *testing.T) {
	tests := []struct {
		input    string
		opts     CleanPathOptions
		expected string
	}{
		{"example.txt", CleanPathOptions{}, "example.txt"},
		{"/folder/file.txt", CleanPathOptions{}, "/folder/file.txt"},
		{"/ -folder- / -file-.txt", CleanPathOptions{}, "/folder/file.txt"},
		{"/MyFolder/File.txt", CleanPathOptions{LowercaseFolders: true}, "/myfolder/File.txt"},
		{"/my folder/File.txt", CleanPathOptions{CaseInsensitive: true}, "/My Folder/File.txt"},
		{"/my folder/file.txt", CleanPathOptions{DotSpace: true}, "/my.folder/file.txt"},
		{"3_seconds_ago.../Mike.webm", CleanPathOptions{}, "3_seconds_ago/Mike.webm"},
		{"test/''/t", CleanPathOptions{}, "test/_/t"},
	}

	for _, tt := range tests {
		got := CleanPath(tt.input, tt.opts)
		if got != tt.expected {
			t.Errorf("CleanPath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestTrimPathSegments(t *testing.T) {
	tests := []struct {
		path          string
		desiredLength int
		expected      string
	}{
		{"/aaaaaaaaaa/fans/001.jpg", 16, "/a/fans/001.jpg"},
		{"/ao/bo/co/do/eo/fo/go/ho", 13, "/a/b/.../g/ho"},
		{"/a/b/c", 10, "/a/b/c"},
	}

	for _, tt := range tests {
		got := TrimPathSegments(tt.path, tt.desiredLength)
		if got != tt.expected {
			t.Errorf("TrimPathSegments(%q, %d) = %q, want %q", tt.path, tt.desiredLength, got, tt.expected)
		}
	}
}

func TestRelativize(t *testing.T) {
	if got := Relativize("/home/user/file"); got != "home/user/file" {
		t.Errorf("Relativize(/home/user/file) = %q, want home/user/file", got)
	}
}

func TestSafeJoin(t *testing.T) {
	base := "/path/to/fakeroot"
	tests := []struct {
		userPath string
		expected string
	}{
		{"etc/passwd", "/path/to/fakeroot/etc/passwd"},
		{"../../etc/passwd", "/path/to/fakeroot/etc/passwd"},
		{"/absolute/path", "/path/to/fakeroot/absolute/path"},
	}

	for _, tt := range tests {
		got := SafeJoin(base, tt.userPath)
		if got != tt.expected {
			t.Errorf("SafeJoin(%q, %q) = %q, want %q", base, tt.userPath, got, tt.expected)
		}
	}
}

func TestPathTupleFromURL(t *testing.T) {
	tests := []struct {
		url              string
		expectedParent   string
		expectedFilename string
	}{
		{"http://example.com/path/to/file.txt", "example.com/path/to", "file.txt"},
		{"https://www.example.org/another/file.jpg", "www.example.org/another", "file.jpg"},
		{"http://example.com/", "example.com", ""},
	}

	for _, tt := range tests {
		gotParent, gotFilename := PathTupleFromURL(tt.url)
		if gotParent != tt.expectedParent || gotFilename != tt.expectedFilename {
			t.Errorf("PathTupleFromURL(%q) = (%q, %q), want (%q, %q)", tt.url, gotParent, gotFilename, tt.expectedParent, tt.expectedFilename)
		}
	}
}
