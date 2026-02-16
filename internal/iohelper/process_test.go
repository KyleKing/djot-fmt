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

func TestProcessFile_Check_AlreadyFormatted(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.djot")

	// Write already-formatted content
	formatted := `# Test

This is formatted.
`
	err := os.WriteFile(inputFile, []byte(formatted), 0600)
	require.NoError(t, err)

	opts := &iohelper.Options{
		Check:      true,
		InputFiles: []string{inputFile},
		SlwMarkers: ".!?",
		SlwWrap:    88,
		SlwMinLine: 40,
	}

	err = iohelper.ProcessFile(opts, inputFile)
	require.NoError(t, err, "already formatted file should not return error")

	// Verify file unchanged
	result, readErr := os.ReadFile(inputFile)
	require.NoError(t, readErr)
	assert.Equal(t, formatted, string(result), "file should not be modified")
}

func TestProcessFile_Check_NeedsFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.djot")

	// Write content that needs formatting (extra spaces in list)
	unformatted := `-  Item 1
-  Item 2
`
	err := os.WriteFile(inputFile, []byte(unformatted), 0600)
	require.NoError(t, err)

	opts := &iohelper.Options{
		Check:      true,
		InputFiles: []string{inputFile},
		SlwMarkers: ".!?",
		SlwWrap:    88,
		SlwMinLine: 40,
	}

	err = iohelper.ProcessFile(opts, inputFile)
	require.Error(t, err, "unformatted file should return error")
	assert.Contains(t, err.Error(), "not formatted")

	// Verify file unchanged
	result, readErr := os.ReadFile(inputFile)
	require.NoError(t, readErr)
	assert.Equal(t, unformatted, string(result), "file should not be modified in check mode")
}

func TestProcessFile_Check_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "empty.djot")

	// Empty file with trailing newline
	err := os.WriteFile(inputFile, []byte("\n"), 0600)
	require.NoError(t, err)

	opts := &iohelper.Options{
		Check:      true,
		InputFiles: []string{inputFile},
		SlwMarkers: ".!?",
		SlwWrap:    88,
		SlwMinLine: 40,
	}

	err = iohelper.ProcessFile(opts, inputFile)
	require.NoError(t, err, "empty file should be considered formatted")

	// Verify file unchanged
	result, readErr := os.ReadFile(inputFile)
	require.NoError(t, readErr)
	assert.Equal(t, "\n", string(result))
}

func TestProcessFile_OutputFile_CreatesNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.djot")
	outputFile := filepath.Join(tmpDir, "output.djot")

	input := `-  Item 1
-  Item 2
`
	expected := `- Item 1
- Item 2
`
	err := os.WriteFile(inputFile, []byte(input), 0600)
	require.NoError(t, err)

	opts := &iohelper.Options{
		OutputFile: outputFile,
		InputFiles: []string{inputFile},
		SlwMarkers: ".!?",
		SlwWrap:    88,
		SlwMinLine: 40,
	}

	err = iohelper.ProcessFile(opts, inputFile)
	require.NoError(t, err)

	// Verify output file created with formatted content
	result, readErr := os.ReadFile(outputFile)
	require.NoError(t, readErr)
	assert.Equal(t, expected, string(result))

	// Verify input file unchanged
	inputResult, inputReadErr := os.ReadFile(inputFile)
	require.NoError(t, inputReadErr)
	assert.Equal(t, input, string(inputResult), "input file should not be modified")
}

func TestProcessFile_OutputFile_OverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.djot")
	outputFile := filepath.Join(tmpDir, "output.djot")

	input := `# New Content
`
	oldOutput := `# Old Content
`
	expected := `# New Content
`

	err := os.WriteFile(inputFile, []byte(input), 0600)
	require.NoError(t, err)

	// Create existing output file with old content
	err = os.WriteFile(outputFile, []byte(oldOutput), 0600)
	require.NoError(t, err)

	opts := &iohelper.Options{
		OutputFile: outputFile,
		InputFiles: []string{inputFile},
		SlwMarkers: ".!?",
		SlwWrap:    88,
		SlwMinLine: 40,
	}

	err = iohelper.ProcessFile(opts, inputFile)
	require.NoError(t, err)

	// Verify output file overwritten with new formatted content
	result, readErr := os.ReadFile(outputFile)
	require.NoError(t, readErr)
	assert.Equal(t, expected, string(result))
}

func TestProcessFile_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	nonexistentFile := filepath.Join(tmpDir, "does-not-exist.djot")

	opts := &iohelper.Options{
		Write:      true,
		InputFiles: []string{nonexistentFile},
		SlwMarkers: ".!?",
		SlwWrap:    88,
		SlwMinLine: 40,
	}

	err := iohelper.ProcessFile(opts, nonexistentFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading input file")
}

func TestProcessFile_WritePermissionDenied(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "readonly.djot")

	input := `# Test
`
	err := os.WriteFile(inputFile, []byte(input), 0600)
	require.NoError(t, err)

	// Make file read-only
	err = os.Chmod(inputFile, 0400)
	require.NoError(t, err)

	opts := &iohelper.Options{
		Write:      true,
		InputFiles: []string{inputFile},
		SlwMarkers: ".!?",
		SlwWrap:    88,
		SlwMinLine: 40,
	}

	err = iohelper.ProcessFile(opts, inputFile)
	require.Error(t, err, "should fail to write to read-only file")
	assert.Contains(t, err.Error(), "writing to file")
}
