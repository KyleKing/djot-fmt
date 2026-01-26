package slw

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test data structures and configuration
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, ".!?", config.Markers)
	assert.Equal(t, 40, config.MinLineLength)
	assert.Equal(t, 88, config.MaxLineWidth)
	assert.NotNil(t, config.Abbreviations)
	assert.True(t, len(config.Abbreviations) > 0)
}

func TestConfigStructure(t *testing.T) {
	config := &Config{
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

func TestAbbreviationsMap(t *testing.T) {
	abbrevs := getDefaultAbbreviations()

	// Test that common abbreviations are present (case-insensitive)
	assert.True(t, abbrevs["dr"])
	assert.True(t, abbrevs["prof"])
	assert.True(t, abbrevs["a.m"])
	assert.True(t, abbrevs["p.m"])
	assert.True(t, abbrevs["e.g"])
	assert.True(t, abbrevs["i.e"])
	assert.True(t, abbrevs["etc"])
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
func readFixtures(filepath string) ([]Fixture, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
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

	return fixtures, scanner.Err()
}

// configFromOptions creates a Config from fixture options
func configFromOptions(options map[string]string) *Config {
	config := DefaultConfig()

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
				result := WrapText(fixture.Input, config)

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

func TestIsAbbreviation(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		name      string
		text      string
		markerPos int
		expected  bool
	}{
		{
			name:      "Dr. is abbreviation",
			text:      "Dr. Smith",
			markerPos: 2,
			expected:  true,
		},
		{
			name:      "regular sentence end",
			text:      "Hello world.",
			markerPos: 11,
			expected:  false,
		},
		{
			name:      "Ph.D. has multiple periods",
			text:      "Ph.D. degree",
			markerPos: 4,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runes := []rune(tt.text)
			result := isAbbreviation(runes, tt.markerPos, config.Abbreviations)
			assert.Equal(t, tt.expected, result)
		})
	}
}
