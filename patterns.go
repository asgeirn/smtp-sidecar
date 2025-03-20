package main

import (
	"regexp"
	"strings"
)

// ConvertToRegexPatterns takes a comma-separated string and returns a slice
// of compiled regex patterns.
func ConvertToRegexPatterns(input string) ([]*regexp.Regexp, error) {
	// Split the input string by commas
	patterns := strings.Split(input, ",")

	// Create a slice to hold the compiled regex patterns
	result := make([]*regexp.Regexp, 0, len(patterns))

	// Compile each pattern and add it to the result slice
	for _, pattern := range patterns {
		// Trim whitespace from the pattern
		trimmedPattern := strings.TrimSpace(pattern)

		// Skip empty patterns
		if trimmedPattern == "" {
			continue
		}

		// Compile the regex pattern
		regex, err := regexp.Compile(trimmedPattern)
		if err != nil {
			return nil, err
		}

		result = append(result, regex)
	}

	return result, nil
}
