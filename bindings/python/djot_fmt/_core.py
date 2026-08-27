"""Format djot markup through the djot-fmt Go library."""

from __future__ import annotations

import ctypes
import sys
from pathlib import Path
from threading import Lock

__version__ = '0.2.0'

_LIB_STEM = '_libdjotfmt'
_LIB_SUFFIXES = {'darwin': '.dylib', 'win32': '.dll'}

_lib: ctypes.CDLL | None = None
_lib_lock = Lock()


class DjotFormatError(Exception):
    """Raised when the formatter rejects its input."""


def _library_path() -> Path:
    suffix = _LIB_SUFFIXES.get(sys.platform, '.so')
    return Path(__file__).parent / f'{_LIB_STEM}{suffix}'


def _load() -> ctypes.CDLL:
    """Open the shared library, binding argument and return types once.

    Loading is deferred to the first call because the Go runtime does not survive
    fork() without exec(). Importing this module stays safe for a process that
    forks; calling into it before forking is not.
    """
    global _lib  # ruff: ignore[global-statement]
    with _lib_lock:
        if _lib is not None:
            return _lib

        lib = ctypes.CDLL(str(_library_path()))
        lib.DjotFormat.argtypes = [
            ctypes.c_char_p,
            ctypes.c_int,
            ctypes.c_char_p,
            ctypes.c_int,
            ctypes.c_int,
            ctypes.POINTER(ctypes.c_char_p),
        ]
        lib.DjotFormat.restype = ctypes.c_void_p
        lib.DjotVersion.argtypes = []
        lib.DjotVersion.restype = ctypes.c_void_p
        lib.DjotFree.argtypes = [ctypes.c_void_p]
        lib.DjotFree.restype = None
        _lib = lib
        return lib


def _take(lib: ctypes.CDLL, pointer: int | None) -> str | None:
    """Copy a Go-allocated C string into Python and free the original."""
    if not pointer:
        return None
    try:
        raw = ctypes.cast(pointer, ctypes.c_char_p).value
        return raw.decode() if raw is not None else None
    finally:
        lib.DjotFree(pointer)


def format(  # ruff: ignore[builtin-variable-shadowing]
    text: str,
    *,
    wrap_sentences: bool = True,
    markers: str = '.!?',
    max_line_width: int = 88,
    min_line_length: int = 40,
) -> str:
    """Format djot source, applying semantic line wrapping unless disabled."""
    lib = _load()
    error = ctypes.c_char_p()
    pointer = lib.DjotFormat(
        text.encode(),
        int(wrap_sentences),
        markers.encode(),
        max_line_width,
        min_line_length,
        ctypes.byref(error),
    )
    message = _take(lib, ctypes.cast(error, ctypes.c_void_p).value)
    if message is not None:
        raise DjotFormatError(message)
    if pointer is None:
        raise DjotFormatError('formatter returned no output')
    return _take(lib, pointer) or ''


def go_version() -> str:
    """Return the version, commit, and build date stamped into the Go library."""
    lib = _load()
    return _take(lib, lib.DjotVersion()) or ''
