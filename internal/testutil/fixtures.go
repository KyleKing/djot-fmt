package testutil

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/KyleKing/djot-fmt/internal/slw"
)

type Fixture struct {
	LineNumber int
	Title      string
	Input      string
	Expected   string
	Options    map[string]string
}

//nolint:cyclop // Test fixture parser has inherent complexity from state machine
func ReadFixtures(filepath string) ([]Fixture, error) {
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

		var inputLines []string

		for scanner.Scan() {
			lineNum++

			line := scanner.Text()
			if line == "." {
				break
			}

			inputLines = append(inputLines, line)
		}

		var expectedLines []string

		for scanner.Scan() {
			lineNum++

			line := scanner.Text()
			if line == "." {
				break
			}

			expectedLines = append(expectedLines, line)
		}

		options := make(map[string]string)

		for scanner.Scan() {
			lineNum++

			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				break
			}

			if strings.HasPrefix(line, "--") {
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

func ConfigFromOptions(options map[string]string) *slw.Config {
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
