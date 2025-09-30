# DS CLI Integration Checklist

## Stage 0 Requirements
- ✅ `schema_version: "ds.v1"` on core endpoints
- ✅ `/api/self-status` includes `nowMs:number`
- ✅ Discovery present: `/.well-known/obs-bridge.json`, `/api/discovery/services`
- ✅ Go client `pkg/dsclient` + example & tests present

## Steps

1) Start DS server (secure mode with CORS)

```bash
DS_TOKEN=<token> DS_CORS=1 mise run serve-secure
# Or directly:
DS_TOKEN=<token> DS_CORS=1 go run ./cmd/ds serve --addr 127.0.0.1:7777
```

2) Verify discovery and well-known endpoints

```bash
# Well-known bridge descriptor
curl -sS http://127.0.0.1:7777/.well-known/obs-bridge.json | jq .

# Discovery services
curl -sS http://127.0.0.1:7777/api/discovery/services | jq '.ds, .ts'
```

3) Verify self-status and capabilities

```bash
# Self-status with nowMs
curl -sS http://127.0.0.1:7777/api/self-status | jq '.nowMs, .schema_version'

# Health check
curl -sS http://127.0.0.1:7777/v1/health | jq '.ok, .schema_version'

# Capabilities with schema version
curl -sS http://127.0.0.1:7777/v1/capabilities | jq '.schema_version, .endpoints'
```

4) Run example Go client

```bash
# Without token
go run ./examples/go-client

# With token (if server requires auth)
DS_TOKEN=<token> go run ./examples/go-client
```

5) Run client tests

```bash
# Run package tests
go test ./pkg/dsclient/...

# Run with verbose output
go test -v ./pkg/dsclient/...
```

6) Optional: Verify contracts script (if present)

```bash
# If script exists
if [ -f ./scripts/verify-ds-services.sh ]; then
  ./scripts/verify-ds-services.sh http://127.0.0.1:7777
fi
```

## Validation Checklist

- [ ] Server starts successfully with token and CORS
- [ ] `/.well-known/obs-bridge.json` returns valid JSON
- [ ] `/api/discovery/services` includes `ds` object and `ts` timestamp
- [ ] `/api/self-status` includes `nowMs` as millisecond timestamp
- [ ] `/api/self-status` includes `schema_version: "ds.v1"`
- [ ] `/v1/health` includes `schema_version: "ds.v1"`
- [ ] `/v1/capabilities` includes `schema_version: "ds.v1"`
- [ ] Example client runs successfully and shows all sections
- [ ] Client tests pass (`go test ./pkg/dsclient/...`)

