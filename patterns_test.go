package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestConvertToRegexPatterns(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string // Expected regex pattern strings
		wantErr bool
	}{
		{
			name:    "Basic patterns",
			input:   "foo.*,bar[0-9]+,baz(a|b)",
			want:    []string{"foo.*", "bar[0-9]+", "baz(a|b)"},
			wantErr: false,
		},
		{
			name:    "Patterns with whitespace",
			input:   " abc , def , ghi ",
			want:    []string{"abc", "def", "ghi"},
			wantErr: false,
		},
		{
			name:    "Empty input",
			input:   "",
			want:    []string{},
			wantErr: false,
		},
		{
			name:    "Single pattern",
			input:   "single-pattern",
			want:    []string{"single-pattern"},
			wantErr: false,
		},
		{
			name:    "Empty patterns",
			input:   ",,",
			want:    []string{},
			wantErr: false,
		},
		{
			name:    "Mixed empty and non-empty patterns",
			input:   "pattern1,,pattern2,",
			want:    []string{"pattern1", "pattern2"},
			wantErr: false,
		},
		{
			name:    "Invalid regex pattern",
			input:   "valid-pattern,[invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Special regex characters",
			input:   `\d+,\w+,\s+`,
			want:    []string{`\d+`, `\w+`, `\s+`},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertToRegexPatterns(tt.input)

			// Check error status
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToRegexPatterns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expect an error, no need to check the result
			if tt.wantErr {
				return
			}

			// Convert []*regexp.Regexp to []string for comparison
			gotPatterns := make([]string, len(got))
			for i, pattern := range got {
				gotPatterns[i] = pattern.String()
			}

			// Compare results
			if !reflect.DeepEqual(gotPatterns, tt.want) {
				t.Errorf("ConvertToRegexPatterns() = %v, want %v", gotPatterns, tt.want)
			}

			// Test that the patterns actually compile and match as expected
			if len(got) > 0 {
				// Test a simple match for the first pattern
				firstPattern := got[0]
				testString := ""

				// Create a test string that should match the pattern
				switch tt.want[0] {
				case `\d+`:
					testString = "123"
				case `\w+`:
					testString = "abc"
				case `\s+`:
					testString = "   "
				case "foo.*":
					testString = "foobar"
				default:
					testString = tt.want[0]
				}

				if !firstPattern.MatchString(testString) {
					t.Errorf("Pattern %s failed to match test string %s", firstPattern, testString)
				}
			}
		})
	}
}

// TestConvertToRegexPatternsEdgeCases tests additional edge cases
func TestConvertToRegexPatternsEdgeCases(t *testing.T) {
	// Test with a very large number of patterns
	t.Run("Large number of patterns", func(t *testing.T) {
		// Create a string with 100 patterns
		var patterns []string
		for i := 0; i < 100; i++ {
			patterns = append(patterns, "pattern"+string(rune('a'+i%26)))
		}
		input := strings.Join(patterns, ",")

		result, err := ConvertToRegexPatterns(input)
		if err != nil {
			t.Errorf("Failed with large number of patterns: %v", err)
		}

		if len(result) != 100 {
			t.Errorf("Expected 100 patterns, got %d", len(result))
		}
	})
}
