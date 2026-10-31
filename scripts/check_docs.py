"""Check repository Markdown links without network requests or dependencies."""
from pathlib import Path
import re
import sys

root = Path(__file__).resolve().parents[1]
errors = []
for document in root.rglob("*.md"):
    if ".git" in document.parts:
        continue
    text = document.read_text(encoding="utf-8")
    if sum(line.startswith("```") for line in text.splitlines()) % 2:
        errors.append(f"{document.relative_to(root)}: unclosed code fence")
    for target in re.findall(r"(?<!!)\[[^\]]+\]\(([^)]+)\)", text):
        if target.startswith(("https://", "http://", "mailto:", "#")):
            continue
        path = target.split("#", 1)[0]
        if path and not (document.parent / path).is_file():
            errors.append(f"{document.relative_to(root)}: missing link {target}")
if errors:
    print("\n".join(errors), file=sys.stderr)
    sys.exit(1)
print("Documentation links and code fences passed.")
