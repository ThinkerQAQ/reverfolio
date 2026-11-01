# Reverfolio v0.6.1

Patch release for the existing v0.6.0 impact dashboard.

- Filter cumulative snapshot charts by the actual last 7/14 observed calendar days, rather than the last 7/14 individual observations; missing measurements stay missing.
- Display interval-based known release download increases alongside unknown asset changes. Downloads remain a proxy, not user/installation counts.
- Limit article charts to exact, attributed page measurements instead of showing unrelated providers.
- Include explanatory source/scope notes in SVG and PNG exports and scale PNG height to avoid clipping.
- Keep the Chromium extension UI byte-identical to the embedded Web UI; validate scope, date-range, export and failed-source behavior.

No schema changes. Upgrading the Windows installation **must** retain the existing `Data/`, `config.toml`, installed extension access key, and same extension identity. Never commit private metrics or provider secrets.
