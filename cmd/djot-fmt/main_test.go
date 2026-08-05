//nolint:gosec // Every path and binary here is built from t.TempDir(), never from user input.
package main

import (
	"context"
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

	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", binaryPath)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build binary: %s", string(output))

	return binaryPath
}

func TestIntegration_MultiFileCheckMode(t *testing.T) {
	t.Parallel()

	binary := buildBinary(t)
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "formatted.dj")
	file2 := filepath.Join(tmpDir, "unformatted.dj")

	err := os.WriteFile(file1, []byte("# Heading\n\nParagraph text.\n"), 0o600)
	require.NoError(t, err)

	err = os.WriteFile(file2, []byte("-  Item 1\n-  Item 2\n"), 0o600)
	require.NoError(t, err)

	cmd := exec.CommandContext(context.Background(), binary, "-c", file1, file2)
	output, err := cmd.CombinedOutput()

	require.Error(t, err, "Expected error in check mode for unformatted files")
	assert.Contains(t, string(output), "unformatted.dj", "Error message should mention unformatted file")
}

func TestIntegration_MultiFileCheckMode_AllFormatted(t *testing.T) {
	t.Parallel()

	binary := buildBinary(t)
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.dj")
	file2 := filepath.Join(tmpDir, "file2.dj")

	err := os.WriteFile(file1, []byte("# Heading\n\nParagraph.\n"), 0o600)
	require.NoError(t, err)

	err = os.WriteFile(file2, []byte("- Item\n"), 0o600)
	require.NoError(t, err)

	cmd := exec.CommandContext(context.Background(), binary, "-c", file1, file2)
	output, err := cmd.CombinedOutput()

	assert.NoError(t, err, "Expected no error for formatted files: %s", string(output))
}

func TestIntegration_MultiFileWriteMode(t *testing.T) {
	t.Parallel()

	binary := buildBinary(t)
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.dj")
	file2 := filepath.Join(tmpDir, "file2.dj")

	err := os.WriteFile(file1, []byte("-  Item 1\n-  Item 2\n"), 0o600)
	require.NoError(t, err)

	err = os.WriteFile(file2, []byte("#  Heading\n\nText.\n"), 0o600)
	require.NoError(t, err)

	cmd := exec.CommandContext(context.Background(), binary, "-w", file1, file2)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Write mode failed: %s", string(output))

	content1, err := os.ReadFile(file1)
	require.NoError(t, err)
	assert.Equal(t, "- Item 1\n- Item 2\n", string(content1))

	content2, err := os.ReadFile(file2)
	require.NoError(t, err)
	assert.Equal(t, "# Heading\n\nText.\n", string(content2))
}

func TestIntegration_OutputFileWithSingleInput(t *testing.T) {
	t.Parallel()

	binary := buildBinary(t)
	tmpDir := t.TempDir()

	inputFile := filepath.Join(tmpDir, "input.dj")
	outputFile := filepath.Join(tmpDir, "output.dj")

	err := os.WriteFile(inputFile, []byte("-  Item\n"), 0o600)
	require.NoError(t, err)

	cmd := exec.CommandContext(context.Background(), binary, "-o", outputFile, inputFile)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Output file mode failed: %s", string(output))

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Equal(t, "- Item\n", string(content))

	inputContent, err := os.ReadFile(inputFile)
	require.NoError(t, err)
	assert.Equal(t, "-  Item\n", string(inputContent))
}

func TestIntegration_OutputFileWithMultipleInputs_Fails(t *testing.T) {
	t.Parallel()

	binary := buildBinary(t)
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.dj")
	file2 := filepath.Join(tmpDir, "file2.dj")
	outputFile := filepath.Join(tmpDir, "output.dj")

	err := os.WriteFile(file1, []byte("Text 1\n"), 0o600)
	require.NoError(t, err)

	err = os.WriteFile(file2, []byte("Text 2\n"), 0o600)
	require.NoError(t, err)

	cmd := exec.CommandContext(context.Background(), binary, "-o", outputFile, file1, file2)
	output, err := cmd.CombinedOutput()

	require.Error(t, err, "Expected error when using -o with multiple files")
	assert.Contains(t, string(output), "single input file", "Error should mention single input file requirement")
}

func TestIntegration_InfoFlags(t *testing.T) {
	t.Parallel()

	binary := buildBinary(t)

	tests := []struct {
		name     string
		flag     string
		contains []string
	}{
		{"version", "--version", []string{"djot-fmt"}},
		{"help", "--help", []string{"Usage:", "-w, --write", "-c, --check"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := exec.CommandContext(context.Background(), binary, tt.flag)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err)

			for _, s := range tt.contains {
				assert.Contains(t, string(output), s)
			}
		})
	}
}
