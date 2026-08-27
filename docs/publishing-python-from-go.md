# Publishing Python from Go

This project ships the same Go code twice: as a binary, and as a Python package on
PyPI. The Python side is a `ctypes` wrapper over a cgo `-buildmode=c-shared`
library. This file explains why that shape, the rules that keep it from crashing
the host interpreter, and what CI has to do.

## Where the boundary sits

`bindings/cshared` holds the only cgo in the repository. Everything reachable
from Python crosses through the `//export` functions in that package, and no
other package imports `"C"`. Keep the shim thin: it converts C types to Go types,
calls into `internal/`, and converts back. Logic that belongs to the project
belongs behind it, where the Go tests can reach it.

`bindings/python` holds the wrapper. `_core.py` owns the `ctypes` bindings and is
the only module that touches the library handle. `__init__.py` re-exports and
does nothing else, so importing the package never loads the library. The
argparse console script lives in `_cli.py` and calls the same public functions a
library caller would.

`pyproject.toml` at the repository root drives the build. The Go sources stay
where they are, so the binary and the wheel compile from one tree.

## Why c-shared and ctypes

The three ways to reach Go from Python are a subprocess, a WASM module, and an
in-process shared library. Measured on djot-fmt, formatting one small document:

| Approach | Per call | Artifact |
|----------|----------|----------|
| ctypes into a c-shared library | 8.7 us | 1.8 MB `.dylib` |
| subprocess to the CLI binary | 2973 us | binary of similar size |
| WASM module | slower than both | 3.1 MB `.wasm` |

A subprocess pays process creation, two pipe copies, and interpreter teardown on
every call, which is roughly 340x the in-process cost here. That is invisible for
one document and dominant for a linter walking a repository. WASM adds a runtime
dependency, a larger artifact, and a sandbox boundary that buys nothing when the
code is already trusted.

The decisive advantage is wheel tagging. `ctypes` and `cffi` call through the C
ABI and never touch the CPython ABI, so a wheel built this way carries no
`cp3XX` ABI tag. It is tagged `py3-none-<platform>`: one wheel per platform,
valid on every Python version the package supports, including free-threaded
builds. A CPython extension module would need a wheel per Python version per
platform, and a new one every October.

`ctypes` also releases the GIL for the duration of a foreign call. Threads
blocked in Go code run in parallel, so a `ThreadPoolExecutor` over this API gets
real concurrency rather than interleaving.

That parallelism is only safe if the Go code underneath is goroutine-safe, and
plenty of libraries are not. A package-level map written during parsing is the
common shape, and two Python threads reaching it at once abort the whole process
with `fatal error: concurrent map writes`, which no `recover()` can catch. Check
the library before advertising concurrency: run its parse or transform entry
point under `go test -race` from several goroutines. When it is not safe, hold a
`sync.Mutex` across the call in the shim and say so in the Python docstring.
Serializing costs throughput, and a crashed interpreter costs the process.
`djot-fmt` binds `godjot` this way for exactly this reason.

## Safety rules at the boundary

Break any of these and the failure mode is a killed interpreter or a slow leak,
never an exception.

Recover from panics inside every exported function. A panic that crosses the cgo
boundary has no Go frame left to unwind into, so the runtime aborts the process.
For a Python caller that means the interpreter dies with no traceback. Each
`//export` function starts with a deferred `recover()` that converts the panic
into a NULL return and an error string:

```go
//export Run
func Run(input *C.char, errOut **C.char) (result *C.char) {
	defer func() {
		if r := recover(); r != nil {
			result = nil
			setError(errOut, panicMessage(r))
		}
	}()
	...
}
```

The named return value matters. A deferred function can only overwrite a result
that has a name.

Never call `os.Exit` from library code. It terminates the host interpreter and
skips every `defer` on the way out. Error paths return a message through the
`errOut` pointer instead.

Free every C string the Go side allocates. `C.CString` calls `malloc`, and Go's
garbage collector does not own that memory. Export one free function, and have
the Python side copy the bytes and release the pointer in a `finally`:

```python
def _take(lib, pointer):
    if not pointer:
        return None
    try:
        raw = ctypes.cast(pointer, ctypes.c_char_p).value
        return raw.decode() if raw is not None else None
    finally:
        lib.Free(pointer)
```

Declare `restype = ctypes.c_void_p` for functions returning strings. With
`c_char_p`, ctypes copies the bytes and discards the pointer, and the allocation
leaks because there is nothing left to free.

Defer `dlopen` to the first call. The Go runtime does not survive `fork()`
without `exec()`: the child inherits one thread out of the parent's scheduler,
and any call into Go deadlocks or crashes. Loading the library lazily keeps
`import <package>` safe in a process that forks (Gunicorn preload, multiprocessing
with the fork start method, Celery). Calling into the library before forking is
still unsafe, and that is the caller's constraint to respect. Guard the lazy load
with a `threading.Lock` so two threads racing on the first call bind the argument
types once.

Two more consequences of loading the Go runtime into someone else's process:

