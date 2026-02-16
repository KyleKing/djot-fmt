package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildBinary(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "djot-fmt")

	cmd := exec.Command("go", "build", "-o", binaryPath)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build binary: %s", string(output))

	return binaryPath
}

func TestIntegration_MultiFileCheckMode(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()

	// Create test files: one formatted, one unformatted
	file1 := filepath.Join(tmpDir, "formatted.djot")
	file2 := filepath.Join(tmpDir, "unformatted.djot")

	err := os.WriteFile(file1, []byte("# Heading\n\nParagraph text.\n"), 0600)
	require.NoError(t, err)

	err = os.WriteFile(file2, []byte("-  Item 1\n-  Item 2\n"), 0600)
	require.NoError(t, err)

	// Run check mode on both files
	cmd := exec.Command(binary, "-c", file1, file2)
	output, err := cmd.CombinedOutput()

	// Should fail because file2 is unformatted
	require.Error(t, err, "Expected error in check mode for unformatted files")
	assert.Contains(t, string(output), "unformatted.djot", "Error message should mention unformatted file")
}

func TestIntegration_MultiFileCheckMode_AllFormatted(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()

	// Create two formatted files
	file1 := filepath.Join(tmpDir, "file1.djot")
	file2 := filepath.Join(tmpDir, "file2.djot")

	err := os.WriteFile(file1, []byte("# Heading\n\nParagraph.\n"), 0600)
	require.NoError(t, err)

	err = os.WriteFile(file2, []byte("- Item\n"), 0600)
	require.NoError(t, err)

	// Run check mode on both files
	cmd := exec.Command(binary, "-c", file1, file2)
	output, err := cmd.CombinedOutput()

	// Should succeed because both files are formatted
	assert.NoError(t, err, "Expected no error for formatted files: %s", string(output))
}

func TestIntegration_MultiFileWriteMode(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()

	// Create unformatted files
	file1 := filepath.Join(tmpDir, "file1.djot")
	file2 := filepath.Join(tmpDir, "file2.djot")

	err := os.WriteFile(file1, []byte("-  Item 1\n-  Item 2\n"), 0600)
	require.NoError(t, err)

	err = os.WriteFile(file2, []byte("#  Heading\n\nText.\n"), 0600)
	require.NoError(t, err)

	// Run write mode on both files
	cmd := exec.Command(binary, "-w", file1, file2)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Write mode failed: %s", string(output))

	// Verify files were formatted
	content1, err := os.ReadFile(file1)
	require.NoError(t, err)
	assert.Equal(t, "- Item 1\n- Item 2\n", string(content1))

	content2, err := os.ReadFile(file2)
	require.NoError(t, err)
	assert.Equal(t, "# Heading\n\nText.\n", string(content2))
}

func TestIntegration_OutputFileWithSingleInput(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()

	inputFile := filepath.Join(tmpDir, "input.djot")
	outputFile := filepath.Join(tmpDir, "output.djot")

	err := os.WriteFile(inputFile, []byte("-  Item\n"), 0600)
	require.NoError(t, err)

	// Run with output file
	cmd := exec.Command(binary, "-o", outputFile, inputFile)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Output file mode failed: %s", string(output))

	// Verify output file was created with formatted content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Equal(t, "- Item\n", string(content))

	// Verify input file was not modified
	inputContent, err := os.ReadFile(inputFile)
	require.NoError(t, err)
	assert.Equal(t, "-  Item\n", string(inputContent))
}

func TestIntegration_OutputFileWithMultipleInputs_Fails(t *testing.T) {
	binary := buildBinary(t)
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.djot")
	file2 := filepath.Join(tmpDir, "file2.djot")
	outputFile := filepath.Join(tmpDir, "output.djot")

	err := os.WriteFile(file1, []byte("Text 1\n"), 0600)
	require.NoError(t, err)

	err = os.WriteFile(file2, []byte("Text 2\n"), 0600)
	require.NoError(t, err)

	// Run with output file and multiple inputs
	cmd := exec.Command(binary, "-o", outputFile, file1, file2)
	output, err := cmd.CombinedOutput()

	// Should fail because -o only works with single input
	require.Error(t, err, "Expected error when using -o with multiple files")
	assert.Contains(t, string(output), "single input file", "Error should mention single input file requirement")
}

func TestIntegration_VersionFlag(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "--version")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	assert.Contains(t, string(output), "djot-fmt", "Version output should contain program name")
}

func TestIntegration_HelpFlag(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "--help")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	assert.Contains(t, string(output), "Usage:", "Help output should contain usage information")
	assert.Contains(t, string(output), "-w, --write", "Help should document write flag")
	assert.Contains(t, string(output), "-c, --check", "Help should document check flag")
}
