# Reverfolio v0.3.1

Fix dashboard interval details reading the wrong report field. Known download increments and unknown assets now display using `last_comparable_changes`, matching MCP and exports. Add a frontend contract test using the actual report shape.

Synchronize workspace navigation with browser history/hash changes, and cover back/forward behavior without double renders. The download UI contract check is now part of release packaging.

No schema or credential changes. All v0.3.0 features are retained.
