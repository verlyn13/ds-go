# Claude AI Integration Guide for ds-go

## Overview

This guide helps Claude AI assistants effectively use the ds-go API for repository management tasks.

## Quick Start

```bash
# 1. Start the ds-go API server
ds serve --addr 127.0.0.1:7777

# 2. Test connection
curl http://127.0.0.1:7777/v1/capabilities
```

## Common Tasks for Claude

### 1. Check Repository Health

```python
# Get overview of all repositories
GET http://127.0.0.1:7777/v1/status

# Focus on problematic repos
GET http://127.0.0.1:7777/v1/status?dirty=true

# Check specific account
GET http://127.0.0.1:7777/v1/status?account=verlyn13
```

**What to look for:**
- Repositories with uncommitted changes (`is_clean: false`)
- Repositories ahead/behind remote
- Repositories with stashed changes
- Missing last_fetch times (never fetched)

### 2. Update Repository Information

```python
# Scan for new repositories
GET http://127.0.0.1:7777/v1/scan

# Update remote information
GET http://127.0.0.1:7777/v1/fetch

# Monitor fetch progress (streaming)
GET http://127.0.0.1:7777/v1/fetch/sse
```

### 3. Run Batch Operations

```python
# Lint all repositories
POST http://127.0.0.1:7777/v1/exec
Body: {"cmd": "mise run lint"}

# Run tests on dirty repos only
POST http://127.0.0.1:7777/v1/exec?dirty=true
Body: {"cmd": "mise run test"}

# Update dependencies for personal projects
POST http://127.0.0.1:7777/v1/exec?account=verlyn13
Body: {"cmd": "mise run deps"}
```

### 4. Check Compliance

```python
# Run policy checks
GET http://127.0.0.1:7777/v1/policy/check

# Fail on critical issues
GET http://127.0.0.1:7777/v1/policy/check?fail_on=critical
```

### 5. Organize Repositories

```python
# Preview organization plan
GET http://127.0.0.1:7777/v1/organize/plan

# Apply organization (safe mode)
POST http://127.0.0.1:7777/v1/organize/apply?require_clean=true
```

## Response Parsing

### Status Response Structure
```json
{
  "success": true,
  "data": {
    "repositories": [
      {
        "name": "ds-go",
        "account": "verlyn13",
        "is_clean": false,
        "uncommitted_files": 5,
        "branch": "main",
        "ahead": 0,
        "behind": 2
      }
    ],
    "summary": {
      "total": 69,
      "clean": 30,
      "dirty": 39
    }
  }
}
```

### Key Indicators
- **Dirty repos**: `is_clean: false` or `uncommitted_files > 0`
- **Out of sync**: `ahead > 0` or `behind > 0`
- **Stashed work**: `has_stash: true`
- **Organization**: Repos directly in base_dir need organizing

## Workflow Examples

### Daily Maintenance Workflow

```python
# 1. Scan for new repositories
scan_result = GET /v1/scan

# 2. Fetch latest remote information
fetch_result = GET /v1/fetch

# 3. Check repository status
status = GET /v1/status

# 4. Report findings
if status.data.summary.dirty > 0:
    print(f"Found {status.data.summary.dirty} repos with uncommitted changes")

if status.data.summary.behind > 0:
    print(f"Found {status.data.summary.behind} repos behind remote")

# 5. Run maintenance tasks
if status.data.summary.dirty > 10:
    # Too many dirty repos, investigate
    dirty_repos = GET /v1/status?dirty=true
    for repo in dirty_repos.data.repositories:
        print(f"{repo.name}: {repo.uncommitted_files} uncommitted files")
```

### CI/CD Integration Workflow

```python
# 1. Check policy compliance
policy = GET /v1/policy/check?fail_on=critical

if not policy.data.failed_threshold:
    # 2. Run tests
    test_result = POST /v1/exec
    Body: {"cmd": "mise run test"}

    # 3. Run linting
    lint_result = POST /v1/exec
    Body: {"cmd": "mise run lint"}

    # 4. Build if tests pass
    if test_result.data.summary.failed == 0:
        build_result = POST /v1/exec
        Body: {"cmd": "mise run build"}
```

### Repository Cleanup Workflow

```python
# 1. Identify repos needing cleanup
dirty_repos = GET /v1/status?dirty=true

# 2. Check organization
org_plan = GET /v1/organize/plan

# 3. Report to user
print(f"Cleanup needed:")
print(f"- {len(dirty_repos.data.repositories)} dirty repositories")
print(f"- {len(org_plan.data.moves)} repositories need reorganization")

# 4. Get user approval, then:
# - Commit changes where appropriate
# - Apply organization plan
if user_approves:
    POST /v1/organize/apply?require_clean=true
```

## Best Practices for Claude

1. **Always check capabilities first**
   - Use `/v1/capabilities` to understand available endpoints
   - Verify feature availability before using

2. **Use filters to reduce data**
   - Filter by account when working with specific projects
   - Use `dirty=true` to focus on repos needing attention

3. **Handle streaming appropriately**
   - Use SSE endpoints for real-time monitoring
   - Use regular endpoints for batch processing

4. **Check before modifying**
   - Use `plan` endpoints before applying changes
   - Use `dry_run` parameters when available

5. **Respect rate limits**
   - Although no current limits, be mindful of parallel operations
   - Use appropriate timeouts for long-running commands

## Error Handling

Common errors and solutions:

### Connection Refused
```json
{
  "error": "connection refused"
}
```
**Solution**: Ensure ds server is running: `ds serve`

### Repository Not Found
```json
{
  "error": "repository not found",
  "code": "REPO_NOT_FOUND"
}
```
**Solution**: Run scan first: `GET /v1/scan`

### Command Timeout
```json
{
  "error": "command timeout after 30s"
}
```
**Solution**: Increase timeout parameter or simplify command

## Advanced Integration

### Using MCP (Model Context Protocol)

```yaml
# mcp-config.yaml
servers:
  ds-repo-manager:
    command: ds
    args: ["serve", "--addr", "127.0.0.1:7777"]
    env:
      DS_CONFIG: ~/.config/ds/config.yaml
```

### Tool Definition for Claude

```json
{
  "name": "repository_status",
  "description": "Get status of Git repositories",
  "input_schema": {
    "type": "object",
    "properties": {
      "dirty_only": {
        "type": "boolean",
        "description": "Show only dirty repositories"
      },
      "account": {
        "type": "string",
        "description": "Filter by account"
      }
    }
  },
  "api_endpoint": "http://127.0.0.1:7777/v1/status"
}
```

## Monitoring & Alerting

Key metrics to track:
- Number of dirty repositories
- Repositories behind remote
- Failed policy checks
- Command execution failures
- Fetch success rate

Alert thresholds:
- Dirty repos > 50% of total
- Any critical policy failures
- Fetch failures > 10%
- Command timeout rate > 20%

## Security Considerations

1. **Local-only by default**: API binds to localhost
2. **No authentication**: Do not expose to network without auth layer
3. **Command injection**: Validate all commands before execution
4. **Path traversal**: API is confined to configured base_dir

## Support & Documentation

- Full API docs: `/API.md`
- OpenAPI spec: `/openapi.yaml`
- Examples: `/examples/` directory
- CLI help: `ds --help`