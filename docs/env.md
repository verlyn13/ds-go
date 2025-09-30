# ds Local API Environment

This document summarizes environment variables and conventions for the ds local API, and the canonical defaults used across repos.

- DS address
  - Prefer CLI flag: `ds serve --addr 127.0.0.1:7777`
  - Convention: bind to loopback only for local agents.

- Concurrency
  - Global: `--workers` flag controls concurrent operations (default 10).

- JSON and exit codes
  - Use `--json` for machine output.
  - `ds status --exit-on-dirty` exits `10` when dirty repos exist.
  - `ds policy check --fail-on <sev>` exits `20` on threshold failures.
  - `ds exec` exits `30` when any repo command fails.

- Response shape and envelope
  - All major DS endpoints include a top-level `schema_version: "ds.v1"` for contract-first typing.
  - Endpoints that return lists (e.g., `/v1/status`, `/v1/organize/plan`) return `{ schema_version, data: [...] }`.
  - `envelope=true` is accepted but no longer required; responses already include `schema_version`.

- Paths and data
  - Config: `${XDG_CONFIG_HOME}/ds/config.yaml` (see `ds init`).
  - Index: `${BaseDir}/.ds-index.json`, Fetch cache: `${BaseDir}/.ds-fetch-cache.json`.
  - BaseDir defaults to `~/Projects` unless overridden by config.

- Integration endpoints
  - OpenAPI: `GET /openapi.yaml` (also `/api/discovery/openapi`).
  - Capabilities: `GET /v1/capabilities` includes `openapi_url`.
  - Discovery: `GET /api/discovery/capabilities` for minimal metadata.

Cross‑Repo Environment Defaults
- Bridge (Agent A): `http://127.0.0.1:7171`
  - Env: `OBS_BRIDGE_URL` (preferred). Server may also accept `PORT` to override.
- MCP Server (Agent D): `http://127.0.0.1:4319`
  - Env: `MCP_URL` (scripts), `MCP_BASE_URL` (codegen).
- DS CLI (Agent B): `http://127.0.0.1:7777`
  - Env: `DS_BASE_URL`, optional `DS_TOKEN` for Authorization.

- Security (optional)
  - Set `DS_TOKEN` to require `Authorization: Bearer $DS_TOKEN` for all endpoints.
  - Or run `ds serve --token <token>` to override env.
