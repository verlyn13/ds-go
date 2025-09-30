# Changelog

All notable changes to ds-go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.1.0] - 2025-09-29

### 🚨 BREAKING CHANGES

#### Response Format Changes
The following endpoints now return wrapped responses with `schema_version` and `data` fields instead of bare arrays:

- **`/v1/status`**: Previously returned `[{repo1}, {repo2}, ...]`
  - Now returns: `{"schema_version": "ds.v1", "data": [{repo1}, {repo2}, ...]}`

- **`/v1/organize/plan`**: Previously returned `[{move1}, {move2}, ...]`
  - Now returns: `{"schema_version": "ds.v1", "data": [{move1}, {move2}, ...]}`

#### Migration Guide
If you have scripts or integrations consuming these endpoints directly:

**Before (v1.0.x):**
```javascript
const repos = await fetch('/v1/status').then(r => r.json());
repos.forEach(repo => console.log(repo.Name));
```

**After (v1.1.0):**
```javascript
const response = await fetch('/v1/status').then(r => r.json());
const repos = response.data;  // Note: access via .data
repos.forEach(repo => console.log(repo.Name));
```

### Added

#### Contract Versioning
- All API endpoints now return `schema_version: "ds.v1"` for version identification
- Contract frozen at v1.1.0 with breaking change protection

#### Discovery Endpoints
- `/.well-known/obs-bridge.json` - Bridge descriptor for service discovery
- `/api/discovery/services` - Service registry with timestamps
- `/api/discovery/capabilities` - Capability descriptor
- `/api/discovery/openapi` - OpenAPI spec alias
- `/api/self-status` - Self-status with `nowMs` (epoch milliseconds)

#### CI/CD
- Comprehensive GitHub Actions workflow (`.github/workflows/ci.yml`)
- Contract validation on all PRs
- OpenAPI linting with Redocly and Spectral
- Automated endpoint verification
- Schema version consistency checks

#### Documentation
- Contract version documentation at `docs/contracts/VERSION.md`
- OpenAPI examples for all major endpoints
- Breaking change migration guide
- Tagging instructions for releases

### Changed

#### Response Envelopes
- Status and plan endpoints wrap arrays in versioned envelopes
- All non-streaming endpoints include `schema_version` field
- Timestamps standardized to epoch milliseconds (`nowMs`, `ts`)

#### OpenAPI Specification
- Added examples for `/v1/status` and `/v1/scan`
- Documented all discovery endpoints
- Enhanced type definitions with proper schemas

#### Client Package
- `pkg/dsclient` updated to handle wrapped responses
- Automatic unwrapping of `{schema_version, data}` envelopes
- Enhanced error handling for version mismatches

### Fixed
- Client compatibility with wrapped array responses
- SSE endpoints correctly stream without version wrapping
- OpenAPI validation errors resolved

### Security
- Authentication enforced via Bearer token when `DS_TOKEN` is set
- CORS properly configured when `DS_CORS=1`
- Security scanning added to CI pipeline

## [v1.0.0] - 2025-07-01

### Added
- Initial release of ds-go
- Core repository management commands
- Status checking across multiple accounts
- Repository organization features
- Policy compliance checking
- HTTP API server with `ds serve`
- Go client package

### Features
- Fast parallel repository scanning
- Multi-account support with SSH configuration
- Clean/dirty repository detection
- Branch and upstream tracking
- Fetch status monitoring
- Command execution across repositories

## Migration Notes

### From v1.0.x to v1.1.0

⚠️ **Required Actions:**

1. **Update API Consumers**: Any code directly consuming `/v1/status` or `/v1/organize/plan` must be updated to handle the wrapped response format.

2. **Update Scripts**: Shell scripts using `jq` need adjustment:
   ```bash
   # Before: curl $URL/v1/status | jq '.[].Name'
   # After:  curl $URL/v1/status | jq '.data[].Name'
   ```

3. **Client Libraries**: If using `pkg/dsclient`, update to latest version which handles wrapping automatically.

4. **Validation**: Test integrations with:
   ```bash
   DS_TOKEN=<token> curl -H "Authorization: Bearer <token>" \
     http://localhost:7777/v1/status | jq '.schema_version, .data | length'
   ```

### Deprecation Notices

None in this release. All v1.0.x features remain supported.

## Support

For questions about migration or breaking changes, please open an issue at:
https://github.com/verlyn13/ds-go/issues