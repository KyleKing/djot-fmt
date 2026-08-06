// Package testutil reads the shared djot-fmt test fixture format.
package testutil

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/KyleKing/djot-fmt/internal/slw"
)

const optionEnabled = "true"

// Fixture is one input/expected pair parsed from a fixture file.
type Fixture struct {
	Options    map[string]string
	Title      string
	Input      string
	Expected   string
	LineNumber int
}

// ReadFixtures parses every fixture defined in the file at filepath.
func ReadFixtures(filepath string) ([]Fixture, error) {
	//nolint:gosec // Fixture paths come from the test suite, not from user input.
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("opening fixture file: %w", err)
	}

	return parseFixtures(strings.Split(string(data), "\n")), nil
}

// cursor walks the fixture file, so a section that stops on a line belonging to
// the next fixture leaves that line for the next read rather than eating it.
type cursor struct {
	lines []string
	i     int
}

func parseFixtures(lines []string) []Fixture {
	var fixtures []Fixture

	c := &cursor{lines: lines}

	for c.i < len(c.lines) {
		title := strings.TrimSpace(c.lines[c.i])
		if title == "" || c.i+1 >= len(c.lines) || strings.TrimSpace(c.lines[c.i+1]) != "." {
			c.i++
			continue
		}

		startLine := c.i + 1
		c.i += 2

		input := c.section()
		expected := c.section()

		fixtures = append(fixtures, Fixture{
			LineNumber: startLine,
			Title:      title,
			Input:      joinLines(input),
			Expected:   joinLines(expected),
			Options:    c.options(),
		})
	}

	return fixtures
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	return strings.Join(lines, "\n") + "\n"
}

func (c *cursor) section() []string {
	var section []string

	for ; c.i < len(c.lines); c.i++ {
		if c.lines[c.i] == "." {
			c.i++
			return section
		}

		section = append(section, c.lines[c.i])
	}

	return section
}

func (c *cursor) options() map[string]string {
	options := make(map[string]string)

	for ; c.i < len(c.lines); c.i++ {
		line := strings.TrimSpace(c.lines[c.i])
		if !strings.HasPrefix(line, "--") {
			return options
		}

		key, value, found := strings.Cut(strings.TrimPrefix(line, "--"), "=")
		if !found {
			options[key] = optionEnabled
			continue
		}

		options[key] = strings.Trim(value, `"`)
	}

	return options
}

// ConfigFromOptions builds a wrapping config from a fixture's option map.
func ConfigFromOptions(options map[string]string) *slw.Config {
	config := slw.DefaultConfig()

	if val, ok := options["no-wrap-sentences"]; ok && val == optionEnabled {
		config.Enabled = false
	}

	if val, ok := options["slw-markers"]; ok {
		config.Markers = val
	}

	if val, ok := options["slw-wrap"]; ok {
		if i, err := strconv.Atoi(val); err == nil {
			config.MaxLineWidth = i
		}
	}

	if val, ok := options["slw-min-line"]; ok {
		if i, err := strconv.Atoi(val); err == nil {
			config.MinLineLength = i
		}
	}

	return config
}
