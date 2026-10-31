#!/usr/bin/env python3
"""Build portable platform archives without ever including live SQLite or local secrets."""
import hashlib
import os
import pathlib
import shutil
import subprocess
import sys
import tarfile
import tempfile
import zipfile

ROOT = pathlib.Path(__file__).resolve().parents[1]
DIST = ROOT / 'dist'
if len(sys.argv) != 2 or not sys.argv[1].startswith('v') or sys.argv[1].count('.') != 2:
    raise SystemExit('Usage: python3 scripts/package_release.py v0.2.0')
version = sys.argv[1]
if DIST.exists() and any(DIST.iterdir()):
    raise SystemExit('Refusing to overwrite existing release artifacts in dist/')
DIST.mkdir(exist_ok=True)
subprocess.run(['python3', 'scripts/build_extension_ui.py'], cwd=ROOT, check=True)
subprocess.run(['node', 'scripts/check_extension.mjs'], cwd=ROOT, check=True)
subprocess.run(['go', 'test', './...'], cwd=ROOT, check=True)
subprocess.run(['python3', 'scripts/check_docs.py'], cwd=ROOT, check=True)
assets = []
for operating_system in ('windows', 'linux', 'darwin'):
    for architecture in ('amd64', 'arm64'):
        with tempfile.TemporaryDirectory(prefix='reverfolio-package-') as temporary:
            root = pathlib.Path(temporary) / 'Reverfolio'
            (root / 'bin').mkdir(parents=True)
            binary = root / 'bin' / ('reverfolio.exe' if operating_system == 'windows' else 'reverfolio')
            env = os.environ.copy()
            env.update(GOOS=operating_system, GOARCH=architecture, CGO_ENABLED='0')
            subprocess.run(['go', 'build', '-trimpath', '-ldflags=-s -w', '-o', str(binary), './cmd/reverfolio'],
                           cwd=ROOT, env=env, check=True)
            if operating_system == 'windows':
                shutil.copytree(ROOT / 'extension', root / 'extension')
                for script in ('Start-Reverfolio.cmd', 'Run-Reverfolio.ps1',
                               'Install-Reverfolio.ps1', 'Initialize-ReverfolioExtension.ps1'):
                    shutil.copy2(ROOT / 'scripts/windows' / script, root / script)
            for name in ('config.example.toml', 'README.md', 'README.zh-CN.md',
                         'LICENSE', 'THIRD_PARTY_NOTICES.md', 'SECURITY.md'):
                shutil.copy2(ROOT / name, root / name)
            (root / 'docs').mkdir()
            shutil.copy2(ROOT / 'docs/windows-layout.md', root / 'docs/windows-layout.md')
            prefix = 'Reverfolio-' + version + '-' + operating_system + '-' + architecture
            if operating_system == 'windows':
                asset = DIST / (prefix + '.zip')
                with zipfile.ZipFile(asset, 'w', compression=zipfile.ZIP_DEFLATED, compresslevel=8) as z:
                    for path in sorted(root.rglob('*')):
                        if path.is_file():
                            z.write(path, path.relative_to(root.parent))
            else:
                asset = DIST / (prefix + '.tar.gz')
                with tarfile.open(asset, 'w:gz') as tf:
                    tf.add(root, arcname='Reverfolio')
            assets.append(asset)
            print('PACKAGED', operating_system, architecture, asset.stat().st_size, flush=True)
checksum_file = DIST / 'SHA256SUMS'
checksum_file.write_text(''.join(
    hashlib.sha256(path.read_bytes()).hexdigest() + '  ' + path.name + '\n'
    for path in sorted(assets)), encoding='ascii')
print('PORTABLE_RELEASE_ASSETS_COMPLETE', len(assets), flush=True)
