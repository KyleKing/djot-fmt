# Next steps

## godjot's tokenizer is not safe for concurrent use

`djot_tokenizer.MatchInlineToken` writes the package-level `StartSymbols` map on
every successful match (`djot_inline_token.go:28`), guarded only by the
`RecordStartSymbol` const, which is `true`. Two goroutines parsing at once corrupt
it. Reached from `BuildDjotAst`, so it affects any concurrent caller, not just
tests.

Proven on 2026-08-31: `go test -count=1 ./...` died with a runtime
`fatal error: concurrent map writes` inside `MatchInlineToken`, and
`go test -count=1 -race ./internal/validate` reproduced it on every run with both
stacks on the same address through `mapaccess2`.

`internal/djotsafe` now serializes every parse behind a mutex, and the two call
sites (`validate.renderHTML` and `iohelper.processFile`) go through it.
`go test -race ./...` is clean.

Two things still owed:

- Fix it upstream in godjot and drop `internal/djotsafe`. `StartSymbols` looks
  like instrumentation left switched on, so making `RecordStartSymbol` a var that
  defaults to false, or guarding the write, would do it
- Add `-race` to the `ci` task. No mise task passes it today, which is why a
  fatal runtime error sat in a dependency unnoticed
