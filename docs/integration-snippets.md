# DS Integration Snippets

Curl with Authorization

```bash
export DS_BASE_URL=${DS_BASE_URL:-http://127.0.0.1:7777}
export DS_TOKEN=yourtoken

curl -H "Authorization: Bearer $DS_TOKEN" "$DS_BASE_URL/v1/health" | jq
curl -H "Authorization: Bearer $DS_TOKEN" "$DS_BASE_URL/v1/status" | jq
curl -H "Authorization: Bearer $DS_TOKEN" "$DS_BASE_URL/api/self-status" | jq
```

Envelope and schema_version

- All major endpoints include `schema_version: "ds.v1"`.
- Array endpoints wrap payload in `{ schema_version, data: [...] }`.
- `envelope=true` remains accepted but is optional.

Generating typed clients

```bash
./scripts/generate-openapi-client.sh examples/dashboard/generated/ds-client internal/server/openapi.yaml
```

