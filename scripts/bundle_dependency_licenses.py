#!/usr/bin/env python3
"""Bundle upstream Go-module license texts into every portable release.

Uses the resolved module graph, not an incomplete hand-curated list.
Includes build/test-only modules (safe superset), preserves exact source texts,
and fails closed when a module has no top-level license notice.
"""
import argparse
import json
from pathlib import Path
import subprocess


def module_inventory():
    payload = subprocess.check_output(['go', 'list', '-m', '-json', 'all'], text=True)
    decoder = json.JSONDecoder()
    results = []
    while payload.strip():
        obj, offset = decoder.raw_decode(payload.lstrip())
        payload = payload.lstrip()[offset:]
        if not obj.get('Main'):
            results.append(obj)
    return sorted(results, key=lambda item: item['Path'])


def licensed_files(directory):
    prefixes = ('LICENSE', 'COPYING', 'NOTICE', 'COPYRIGHT')
    allowed = []
    for p in sorted(directory.iterdir()):
        if not p.is_file():
            continue
        name = p.name.upper()
        if not name.startswith(prefixes):
            continue
        # Source files such as copyright_test.go are not license notices.
        if p.suffix.lower() in ('.go', '.py', '.js', '.c', '.h'):
            continue
        if p.stat().st_size > 3 * 1024 * 1024:
            raise ValueError(f'Unexpectedly large notice: {directory.name}/{p.name}')
        allowed.append(p)
    return allowed


def build_bundle():
    modules = module_inventory()
    sections = [
        'REVERFOLIO - THIRD-PARTY DEPENDENCY LICENSES\n',
        'This file reproduces original upstream license/notice texts from the\n',
        'resolved Go module graph at packaging time. This graph includes some\n',
        'build/test-only modules, so it is an intentional superset of modules\n',
        'linked into the application. Refer to go.mod and go.sum for versions.\n',
        'The Reverfolio application itself is MIT licensed (see LICENSE).\n',
    ]
    for module in modules:
        root = Path(module.get('Dir', ''))
        if not root.is_dir():
            raise RuntimeError(f'Module sources unavailable: {module["Path"]}')
        licenses = licensed_files(root)
        if not licenses:
            raise RuntimeError(f'No root-level license for {module["Path"]}@{module["Version"]}')
        for license_file in licenses:
            content = license_file.read_bytes().decode('utf-8-sig')
            sections += [
                '\n' + '=' * 76 + '\n',
                f'Module: {module["Path"]}\n',
                f'Version: {module["Version"]}\n',
                f'Upstream notice: {license_file.name}\n',
                '=' * 76 + '\n',
                content.rstrip('\n') + '\n',
            ]
    return ''.join(sections), len(modules)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--output', type=Path, help='Write bundled license texts to this path')
    parser.add_argument('--check', action='store_true', help='Verify all upstream notices are available')
    args = parser.parse_args()
    if not args.output and not args.check:
        parser.error('Specify --output or --check')
    result, count = build_bundle()
    if args.output:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(result, encoding='utf-8')
    print(f'RESOLVED_MODULE_LICENSES_OK modules={count} bundle_bytes={len(result.encode("utf-8"))}')


if __name__ == '__main__':
    main()
