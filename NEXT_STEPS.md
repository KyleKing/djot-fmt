# Next steps

## godjot's tokenizer is not safe for concurrent use

`djot_tokenizer.MatchInlineToken` writes a package-level map on every call
(`djot_inline_token.go:28`, reached through `BuildDjotTokens` and `BuildDjotAst`).
Two goroutines parsing at once corrupt it. This is a defect in
`github.com/sivukhin/godjot/v2`, not here, but it reaches any caller of
`validate.Compare` or the formatter that runs in parallel.

Proven on 2026-08-31: `go test -count=1 ./...` died with a runtime
`fatal error: concurrent map writes` inside `MatchInlineToken`, and
`go test -count=1 -race ./internal/validate` reproduces it on every run, with both
stacks landing on the same address through `mapaccess2`. `internal/formatter` shows
the same warning.

Nothing catches it today because no mise task passes `-race`. The tasks are
`ci`, `test`, `verify-released`, and `coverage`, all plain `go test`.

Three ways out, in order of preference:

- Fix it upstream in godjot and pin the release. The map looks like a lazily
  populated lookup table, so a `sync.Once` or a precomputed table would do it
- Serialize godjot behind a mutex in `internal/validate` and the formatter, which
  costs parallelism in exactly the hot path
- Add `-race` to the `ci` task so it fails loudly. Do this only after one of the
  above, because it turns CI red immediately

Whichever lands, add `-race` to `ci` afterwards so a regression cannot go quiet
again.
