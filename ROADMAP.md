# Roadmap

Where djot-fmt goes from `v0.1.0rc1` to a formatter that yak-shears can depend on.
Seven milestones, ordered by what blocks what.

## Verified starting point

Everything below was reproduced against the current `main` before planning, because
several of these are worse than the code reads.

| Input | Output | Verdict |
| --- | --- | --- |
| `---`/`title: Hi`/`---` frontmatter | `***` thematic breaks, keys emitted as prose | Data loss, blocks yak-shears |
| `- a`/`- b`/blank/`  nested para` | `nested para` emitted at top level | Data loss, content escapes the list item |
| `5. five` / `6. six` | `1. five` / `1. six` | Start number lost, and djot says the start number is significant |
| `i. one` / `ii. two` | `a. one` / `a. two` | godjot parses this as lower-alpha `start="9"`, so roman is lost upstream |
| `1) one` / `(1) one` | `1. one` | Delimiter style normalized, no way to opt out |
| 160-char line, no sentence end | Unchanged | `--slw-wrap` is accepted, stored, and never read |
| Sentence break next to `` `code. inside` `` and `[link with. dots](u)` | Break inserted, no protection | mdformat-slw protects both |
| Loose list item, `--slw-min-line 0` | Continuation line at column 0 | Legal djot (lazy continuation), reads badly in a diff |
| ` ``` python ` block | Fence normalized, body untouched | No code formatting yet |

Two spec points I checked rather than assumed, from `jgm/djot doc/syntax.md`:

- djot allows lazy continuation, so the unindented list continuation above is valid.
  It is a style question, not a correctness bug
- djot defines no frontmatter. Any support here is an extension, which is exactly why
  it belongs behind configuration

`--slw-wrap` being inert is the clearest signal of the structural problem: `slw.WrapText`
runs inside `formatText`, on one `TextNode` at a time. It cannot see the rendered width
of a line because emphasis, links, and code spans are siblings it never sees, and it
cannot see the active indent or `>` prefix. Width-aware wrapping, link protection, and
indent-aware continuation all need the same fix.

## M1: validation, so corruption cannot ship silently

This comes first because it is what turns every remaining milestone from "hope the
fixtures cover it" into a hard guarantee, and because it found a bug the manual probing
missed.

mdformat's approach (`mdformat/_util.py:is_md_equal`, called from `_cli.py:157`): render
both the input and the output to HTML, collapse whitespace, drop empty paragraphs,
compare. Not equal means abort the write and exit non-zero with "this is a bug in
mdformat". It is skipped when a plugin declares `CHANGES_AST`, code-formatted languages
are stripped from the HTML before comparing, and `--no-validate` opts out.

djot-fmt can do exactly this. godjot ships `djot_html`, so the check is about 25 lines.
I prototyped it and ran it against the cases above:

```
fm     DIFF   in: <hr> <p><hr> title: Hi</p> …    out: <hr> <hr> <p>title: Hi</p> …
list   DIFF   in: <li> <p>b</p> <p>nested para</p> </li>
              out: <li> <p>b</p> </li> </ul> <p>nested para</p>
