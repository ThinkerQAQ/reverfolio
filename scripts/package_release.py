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
import json
manifest_version=json.loads((ROOT / 'extension' / 'manifest.json').read_text())['version']
if version.removeprefix('v') != manifest_version:
    raise SystemExit(f'Release/extension mismatch: {version} != {manifest_version}')
if DIST.exists() and any(DIST.iterdir()):
    raise SystemExit('Refusing to overwrite existing release artifacts in dist/')
DIST.mkdir(exist_ok=True)
subprocess.run(['python3', 'scripts/build_extension_ui.py'], cwd=ROOT, check=True)
subprocess.run(['node', 'scripts/check_extension.mjs'], cwd=ROOT, check=True)
subprocess.run(['node', 'scripts/check_download_ui.mjs'], cwd=ROOT, check=True)
subprocess.run(['node', 'scripts/check_dashboard.mjs'], cwd=ROOT, check=True)
subprocess.run(['node', 'scripts/check_propagation.mjs'], cwd=ROOT, check=True)
subprocess.run(['node', 'scripts/check_impact.mjs'], cwd=ROOT, check=True)
subprocess.run(['node', 'scripts/check_discovery.mjs'], cwd=ROOT, check=True)
subprocess.run(['node', 'scripts/check_native_host.mjs'], cwd=ROOT, check=True)
subprocess.run(['node', 'scripts/check_mcp_guide.mjs'], cwd=ROOT, check=True)
subprocess.run(['go', 'test', './...'], cwd=ROOT, check=True)
subprocess.run(['python3', 'scripts/check_docs.py'], cwd=ROOT, check=True)
# Ship actual dependency license texts, not only a notice pointing upstream.
# This is a generated distribution artifact and must not enter the source tree.
license_temp = tempfile.TemporaryDirectory(prefix='reverfolio-licenses-')
licenses = pathlib.Path(license_temp.name) / 'THIRD_PARTY_LICENSES.txt'
subprocess.run(['python3', 'scripts/bundle_dependency_licenses.py', '--output', str(licenses)],
               cwd=ROOT, check=True)
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
            if operating_system != 'windows':
                subprocess.run(['go','build','-trimpath','-ldflags=-s -w',
                                '-o',str(root/'bin/reverfolio-host'),'./cmd/reverfolio-host'],
                               cwd=ROOT,env=env,check=True)
            if operating_system == 'windows':
                # The browser launches the host in GUI subsystem mode, preventing
                # a flashing console. It starts the separate service hidden.
                host = root / 'bin' / 'reverfolio-host.exe'
                subprocess.run(['go','build','-trimpath','-ldflags=-s -w -H=windowsgui',
                                '-o',str(host),'./cmd/reverfolio-host'],cwd=ROOT,env=env,check=True)

            shutil.copytree(ROOT / 'extension', root / 'extension')
            for name in ('config.example.toml', 'README.md', 'README_ZH.md', 'LICENSE'):
                shutil.copy2(ROOT / name, root / name)
            shutil.copy2(licenses, root / 'THIRD_PARTY_LICENSES.txt')
            (root / 'scripts').mkdir()
            shutil.copytree(ROOT / 'third_party', root / 'third_party')
            shutil.copy2(ROOT / 'scripts/publications_to_catalog.py', root / 'scripts/publications_to_catalog.py')
            shutil.copy2(ROOT / 'scripts/check_mcp_contract.py', root / 'scripts/check_mcp_contract.py')
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
subprocess.run(['python3', 'scripts/check_release_assets.py', version, '--dist', str(DIST)],
               cwd=ROOT, check=True)
license_temp.cleanup()
