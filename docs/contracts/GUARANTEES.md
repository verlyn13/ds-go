# DS API Contract Guarantees

## Version: v1.1.0 (Frozen 2025-09-29)

This document provides comprehensive contract guarantees for the ds-go API.

## Core Contract Principles

1. **No Breaking Changes**: Once frozen, the v1.1.0 contract cannot be broken without a major version bump (v2.0.0)
2. **Consistent Versioning**: All endpoints return `schema_version: "ds.v1"`
3. **Predictable Envelopes**: Array responses are wrapped in `{schema_version, data}`
4. **Backwards Compatibility**: New fields may be added (non-breaking), but existing fields cannot be removed or changed

## Endpoint-Specific Guarantees

### 1. Versioned Response Endpoints

All endpoints MUST include `schema_version: "ds.v1"` in their response.

| Endpoint | Response Structure | Code Reference |
|----------|-------------------|----------------|
| `/v1/capabilities` | `{schema_version, version, endpoints, ...}` | server.go:56 |
| `/v1/health` | `{schema_version, ok, version, ...}` | server.go:70 |
| `/v1/status` | `{schema_version, data: [...]}` | server.go:174, writeJSONVersioned |
| `/v1/scan` | `{schema_version, count}` | server.go:218 |
| `/v1/organize/plan` | `{schema_version, data: [...]}` | server.go:233 |
| `/v1/organize/apply` | `{schema_version, moved, failed, ...}` | server.go:250 |
| `/v1/fetch` | `{schema_version, results}` | server.go:268 |
| `/v1/policy/check` | `{schema_version, report, ...}` | server.go:302 |
| `/v1/exec` | `{schema_version, results}` | server.go:329 |

### 2. Self-Status Guarantees

`/api/self-status` endpoint (server.go:135):

**Required Fields:**
- `service`: string = "ds"
- `ok`: boolean
- `nowMs`: number (epoch milliseconds)
- `schema_version`: string = "ds.v1"
- `auth`: object containing:
  - `tokenRequired`: boolean
  - `corsEnabled`: boolean
- `endpoints`: object with normalized paths including:
  - `well_known`: string
  - `openapi`: string
  - `capabilities`: string
  - `health`: string

### 3. Discovery Guarantees

#### `/api/discovery/services` (server.go:89)

**Required Fields:**
- `ds`: object containing:
  - `url`: string (base URL)
  - `well_known`: string (well-known path)
  - `openapi`: string (OpenAPI path)
  - `capabilities`: string
  - `health`: string
  - `self_status`: string
- `ds_token_present`: boolean
- `ts`: number (epoch milliseconds)

#### `/.well-known/obs-bridge.json` (server.go:107)

**Required Fields:**
- `contractVersion`: number
- `schemaVersion`: string
- `openapi_url`: string
- `capabilities_url`: string
- `endpoints`: object with:
  - `openapi`: string
  - `capabilities`: string
  - `health`: string
- `all`: array of endpoint paths

### 4. Response Wrapping Rules

The `writeJSONVersioned` function (server.go:436-450) enforces:

1. **Maps/Objects**: If the response doesn't have `schema_version`, it's injected
2. **Arrays**: Always wrapped as `{schema_version: "ds.v1", data: [...]}`
3. **Other Types**: Wrapped as `{schema_version: "ds.v1", data: <value>}`

### 5. Streaming Endpoint Exceptions

These endpoints stream data without version wrapping:
- `/v1/status/sse` - Server-Sent Events
- `/v1/status/stream` - NDJSON stream
- `/v1/fetch/sse` - Server-Sent Events

**Rationale**: Streaming protocols have their own framing; wrapping would break standard consumers.

## Authentication & CORS

### Authentication
- When `DS_TOKEN` environment variable is set, all endpoints require `Authorization: Bearer <token>`
- The `wrapAuth` middleware (server.go:391-403) enforces this

### CORS
- When `DS_CORS=1`, permissive CORS headers are added (server.go:407-417)
- Allows cross-origin requests from dashboards

## Contract Enforcement

### CI/CD Pipeline

The contract is enforced through multiple layers:

1. **Build & Test** (`.github/workflows/ci.yml`)
   - `go build ./...` - Ensures code compiles
   - `go test ./...` - Runs all unit tests

2. **OpenAPI Validation**
   - Redocly linting (`.redocly.yaml`)
   - Spectral linting (`.spectral.yml`)
   - Swagger CLI validation

3. **Contract Validation**
   - Live endpoint testing
   - Schema version consistency checks
   - Response envelope validation

4. **Code Review** (`.github/CODEOWNERS`)
   - Critical paths require review
   - `internal/server/server.go` - Contract implementation
   - `internal/server/openapi.yaml` - Contract definition

### Validation Scripts

- `scripts/validate-contract.sh` - Comprehensive contract validation
- `scripts/verify-stage0.sh` - Stage 0 requirements
- `scripts/verify-ds-services.sh` - Service endpoint verification

## Breaking Change Policy

### What Constitutes a Breaking Change

1. **Removing** any field from a response
2. **Changing** the type of any field
3. **Changing** the structure of response envelopes
4. **Removing** any endpoint
5. **Changing** authentication requirements (making stricter)

### What's Allowed (Non-Breaking)

1. **Adding** new optional fields to responses
2. **Adding** new endpoints
3. **Adding** new optional query parameters
4. **Extending** enums with new values
5. **Adding** new examples or documentation

### Deprecation Process

1. **Announce** deprecation in CHANGELOG
2. **Mark** deprecated in OpenAPI spec
3. **Maintain** for minimum 2 major versions
4. **Remove** only in major version bump

## Version History

### v1.1.0 (2025-09-29) - Contract Freeze
- Initial contract freeze
- All endpoints versioned with `schema_version: "ds.v1"`
- Array responses wrapped in envelopes
- Complete discovery system implemented

## Testing Contract Compliance

### Manual Testing
```bash
# Run comprehensive validation
./scripts/validate-contract.sh

# Test specific endpoint
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:7777/v1/health | \
  jq '.schema_version'
```

### Automated Testing
```bash
# Run CI locally
go test ./...
npm install -g @redocly/cli @stoplight/spectral-cli
spectral lint internal/server/openapi.yaml
redocly lint internal/server/openapi.yaml
```

## Support

For contract-related questions or issues:
- GitHub Issues: https://github.com/verlyn13/ds-go/issues
- Contract Documentation: This file
- OpenAPI Spec: `/internal/server/openapi.yaml`

---

**This contract is legally binding for the v1.x release series.**