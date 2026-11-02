#!/usr/bin/env python3
"""Mirror the exact embedded Reverfolio UI into its own Edge/Chrome extension.

Use the same app logic/CSS, not a separately implemented launcher UI.
Static resources come from this repository; no external CDN dependency.
"""
from pathlib import Path
import re
import json

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "internal/echo/web"
listen=re.search(r'(?m)^listen\s*=\s*"([^"]+)"',(ROOT/"config.example.toml").read_text()).group(1)
origin="http://"+listen
DEST = ROOT / "extension/ui"
DEST.mkdir(parents=True, exist_ok=True)

html = (SOURCE / "app.html").read_text(encoding="utf-8")
for src, dest in (
    ('href="/assets/favicon.svg"', 'href="./favicon.svg"'),
    ('src="/assets/favicon.svg"', 'src="./favicon.svg"'),
    ('href="/assets/app.css"', 'href="./app.css"'),
    ('src="/assets/app.js"', 'src="./app.js"'),
    ('src="/assets/dashboard.js"', 'src="./dashboard.js"'),
    ('src="/assets/settings.js"', 'src="./settings.js"'),
):
    if src not in html:
        raise SystemExit(f"Required UI resource is missing: {src}")
    html = html.replace(src, dest)
html = html.replace(
    '<title>Reverfolio',
    f'<meta name="reverfolio-api-origin" content="{origin}">\n'
    '<title>Reverfolio',
    1,
)
html = html.replace(
    '<script defer src="./app.js"></script>',
    '<script defer src="./local-auth.js"></script>\n  '
    '<script defer src="./app.js"></script>',
    1,
)
(DEST / "index.html").write_text(html, encoding="utf-8")
for name in ("app.css", "app.js", "settings.js", "dashboard.js", "favicon.svg"):
    (DEST / name).write_bytes((SOURCE / name).read_bytes())
print("Reverfolio extension UI synchronized with main embedded Web UI")

manifestPath=ROOT/"extension/manifest.json"
manifest=json.loads(manifestPath.read_text())
manifest["host_permissions"]=[origin+"/*"]
manifest["content_security_policy"]["extension_pages"]="script-src 'self'; object-src 'self'; connect-src 'self' "+origin
manifestPath.write_text(json.dumps(manifest,indent=2)+"\n")
