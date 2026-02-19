package utils

import (
	"reflect"
	"testing"
)

func TestGetString(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{"hello", "hello"},
		{123, ""},
		{nil, ""},
	}
	for _, tt := range tests {
		if got := GetString(tt.input); got != tt.expected {
			t.Errorf("GetString(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		input    any
		expected int
	}{
		{int64(123), 123},
		{"123", 0},
		{nil, 0},
	}
	for _, tt := range tests {
		if got := GetInt(tt.input); got != tt.expected {
			t.Errorf("GetInt(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestGetInt64(t *testing.T) {
	tests := []struct {
		input    any
		expected int64
	}{
		{int64(123), 123},
		{123, 0},
		{nil, 0},
	}
	for _, tt := range tests {
		if got := GetInt64(tt.input); got != tt.expected {
			t.Errorf("GetInt64(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestStringValue(t *testing.T) {
	s := "hello"
	if got := StringValue(&s); got != "hello" {
		t.Errorf("StringValue(&s) = %v, want hello", got)
	}
	if got := StringValue(nil); got != "" {
		t.Errorf("StringValue(nil) = %v, want empty string", got)
	}
}

func TestInt64Value(t *testing.T) {
	var i int64 = 123
	if got := Int64Value(&i); got != 123 {
		t.Errorf("Int64Value(&i) = %v, want 123", got)
	}
	if got := Int64Value(nil); got != 0 {
		t.Errorf("Int64Value(nil) = %v, want 0", got)
	}
}

func TestParseSlice(t *testing.T) {
	tests := []struct {
		input    string
		expected Slice
	}{
		{"1:5:2", Slice{Start: new(1), Stop: new(5), Step: new(2)}},
		{"3", Slice{Start: new(3)}},
		{"3:4", Slice{Start: new(3), Stop: new(4)}},
		{":4", Slice{Stop: new(4)}},
	}

	for _, tt := range tests {
		got, err := ParseSlice(tt.input)
		if err != nil {
			t.Errorf("ParseSlice(%q) error: %v", tt.input, err)
			continue
		}
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("ParseSlice(%q) = %+v, want %+v", tt.input, got, tt.expected)
		}
	}
}

func TestDictFilterBool(t *testing.T) {
	data := map[string]any{
		"a": 1,
		"b": 0,
		"c": nil,
		"d": "",
		"e": false,
	}
	got := DictFilterBool(data)
	if len(got) != 1 || got["a"] != 1 {
		t.Errorf("DictFilterBool() = %v, want {a: 1}", got)
	}
}
