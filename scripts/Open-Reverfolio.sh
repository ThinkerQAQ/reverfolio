#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")" && pwd)"
cd "$root"
# noclobber keeps an existing configuration, including during concurrent launches.
if [[ ! -f config.toml ]]; then
 (set -o noclobber; cat config.example.toml > config.toml) || true
 chmod 600 config.toml
fi
mkdir -p Data
chmod 700 Data
addr="$(sed -n 's/^listen *= *"\(127\.0\.0\.1:[0-9]*\|\[::1\]:[0-9]*\)".*/\1/p' config.toml | head -1)"
[[ -n "$addr" ]] || { echo "Expected loopback listen address" >&2; exit 1; }
url="http://$addr"
if ! curl --fail --silent --max-time 2 "$url/healthz" >/dev/null; then
 nohup "$root/bin/reverfolio" --config "$root/config.toml" >Data/service.log 2>&1 &
fi
for ((i=0;i<40;i++)); do
 if curl --fail --silent --max-time 1 "$url/healthz" >/dev/null; then
  if [[ "$(uname -s)" == Darwin ]]; then open "$url"; else xdg-open "$url"; fi
  exit 0
 fi
 sleep 0.25
done
echo "Reverfolio did not become ready; inspect Data/service.log" >&2
exit 1
