package slw_test

import (
	"path/filepath"
	"testing"

	"github.com/KyleKing/djot-fmt/internal/slw"
	"github.com/KyleKing/djot-fmt/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := slw.DefaultConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, ".!?", config.Markers)
	assert.Equal(t, 40, config.MinLineLength)
	assert.Equal(t, 88, config.MaxLineWidth)
	assert.NotNil(t, config.Abbreviations)
	assert.NotEmpty(t, config.Abbreviations)
}

func TestFixtures(t *testing.T) {
	fixtureFiles := []string{
		"basic.txt",
	}

	for _, filename := range fixtureFiles {
		path := filepath.Join("../../testdata/slw", filename)

		fixtures, err := testutil.ReadFixtures(path)
		if err != nil {
			t.Fatalf("Failed to read fixtures from %s: %v", filename, err)
		}

		for _, fixture := range fixtures {
			t.Run(fixture.Title, func(t *testing.T) {
				config := testutil.ConfigFromOptions(fixture.Options)
				result := slw.WrapText(fixture.Input, config)

				if !assert.Equal(t, fixture.Expected, result) {
					t.Logf("Fixture: %s (line %d)", fixture.Title, fixture.LineNumber)
					t.Logf("Input: %q", fixture.Input)
				}
			})
		}
	}
}
