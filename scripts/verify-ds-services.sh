#!/usr/bin/env bash
set -euo pipefail

BASE_URL=${1:-http://127.0.0.1:7777}
AUTH=${DS_TOKEN:-}

hdrs=("-H" "Accept: application/json")
if [[ -n "$AUTH" ]]; then
  hdrs+=("-H" "Authorization: Bearer $AUTH")
fi

echo "Probing ds services at $BASE_URL"
curl -fsS "${hdrs[@]}" "$BASE_URL/v1/health" | jq . >/dev/null && echo "✓ /v1/health"
curl -fsS "${hdrs[@]}" "$BASE_URL/api/self-status" | jq -e '.ok == true and (.nowMs|type=="number")' >/dev/null && echo "✓ /api/self-status"
curl -fsS "${hdrs[@]}" "$BASE_URL/v1/capabilities" | jq . >/dev/null && echo "✓ /v1/capabilities"
curl -fsS "${hdrs[@]}" "$BASE_URL/openapi.yaml" >/dev/null && echo "✓ /openapi.yaml"
curl -fsS "${hdrs[@]}" "$BASE_URL/api/discovery/openapi" >/dev/null && echo "✓ /api/discovery/openapi"
curl -fsS "${hdrs[@]}" "$BASE_URL/.well-known/obs-bridge.json" | jq -e '.endpoints.openapi and .endpoints.capabilities and .endpoints.health' >/dev/null && echo "✓ /.well-known/obs-bridge.json"
curl -fsS "${hdrs[@]}" "$BASE_URL/api/discovery/capabilities" | jq . >/dev/null && echo "✓ /api/discovery/capabilities"
curl -fsS "${hdrs[@]}" "$BASE_URL/api/discovery/services" | jq -e '.ds.self_status and .ds.openapi and .ds.capabilities and .ds.health and (.ds_token_present|type=="boolean") and (.ts|type=="number")' >/dev/null && echo "✓ /api/discovery/services"

echo "All probes completed."
