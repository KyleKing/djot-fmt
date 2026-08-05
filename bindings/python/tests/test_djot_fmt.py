from __future__ import annotations

import subprocess
import sys

import pytest

import djot_fmt


def test_format_normalizes_list_spacing() -> None:
    assert djot_fmt.format('- a\n- b\n') == '- a\n- b\n'


def test_format_is_idempotent() -> None:
    once = djot_fmt.format('# Title\n\nA paragraph.\n')
    assert djot_fmt.format(once) == once


def test_format_accepts_empty_input() -> None:
    """Empty input yields a bare newline, matching the Go binary."""
    assert djot_fmt.format('') == '\n'


def test_format_round_trips_unicode() -> None:
    assert 'héllo' in djot_fmt.format('héllo wörld\n')


@pytest.mark.parametrize('wrap_sentences', [True, False])
def test_wrap_sentences_toggle_returns_text(wrap_sentences: bool) -> None:
    source = 'One sentence here. Another sentence follows it. A third one closes.\n'
    assert djot_fmt.format(source, wrap_sentences=wrap_sentences)


def test_go_version_is_stamped() -> None:
    assert len(djot_fmt.go_version().split()) == 3


def test_cli_formats_stdin() -> None:
    result = subprocess.run(
        [sys.executable, '-m', 'djot_fmt._cli'],
        input='- a\n- b\n',
        capture_output=True,
        text=True,
        check=True,
    )
    assert result.stdout == '- a\n- b\n'


def test_cli_check_flags_unformatted_input() -> None:
    result = subprocess.run(
        [sys.executable, '-m', 'djot_fmt._cli', '--check'],
        input='-  a\n',
        capture_output=True,
        text=True,
        check=False,
    )
    assert result.returncode == 1