num    DIFF   in: <ol start="5">                  out: <ol>
roman  DIFF   in: <ol start="9" type="a">         out: <ol type="a">
slw    OK
misc   OK     (tables, blockquotes, and ::: divs)
```

It caught every known bug, found the nested-paragraph one, and did not false-positive on
the formatting djot-fmt gets right. That is the whole case for building it first.

Design notes:

- Normalization has to match mdformat's list, plus djot's `<section>` wrappers around
  headings
- Semantic line wrapping only inserts whitespace, so it is HTML-invisible and needs no
  exemption. Code formatting (M6) does change content, so strip `<code
  class="language-X">` for any language with a configured formatter, same as mdformat
- Frontmatter is stripped before both renders, since it is not djot and must compare
  byte-for-byte instead
- On by default, `--no-validate` to opt out, and the failure message should name
  djot-fmt as the bug rather than blaming the input
- Wire the same check into the fuzz corpus. godjot already has `djot_fuzz_test.go` to
  copy from, and "parse, format, reparse, compare HTML" is a natural fuzz property

## M2: buffer inline output per block

`Writer` gains an inline buffer. Block formatters open the buffer, let children render
into it, then hand the finished string to a post-render pass before writing it out with
the current indent and prefix stack applied. This is the same shape as mdformat's
`POSTPROCESSORS` map, which is why mdformat-slw can protect link syntax at all.

Concretely:

- `Writer.BeginInline()` / `EndInline() string` around the `next(nil)` call
- `slw.Wrap(text, startColumn, cfg)` replaces the per-`TextNode` call
- `formatText` goes back to writing text verbatim

No intended behavior change beyond the wrapping that starts working as documented. M1
makes this refactor safe to do.

## M3: frontmatter

The blocker for yak-shears, and the one that currently destroys files.

Detect a delimited block at byte 0 before `BuildDjotAst` runs, hold it aside, format the
remaining body, and re-emit the block ahead of the output. Recognize `---` (YAML),
`+++` (TOML), and a leading `{` object (JSON), matching mdformat-front-matters.

Detect and preserve verbatim by default, and normalize only when asked. The alternative
default is corrupting input. This keeps djot-fmt spec-consistent in the sense that
counts: it does not invent syntax, it declines to mangle bytes it does not own.

Two details to pin early. A `---` at byte 0 with no closing delimiter is a thematic
break, so require the close before treating it as frontmatter. And yak-shears' Apple
Notes export format (`: key=value\`) is real djot syntax, so leave it to yak-shears.

## M4: ordered list numbering

godjot already gives us most of what we need. It sets `start` on the list node when the
first item is not `1`, and `type` carries the enumerator marker. `formatListItem` reads
`type` and ignores `start`, which is where `5.` becomes `1.`.

- Read `start` and count from it
- Map the marker to its style and render item `n` in that style instead of echoing the
  first marker
- Preserve the delimiter (`1.`, `1)`, `(1)`) rather than normalizing to `.`

Roman is a separate problem and it is upstream. godjot renders `i.`/`ii.` as
`<ol start="9" type="a">`, so it never recognizes roman enumerators at all. Either fix
`detectListProps` in godjot and upstream it, or detect the style in djot-fmt before
parsing. Worth scoping before committing to the fix, and worth noting that M1 will flag
any half-fix as a validation failure rather than letting it through.

Then the one genuinely optional part:

```toml
[lists]
numbering = "one"  # one | consecutive
```

`one` writes `1.` for every item, keeping diffs small when items are inserted.
`consecutive` writes `1.`, `2.`, `3.`, mirroring mdformat-mkdocs `--number`. Default to
`one` for the same diff-size reason that motivates semantic line wrapping. The start
number is honored in both modes because the spec says it is meaningful.

## M5: semantic line wrapping parity

The Go implementation is 152 lines against 772 in Python. The gap, in the order the
Python pipeline applies it:

| mdformat-slw stage | Go status |
| --- | --- |
| Collapse whitespace, preserve U+00A0 | Missing |
| Find protected regions (code spans, inline and reference links) | Missing |
| Sentence break, with closing `"'` `)]}` carried past the marker | Marker only |
| Min-line measured against the accumulated output line | Measured against the whole input line, in bytes |
| Abbreviation suppression | 20 hardcoded English entries vs 6 language lists |
| Replace spaces in link text so links do not split | Missing |
| Protect spaces inside code spans | Missing |
| Wrap to `--slw-wrap` using display width | Not implemented at all |
| Pull block markers off wrapped line starts | Missing |

M2 makes stages 2, 7, 8, and 9 possible. Stages 1, 3, 4, and 5 are self-contained.

Two djot-specific additions with no mdformat counterpart: protect attribute blocks
(`{.class #id key="value"}`) and inline math (`$...$`, `$$...$$`) as unwrappable, and
treat `:::` fences and `|` table rows as never-wrap.

One new option:

```toml
[lists]
align_continuation = false  # indent wrapped lines to the marker width
```

This is mdformat-mkdocs `--align-semantic-breaks-in-lists`. Optional because both
outputs are valid djot and the choice is taste.

### Sharing with mdformat-slw

Share the data and the tests, not the engine.

Sharing the engine through cgo/ctypes is available (djot-fmt already builds
`_libdjotfmt` and ships it in a wheel) and I would still not do it. The protection rules
are format-specific: mdformat-slw's protected regions encode CommonMark link syntax and
its `BLOCK_MARKER_PATTERN` encodes setext underlines and CommonMark bullets, while djot
needs attribute blocks, `:::`, and `$math$`. A shared engine carries both dialects and
each side pays for the other's syntax. mdformat-slw is also installed as a pre-commit
`additional_dependencies` entry, where a compiled dependency means wheels for every
platform and Python version plus an sdist fallback needing a Go toolchain. A rules DSL
has the same problem moved one level down: it needs an interpreter written twice, which
is more code than the 772 lines it would replace.

