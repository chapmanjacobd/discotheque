package utils

import (
	"reflect"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/models"
)

func TestTextProcessor(t *testing.T) {
	lines := []string{
		"apple banana cherry",
		"banana apple apple",
		"cherry cherry cherry",
	}

	flags := models.GlobalFlags{
		WordSorts: []string{"alpha"},
		LineSorts: []string{"count"},
	}

	// All lines have 3 words, so count sorting won't change order if stable.
	// But let's check word sorting within each line if we were to expose it.
	// TextProcessor returns original lines sorted.
	
	got := TextProcessor(flags, lines)
	if !reflect.DeepEqual(got, lines) {
		t.Errorf("TextProcessor failed, got %v, want %v", got, lines)
	}
}

func TestFilterCorpus(t *testing.T) {
	corpusStats := map[string]int{
		"apple":  2,
		"banana": 1,
		"cherry": 3,
	}

	trueVal := true
	falseVal := false

	tests := []struct {
		words    []string
		unique   *bool
		dups     *bool
		expected bool
	}{
		{[]string{"banana"}, &trueVal, nil, true},
		{[]string{"apple"}, &trueVal, nil, false},
		{[]string{"apple"}, nil, &trueVal, true},
		{[]string{"banana"}, nil, &trueVal, false},
		{[]string{"apple", "banana"}, &trueVal, &falseVal, false}, // not all unique
		{[]string{"banana"}, &trueVal, &falseVal, true},           // all unique
	}

	for _, tt := range tests {
		if got := filterCorpus(corpusStats, tt.words, tt.unique, tt.dups); got != tt.expected {
			t.Errorf("filterCorpus(%v, %v, %v) = %v, want %v", tt.words, tt.unique, tt.dups, got, tt.expected)
		}
	}
}
