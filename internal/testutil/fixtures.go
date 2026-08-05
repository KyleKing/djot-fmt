// Package testutil reads the shared djot-fmt test fixture format.
package testutil

import (
	"bufio"
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
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("opening fixture file: %w", err)
	}
	//nolint:errcheck // The fixture file is read-only, so a close failure cannot affect the parsed result.
	defer file.Close()

	var fixtures []Fixture

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		title := line
		startLine := lineNum
		lineNum++

		if !scanner.Scan() {
			break
		}

		if strings.TrimSpace(scanner.Text()) != "." {
			continue
		}

		lineNum++

		inputLines := readSection(scanner, &lineNum)
		expectedLines := readSection(scanner, &lineNum)

		options := readOptions(scanner, &lineNum)

		fixtures = append(fixtures, Fixture{
			LineNumber: startLine,
			Title:      title,
			Input:      joinLines(inputLines),
			Expected:   joinLines(expectedLines),
			Options:    options,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning fixture file: %w", err)
	}

	return fixtures, nil
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	return strings.Join(lines, "\n") + "\n"
}

func readSection(scanner *bufio.Scanner, lineNum *int) []string {
	var lines []string

	for scanner.Scan() {
		*lineNum++

		line := scanner.Text()
		if line == "." {
			break
		}

		lines = append(lines, line)
	}

	return lines
}

func readOptions(scanner *bufio.Scanner, lineNum *int) map[string]string {
	options := make(map[string]string)

	for scanner.Scan() {
		*lineNum++

		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "--") {
			break
		}

		option := strings.TrimPrefix(line, "--")

		key, value, found := strings.Cut(option, "=")
		if !found {
			options[option] = optionEnabled
			continue
		}

		options[key] = strings.Trim(value, "\"")
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
