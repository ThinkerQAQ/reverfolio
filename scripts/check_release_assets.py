#!/usr/bin/env python3
"""Verify portable release completeness, checksums and privacy boundaries."""
import argparse
import hashlib
from pathlib import Path
import tarfile
import zipfile


PLATFORMS = {f'{os_name}-{arch}' for os_name in ('windows', 'linux', 'darwin')
             for arch in ('amd64', 'arm64')}


def validate_archive(archive, platform):
    if archive.suffix == '.zip':
        with zipfile.ZipFile(archive) as z:
            names = z.namelist()
            read = z.read
            assert z.testzip() is None, f'Corrupt zip: {archive}'
            return validate_members(names, read, archive, platform)
    with tarfile.open(archive, 'r:gz') as t:
        names = t.getnames()
        def read(name):
            file = t.extractfile(name)
            assert file is not None, f'Missing file: {name}'
            return file.read()
        return validate_members(names, read, archive, platform)


def validate_members(names, read, archive, platform):
    paths = set(names)
    required = {
        'Reverfolio/LICENSE', 'Reverfolio/THIRD_PARTY_LICENSES.txt',
        'Reverfolio/README.md', 'Reverfolio/README_ZH.md',
        'Reverfolio/config.example.toml',
    }
    if platform.startswith('windows'):
        required |= {'Reverfolio/bin/reverfolio.exe', 'Reverfolio/bin/reverfolio-host.exe',
                     'Reverfolio/Install-Reverfolio.ps1', 'Reverfolio/Start-Reverfolio.cmd',
                     'Reverfolio/extension/manifest.json', 'Reverfolio/extension/ui/app.js'}
    else:
        required.add('Reverfolio/bin/reverfolio')
    missing = required - paths
    if missing:
        raise AssertionError(f'{archive.name}: required assets missing: {sorted(missing)}')
    for name in names:
        chunks = Path(name).parts
        if not name.startswith('Reverfolio/') and name != 'Reverfolio':
            raise AssertionError(f'{archive.name}: unexpected archive root: {name}')
        if '..' in chunks or 'Data' in chunks or 'data' in chunks:
            raise AssertionError(f'{archive.name}: suspicious/private path: {name}')
        if name.endswith(('config.toml', '.sqlite', '.sqlite3', '.db', '.secret', '.env')):
            raise AssertionError(f'{archive.name}: contains private config/data: {name}')
    text = read('Reverfolio/THIRD_PARTY_LICENSES.txt').decode('utf-8')
    if text.count('Module: ') < 30 or 'modernc.org/sqlite' not in text or 'github.com/modelcontextprotocol/go-sdk' not in text:
        raise AssertionError(f'{archive.name}: incomplete dependency license bundle')
    return len(paths)


def check(dist, version):
    manifest = dist / 'SHA256SUMS'
    if not manifest.is_file():
        raise AssertionError('Missing SHA256SUMS')
    checksums = {}
    for line in manifest.read_text(encoding='ascii').splitlines():
        digest, filename = line.split('  ', 1)
        if filename in checksums or '/' in filename or '\\' in filename:
            raise AssertionError(f'Unexpected asset filename: {filename}')
        checksums[filename] = digest
    if len(checksums) != 6:
        raise AssertionError(f'Expected six platform archives; found {len(checksums)}')
    seen = set()
    for name, digest in checksums.items():
        prefix = f'Reverfolio-{version}-'
        if not name.startswith(prefix):
            raise AssertionError(f'Version mismatch: {name}')
        rest = name[len(prefix):]
        platform = rest.removesuffix('.tar.gz').removesuffix('.zip')
        if platform not in PLATFORMS or platform in seen:
            raise AssertionError(f'Duplicate or unknown platform {platform}')
        seen.add(platform)
        expect_ext = '.zip' if platform.startswith('windows') else '.tar.gz'
        if not name.endswith(expect_ext):
            raise AssertionError(f'Unexpected archive type: {name}')
        archive = dist / name
        if not archive.is_file() or hashlib.sha256(archive.read_bytes()).hexdigest() != digest:
            raise AssertionError(f'Checksum mismatch: {name}')
        n = validate_archive(archive, platform)
        print(f'RELEASE_ASSET_OK {platform} files={n}')
    if seen != PLATFORMS:
        raise AssertionError('Release missing platform/architecture')
    print('PORTABLE_RELEASE_VALIDATED 6 targets; checksums, licenses and privacy gates PASS')


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('version', help='Version name including v, e.g. v0.9.2')
    parser.add_argument('--dist', type=Path, default=Path('dist'))
    args = parser.parse_args()
    check(args.dist, args.version)
