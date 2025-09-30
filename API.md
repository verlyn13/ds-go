# ds-go API Documentation

## Overview

The ds-go API provides programmatic access to repository management features via HTTP REST endpoints. Designed for automation, CI/CD integration, and agent-based workflows.

**Base URL**: `http://127.0.0.1:7777/v1` (configurable via `--addr`)

## Quick Start

```bash
# Start the API server
ds serve --addr 127.0.0.1:7777

# Test connection
curl http://127.0.0.1:7777/v1/capabilities
```

## Authentication

Currently, the API is unauthenticated and bound to localhost only. For production use, consider:
- Unix socket binding: `ds serve --socket /tmp/ds.sock`
- Token authentication: `ds serve --token <secret>` (planned)

## Response Format

All endpoints return JSON with consistent structure:

```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "timestamp": "2025-09-28T12:00:00Z"
}
```

Error responses:
```json
{
  "success": false,
  "data": null,
  "error": "error message",
  "timestamp": "2025-09-28T12:00:00Z"
}
```

## Endpoints

### Discovery

#### GET /v1/capabilities
Returns available endpoints and API metadata.

**Response:**
```json
{
  "success": true,
  "data": {
    "version": "1.0.0",
    "endpoints": [
      {
        "path": "/v1/status",
        "method": "GET",
        "description": "Get repository status",
        "parameters": ["dirty", "account", "path"]
      },
      ...
    ],
    "features": {
      "streaming": true,
      "policy": true,
      "exec": true,
      "organize": true
    }
  }
}
```

### Repository Status

#### GET /v1/status
Retrieve status of all repositories with optional filters.

**Query Parameters:**
- `dirty` (bool): Show only repositories with uncommitted changes
- `account` (string): Filter by account/organization
- `path` (string): Scan specific path (default: configured base_dir)

**Example:**
```bash
curl "http://127.0.0.1:7777/v1/status?dirty=true&account=verlyn13"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "repositories": [
      {
        "name": "ds-go",
        "path": "/Users/verlyn13/Development/personal/ds-go",
        "account": "verlyn13",
        "is_clean": false,
        "uncommitted_files": 9,
        "branch": "main",
        "ahead": 0,
        "behind": 0,
        "last_commit": "2025-09-28T08:00:00Z",
        "last_fetch": "2025-09-28T09:00:00Z"
      }
    ],
    "summary": {
      "total": 1,
      "clean": 0,
      "dirty": 1,
      "ahead": 0,
      "behind": 0
    }
  }
}
```

#### GET /v1/status/stream
NDJSON stream of repository status (one JSON object per line).

**Example:**
```bash
curl -N "http://127.0.0.1:7777/v1/status/stream"
```

#### GET /v1/status/sse
Server-Sent Events stream for real-time status updates.

**Example:**
```bash
curl -N -H "Accept: text/event-stream" \
  "http://127.0.0.1:7777/v1/status/sse"
```

**Event Format:**
```
event: repository
data: {"name":"ds-go","account":"verlyn13","is_clean":false,...}

event: complete
data: {"total":69,"processed":69}
```

### Repository Scanning

#### GET /v1/scan
Scan filesystem for repositories and update index.

**Query Parameters:**
- `path` (string): Path to scan (default: configured base_dir)

**Response:**
```json
{
  "success": true,
  "data": {
    "count": 69,
    "duration_ms": 1234,
    "index_updated": true
  }
}
```

### Fetch Operations

#### GET /v1/fetch
Fetch remote information for repositories.

**Query Parameters:**
- `account` (string): Filter by account
- `dirty` (bool): Fetch only dirty repositories

**Example:**
```bash
curl "http://127.0.0.1:7777/v1/fetch?account=verlyn13"
```

#### GET /v1/fetch/sse
Stream fetch progress via Server-Sent Events.

**Example:**
```bash
curl -N -H "Accept: text/event-stream" \
  "http://127.0.0.1:7777/v1/fetch/sse?account=verlyn13"
```

