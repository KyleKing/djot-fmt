package slw

import (
	"strings"
	"unicode"
)

type Config struct {
	Enabled       bool
	Markers       string
	MinLineLength int
	MaxLineWidth  int
	Abbreviations map[string]bool
}

func DefaultConfig() *Config {
	return &Config{
		Enabled:       true,
		Markers:       ".!?",
		MinLineLength: 40,
		MaxLineWidth:  88,
		Abbreviations: getDefaultAbbreviations(),
	}
}

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

		if len(strings.TrimSpace(line)) == 0 {
			result.WriteString(line)
			continue
		}

		wrapped := wrapLine(line, config)
		result.WriteString(wrapped)
	}

	return result.String()
}

func wrapLine(line string, config *Config) string {
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

	if currentLineStart < len(runes) {
		result.WriteString(string(runes[currentLineStart:]))
	}

	return result.String()
}

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

func skipWhitespace(runes []rune, pos int) int {
	for pos < len(runes) && unicode.IsSpace(runes[pos]) {
		pos++
	}

	return pos
}

func isAbbreviation(runes []rune, markerPos int, abbreviations map[string]bool) bool {
	start := markerPos - 1
	for start >= 0 && (unicode.IsLetter(runes[start]) || runes[start] == '.') {
		start--
	}

	start++

	word := strings.ToLower(string(runes[start:markerPos]))

	return abbreviations[word]
}