mdformat-slw owns both shared artifacts, and djot-fmt consumes them.

**The data.** `_language_data.py` is 327 lines across six languages, it is the most
valuable asset here, and it is genuinely static. Move it to
`slw_data/abbreviations.toml` in mdformat-slw and generate the Python from it, so the
Python package stays the source of truth rather than a second copy.

**The corpus.** A directory of parity cases, each with input, config, and expected
output, in a format neither language owns. Each repo writes a thin runner. This is what
actually proves parity, and it is the artifact that makes the "no shared engine" call
safe.

**The sync mechanism**, which is where "pull from main" needs care. Fetching at build
time puts the network in the build and makes it non-reproducible, and tracking `main`
means an upstream mid-flight commit can break djot-fmt's CI for reasons unrelated to
djot-fmt. Instead:

- djot-fmt vendors `internal/slw/data/abbreviations.toml` (with `go:embed`) and
  `testdata/slw-conformance/`, plus a `.slw-upstream` file recording the mdformat-slw
  tag it came from
- `mise run slw:sync` fetches that tag and rewrites the vendored copies
- A weekly CI job runs the sync against the newest tag and opens a PR when the diff is
  non-empty, so parity drift arrives as a reviewable change
- Per-PR CI verifies the vendored files still match the recorded tag, so a local edit to
  the vendored copy fails rather than silently forking

Track tags, not `main`. mdformat-slw cuts a tag when it wants djot-fmt to follow, which
also gives the two repos a way to disagree deliberately.

If the two implementations still agree after a year of the corpus, extracting a real
`slw-core` becomes defensible. Doing it first buys the packaging cost before knowing the
algorithms converge.

## M6: code block formatting

mdformat-hooks runs one shell command over the whole document, usually
`mdsf format --stdin`. djot-fmt should do both levels, because it already has the block
boundaries that mdsf has to rediscover. Per-language commands win over the mdsf
fallback, a block with no match is left alone, and the formatter's output is reindented
to the block's current column since a block inside a list item is not at column 0.

This is the only milestone that runs other people's code, so the rest of this section is
about that.

### The exposure

The risk is not the config format. TOML is inert data with no execution semantics, and
switching to JSON changes nothing. The problem is the combination of two things that are
individually reasonable:

1. The config value is a command djot-fmt executes
2. The config file is discovered by walking up from the input path

Put together, cloning a repository and opening a file with format-on-save runs whatever
that repository's `.djot-fmt.toml` says. No prompt, no diff, no review step. The attack
needs a pull request to a repo you already format, or a clone of anything.

This is a known class rather than something novel. direnv's `.envrc`, Vim modelines,
`.git/config` `core.fsmonitor`, and `.vscode/settings.json` have all shipped the same
shape. It is worth flagging that mdformat-hooks has this exact exposure today via
`post_command` in a discovered `.mdformat.toml`, which is your package, not a hypothetical.

Worth naming what does not fix it. A safer config format does not, since TOML is already
safe and the danger is the value's meaning. An allowlist of known formatter names does
not on its own, because resolving `ruff` through `PATH` in a repo that ships a
`node_modules/.bin` or a shim is still attacker-controlled. Sandboxing the subprocess
would work and is platform-specific enough that it is a project rather than a feature.

### What I would do instead

Trust, recorded per config file, in the direnv shape:

- Commands are read only from a config file whose path and content hash are in a
  user-level trust store (`~/.config/djot-fmt/trust.toml`)
- An untrusted or changed config with a `[code]` section is refused, the run continues
  with code formatting skipped, and stderr prints the offending commands plus
  `djot-fmt trust <path>` to allow them
- Editing the config invalidates the trust, so approval covers the content you read and
  not the file forever
- Everything else in the config (frontmatter, numbering, slw) stays freely discoverable,
  since none of it executes anything

Two smaller measures on top. Parse the command into argv and `exec` it directly with no
`sh -c`, so shell metacharacters cannot chain a second command even in a trusted config.
And skip discovery entirely in stdin mode, where there is no input path to anchor to and
the ambient working directory is a poor proxy.

The escape hatch for CI is `--allow-code-commands`, which is explicit, appears in the
job definition, and is reviewable in the same diff as any change to it.

## M7: Python API beyond `format()`

`djot_fmt` currently exports `format()` and `go_version()`. yak-shears reads djot with
regexes (`links.py`, `frontmatter.py`) because there is nothing else to use.

