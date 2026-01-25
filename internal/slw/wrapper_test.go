package slw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapText_BasicWrapping(t *testing.T) {
	config := DefaultConfig()
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple sentence wrapping",
			input:    "This is a long sentence. It contains multiple clauses! Does it work? Yes it does.",
			expected: "This is a long sentence.\nIt contains multiple clauses!\nDoes it work?\nYes it does.",
		},
		{
			name:     "single sentence",
			input:    "This is a single sentence.",
			expected: "This is a single sentence.",
		},
		{
			name:     "no wrapping for short text",
			input:    "Short text.",
			expected: "Short text.",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.input, config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWrapText_Abbreviations(t *testing.T) {
	config := DefaultConfig()
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "common title abbreviations",
			input:    "Dr. Smith met with Prof. Johnson yesterday morning. They discussed the project etc. and other important topics.",
			expected: "Dr. Smith met with Prof. Johnson yesterday morning.\nThey discussed the project etc. and other important topics.",
		},
		{
			name:     "time abbreviations",
			input:    "The meeting is at 9 a.m. tomorrow. Please arrive on time.",
			expected: "The meeting is at 9 a.m. tomorrow.\nPlease arrive on time.",
		},
		{
			name:     "academic abbreviations",
			input:    "She earned her Ph.D. in computer science. Her research is impressive.",
			expected: "She earned her Ph.D. in computer science.\nHer research is impressive.",
		},
		{
			name:     "latin terms",
			input:    "Use proper citations, e.g. APA or MLA format. The document should be clear, i.e. well-written.",
			expected: "Use proper citations, e.g. APA or MLA format.\nThe document should be clear, i.e. well-written.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.input, config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWrapText_SoftWrapMode(t *testing.T) {
	tests := []struct {
		name          string
		minLineLength int
		input         string
		expected      string
	}{
		{
			name:          "default soft wrap (40 chars)",
			minLineLength: 40,
			input:         "Short sentence. Another short one.",
			expected:      "Short sentence. Another short one.",
		},
		{
			name:          "aggressive mode (min 0)",
			minLineLength: 0,
			input:         "Short sentence. Another one.",
			expected:      "Short sentence.\nAnother one.",
		},
		{
			name:          "custom threshold",
			minLineLength: 60,
			input:         "This is a sentence that is definitely more than sixty characters long! Next sentence.",
			expected:      "This is a sentence that is definitely more than sixty characters long!\nNext sentence.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.MinLineLength = tt.minLineLength
			result := WrapText(tt.input, config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWrapText_MultipleLines(t *testing.T) {
	config := DefaultConfig()
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "preserve existing line breaks",
			input:    "Line one. Line two.\nLine three. Line four.",
			expected: "Line one. Line two.\nLine three. Line four.",
		},
		{
			name:     "empty lines",
			input:    "First paragraph. Second sentence.\n\nNew paragraph. More text.",
			expected: "First paragraph. Second sentence.\n\nNew paragraph. More text.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.input, config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWrapText_DisabledConfig(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = false
	
	input := "This is a long sentence. It should not be wrapped! Even though it's long?"
	result := WrapText(input, config)
	assert.Equal(t, input, result)
}

func TestWrapText_CustomMarkers(t *testing.T) {
	config := DefaultConfig()
	config.Markers = ".!?"
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "all default markers",
			input:    "First sentence. Second sentence! Third sentence? Fourth sentence.",
			expected: "First sentence.\nSecond sentence!\nThird sentence?\nFourth sentence.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.input, config)
			assert.Equal(t, tt.expected, result)
		})
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

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	assert.True(t, config.Enabled)
	assert.Equal(t, ".!?", config.Markers)
	assert.Equal(t, 40, config.MinLineLength)
	assert.Equal(t, 88, config.MaxLineWidth)
	assert.NotNil(t, config.Abbreviations)
	assert.True(t, len(config.Abbreviations) > 0)
}
