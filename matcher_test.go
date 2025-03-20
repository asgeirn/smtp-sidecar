package main

import (
	"fmt"
	"regexp"
	"testing"
)

func TestMatchAnyPattern(t *testing.T) {
	// Helper function to compile regex patterns
	compilePatterns := func(patterns []string) []*regexp.Regexp {
		result := make([]*regexp.Regexp, 0, len(patterns))
		for _, p := range patterns {
			re, err := regexp.Compile(p)
			if err != nil {
				t.Fatalf("Failed to compile pattern %s: %v", p, err)
			}
			result = append(result, re)
		}
		return result
	}

	tests := []struct {
		name     string
		input    string
		patterns []string
		want     bool
	}{
		{
			name:     "Empty patterns",
			input:    "test string",
			patterns: []string{},
			want:     true,
		},
		{
			name:     "Single matching pattern",
			input:    "test string",
			patterns: []string{"test.*"},
			want:     true,
		},
		{
			name:     "Single non-matching pattern",
			input:    "test string",
			patterns: []string{"foo.*"},
			want:     false,
		},
		{
			name:     "Multiple patterns with one match",
			input:    "test string",
			patterns: []string{"foo.*", "bar.*", "test.*"},
			want:     true,
		},
		{
			name:     "Multiple patterns with no match",
			input:    "test string",
			patterns: []string{"foo.*", "bar.*", "baz.*"},
			want:     false,
		},
		{
			name:     "Empty input string",
			input:    "",
			patterns: []string{".*"},
			want:     true,
		},
		{
			name:     "Empty input string with specific patterns",
			input:    "",
			patterns: []string{"foo", "bar"},
			want:     false,
		},
		{
			name:     "Case sensitivity",
			input:    "Test String",
			patterns: []string{"test string"},
			want:     false,
		},
		{
			name:     "Case insensitivity",
			input:    "Test String",
			patterns: []string{"(?i)test string"},
			want:     true,
		},
		{
			name:     "Word boundaries",
			input:    "testing string",
			patterns: []string{`\btest\b`},
			want:     false,
		},
		{
			name:     "Special characters",
			input:    "price: $100.50",
			patterns: []string{`\$\d+\.\d+`},
			want:     true,
		},
		{
			name:     "Match at beginning",
			input:    "test at beginning",
			patterns: []string{"^test"},
			want:     true,
		},
		{
			name:     "Match at end",
			input:    "end with test",
			patterns: []string{"test$"},
			want:     true,
		},
		{
			name:     "Complex patterns",
			input:    "user@example.com",
			patterns: []string{`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`},
			want:     true,
		},
		{
			name:  "Multiple complex patterns",
			input: "123-456-7890",
			patterns: []string{
				`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, // Email
				`^\d{3}-\d{3}-\d{4}$`,                              // Phone
				`^\d{5}(-\d{4})?$`,                                 // ZIP code
			},
			want: true,
		},
		{
			name:     "First pattern matches",
			input:    "abcdef",
			patterns: []string{"abc.*", "xyz.*"},
			want:     true,
		},
		{
			name:     "Last pattern matches",
			input:    "xyzdef",
			patterns: []string{"abc.*", "xyz.*"},
			want:     true,
		},
		{
			name:     "Unicode support",
			input:    "こんにちは世界",
			patterns: []string{"世界$"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := compilePatterns(tt.patterns)
			got := MatchAnyPattern(tt.input, patterns)
			if got != tt.want {
				t.Errorf("MatchAnyPattern(%q, %v) = %v, want %v",
					tt.input, tt.patterns, got, tt.want)
			}
		})
	}
}

// TestMatchAnyPatternPerformance tests the performance with a large number of patterns
func TestMatchAnyPatternPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create a large number of patterns
	var patternStrings []string
	for i := 0; i < 1000; i++ {
		patternStrings = append(patternStrings, fmt.Sprintf("pattern%d.*", i))
	}

	// Compile all patterns
	patterns := make([]*regexp.Regexp, 0, len(patternStrings))
	for _, p := range patternStrings {
		re, err := regexp.Compile(p)
		if err != nil {
			t.Fatalf("Failed to compile pattern %s: %v", p, err)
		}
		patterns = append(patterns, re)
	}

	// Test cases
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "Match first pattern",
			input: "pattern0-test",
			want:  true,
		},
		{
			name:  "Match last pattern",
			input: "pattern999-test",
			want:  true,
		},
		{
			name:  "Match middle pattern",
			input: "pattern500-test",
			want:  true,
		},
		{
			name:  "No match",
			input: "no-match-pattern",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchAnyPattern(tt.input, patterns)
			if got != tt.want {
				t.Errorf("MatchAnyPattern(%q, [1000 patterns]) = %v, want %v",
					tt.input, got, tt.want)
			}
		})
	}
}

// TestMatchAnyPatternWithNil tests behavior with nil patterns
func TestMatchAnyPatternWithNil(t *testing.T) {
	t.Run("Nil patterns slice", func(t *testing.T) {
		result := MatchAnyPattern("test", nil)
		if result != true {
			t.Errorf("Expected true for nil patterns, got %v", result)
		}
	})

	t.Run("Slice with nil pattern", func(t *testing.T) {
		patterns := []*regexp.Regexp{nil}

		// This should panic, so we recover and check
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic with nil pattern, but no panic occurred")
			}
		}()

		MatchAnyPattern("test", patterns)
	})
}

// TestMatchAnyPatternIntegration tests integration with ConvertToRegexPatterns
func TestMatchAnyPatternIntegration(t *testing.T) {
	tests := []struct {
		name          string
		patternString string
		input         string
		want          bool
		wantErr       bool
	}{
		{
			name:          "Basic integration",
			patternString: "foo.*,bar[0-9]+,baz(a|b)",
			input:         "foobar",
			want:          true,
			wantErr:       false,
		},
		{
			name:          "No match integration",
			patternString: "foo.*,bar[0-9]+,baz(a|b)",
			input:         "nope",
			want:          false,
			wantErr:       false,
		},
		{
			name:          "Invalid pattern",
			patternString: "foo.*,[invalid",
			input:         "foobar",
			want:          false,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns, err := ConvertToRegexPatterns(tt.patternString)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToRegexPatterns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			got := MatchAnyPattern(tt.input, patterns)
			if got != tt.want {
				t.Errorf("Integration test: MatchAnyPattern(%q, patterns from %q) = %v, want %v",
					tt.input, tt.patternString, got, tt.want)
			}
		})
	}
}
