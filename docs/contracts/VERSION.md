# DS CLI Contract Version

## Current Version: v1.1.0

**Frozen on**: 2025-09-29
**Stage**: Stage 1 - Contract Freeze & CI Gates

## Contract Details

### API Version
- Schema Version: `ds.v1`
- OpenAPI Version: `1.0.0`
- Contract Version: `1.1.0`

### Frozen Endpoints

All endpoints are frozen with the following contract guarantees:

#### Core Endpoints
- `/v1/capabilities` - Returns `schema_version: "ds.v1"`
- `/v1/health` - Returns `schema_version: "ds.v1"`
- `/v1/status` - Returns wrapped `{schema_version: "ds.v1", data: [...]}`
- `/v1/scan` - Returns `{schema_version: "ds.v1", ...}`
- `/v1/organize/plan` - Returns wrapped response
- `/v1/organize/apply` - Returns versioned response
- `/v1/fetch` - Returns versioned response
- `/v1/policy/check` - Returns versioned response
- `/v1/exec` - Returns versioned response

#### Discovery Endpoints
- `/.well-known/obs-bridge.json` - Bridge descriptor
- `/api/discovery/services` - Service registry with `ts` and `ds_token_present`
- `/api/discovery/capabilities` - Minimal capabilities
- `/api/discovery/openapi` - OpenAPI spec alias
- `/api/self-status` - Self-status with `nowMs` and `schema_version`

#### SSE Endpoints (Streaming, not wrapped)
- `/v1/status/sse` - Repository status stream
- `/v1/fetch/sse` - Fetch progress stream
- `/v1/status/stream` - NDJSON stream

### Response Format Guarantees

1. **Versioned Responses**: All non-streaming endpoints return `schema_version: "ds.v1"`
2. **Wrapped Arrays**: Status and plan endpoints wrap arrays in `{schema_version, data}`
3. **Timestamps**: All timestamps use epoch milliseconds (`nowMs`, `ts`)
4. **Authentication**: Bearer token support when `DS_TOKEN` is set
5. **CORS**: Enabled when `DS_CORS=1` is set

### Breaking Change Policy

From v1.1.0 onwards:
- No breaking changes to frozen endpoints
- New endpoints may be added (non-breaking)
- Optional fields may be added to responses (non-breaking)
- Deprecated features will be maintained for at least 2 major versions

### OpenAPI Specification

The complete OpenAPI specification is available at:
- `/openapi.yaml` (main spec)
- `/internal/server/openapi.yaml` (internal implementation)

### Validation

Contract compliance can be validated using:
```bash
# Stage 0 verification
./scripts/verify-stage0.sh

# DS services verification
./scripts/verify-ds-services.sh

# Cross-repo validation (from Agent A)
DS_BASE_URL=http://127.0.0.1:7777 DS_TOKEN=<token> node scripts/ds-validate.mjs
```

## Version History

### v1.1.0 (2025-09-29)
- Initial contract freeze for Stage 1
- All Stage 0 requirements implemented
- Discovery endpoints added
- Schema versioning standardized
- SSE endpoints operational

### v1.0.0 (Initial)
- Basic repository management API
- Core status and fetch operations

## Contact

Maintained by Agent B Team
Repository: https://github.com/verlyn13/ds-go