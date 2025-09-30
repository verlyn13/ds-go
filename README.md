# ds

Dead simple Git repository manager. Fast status checks across multiple accounts.

## Install

```bash
git clone https://github.com/verlyn13/ds-go.git
cd ds-go
make install
```

## Usage

```bash
ds status         # show all repos
ds status -d      # show dirty repos only
ds fetch          # update remote info
ds scan           # rebuild index
ds organize --plan   # preview repo moves (use --json for machine output)
ds organize --require-clean  # enforce no uncommitted changes
ds policy check --json --fail-on critical  # policy/compliance gate
ds exec -a verlyn13 -- 'mise run lint'     # run command across repos
ds serve --addr 127.0.0.1:7777             # start local API for agents
```

## API Server

The `ds serve` command starts an HTTP API server with contract guarantees:

### Contract Version
- **Schema Version**: `ds.v1`
- **Contract Version**: `v1.1.0` (frozen as of 2025-09-29)

### Key Features
- All endpoints return `schema_version: "ds.v1"`
- Status endpoints wrap arrays in `{schema_version, data}` envelope
- Self-status includes `nowMs` as epoch milliseconds
- Discovery endpoints at `/.well-known/obs-bridge.json` and `/api/discovery/services`
- Optional authentication with `DS_TOKEN` environment variable
- CORS support with `DS_CORS=1` environment variable

### Example
```bash
# Start with authentication and CORS
DS_TOKEN=secret DS_CORS=1 ds serve --addr 127.0.0.1:7777

# Verify endpoints
curl -H "Authorization: Bearer secret" http://127.0.0.1:7777/v1/health
curl -H "Authorization: Bearer secret" http://127.0.0.1:7777/api/self-status
```

See [Contract Documentation](docs/contracts/VERSION.md) for full API details.

## Config

Creates `~/.config/ds/config.yaml` on first run:

```yaml
base_dir: ~/Projects

accounts:
  verlyn13:
    type: personal
    ssh_host: github-personal
  jjohnson-47:
    type: school
    ssh_host: github-work
```

## Build

```bash
make build        # build binary
make test         # run tests
make install      # install to /usr/local/bin
```

## Local Agent API

Start the API:

```bash
ds serve --addr 127.0.0.1:7777
```

Endpoints (JSON):

- GET `/v1/capabilities` — list supported endpoints and schema version
- GET `/v1/health` — basic health with uptime, workers, auth-enabled
- GET `/v1/status?dirty=true&account=verlyn13&path=~/Projects` — repo status with filters
- GET `/v1/status/stream` — NDJSON stream of repositories
- GET `/v1/status/sse` — Server-Sent Events stream of repositories
- GET `/v1/scan?path=~/Projects` — scan and update index, returns count
- GET `/v1/organize/plan?require_clean=true` — list planned moves
- POST/GET `/v1/organize/apply?require_clean=true&force=false&dry_run=false` — apply organize plan
- GET `/v1/fetch?account=verlyn13` — fetch remotes for filtered repos
- GET `/v1/fetch/sse?account=verlyn13` — SSE streaming of fetch results
- GET `/v1/policy/check?file=.project-compliance.yaml&fail_on=high` — run policy checks
- POST `/v1/exec?account=verlyn13&dirty=false&timeout=30` with JSON `{ "cmd": "mise run lint" }` — run a command across repos

Discovery:
- GET `/openapi.yaml` — OpenAPI 3.1 spec (also `/api/discovery/openapi`)
- GET `/api/discovery/capabilities` — minimal discovery metadata
- GET `/.well-known/obs-bridge.json` — well-known bridge descriptor

Notes:
- Use `--json` on CLI commands for machine output; `ds status --exit-on-dirty` exits 10 when dirty is found.
- Organize supports `--plan` (no changes) and `--require-clean` for safety.

MIT License

## Ports & Env Conventions

- DS (Agent B): runs on port 7777 by default.
  - Base URL env: `DS_BASE_URL` (default `http://127.0.0.1:7777`)
  - Auth env: `DS_TOKEN` (adds `Authorization: Bearer`)
- Do not use Bridge/MCP ports in DS scripts/docs.
  - Bridge (Agent A) defaults to 7171 (OBS_BRIDGE_URL) and MCP (Agent D) to 4319 (MCP_URL/MCP_BASE_URL), but DS materials must reference DS only.
\n## Ports & Env Conventions (Required)

- Canonical DS port: `7777`\n- Use `DS_BASE_URL` (default `http://127.0.0.1:7777`) and `DS_TOKEN` where required.\n- Do not use Bridge/MCP ports in DS scripts or docs.

See policy: docs/policies/ports-and-env.md
