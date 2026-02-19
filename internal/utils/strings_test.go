package utils

import (
	"reflect"
	"testing"
)

func TestCompareBlockStrings(t *testing.T) {
	tests := []struct {
		pattern  string
		value    string
		expected bool
	}{
		{"abc%", "abcdef", true},
		{"%def", "abcdef", true},
		{"%bc%", "abcdef", true},
		{"a%f", "abcdef", true},
		{"missing", "abcdef", false},
	}
	for _, tt := range tests {
		if got := CompareBlockStrings(tt.pattern, tt.value); got != tt.expected {
			t.Errorf("CompareBlockStrings(%q, %q) = %v, want %v", tt.pattern, tt.value, got, tt.expected)
		}
	}
}

func TestMatchesAny(t *testing.T) {
	tests := []struct {
		path     string
		patterns []string
		expected bool
	}{
		{"/home/user/movie.mp4", []string{"%.mp4"}, true},
		{"/home/user/movie.mp4", []string{"%.mkv"}, false},
		{"/home/user/movie.mp4", []string{"%user%"}, true},
		{"/home/user/movie.mp4", []string{"%missing%"}, false},
	}
	for _, tt := range tests {
		if got := MatchesAny(tt.path, tt.patterns); got != tt.expected {
			t.Errorf("MatchesAny(%q, %v) = %v, want %v", tt.path, tt.patterns, got, tt.expected)
		}
	}
}

func TestNaturalLess(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected bool
	}{
		{"file1.txt", "file2.txt", true},
		{"file2.txt", "file1.txt", false},
		{"file1.txt", "file10.txt", true},
		{"file10.txt", "file2.txt", false},
		{"Season 1 Episode 1", "Season 1 Episode 10", true},
		{"S01E01", "S01E02", true},
		{"S01E02", "S01E01", false},
		{"S01E09", "S01E10", true},
	}

	for _, tt := range tests {
		result := NaturalLess(tt.s1, tt.s2)
		if result != tt.expected {
			t.Errorf("NaturalLess(%q, %q) = %v, want %v", tt.s1, tt.s2, result, tt.expected)
		}
	}
}

func TestExtractNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected []chunk
	}{
		{"abc123def", []chunk{{"abc", 0, false}, {"", 123, true}, {"def", 0, false}}},
		{"123", []chunk{{"", 123, true}}},
		{"abc", []chunk{{"abc", 0, false}}},
	}
	for _, tt := range tests {
		got := extractNumbers(tt.input)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("extractNumbers(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestCleanString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "hello_world"},
		{"hello.world", "hello.world"},
		{"hello (world)", "hello"},
		{"Hello & World!", "Hello World"},
	}
	for _, tt := range tests {
		if got := CleanString(tt.input); got != tt.expected {
			t.Errorf("CleanString(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestShorten(t *testing.T) {
	tests := []struct {
		input    string
		width    int
		expected string
	}{
		{"hello world", 5, "hell…"},
		{"hello world", 20, "hello world"},
		{"こんにちは", 6, "こん…"},
	}
	for _, tt := range tests {
		if got := Shorten(tt.input, tt.width); got != tt.expected {
			t.Errorf("Shorten(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.expected)
		}
	}
}

func TestShortenMiddle(t *testing.T) {
	tests := []struct {
		input    string
		width    int
		expected string
	}{
		{"hello world", 8, "hel...ld"},
		{"こんにちは世界", 8, "こ...界"},
	}
	for _, tt := range tests {
		if got := ShortenMiddle(tt.input, tt.width); got != tt.expected {
			t.Errorf("ShortenMiddle(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.expected)
		}
	}
}

func TestIsMimeMatch(t *testing.T) {
	tests := []struct {
		search   []string
		mime     string
		expected bool
	}{
		{[]string{"video"}, "video/mp4", true},
		{[]string{"mp4"}, "video/mp4", true},
		{[]string{"audio"}, "video/mp4", false},
		{[]string{"VIDEO"}, "video/mp4", true},
		{[]string{"video"}, "VIDEO/MP4", true},
		{[]string{"plain"}, "text/plain", true},
		{[]string{"text"}, "text/plain", true},
		{[]string{}, "text/plain", false},
		{[]string{"video"}, "", false},
	}
	for _, tt := range tests {
		if got := IsMimeMatch(tt.search, tt.mime); got != tt.expected {
			t.Errorf("IsMimeMatch(%v, %q) = %v, want %v", tt.search, tt.mime, got, tt.expected)
		}
	}
}

func TestStripEnclosingQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{"'hello'", "hello"},
		{"«hello»", "hello"},
		{`"'hello'"`, "hello"},
	}
	for _, tt := range tests {
		if got := StripEnclosingQuotes(tt.input); got != tt.expected {
			t.Errorf("StripEnclosingQuotes(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCombine(t *testing.T) {
	if got := Combine("a", "b"); got != "a;b" {
		t.Errorf("Combine(a, b) = %q, want a;b", got)
	}
	if got := Combine("a", []string{"b", "c"}); got != "a;b;c" {
		t.Errorf("Combine(a, [b, c]) = %q, want a;b;c", got)
	}
	if got := Combine("a,b", "c;d"); got != "a;b;c;d" {
		t.Errorf("Combine(a,b, c;d) = %q, want a;b;c;d", got)
	}
}

func TestFromTimestampSeconds(t *testing.T) {
	if got := FromTimestampSeconds(":30"); got != 30 {
		t.Errorf("FromTimestampSeconds(:30) = %v, want 30", got)
	}
	if got := FromTimestampSeconds("1:30"); got != 90 {
		t.Errorf("FromTimestampSeconds(1:30) = %v, want 90", got)
	}
}

func TestPartialStartswith(t *testing.T) {
	list := []string{"daily", "weekly", "monthly"}
	if got, _ := PartialStartswith("da", list); got != "daily" {
		t.Errorf("PartialStartswith(da) = %q, want daily", got)
	}
}

func TestGlobMatchAny(t *testing.T) {
	if !GlobMatchAny("test", []string{"*test*"}) {
		t.Error("GlobMatchAny failed")
	}
}

func TestGlobMatchAll(t *testing.T) {
	if !GlobMatchAll("test", []string{"*test*", "t*"}) {
		t.Error("GlobMatchAll failed")
	}
}

func TestDurationShort(t *testing.T) {
	if got := DurationShort(60); got != "1 minute" {
		t.Errorf("DurationShort(60) = %q, want 1 minute", got)
	}
}

func TestExtractWords(t *testing.T) {
	input := "UniqueTerm, AnotherTerm! 123"
	expected := []string{"uniqueterm", "anotherterm", "123"}
	got := ExtractWords(input)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ExtractWords() = %v, want %v", got, expected)
	}
}

func TestSafeJSONLoads(t *testing.T) {
	input := `{"a": 1}`
	got := SafeJSONLoads(input)
	expected := map[string]any{"a": float64(1)}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("SafeJSONLoads() = %v, want %v", got, expected)
	}
}

func TestLoadString(t *testing.T) {
	input := `{'a': 1}`
	got := LoadString(input)
	expected := map[string]any{"a": float64(1)}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("LoadString() = %v, want %v", got, expected)
	}
	
	if got := LoadString("just string"); got != "just string" {
		t.Errorf("LoadString() = %v, want just string", got)
	}
}
