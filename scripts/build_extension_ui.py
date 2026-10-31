#!/usr/bin/env python3
"""Mirror the exact embedded OpenEcho UI into its own Edge/Chrome extension.

Use the same app logic/CSS, not a separately implemented launcher UI.
Static resources come from this repository; no external CDN dependency.
"""
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "internal/echo/web"
DEST = ROOT / "extension/ui"
DEST.mkdir(parents=True, exist_ok=True)

html = (SOURCE / "app.html").read_text(encoding="utf-8")
for src, dest in (
    ('href="/assets/favicon.svg"', 'href="./favicon.svg"'),
    ('href="/assets/app.css"', 'href="./app.css"'),
    ('src="/assets/app.js"', 'src="./app.js"'),
    ('src="/assets/settings.js"', 'src="./settings.js"'),
):
    if src not in html:
        raise SystemExit(f"Required UI resource is missing: {src}")
    html = html.replace(src, dest)
html = html.replace(
    '<title>OpenEcho',
    '<meta name="openecho-api-origin" content="http://127.0.0.1:19452">\n'
    '<title>OpenEcho',
    1,
)
html = html.replace(
    '<script defer src="./app.js"></script>',
    '<script defer src="./local-auth.js"></script>\n  '
    '<script defer src="./app.js"></script>',
    1,
)
(DEST / "index.html").write_text(html, encoding="utf-8")
for name in ("app.css", "app.js", "settings.js", "favicon.svg"):
    (DEST / name).write_bytes((SOURCE / name).read_bytes())
print("OpenEcho extension UI synchronized with main embedded Web UI")
