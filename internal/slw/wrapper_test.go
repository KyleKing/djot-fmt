package slw_test

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/KyleKing/djot-fmt/internal/slw"
	"github.com/stretchr/testify/assert"
)

// Test data structures and configuration
func TestDefaultConfig(t *testing.T) {
	config := slw.DefaultConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, ".!?", config.Markers)
	assert.Equal(t, 40, config.MinLineLength)
	assert.Equal(t, 88, config.MaxLineWidth)
	assert.NotNil(t, config.Abbreviations)
	assert.NotEmpty(t, config.Abbreviations)
}

func TestConfigStructure(t *testing.T) {
	config := &slw.Config{
		Enabled:       false,
		Markers:       ".!",
		MinLineLength: 50,
		MaxLineWidth:  100,
		Abbreviations: map[string]bool{"test": true},
	}

	assert.False(t, config.Enabled)
	assert.Equal(t, ".!", config.Markers)
	assert.Equal(t, 50, config.MinLineLength)
	assert.Equal(t, 100, config.MaxLineWidth)
	assert.True(t, config.Abbreviations["test"])
}

// Fixture represents a single test case from a fixture file
type Fixture struct {
	LineNumber int
	Title      string
	Input      string
	Expected   string
	Options    map[string]string
}

// readFixtures reads test fixtures from a file in the format:
// title
// .
// input text
// .
// expected output
// .
// --option=value (optional)
//
//nolint:cyclop // Test fixture parser has inherent complexity from state machine
func readFixtures(filepath string) ([]Fixture, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("opening fixture file: %w", err)
	}
	defer file.Close()

	var fixtures []Fixture

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Read title
		title := line
		startLine := lineNum
		lineNum++

		// Expect '.'
		if !scanner.Scan() {
			break
		}

		if strings.TrimSpace(scanner.Text()) != "." {
			continue
		}

		lineNum++

		// Read input until '.'
		var inputLines []string

		for scanner.Scan() {
			lineNum++

			line := scanner.Text()
			if line == "." {
				break
			}

			inputLines = append(inputLines, line)
		}

		// Read expected until '.'
		var expectedLines []string

		for scanner.Scan() {
			lineNum++

			line := scanner.Text()
			if line == "." {
				break
			}

			expectedLines = append(expectedLines, line)
		}

		// Read options (optional)
		options := make(map[string]string)

		for scanner.Scan() {
			lineNum++

			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				break
			}

			if strings.HasPrefix(line, "--") {
				// Parse option
				option := strings.TrimPrefix(line, "--")
				if strings.Contains(option, "=") {
					parts := strings.SplitN(option, "=", 2)
					options[parts[0]] = strings.Trim(parts[1], "\"")
				} else {
					options[option] = "true"
				}
			} else {
				break
			}
		}

		input := strings.Join(inputLines, "\n")
		if len(inputLines) > 0 {
			input += "\n"
		}

		expected := strings.Join(expectedLines, "\n")
		if len(expectedLines) > 0 {
			expected += "\n"
		}

		fixtures = append(fixtures, Fixture{
			LineNumber: startLine,
			Title:      title,
			Input:      input,
			Expected:   expected,
			Options:    options,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning fixture file: %w", err)
	}

	return fixtures, nil
}

// configFromOptions creates a Config from fixture options
func configFromOptions(options map[string]string) *slw.Config {
	config := slw.DefaultConfig()

	if val, ok := options["no-wrap-sentences"]; ok && val == "true" {
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

func TestFixtures(t *testing.T) {
	fixtureFiles := []string{
		"basic.txt",
	}

	for _, filename := range fixtureFiles {
		path := filepath.Join("../../testdata/slw", filename)

		fixtures, err := readFixtures(path)
		if err != nil {
			t.Fatalf("Failed to read fixtures from %s: %v", filename, err)
		}

		for _, fixture := range fixtures {
			t.Run(fixture.Title, func(t *testing.T) {
				config := configFromOptions(fixture.Options)
				result := slw.WrapText(fixture.Input, config)

				if !assert.Equal(t, fixture.Expected, result) {
					t.Logf("Fixture: %s (line %d)", fixture.Title, fixture.LineNumber)
					t.Logf("Input: %q", fixture.Input)
					t.Logf("Expected: %q", fixture.Expected)
					t.Logf("Got: %q", result)
				}
			})
		}
	}
}
