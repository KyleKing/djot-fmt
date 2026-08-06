package slw_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/KyleKing/djot-fmt/internal/slw"
	"github.com/KyleKing/djot-fmt/internal/testutil"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := slw.DefaultConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, ".!?", config.Markers)
	assert.Equal(t, 40, config.MinLineLength)
	assert.Equal(t, 88, config.MaxLineWidth)
	assert.NotNil(t, config.Abbreviations)
	assert.NotEmpty(t, config.Abbreviations)
}

func TestFixtures(t *testing.T) {
	t.Parallel()

	fixtureFiles := []string{
		"basic.txt",
	}

	for _, filename := range fixtureFiles {
		path := filepath.Join("..", "..", "testdata", "slw", filename)

		fixtures, err := testutil.ReadFixtures(path)
		if err != nil {
			t.Fatalf("Failed to read fixtures from %s: %v", filename, err)
		}

		for _, fixture := range fixtures {
			t.Run(fixture.Title, func(t *testing.T) {
				t.Parallel()

				config := testutil.ConfigFromOptions(fixture.Options)
				result := slw.Wrap(strings.TrimSuffix(fixture.Input, "\n"), slw.Layout{}, config) + "\n"

				if !assert.Equal(t, fixture.Expected, result) {
					t.Logf("Fixture: %s (line %d)", fixture.Title, fixture.LineNumber)
					t.Logf("Input: %q", fixture.Input)
				}
			})
		}
	}
}
