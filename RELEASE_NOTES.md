# Release Notes - v1.1.0

## Contract Freeze Release
**Release Date**: September 29, 2025
**Type**: Major Release with Breaking Changes

## 🚨 Breaking Changes Alert

This release introduces **breaking changes** to API response formats. Please review the migration guide below before upgrading.

### Key Breaking Changes

1. **Response Format Changes**
   - `/v1/status` now returns `{schema_version, data}` instead of bare array
   - `/v1/organize/plan` now returns `{schema_version, data}` instead of bare array

2. **Required Header Changes**
   - Bearer token authentication now strictly enforced when `DS_TOKEN` is set
   - CORS headers require explicit `DS_CORS=1` environment variable

## ✨ New Features

### API Versioning
- **Schema Version**: All endpoints now return `schema_version: "ds.v1"`
- **Contract Guarantee**: API contract frozen at v1.1.0 - no breaking changes without major version bump
- **Backwards Compatibility**: Future minor versions will maintain compatibility

### Discovery System
- **Well-Known Endpoint**: `/.well-known/obs-bridge.json` for service discovery
- **Service Registry**: `/api/discovery/services` with service metadata
- **Self-Status**: `/api/self-status` with server time (`nowMs`) and configuration

### Enhanced CI/CD
- **Automated Validation**: GitHub Actions workflow validates contracts on every PR
- **OpenAPI Linting**: Redocly and Spectral integration for API quality
- **Contract Enforcement**: Breaking changes automatically detected and blocked

## 🔄 Migration Guide

### For API Consumers

#### JavaScript/TypeScript
```javascript
// Old (v1.0.x)
const repos = await fetch('/v1/status').then(r => r.json());

// New (v1.1.0)
const response = await fetch('/v1/status').then(r => r.json());
const repos = response.data;  // Access via .data field
```

#### Shell/curl with jq
```bash
# Old (v1.0.x)
curl $URL/v1/status | jq '.[].Name'

# New (v1.1.0)
curl $URL/v1/status | jq '.data[].Name'
```

#### Go Client
```go
// The pkg/dsclient package handles this automatically
// Just update to latest version
import "github.com/verlyn13/ds-go/pkg/dsclient"
```

### For Server Operators

#### Starting the Server
```bash
# Recommended: Enable auth and CORS
DS_TOKEN=your-secret-token DS_CORS=1 ds serve --addr 127.0.0.1:7777

# Token is now enforced on all endpoints
# CORS headers only sent when DS_CORS=1
```

#### Validation
```bash
# Verify your server is v1.1.0 compliant
./scripts/verify-stage0.sh

# Test contract compliance
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:7777/v1/health | \
  jq '.schema_version'  # Should return "ds.v1"
```

## 📦 What's Included

### Files Changed
- `internal/server/server.go` - Core API implementation with versioning
- `internal/server/openapi.yaml` - Complete OpenAPI specification
- `pkg/dsclient/client.go` - Updated client with envelope handling
- `.github/workflows/ci.yml` - Comprehensive CI pipeline
- `CHANGELOG.md` - Detailed change history
- `.github/CODEOWNERS` - Review requirements for contract paths

### New Scripts
- `scripts/verify-stage0.sh` - Stage 0 compliance verification
- `scripts/verify-ds-services.sh` - Service endpoint verification

### Documentation
- `docs/contracts/VERSION.md` - Contract version documentation
- `README.md` - Updated with API versioning information
- Migration guides and examples

## 🔒 Contract Guarantee

From v1.1.0 forward, we guarantee:
- No breaking changes without major version bump (v2.0.0)
- New fields may be added (non-breaking)
- Deprecated features maintained for 2 major versions minimum
- Response format stability for all versioned endpoints

## 🧪 Testing

Before deploying v1.1.0:

```bash
# Run all tests
go test ./...

# Verify contract compliance
DS_TOKEN=test ./scripts/verify-stage0.sh

# Test with example client
DS_TOKEN=test go run ./examples/go-client
```

## 📝 Known Issues

- SSE endpoints (`/v1/status/sse`, `/v1/fetch/sse`) stream raw data without version wrapping - this is intentional for streaming protocols
- Some legacy scripts may need updates to handle wrapped responses

## 🆘 Support

For migration assistance or to report issues:
- GitHub Issues: https://github.com/verlyn13/ds-go/issues
- Documentation: See `CHANGELOG.md` for detailed changes
- Examples: Check `examples/` directory for updated usage patterns

## 🔄 Upgrade Checklist

Before upgrading production systems:

- [ ] Review breaking changes above
- [ ] Update API consumer code to handle wrapped responses
- [ ] Test with new response format
- [ ] Update monitoring/alerting for new endpoints
- [ ] Verify authentication headers if using tokens
- [ ] Test CORS configuration if needed
- [ ] Run integration tests
- [ ] Update documentation for your team

## 🎯 Next Release

v1.2.0 (planned) will include:
- Additional discovery endpoints
- Performance improvements
- Extended policy checking
- No breaking changes (per contract guarantee)

---

**Note**: This is a significant release with breaking changes. Please test thoroughly in development before deploying to production.