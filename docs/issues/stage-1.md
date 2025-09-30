# Stage 1 — Contract Freeze & CI Gates

Epic: <link to MVP Orchestration Epic>
Owners: @AgentA @AgentB @AgentC @AgentD

## Entrance Criteria (READY)
- [x] Agent A Stage 0 complete (Bridge endpoints, aliases, SSE, tools, schemas, OpenAPI)
- [x] DS validation script present (scripts/ds-validate.mjs)
- [x] MCP alias endpoints planned and runbook present
- [x] Orchestration scaffolds applied across repos

## Agent A — Bridge/Contracts
- [ ] Freeze contracts (OpenAPI + JSON Schemas) and annotate version (contracts/VERSION)
- [ ] Ensure all example timestamps use epoch ms consistently
- [ ] CI gates enforced on PRs:
  - [ ] Ajv schema validation
  - [ ] OpenAPI lint
  - [ ] Endpoint validation (health, discovery, well-known, tools)

## Agent B — DS CLI
- [ ] `/api/self-status` includes `schema_version: "ds.v1"` and `nowMs:number`
- [ ] `/v1/health` and `/v1/capabilities` available and versioned
- [ ] Discovery present: `/.well-known/obs-bridge.json`, `/api/discovery/services`
- [ ] Readme/docs note `schema_version` + envelope behavior

## Agent C — Dashboard
- [ ] Docs page links to discovery, openapi, registry
- [ ] Contracts page: ETag-aware schema fetch + raw JSON toggle

## Agent D — MCP
- [ ] CI alias parity tests for `/api/obs/*`
- [ ] Endpoint smoke: discovery services + openapi + self-status (`schemaVersion`, `nowMs`)

## Validation Steps

```
# DS validation
DS_BASE_URL=http://127.0.0.1:7777 DS_TOKEN=<token> node scripts/ds-validate.mjs

# MCP smoke
curl -sS http://127.0.0.1:4319/api/obs/discovery/services | jq '.ts|type'
curl -sS http://127.0.0.1:4319/api/obs/discovery/openapi | head -n 3
curl -sS http://127.0.0.1:4319/api/self-status | jq '.schemaVersion, .nowMs'
```

## Acceptance
- [ ] Contracts frozen and tagged
- [ ] CI gates enforced across repos
- [ ] DS and MCP validations pass

