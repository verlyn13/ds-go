# SSE Validation (Agent B — DS CLI)

This guide shows how to validate DS Server‑Sent Events (SSE) locally.

Endpoints
- Status SSE: `GET /v1/status/sse` (emits `repo` events with Repository JSON)
- Fetch SSE: `GET /v1/fetch/sse` (emits `fetch` events with FetchResult JSON)

Quick smoke test
```bash
DS_BASE_URL=http://127.0.0.1:7777 DS_TOKEN=... node scripts/sse-validate.mjs
```
- Env:
  - `SSE_TIMEOUT_MS` (default 5000)
  - `SSE_MAX_EVENTS` (default 1)

What it checks
- Status SSE: at least one event contains required Repository fields (Path, Name, Account, IsClean, Ahead, Behind, HasUpstream)
- Fetch SSE: if events are present, they include required fields (RepoName, Success, Duration)

Notes
- For fetch SSE to emit events, a prior fetch must have run (`ds fetch`) or repositories with remotes must be available.
- SSE is text/event-stream; the script parses `event:` and `data:` lines and validates JSON.

