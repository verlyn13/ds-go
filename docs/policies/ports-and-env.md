# Ports & Environment Policy (Agent B — DS CLI)

Canonical Defaults
- DS (this repo): `http://127.0.0.1:7777`
  - `DS_BASE_URL` defaults to `http://127.0.0.1:7777`
  - `DS_TOKEN` adds `Authorization: Bearer` headers

Scope & Rules
- DS scripts and docs must not default to Bridge or MCP ports (7171 or 4319).
- Cross‑repo documentation may mention Bridge (7171) or MCP (4319) for context, but DS run commands, examples, and scripts must use DS ports/env only.
- Prefer loopback addresses for local development.

Quick Reference (system‑wide)
- Bridge (Agent A): `http://127.0.0.1:7171` via `OBS_BRIDGE_URL`
- MCP (Agent D): `http://127.0.0.1:4319` via `MCP_URL`/`MCP_BASE_URL`
- DS (Agent B): `http://127.0.0.1:7777` via `DS_BASE_URL`

Enforcement
- CI runs `scripts/validate-conventions.mjs` to ensure DS materials adhere to these ports/env conventions.
