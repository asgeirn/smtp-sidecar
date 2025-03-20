package main

import (
	"regexp"
)

// MatchAnyPattern checks if the input string matches any of the provided regex patterns.
// Returns true if any pattern matches, false otherwise.
func MatchAnyPattern(input string, patterns []*regexp.Regexp) bool {
	// If there are no patterns, return true
	if len(patterns) == 0 {
		return true
	}

	// Check each pattern against the input string
	for _, pattern := range patterns {
		if pattern.MatchString(input) {
			return true // Return true on the first match
		}
	}

	// No patterns matched
	return false
}
