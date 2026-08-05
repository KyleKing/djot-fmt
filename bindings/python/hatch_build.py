"""Build the Go c-shared library into the wheel as a platform-tagged artifact."""

from __future__ import annotations

import os
import re
import subprocess
import sysconfig
from pathlib import Path
from typing import Any

from hatchling.builders.hooks.plugin.interface import BuildHookInterface

LIB_STEM = '_libdjotfmt'
MACOS_SINGLE_VERSION_MAJOR = 11
SHARED_LIB_SUFFIXES = {'darwin': '.dylib', 'windows': '.dll'}
PACKAGE_DIR = Path('bindings') / 'python' / 'djot_fmt'
CSHARED_PKG = './bindings/cshared'


def _goos() -> str:
    if goos := os.environ.get('GOOS'):
        return goos
    platform = sysconfig.get_platform()
    if platform.startswith('macosx'):
        return 'darwin'
    return 'windows' if platform.startswith('win') else 'linux'


def _wheel_platform_tag() -> str:
    """Return the local platform tag.

    Linux wheels are tagged `linux_*` on purpose. `auditwheel repair` retags them to
    the manylinux level the build actually satisfies, which cannot be known here.
    """
    tag = sysconfig.get_platform().replace('-', '_').replace('.', '_')
    match = re.fullmatch(r'macosx_(\d+)_\d+_(\w+)', tag)
    if match and int(match[1]) >= MACOS_SINGLE_VERSION_MAJOR:
        # macOS 11 dropped minor versions from the ABI, so installers only accept `_0`
        return f'macosx_{match[1]}_0_{match[2]}'
    return tag


class GoSharedLibraryHook(BuildHookInterface):
    """Compile the Go shared library and mark the wheel as platform specific."""

    PLUGIN_NAME = 'go-shared-library'

    def initialize(self, version: str, build_data: dict[str, Any]) -> None:  # ruff: ignore[unused-method-argument]
        """Build the library into the package directory before the wheel is assembled."""
        if self.target_name != 'wheel':
            return

        suffix = SHARED_LIB_SUFFIXES.get(_goos(), '.so')
        output = Path(self.root) / PACKAGE_DIR / f'{LIB_STEM}{suffix}'
        output.parent.mkdir(parents=True, exist_ok=True)

        ldflags = ' '.join(
            [
                '-s',
                '-w',
                f'-X main.version={self.metadata.version}',
                f'-X main.commit={os.environ.get("GIT_COMMIT", "none")}',
                f'-X main.date={os.environ.get("BUILD_DATE", "unknown")}',
            ],
        )
        # ruff: ignore[subprocess-without-shell-equals-true, start-process-with-partial-path]
        # the Go toolchain is resolved from PATH by mise or the CI runner
        subprocess.run(
            [
                'go',
                'build',
                '-buildmode=c-shared',
                '-trimpath',
                '-ldflags',
                ldflags,
                '-o',
                str(output),
                CSHARED_PKG,
            ],
            check=True,
            cwd=self.root,
            env={**os.environ, 'CGO_ENABLED': '1'},
        )
        output.with_suffix('.h').unlink(missing_ok=True)

        build_data['pure_python'] = False
        build_data['tag'] = f'py3-none-{_wheel_platform_tag()}'
        build_data['force_include'][str(output)] = f'djot_fmt/{output.name}'
