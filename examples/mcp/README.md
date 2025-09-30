# MCP Integration with ds

This repo includes an example MCP configuration to register ds as an HTTP-backed toolset for agents.

Quick start
1) Start ds with auth (recommended):
   DS_TOKEN=yourtoken ds serve --addr 127.0.0.1:7777

2) Copy and edit the example config:
   cp examples/mcp/config.example.toml ~/.config/devops-mcp/ds-tools.toml
   # Set base_url if different, and ensure your MCP runtime loads this file

3) Export DS_TOKEN so the MCP runtime can attach Authorization:
   export DS_TOKEN=yourtoken

4) Verify tools work via curl:
   curl -H "Authorization: Bearer $DS_TOKEN" 'http://127.0.0.1:7777/v1/status?dirty=true' | jq
   curl -H "Authorization: Bearer $DS_TOKEN" 'http://127.0.0.1:7777/v1/fetch/sse?account=verlyn13'

Tool mapping summary (from examples/mcp/config.example.toml)
- status → GET /v1/status?account=&dirty=&path=
- scan → GET /v1/scan?path=
- fetch → GET /v1/fetch?account=&dirty=&path=
- fetch_sse → GET /v1/fetch/sse (SSE stream)
- organize_plan → GET /v1/organize/plan?require_clean=&path=
- organize_apply → POST /v1/organize/apply?require_clean=&force=&dry_run=&path=
- policy_check → GET /v1/policy/check?file=&fail_on=
- exec → POST /v1/exec?account=&dirty=&timeout=&path= body { cmd }

Notes
- You can add an `envelope=true` query parameter to endpoints for an explicit envelope `{ schema_version, data }`.
- If your MCP runtime needs a typed client, you can generate types from `internal/server/openapi.yaml` or use the dashboard Ts client in `examples/dashboard/types/ds.ts` as reference.