**Events:**
```
event: fetch
data: {"repo":"ds-go","status":"fetching","account":"verlyn13"}

event: fetch
data: {"repo":"ds-go","status":"complete","account":"verlyn13","duration_ms":500}

event: complete
data: {"total":10,"success":10,"failed":0}
```

### Organization

#### GET /v1/organize/plan
Preview repository organization changes.

**Query Parameters:**
- `require_clean` (bool): Only show moves for clean repositories

**Response:**
```json
{
  "success": true,
  "data": {
    "moves": [
      {
        "name": "project",
        "account": "verlyn13",
        "is_org": false,
        "old_path": "/Users/verlyn13/Development/project",
        "new_path": "/Users/verlyn13/Development/personal/project"
      }
    ],
    "summary": {
      "total_moves": 1,
      "user_repos": 1,
      "org_repos": 0
    }
  }
}
```

#### POST /v1/organize/apply
Apply organization changes to move repositories.

**Query Parameters:**
- `require_clean` (bool): Abort if any repo has uncommitted changes
- `force` (bool): Overwrite existing destinations
- `dry_run` (bool): Preview without making changes

**Response:**
```json
{
  "success": true,
  "data": {
    "results": [
      {
        "name": "project",
        "old_path": "/Users/verlyn13/Development/project",
        "new_path": "/Users/verlyn13/Development/personal/project",
        "applied": true,
        "error": null
      }
    ],
    "summary": {
      "moved": 1,
      "failed": 0,
      "dry_run": false
    }
  }
}
```

### Policy & Compliance

#### GET /v1/policy/check
Run policy compliance checks.

**Query Parameters:**
- `file` (string): Policy file path (default: .project-compliance.yaml)
- `fail_on` (string): Severity threshold (critical|high|medium|low)

**Response:**
```json
{
  "success": true,
  "data": {
    "results": [
      {
        "name": "directory_location",
        "description": "Project in correct directory",
        "severity": "critical",
        "passed": true
      },
      {
        "name": "mise_configured",
        "description": "Mise configuration present",
        "severity": "high",
        "passed": true
      }
    ],
    "summary": {
      "total": 10,
      "passed": 8,
      "failed": 2,
      "by_severity": {
        "critical": {"passed": 3, "failed": 0},
        "high": {"passed": 3, "failed": 1},
        "medium": {"passed": 2, "failed": 1},
        "low": {"passed": 0, "failed": 0}
      }
    },
    "failed_threshold": false
  }
}
```

### Command Execution

#### POST /v1/exec
Execute a command across multiple repositories.

**Query Parameters:**
- `account` (string): Filter by account
- `dirty` (bool): Execute only on dirty repositories
- `timeout` (int): Command timeout in seconds (default: 30)

**Request Body:**
```json
{
  "cmd": "mise run lint"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "results": [
      {
        "repo": "ds-go",
        "success": true,
        "stdout": "✓ All checks passed",
        "stderr": "",
        "exit_code": 0,
        "duration_ms": 1500
      }
    ],
    "summary": {
      "total": 1,
      "success": 1,
      "failed": 0,
      "total_duration_ms": 1500
    }
  }
}
```

#### GET /v1/exec (Alternative)
Same as POST but with command in query parameter.

**Query Parameters:**
- `cmd` (string): URL-encoded command
- Other parameters same as POST

**Example:**
```bash
curl "http://127.0.0.1:7777/v1/exec?cmd=git%20status&account=verlyn13"
```

## Streaming Responses

The API supports three types of streaming:

1. **NDJSON**: Line-delimited JSON objects
   - Content-Type: `application/x-ndjson`
   - Each line is a complete JSON object

2. **Server-Sent Events (SSE)**: W3C EventSource protocol
   - Content-Type: `text/event-stream`
   - Supports event types and reconnection

3. **Standard JSON**: Complete response after operation
   - Content-Type: `application/json`

## Error Handling

HTTP status codes:
- `200 OK`: Success
- `400 Bad Request`: Invalid parameters
- `404 Not Found`: Endpoint not found
- `500 Internal Server Error`: Server error