- Go installs its own signal handlers when the library loads, and it uses SIGURG
  to preempt goroutines. A host that installs handlers of its own, or an
  extension that treats an unexpected SIGURG as fatal, sees traffic it did not
  expect. Debuggers and profilers attached to the interpreter report the same
  signal
- Go's scheduler starts threads that outlive any single call, so process-level
  thread counts and any code asserting on them will see more threads after the
  first call than before

## Building the wheel

The hatchling build hook shells out to `go build -buildmode=c-shared`, writes the
library next to the Python package, and then sets three things on `build_data`:

- `pure_python = False`, so hatchling emits a platform wheel rather than
  `py3-none-any`
- `tag = f'py3-none-{platform}'`, the ABI-free tag described above
- `force_include`, mapping the built library into the wheel

The macOS tag needs one correction. Since macOS 11 the ABI carries no minor
version, so installers only accept `macosx_15_0_arm64` and reject the
`macosx_15_4_arm64` that `sysconfig.get_platform()` reports on a 15.4 machine.
The hook rewrites the minor to `0` for major 11 and above.

Linux wheels are tagged `linux_*` at build time on purpose. The manylinux level a
wheel satisfies depends on the glibc symbols the library actually references,
which is only knowable by inspecting the finished binary. `auditwheel repair`
does that inspection and retags.

## CI shape

cgo cannot cross-compile without a target toolchain, so every wheel builds on a
runner of its own architecture and operating system. The matrix in
`.github/workflows/publish.yml` covers six targets:

| Target | Runner | Container |
|--------|--------|-----------|
| linux-x86_64 | ubuntu-24.04 | `quay.io/pypa/manylinux_2_28_x86_64` |
| linux-aarch64 | ubuntu-24.04-arm | `quay.io/pypa/manylinux_2_28_aarch64` |
| macos-x86_64 | macos-15-intel | none |
| macos-arm64 | macos-15 | none |
| windows-x86_64 | windows-2025 | none |
| windows-arm64 | windows-11-arm | none |

The matrix is not permanent. `macos-15-intel` is the last x86_64 macOS image
GitHub provides, available until August 2027, after which no runner can build the
macos-x86_64 wheel natively.

The Linux jobs run inside manylinux containers because the wheel must link
against a glibc old enough for the distributions users run. Building on the bare
runner links against its glibc and produces a wheel that fails to load on
anything older.

macOS jobs export `MACOSX_DEPLOYMENT_TARGET=13.0` before building. The wheel tag
claims a minimum macOS version, and without the pin the toolchain targets the
runner's own release, so the tag promises support the binary does not have.

Publishing uses PyPI trusted publishing over OIDC. The job requests
`permissions: id-token: write` and runs in a GitHub environment bound to the
PyPI publisher, so there is no long-lived API token in repository secrets.

`publish.yml` is triggered by `workflow_dispatch` and takes the tag to publish.
`bump_version.yml` dispatches it with `gh workflow run` after cutting a tag,
which works because `workflow_dispatch` is the one event `GITHUB_TOKEN` may
raise. Two constraints force this shape rather than a tag trigger or a
`workflow_call`:

- a tag pushed with `GITHUB_TOKEN` starts no workflow run, so an on-tag trigger
  never fires for an automated release
- PyPI validates the attestation against the top-level `workflow_ref`, so
  publishing from inside a called workflow fails with a Build Config URI
  mismatch even when the OIDC exchange succeeds

Versions are semver (`v0.1.0-rc.2`), because goreleaser and the Go module proxy
both reject PEP 440 prereleases like `0.1.0rc2`. Python normalizes semver to
`0.1.0rc2` for the distribution, so the tag, `__version__`, and the installed
version differ in form and must be compared through
`packaging.utils.canonicalize_version`.

## Verify after publishing, on every platform

A wheel that builds is not a wheel that runs. Three failures reach PyPI looking
healthy and only surface on install:

- a wheel tagged for a platform it cannot run on, which pip either refuses or
  installs into a broken import
- a shared library missing a symbol, or linked against a glibc newer than the
  installing machine has
- a console script entry point that resolves to nothing, because the package
  imports but the script name is wrong

None of these fail the build job. The `verify` matrix runs after publish on all
six platforms, installs from the real index with `uvx`, runs the console script,
and diffs its output against an expected file. It also imports the package and
asserts that `__version__` and the version stamped into the Go library both match
the tag, which catches a release built from a stale checkout.

PyPI serves its index through a CDN that lags an upload by minutes, so the verify
job retries the download before failing.

## Releasing

The version lives in three files and commitizen bumps all three together:
`.cz.toml`, `pyproject.toml`, and `__version__` in the package's `_core.py`. The
Go library gets its version through `-ldflags -X main.version=...` at build time,
read from the package metadata. Never edit any of them by hand, and after a
`copier update` check that `.cz.toml` still lists all three, since the template
re-renders that file.

## Local tasks

```bash
mise run python:build   # build the wheel, compiling the c-shared library
mise run python:test    # pytest against the built library
mise run python:lint    # ruff
mise run python:types   # ty
```

The pre-commit hooks run ruff and pytest on any change under `bindings/`.
