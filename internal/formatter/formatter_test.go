package formatter_test

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/KyleKing/djot-fmt/internal/formatter"
	"github.com/KyleKing/djot-fmt/internal/slw"
	"github.com/sivukhin/godjot/v2/djot_parser"
	"github.com/stretchr/testify/assert"
)

func TestFormat_UnsupportedNodePanics(t *testing.T) {
	unsupportedInputs := []struct {
		name  string
		input string
	}{
		// No unsupported types tested yet - add when implementing definition lists
	}

	for _, tt := range unsupportedInputs {
		t.Run(tt.name, func(t *testing.T) {
			ast := djot_parser.BuildDjotAst([]byte(tt.input))

			assert.Panics(t, func() { formatter.Format(ast) })
		})
	}
}

// Test basic data structure manipulation
func TestFormat_SimpleParagraphAST(t *testing.T) {
	// Test with manually constructed AST to ensure the data structure handling works
	ast := []djot_parser.TreeNode[djot_parser.DjotNode]{
		{
			Type: djot_parser.ParagraphNode,
			Children: []djot_parser.TreeNode[djot_parser.DjotNode]{
				{Type: djot_parser.TextNode, Text: []byte("Hello, world!")},
			},
		},
	}

	result := formatter.Format(ast)
	expected := "Hello, world!\n"
	assert.Equal(t, expected, result)
}

// Fixture represents a single test case from a fixture file
type Fixture struct {
	LineNumber int
	Title      string
	Input      string
	Expected   string
	Options    map[string]string
}

// readFixtures reads test fixtures from a file in the format used by mdformat-slw
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

// configFromOptions creates a SLW Config from fixture options
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

func TestFormat_BasicFixtures(t *testing.T) {
	path := filepath.Join("../../testdata/formatter", "basic.txt")

	fixtures, err := readFixtures(path)
	if err != nil {
		t.Fatalf("Failed to read fixtures: %v", err)
	}

	for _, fixture := range fixtures {
		t.Run(fixture.Title, func(t *testing.T) {
			ast := djot_parser.BuildDjotAst([]byte(fixture.Input))
			result := formatter.Format(ast)

			if !assert.Equal(t, fixture.Expected, result) {
				t.Logf("Fixture: %s (line %d)", fixture.Title, fixture.LineNumber)
				t.Logf("Input: %q", fixture.Input)
				t.Logf("Expected: %q", fixture.Expected)
				t.Logf("Got: %q", result)
			}
		})
	}
}

func TestFormat_SLWFixtures(t *testing.T) {
	path := filepath.Join("../../testdata/formatter", "slw.txt")

	fixtures, err := readFixtures(path)
	if err != nil {
		t.Fatalf("Failed to read fixtures: %v", err)
	}

	for _, fixture := range fixtures {
		t.Run(fixture.Title, func(t *testing.T) {
			config := configFromOptions(fixture.Options)
			ast := djot_parser.BuildDjotAst([]byte(fixture.Input))
			result := formatter.FormatWithConfig(ast, config)

			if !assert.Equal(t, fixture.Expected, result) {
				t.Logf("Fixture: %s (line %d)", fixture.Title, fixture.LineNumber)
				t.Logf("Input: %q", fixture.Input)
				t.Logf("Expected: %q", fixture.Expected)
				t.Logf("Got: %q", result)
			}
		})
	}
}

func TestFormat_InlineFixtures(t *testing.T) {
	path := filepath.Join("../../testdata/formatter", "inline.txt")

	fixtures, err := readFixtures(path)
	if err != nil {
		t.Fatalf("Failed to read fixtures: %v", err)
	}

	for _, fixture := range fixtures {
		t.Run(fixture.Title, func(t *testing.T) {
			ast := djot_parser.BuildDjotAst([]byte(fixture.Input))
			result := formatter.Format(ast)

			if !assert.Equal(t, fixture.Expected, result) {
				t.Logf("Fixture: %s (line %d)", fixture.Title, fixture.LineNumber)
				t.Logf("Input: %q", fixture.Input)
				t.Logf("Expected: %q", fixture.Expected)
				t.Logf("Got: %q", result)
			}
		})
	}
}
