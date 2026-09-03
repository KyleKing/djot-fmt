"""Command line entry point mirroring the djot-fmt Go binary."""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

from ._core import DjotFormatError, __version__
from ._core import format as format_djot


def _parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog='djot-fmt', description='Format djot markup files.')
    parser.add_argument('files', nargs='*', help='files to format (reads stdin when omitted)')
    parser.add_argument('-V', '--version', action='version', version=f'djot-fmt {__version__}')
    parser.add_argument('-w', '--write', action='store_true', help='rewrite files in place')
    parser.add_argument('-c', '--check', action='store_true', help='exit non-zero if reformatting is needed')
    parser.add_argument('-o', '--output', help='write output to this file')
    parser.add_argument('--no-wrap-sentences', action='store_true', help='disable semantic line wrapping')
    parser.add_argument('--slw-markers', default='.!?', help='sentence-ending characters')
    parser.add_argument('--slw-wrap', type=int, default=88, help='maximum line width')
    parser.add_argument('--slw-min-line', type=int, default=40, help='minimum line length before wrapping')
    return parser


def _validate(args: argparse.Namespace, parser: argparse.ArgumentParser) -> None:
    if args.write and args.output:
        parser.error('cannot use both -w and -o')
    if args.write and not args.files:
        parser.error('-w requires at least one input file (cannot use with stdin)')
    if args.output and len(args.files) > 1:
        parser.error('-o can only be used with a single input file')
    if args.check and (args.write or args.output):
        parser.error('-c cannot be used with -w or -o')


def main(argv: list[str] | None = None) -> int:
    parser = _parser()
    args = parser.parse_args(argv)
    _validate(args, parser)

    # Without this, Windows text-mode writes turn the formatter's LF output into CRLF.
    sys.stdout.reconfigure(newline='')

    options = {
        'wrap_sentences': not args.no_wrap_sentences,
        'markers': args.slw_markers,
        'max_line_width': args.slw_wrap,
        'min_line_length': args.slw_min_line,
    }
    read = [(Path(name), Path(name).read_text(encoding='utf-8')) for name in args.files]
    sources = read or [(None, sys.stdin.read())]

    needs_format = False
    for path, source in sources:
        try:
            formatted = format_djot(source, **options)  # type: ignore[arg-type]
        except DjotFormatError as exc:
            label = path or '<stdin>'
            print(f'{label}: {exc}', file=sys.stderr)
            return 1

        if args.check:
            if formatted != source:
                needs_format = True
                print(f'{path or "<stdin>"} is not formatted', file=sys.stderr)
        elif args.write and path is not None:
            path.write_text(formatted, encoding='utf-8', newline='')
        elif args.output:
            Path(args.output).write_text(formatted, encoding='utf-8', newline='')
        else:
            sys.stdout.write(formatted)

    return 1 if needs_format else 0


if __name__ == '__main__':
    sys.exit(main())
