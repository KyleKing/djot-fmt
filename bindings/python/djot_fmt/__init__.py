"""Format djot markup through the djot-fmt Go library."""

from ._core import (
    DjotFormatError,
    __version__,
    format,  # ruff: ignore[builtin-import-shadowing] - the module-qualified name djot_fmt.format is the public API
    go_version,
)

__all__ = ['DjotFormatError', '__version__', 'format', 'go_version']