Error response includes details:
```json
{
  "success": false,
  "error": "repository not found: /path/to/repo",
  "code": "REPO_NOT_FOUND",
  "timestamp": "2025-09-28T12:00:00Z"
}
```

## Rate Limiting

Currently no rate limiting. For production deployment, consider:
- Request throttling per IP
- Concurrent operation limits
- Queue for long-running operations

## Examples

### Python Client
```python
import requests
import json

class DsClient:
    def __init__(self, base_url="http://127.0.0.1:7777/v1"):
        self.base_url = base_url

    def status(self, dirty=None, account=None):
        params = {}
        if dirty is not None:
            params['dirty'] = dirty
        if account:
            params['account'] = account

        response = requests.get(f"{self.base_url}/status", params=params)
        return response.json()

    def exec_command(self, cmd, account=None, timeout=30):
        params = {'timeout': timeout}
        if account:
            params['account'] = account

        response = requests.post(
            f"{self.base_url}/exec",
            params=params,
            json={'cmd': cmd}
        )
        return response.json()

# Usage
client = DsClient()
status = client.status(dirty=True)
print(f"Dirty repos: {status['data']['summary']['dirty']}")

result = client.exec_command("mise run lint", account="verlyn13")
for r in result['data']['results']:
    print(f"{r['repo']}: {'✓' if r['success'] else '✗'}")
```

### Node.js EventSource
```javascript
const EventSource = require('eventsource');

const url = 'http://127.0.0.1:7777/v1/fetch/sse?account=verlyn13';
const es = new EventSource(url);

es.addEventListener('fetch', (event) => {
  const data = JSON.parse(event.data);
  console.log(`${data.repo}: ${data.status}`);
});

es.addEventListener('complete', (event) => {
  const summary = JSON.parse(event.data);
  console.log(`Fetched ${summary.success}/${summary.total} repos`);
  es.close();
});

es.onerror = (err) => {
  console.error('Error:', err);
  es.close();
};
```

### Shell Script Automation
```bash
#!/bin/bash
# Check and fix all dirty repos

API_BASE="http://127.0.0.1:7777/v1"

# Get dirty repos
dirty_repos=$(curl -s "$API_BASE/status?dirty=true" | jq -r '.data.repositories[].name')

for repo in $dirty_repos; do
  echo "Processing $repo..."

  # Run lint
  curl -s -X POST "$API_BASE/exec" \
    -H "Content-Type: application/json" \
    -d "{\"cmd\": \"cd $repo && mise run lint\"}" \
    | jq -r '.data.results[0].success'
done
```

## Webhooks (Planned)

Future webhook support for events:
- Repository status changes
- Fetch completion
- Policy check failures
- Organization changes

## API Versioning

The API uses URL versioning: `/v1/`, `/v2/`, etc.

Breaking changes will increment the major version. The `/capabilities` endpoint will always return supported versions.

## Performance

- Repository scanning: ~50ms per repository
- Status queries use cached index (instant)
- Fetch operations: parallel with configurable workers
- Command execution: parallel with timeout protection

## Security Considerations

1. **Local-only by default**: Binds to 127.0.0.1
2. **No authentication**: Add token/key for network exposure
3. **Command injection**: Commands are executed as-is, validate input
4. **Path traversal**: Paths are confined to configured base_dir
5. **Resource limits**: Set timeouts and worker counts

## Monitoring

Recommended monitoring points:
- API response times
- Error rates by endpoint
- Repository scan duration
- Fetch success rates
- Command execution failures

## Troubleshooting

Common issues:

1. **"Connection refused"**
   - Ensure server is running: `ds serve`
   - Check address binding: `--addr 127.0.0.1:7777`

2. **"Repository not found"**
   - Run scan first: `GET /v1/scan`
   - Verify base_dir in config

3. **SSE not streaming**
   - Add header: `Accept: text/event-stream`
   - Use `-N` flag with curl

4. **Command execution timeout**
   - Increase timeout parameter
   - Check command complexity

## Support

- GitHub Issues: https://github.com/verlyn13/ds-go/issues
- Documentation: https://github.com/verlyn13/ds-go/wiki
- Examples: `/examples` directory in repository