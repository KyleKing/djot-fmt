# Roadmap

Where djot-fmt goes from `v0.1.0rc1` to a formatter that yak-shears can depend on,
ordered by what blocks what. Validation and the inline buffer have landed; what is
below has not.

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
| Loose list item, `--slw-min-line 0` | Continuation line at column 0 | Legal djot (lazy continuation), reads badly in a diff |
| ` ``` python ` block | Fence normalized, body untouched | No code formatting yet |

Two spec points I checked rather than assumed, from `jgm/djot doc/syntax.md`:

- djot allows lazy continuation, so the unindented list continuation above is valid.
  It is a style question, not a correctness bug
- djot defines no frontmatter. Any support here is an extension, which is exactly why
  it belongs behind configuration

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

Roman enumerators and the delimiter are both upstream problems, written up in
[docs/upstream-godjot.md](docs/upstream-godjot.md). godjot never recognizes roman at all,
and it tracks the delimiter well enough to split lists on a style change but does not
record it on the node. Delimiter preservation depends on that upstream fix or on
pre-scanning the source. M1 will flag any half-fix as a validation failure rather than
letting it through.

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

The pipeline now mirrors mdformat-slw: collapse spaces, protect regions, break
sentences against the accumulated output line, wrap to width, pull block markers off
wrapped line starts. Every stage is checked against real mdformat-slw output. What is
left:

| mdformat-slw stage | Go status |
| --- | --- |
| Abbreviation suppression | 20 hardcoded English entries vs 6 language lists |
| Display width | Rune count, so CJK and emoji measure short |

Two djot-specific additions with no mdformat counterpart: attribute blocks
(`{.class #id key="value"}`) and math (`$...$`, `$$...$$`) are already protected, and
`:::` fences and `|` table rows never reach the wrapper because they are not paragraphs.

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

Copy `mise trust`, which already draws the line in the right place. Its own help text
states the principle:

> Safe config files do not require trust: files that only contain `min_version`,
> `[tools]` entries with plain version strings (or arrays of them), and `[tasks]` (no
> templates and no tool options) are loaded without prompting, since nothing in them
> executes code at load time.

So trust is keyed on what the file contains, not on the file existing. Applied here:

- A config with no `[code]` section is loaded without prompting, always. Frontmatter,
  numbering, and slw settings execute nothing, so gating them on trust would train
  people to approve reflexively, which is how a trust prompt stops working
- A config with a `[code]` section is checked against a user-level store of path and
  content hash. Untrusted means the run continues with code formatting skipped and
  stderr naming the commands it refused
- Editing the config invalidates the trust, so approval covers content that was read
  rather than the file forever

Worth taking from mise beyond the core idea: `djot-fmt trust --show` to report the trust
state of every config on the path from cwd upward, and `--ignore` to record a permanent
deny so a repo you do not control stops asking. Its CI behavior is the one piece I would
invert. mise assumes trust in detected CI, which suits a tool people invoke deliberately.
A formatter running in CI over a pull request from a fork is the exact scenario this
protects against, so require the explicit flag there.

Two smaller measures on top. Parse the command into argv and `exec` it directly with no
`sh -c`, so shell metacharacters cannot chain a second command even in a trusted config.
And skip discovery entirely in stdin mode, where there is no input path to anchor to and
the ambient working directory is a poor proxy.

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

M3 first. Frontmatter wraps the pipeline rather than living inside it,
so it does not conflict, and it unblocks yak-shears earliest.

M4 next, small and self-contained, with the roman question scoped before it starts.

M5 begins with the shared TOML data and the conformance corpus in mdformat-slw, since
those set the target the Go code is written against.

M6 and M7 in either order. M7 is smaller and yak-shears wants it.

Configuration lands incrementally with each milestone that adds a key, and the file
format freezes once the option set stops moving.

## Open questions

1. Does yak-shears want `[[wikilinks]]` parsed by djot-fmt, or only read?
2. Is mdformat-slw's language data allowed to become generated-from-TOML, or does it
   need to stay hand-editable Python?
3. Both godjot findings in [docs/upstream-godjot.md](docs/upstream-godjot.md): upstream a
   fix, work around locally, or accept the loss and document it? The delimiter one is
   cheap enough that it is probably worth a PR regardless
4. Should `--check` grow a `--diff` split, so CI can print the diff without the exit
   code and vice versa?
5. Validation compares godjot's derived `<section id>`, so normalizing whitespace inside
   a heading is reported as a difference. Preserving an authored `{#custom-id}` and
   ignoring a derived one are indistinguishable in the HTML. Reject is the safe default
   and the trigger is obscure (`# ` followed by a lazy continuation line), so this stands
   until someone hits it
