package slw

import (
	"strings"
	"unicode"
)

// Config holds SLW (Semantic Line Wrap) configuration
type Config struct {
	// Enabled controls whether SLW is active
	Enabled bool
	// Markers are the characters that mark sentence endings (default: ".!?")
	Markers string
	// MinLineLength is the minimum line length before wrapping (default: 40)
	// Set to 0 for aggressive mode (always wrap after sentences)
	MinLineLength int
	// MaxLineWidth is the maximum line width for wrapping (default: 88)
	// Set to 0 to disable max width wrapping
	MaxLineWidth int
	// Abbreviations is a set of abbreviations that shouldn't trigger wrapping
	Abbreviations map[string]bool
}

// DefaultConfig returns the default SLW configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:       true,
		Markers:       ".!?",
		MinLineLength: 40,
		MaxLineWidth:  88,
		Abbreviations: getDefaultAbbreviations(),
	}
}

// getDefaultAbbreviations returns a basic set of common abbreviations
func getDefaultAbbreviations() map[string]bool {
	abbrevs := []string{
		// Titles
		"Dr", "Mr", "Mrs", "Ms", "Prof", "Sr", "Jr",
		// Time
		"a.m", "p.m", "A.M", "P.M",
		// Latin terms
		"e.g", "i.e", "etc", "vs", "cf",
		// Academic
		"Ph.D", "M.D", "B.A", "M.A", "B.S", "M.S",
	}

	result := make(map[string]bool)
	for _, abbrev := range abbrevs {
		// Store both with and without period for easier matching
		result[strings.ToLower(abbrev)] = true
		result[strings.ToLower(strings.TrimSuffix(abbrev, "."))] = true
	}

	return result
}

// WrapText applies semantic line wrapping to the given text
func WrapText(text string, config *Config) string {
	if !config.Enabled || text == "" {
		return text
	}

	var result strings.Builder

	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		// Skip empty lines or very short lines
		if len(strings.TrimSpace(line)) == 0 {
			result.WriteString(line)
			continue
		}

		wrapped := wrapLine(line, config)
		result.WriteString(wrapped)
	}

	return result.String()
}

// wrapLine wraps a single line according to SLW rules
func wrapLine(line string, config *Config) string {
	// Don't wrap if line is below minimum length threshold (soft wrap mode)
	if config.MinLineLength > 0 && len(line) < config.MinLineLength {
		return line
	}

	var result strings.Builder

	runes := []rune(line)
	currentLineStart := 0

	for i := 0; i < len(runes); i++ {
		if isSentenceBoundary(runes, i, config) {
			segment := string(runes[currentLineStart : i+1])
			result.WriteString(segment)

			j := skipWhitespace(runes, i+1)

			if j < len(runes) {
				result.WriteString("\n")

				currentLineStart = j
				i = j - 1
			}
		}
	}

	// Write any remaining content
	if currentLineStart < len(runes) {
		result.WriteString(string(runes[currentLineStart:]))
	}

	return result.String()
}

// isSentenceBoundary checks if position i is a sentence boundary
func isSentenceBoundary(runes []rune, i int, config *Config) bool {
	if i >= len(runes) {
		return false
	}

	char := runes[i]
	if !strings.ContainsRune(config.Markers, char) {
		return false
	}

	if i+1 >= len(runes) || !unicode.IsSpace(runes[i+1]) {
		return false
	}

	return !isAbbreviation(runes, i, config.Abbreviations)
}

// skipWhitespace returns the position after all whitespace starting at pos
func skipWhitespace(runes []rune, pos int) int {
	for pos < len(runes) && unicode.IsSpace(runes[pos]) {
		pos++
	}

	return pos
}

// isAbbreviation checks if the marker is part of an abbreviation
func isAbbreviation(runes []rune, markerPos int, abbreviations map[string]bool) bool {
	// Look backwards to extract the word before the marker
	start := markerPos - 1
	for start >= 0 && (unicode.IsLetter(runes[start]) || runes[start] == '.') {
		start--
	}

	start++

	// Extract the potential abbreviation (without the current marker)
	// This includes any internal periods (e.g., "ph.d" from "Ph.D.")
	word := strings.ToLower(string(runes[start:markerPos]))

	// Check if it's in our abbreviations list
	return abbreviations[word]
}
