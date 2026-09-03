## v0.2.4 (2026-09-03)

### Fix

- **cli**: write LF output on Windows instead of CRLF

## v0.2.3 (2026-09-03)

### Fix

- **ci**: correct the Publish verify fixture and Python smoke check

## v0.2.2 (2026-08-31)

### Fix

- **scripts**: generate the tap deploy key inside 1Password

## v0.2.1 (2026-08-31)

### Fix

- serialize godjot parsing, which races on a package-level map

## v0.2.0 (2026-08-27)

### Feat

- **slw**: buffer inline output per block so wrapping sees the whole line
- **validate**: reject formatting that changes what a document means

### Fix

- use portable arithmetic in the PyPI wait loop for Git Bash

## v0.1.2 (2026-08-05)

### Fix

- probe PyPI availability with a command uv actually has

## v0.1.1 (2026-08-05)

### Fix

- dispatch publish.yml so PyPI attestations validate, and tag with semver

## v0.1.0 (2026-08-05)

## v0.1.0rc2 (2026-08-05)

### Feat

- allow bump_version to cut a prerelease on manual dispatch

### Fix

- publish Go binaries and the Python package from one called workflow

## v0.1.0rc1 (2026-08-05)

### Fix

- build wheels on manylinux containers and windows arm64

## v0.1.0rc0 (2026-08-05)

### Feat

- publish djot-fmt to PyPI through cgo bindings
- address bugs and gaps
- expand write implementation
- support multiple input files
- implement semantic line wrapping (SLW) (#1)
- init djot formatter

### Fix

- **hk**: skip the commitizen branch check on an empty rev-range
- address space issues
- format custom attributes
- finish addressing low priority bugs
- prevent silently dropping unknown nodes
- write to output file when headers
- resolve linting errors
- rename for consistency with GitHub repo

### Refactor

- cleanup overall code quality