Add one cgo function, not a query API:

```c
char* DjotParseJSON(const char* input, char** errOut);
```

It serializes the godjot AST to JSON and Python walks it. One new export, no schema
negotiation across the boundary, and yak-shears builds whatever accessors it wants on
top. The existing `formatMu` mutex covers it, since the constraint is the tokenizer's
package-level map rather than the formatter.

On the Python side:

```py
djot_fmt.parse(text) -> Document          # frontmatter + node tree
djot_fmt.format(text, options=Options())  # replaces the growing keyword list
```

`Options` as a dataclass matters. `format()` already takes four keyword arguments and
this roadmap adds frontmatter, numbering, alignment, validation, and code settings.
A config object also means the CLI, the config file, and the Python API validate one
struct instead of three.

Open question before building: yak-shears' `[[wikilinks]]` and `#tags` are not djot
syntax. Either they stay regexes in yak-shears, or djot-fmt grows an extension parsing
them into span nodes. The second makes djot-fmt less general, so I lean toward the first
unless yak-shears needs them formatted rather than just read.

## Configuration

Discovery walks up from each input file for `.djot-fmt.toml` and stops at the first hit
or the repository root. `--config PATH` overrides discovery. Precedence is CLI flag,
then config file, then default.

### The opt-in pattern

Not everything optional is optional in the same way, and one flat `extensions = [...]`
list would flatten a distinction that matters. Three tiers:

**Safety, on by default, no opt-in.** Frontmatter detection and output validation. These
exist to stop djot-fmt corrupting a file, so making them opt-in means shipping the
corruption to everyone who did not read the docs. They get an opt-out flag
(`--no-validate`) rather than an opt-in.

**Style, always available, configured by value.** Numbering mode, continuation
alignment, and every slw knob. There is no enabling to do, only choosing, so these are
plain keys with defaults and no `enabled` field:

```toml
[lists]
numbering = "one"
align_continuation = false

[slw]
enabled = true
wrap = 88
min_line = 40
```

`slw.enabled` is the one apparent exception, and it is a value like the others: wrapping
has a genuine off state, unlike numbering.

**Extras, off by default, named explicitly.** Anything reaching outside djot-fmt. Right
now that is only code formatting, which is exactly the thing needing a deliberate act:

```toml
extras = ["code"]

[code]
mdsf = true
timeout = 30
strict = false

[code.formatters]
python = "ruff format -"
go = "gofmt"
```

The section alone does not enable it. Naming it in `extras` does, and for `code` that
still is not sufficient, since the trust store has to agree too. Section-presence-implies-
enabled was the alternative and it is too magic: you cannot leave a configured section
switched off, and adding a section becomes a behavior change that is easy to miss in
review.

Frontmatter normalization sits under the safety tier because detection is not optional,
while what to do beyond preserving bytes is:

```toml
[frontmatter]
normalize = "none"  # none | minimal | sort-keys
```

### What stays fixed

Configure what the spec leaves genuinely open or what djot does not define. Everything
else is opinionated and fixed: `_` for emphasis and `*` for strong, backtick fences at
the shortest length the content permits, `***` for thematic breaks, attribute ordering
inside `{...}`, and nested indent width. Each has one sensible answer, and a formatter
that lets you pick loses the property that makes it useful.

## Sequencing

M1 first. It is small, it is the safety net every other milestone leans on, and it
already earned its place by finding the nested-paragraph bug.

M2 second, since M5 is mostly blocked on it and M1 makes the refactor verifiable.

M3 in parallel with either. Frontmatter wraps the pipeline rather than living inside it,
so it does not conflict, and it unblocks yak-shears earliest.

M4 next, small and self-contained, with the roman question scoped before it starts.

M5 after M2. Begin with the shared TOML data and the conformance corpus in mdformat-slw,
since those set the target the Go code is written against.

M6 and M7 in either order. M7 is smaller and yak-shears wants it.

Configuration lands incrementally with each milestone that adds a key, and the file
format freezes once the option set stops moving.

## Open questions

1. Does yak-shears want `[[wikilinks]]` parsed by djot-fmt, or only read?
2. Is mdformat-slw's language data allowed to become generated-from-TOML, or does it
   need to stay hand-editable Python?
3. Roman list enumerators are broken in godjot itself. Upstream a fix, work around it
   locally, or accept the loss and document it?
4. Should `--check` grow a `--diff` split, so CI can print the diff without the exit
   code and vice versa?
