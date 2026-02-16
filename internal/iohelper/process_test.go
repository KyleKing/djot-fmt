package iohelper_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KyleKing/djot-fmt/internal/iohelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessFile_CodeBlockSupported(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.djot")

	input := "```\ncode block\n```\n"
	err := os.WriteFile(inputFile, []byte(input), 0600)
	require.NoError(t, err)

	opts := &iohelper.Options{
		Write:      true,
		InputFiles: []string{inputFile},
		SlwMarkers: ".!?",
		SlwWrap:    88,
		SlwMinLine: 40,
	}

	err = iohelper.ProcessFile(opts, inputFile)
	require.NoError(t, err, "code blocks are now supported and should not error")

	result, readErr := os.ReadFile(inputFile)
	require.NoError(t, readErr)
	assert.Equal(t, input, string(result))
}

func TestProcessFile_Write(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "simple paragraph",
			input: `# Test File

This is a paragraph with some content.

- Item 1
- Item 2
- Item 3

Another paragraph here.
`,
			expected: `# Test File

This is a paragraph with some content.


- Item 1
- Item 2
- Item 3

Another paragraph here.
`,
		},
		{
			name: "paragraph with emphasis",
			input: `This is a _paragraph_ with *emphasis*.
`,
			expected: `This is a _paragraph_ with *emphasis*.
`,
		},
		{
			name: "multiple paragraphs",
			input: `First paragraph.

Second paragraph.

Third paragraph.
`,
			expected: `First paragraph.

Second paragraph.

Third paragraph.
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "test.djot")

			err := os.WriteFile(inputFile, []byte(tt.input), 0600)
			require.NoError(t, err)

			// Process with write flag
			opts := &iohelper.Options{
				Write:      true,
				InputFiles: []string{inputFile},
				SlwMarkers: ".!?",
				SlwWrap:    88,
				SlwMinLine: 40,
			}

			err = iohelper.ProcessFile(opts, inputFile)
			require.NoError(t, err)

			// Read back the file
			result, err := os.ReadFile(inputFile)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, string(result))
		})
	}
}
