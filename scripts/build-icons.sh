#!/usr/bin/env bash
# Rebuild branded extension icons, Windows ICO and Windows executable resources.
# Prerequisites for maintainers: ffmpeg, Python 3, Go (module downloads permitted).
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p extension/icons
cp assets/reverfolio-icon.svg internal/echo/web/favicon.svg
for size in 16 32 48 128 256; do
  ffmpeg -hide_banner -loglevel error -y -i assets/reverfolio-icon.svg \
    -vf "scale=$size:$size:flags=lanczos" -frames:v 1 "extension/icons/icon$size.png"
done
python3 - <<'PY'
from pathlib import Path
import struct
directory=Path("extension/icons")
sizes=[16,32,48,256]
parts=[(n,(directory/f"icon{n}.png").read_bytes()) for n in sizes]
offset=6+16*len(parts)
records=[]
for n,data in parts:
    records.append(struct.pack("<BBBBHHII", 0 if n==256 else n,0 if n==256 else n,0,0,1,32,len(data),offset))
    offset+=len(data)
(directory/"reverfolio.ico").write_bytes(
    struct.pack("<HHH",0,1,len(parts))+b"".join(records)+b"".join(data for _,data in parts))
PY
go run github.com/akavel/rsrc@v0.10.2 -arch amd64 -ico extension/icons/reverfolio.ico -o cmd/reverfolio/rsrc_windows_amd64.syso
go run github.com/akavel/rsrc@v0.10.2 -arch arm64 -ico extension/icons/reverfolio.ico -o cmd/reverfolio/rsrc_windows_arm64.syso
echo "Reverfolio icons regenerated."
