package validate

import (
	"regexp"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

// Category names the kind of difference found between two renderings.
type Category string

// Difference categories, ordered from the most to the least serious.
const (
	CategoryText         Category = "text"
	CategoryBlockKind    Category = "block-kind"
	CategoryListAttrs    Category = "list-attrs"
	CategoryListNesting  Category = "list-nesting"
	CategoryLooseTight   Category = "loose-tight"
	CategoryUnclassified Category = "unclassified"
)

// Difference is one classified change between the input and the output rendering.
type Difference struct {
	Category Category
	Detail   string
	waivable bool
}

// Options controls which differences are tolerated.
type Options struct {
	// CodeLanguages names the languages with a configured formatter. Their code
	// block bodies are rewritten deliberately, so they are excluded before comparing.
	CodeLanguages []string
	// FromMD tolerates the structural change that formatting markdown as djot
	// legitimately causes, because djot requires a blank line before a nested
	// list and markdown does not.
	FromMD bool
}

// Result separates differences that must fail the run from those a flag waived.
type Result struct {
	Rejected []Difference
	Waived   []Difference
}

// OK reports whether the output preserved everything the policy requires.
func (r Result) OK() bool { return len(r.Rejected) == 0 }

// Compare renders both documents to HTML and classifies every difference.
// The word multiset is the floor: no option lets a word disappear except a
// leading list marker that became structure.
func Compare(input, output []byte, opts Options) Result {
	in := tokenize(renderHTML(input, opts.CodeLanguages))
	out := tokenize(renderHTML(output, opts.CodeLanguages))

	diffs := textDifferences(in, out)
	diffs = append(diffs, structureDifferences(in, out)...)

	return partition(diffs, opts.FromMD)
}

func partition(diffs []Difference, fromMD bool) Result {
	var result Result

	for _, d := range diffs {
		if fromMD && d.waivable {
			result.Waived = append(result.Waived, d)
			continue
		}

		result.Rejected = append(result.Rejected, d)
	}

	return result
}

var listMarker = regexp.MustCompile(`^\(?(?:[-*+]|\d+|[a-zA-Z]|[ivxlcdmIVXLCDM]+)[.)]?$`)

func textDifferences(in, out []token) []Difference {
	counts := map[string]int{}

	for _, t := range in {
		if t.kind == kindWord {
			counts[t.name]++
		}
	}

	var diffs []Difference

	for _, t := range out {
		if t.kind != kindWord {
			continue
		}

		if counts[t.name] == 0 {
			diffs = append(diffs, Difference{Category: CategoryText, Detail: "output invented the word " + t.name})
			continue
		}

		counts[t.name]--
	}

	return append(diffs, missingWords(counts)...)
}

func missingWords(counts map[string]int) []Difference {
	var diffs []Difference

	for word, n := range counts {
		if n == 0 || listMarker.MatchString(word) {
			continue
		}

		diffs = append(diffs, Difference{Category: CategoryText, Detail: "output dropped the word " + word})
	}

	return diffs
}

func structureDifferences(in, out []token) []Difference {
	matcher := difflib.NewMatcher(keys(in), keys(out))

	var diffs []Difference

	for _, op := range matcher.GetOpCodes() {
		if op.Tag == 'e' {
			continue
		}

		deleted, inserted := onlyTags(in[op.I1:op.I2]), onlyTags(out[op.J1:op.J2])
		if len(deleted) == 0 && len(inserted) == 0 {
			continue
		}

		diffs = append(diffs, classify(deleted, inserted))
	}

	return diffs
}

var (
	listTags  = set("ul", "/ul", "ol", "/ol", "li", "/li")
	paraTags  = set("p", "/p")
	blockTags = set(
		"p", "/p", "h1", "/h1", "h2", "/h2", "h3", "/h3", "h4", "/h4", "h5", "/h5", "h6", "/h6",
		"hr", "blockquote", "/blockquote", "section", "/section", "div", "/div", "pre", "/pre",
		"ul", "/ul", "ol", "/ol", "li", "/li", "dl", "/dl", "dt", "/dt", "dd", "/dd",
	)
)

func set(names ...string) map[string]bool {
	m := make(map[string]bool, len(names))
	for _, n := range names {
		m[n] = true
	}

	return m
}

func classify(deleted, inserted []token) Difference {
	detail := render(deleted) + " -> " + render(inserted)

	if attrsOnly(deleted, inserted) {
		if deleted[0].name == "ol" || deleted[0].name == "ul" {
			return Difference{Category: CategoryListAttrs, Detail: detail}
		}

		return Difference{Category: CategoryUnclassified, Detail: detail}
	}

	gone, added := tagNames(deleted), tagNames(inserted)
	names := make([]string, 0, len(gone)+len(added))
	names = append(append(names, gone...), added...)

	switch {
	// Formatting markdown as djot only ever adds structure, because the fix is
	// inserting a blank line. Structure that disappeared merged or freed content
	// that used to be contained, so no flag waives it.
	case within(names, paraTags):
		return Difference{
			Category: CategoryLooseTight,
			Detail:   detail,
			waivable: !containsAny(gone, paraTags),
		}
	case within(names, union(listTags, paraTags)):
		return Difference{
			Category: CategoryListNesting,
			Detail:   detail,
			waivable: !containsAny(gone, union(listTags, paraTags)),
		}
	case within(names, blockTags):
		return Difference{Category: CategoryBlockKind, Detail: detail}
	default:
		return Difference{Category: CategoryUnclassified, Detail: detail}
	}
}

func containsAny(names []string, wanted map[string]bool) bool {
	for _, n := range names {
		if wanted[n] {
			return true
		}
	}

	return false
}

func attrsOnly(deleted, inserted []token) bool {
	if len(deleted) == 0 || len(deleted) != len(inserted) {
		return false
	}

	for i := range deleted {
		if deleted[i].name != inserted[i].name {
			return false
		}
	}

	return true
}

func tagNames(tokens []token) []string {
	out := make([]string, len(tokens))
	for i, t := range tokens {
		out[i] = t.name
	}

	return out
}

func within(names []string, allowed map[string]bool) bool {
	for _, n := range names {
		if !allowed[n] {
			return false
		}
	}

	return true
}

func union(a, b map[string]bool) map[string]bool {
	m := make(map[string]bool, len(a)+len(b))
	for k := range a {
		m[k] = true
	}

	for k := range b {
		m[k] = true
	}

	return m
}

func render(tokens []token) string {
	if len(tokens) == 0 {
		return "(nothing)"
	}

	parts := make([]string, len(tokens))
	for i, t := range tokens {
		parts[i] = t.String()
	}

	return strings.Join(parts, "")
}
