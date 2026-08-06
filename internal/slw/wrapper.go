// Package slw implements semantic line wrapping, breaking text after sentence endings.
package slw

import (
	"regexp"
	"strings"
)

const (
	defaultMinLineLength = 40
	defaultMaxLineWidth  = 88
)

// Config controls how text is split at sentence boundaries.
type Config struct {
	Abbreviations map[string]bool
	Markers       string
	MinLineLength int
	MaxLineWidth  int
	Enabled       bool
}

// Layout gives the columns a wrapped block occupies, so wrapping accounts for
// the indent and line prefixes the writer adds around it.
type Layout struct {
	FirstColumn        int
	ContinuationColumn int
}

// DefaultConfig returns the semantic line wrapping defaults.
func DefaultConfig() *Config {
	return &Config{
		Enabled:       true,
		Markers:       ".!?",
		MinLineLength: defaultMinLineLength,
		MaxLineWidth:  defaultMaxLineWidth,
		Abbreviations: getDefaultAbbreviations(),
	}
}

func getDefaultAbbreviations() map[string]bool {
	abbrevs := []string{
		// Titles
		"Dr", "Mr", "Mrs", "Ms", "Prof", "Sr", "Jr",
		// Time
		"a.m", "p.m", "A.M", "P.M",
		// Latin terms
		"e.g", "i.e", "etc", "vs", "cf",
		// Academic
		"Ph.D", "M.D", "B.A", "M.A", "B.S", "M.S",
	}

	result := make(map[string]bool)
	for _, abbrev := range abbrevs {
		result[strings.ToLower(abbrev)] = true
		result[strings.ToLower(strings.TrimSuffix(abbrev, "."))] = true
	}

	return result
}

// Wrap breaks one block of rendered inline text after sentence endings and at
// the configured width. Existing line breaks are kept, because an author who
// broke a line meant it.
func Wrap(text string, layout Layout, cfg *Config) string {
	if !cfg.Enabled || text == "" {
		return text
	}

	var out []string

	column := layout.FirstColumn

	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == "" {
			out = append(out, line)
			column = layout.ContinuationColumn

			continue
		}

		wrapped := wrapLine(line, layout.withFirst(column), cfg)
		out = append(out, wrapped...)
		column = layout.ContinuationColumn
	}

	return pullBlockMarkers(out)
}

func (l Layout) withFirst(column int) Layout {
	l.FirstColumn = column
	return l
}

func wrapLine(line string, layout Layout, cfg *Config) []string {
	atoms := splitAtoms(collapseSpaces(line), cfg)
	if len(atoms) == 0 {
		return []string{""}
	}

	var (
		lines   []string
		current strings.Builder
	)

	width := 0
	column := layout.FirstColumn

	breakLine := func() {
		lines = append(lines, current.String())
		current.Reset()

		width = 0
		column = layout.ContinuationColumn
	}

	for i, a := range atoms {
		if width > 0 && exceedsWidth(column+width+1+a.width, cfg) {
			breakLine()
		}

		if width > 0 {
			current.WriteString(" ")

			width++
		}

		current.WriteString(a.text)

		width += a.width

		if a.sentenceEnd && i < len(atoms)-1 && breaksAfterSentence(column+width, cfg) {
			breakLine()
		}
	}

	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	return lines
}

func exceedsWidth(column int, cfg *Config) bool {
	return cfg.MaxLineWidth > 0 && column > cfg.MaxLineWidth
}

func breaksAfterSentence(column int, cfg *Config) bool {
	return cfg.MinLineLength <= 0 || column >= cfg.MinLineLength
}

// collapseSpaces reduces runs of spaces and tabs to one space. A non-breaking
// space is content rather than layout, so it survives untouched.
func collapseSpaces(line string) string {
	var (
		b       strings.Builder
		lastGap bool
	)

	for _, r := range line {
		gap := r == ' ' || r == '\t'
		if gap && lastGap {
			continue
		}

		if gap {
			b.WriteRune(' ')
		} else {
			b.WriteRune(r)
		}

		lastGap = gap
	}

	return strings.TrimSpace(b.String())
}

// RE2 has no backreferences, so the thematic break alternative is spelled out
// once per marker character instead of matching the first one against itself.
var blockMarker = regexp.MustCompile(
	`^ {0,3}((-+|=+) *$|>|#{1,6}( |$)|\*( *\*){2,} *$|_( *_){2,} *$|[-+*]( |$)|\d{1,9}[.)]( |$))`,
)

// pullBlockMarkers moves a token back onto the line above when wrapping left it
// at a line start where djot would reparse it as a block construct.
func pullBlockMarkers(lines []string) string {
	var merged []string

	for _, line := range lines {
		current := line
		pulled := false

		for len(merged) > 0 && blockMarker.MatchString(current) {
			head, tail, _ := strings.Cut(strings.TrimSpace(current), " ")
			merged[len(merged)-1] += " " + head
			current = tail
			pulled = true

			if current == "" {
				break
			}
		}

		if current == "" && pulled {
			continue
		}

		merged = append(merged, current)
	}

	return strings.Join(merged, "\n")
}
