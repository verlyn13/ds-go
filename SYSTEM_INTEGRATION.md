# System Integration Requirements for ds-go

## Overview

This document defines how the `ds` CLI tool must integrate with verlyn13's system configuration as defined in the `system-setup-update` repository (the single source of truth). The ds tool MUST respect and enforce the system's organizational principles, directory structure, and configuration management patterns.

## Core System Principles

The ds tool must align with three fundamental principles:

1. **Clarity through separation** - Repositories organized by GitHub account/organization
2. **Discipline through constraint** - Strict adherence to directory structure
3. **Intelligence through augmentation** - Integration with MCP server and system tools

## Directory Structure Requirements

### Repository Organization

The ds tool MUST scan and organize repositories according to this STRICT structure:

```
~/Development/
├── personal/        # GitHub: verlyn13
├── work/           # GitHub: jjohnson-47
├── business/       # GitHub: happy-patterns
├── business-org/   # GitHub: happy-patterns-org
└── hubofwyn/       # GitHub: hubofwyn
```

**CRITICAL**:
- The ds tool MUST NOT create repositories outside these directories
- Each directory corresponds to a specific GitHub account with its own SSH configuration
- Mixed ownership within a directory is FORBIDDEN

### Configuration Paths

The ds configuration MUST use these standard paths:

```yaml
# Primary configuration location (XDG compliant)
~/.config/ds/config.yaml

# Integration with system tools
~/.config/devops-mcp/     # MCP server configuration
~/.config/mise/            # Version management
~/.config/chezmoi/         # Dotfiles management
```

## GitHub Account Mapping

The ds tool MUST recognize and properly map these accounts:

```yaml
accounts:
  verlyn13:
    type: personal
    ssh_host: github.com  # Default SSH key
    directory: ~/Development/personal

  jjohnson-47:
    type: work
    ssh_host: github-work  # Maps to Host in ~/.ssh/config
    directory: ~/Development/work

  happy-patterns:
    type: business
    ssh_host: github-business
    directory: ~/Development/business

  happy-patterns-org:
    type: business-org
    ssh_host: github-business-org
    directory: ~/Development/business-org

  hubofwyn:
    type: business
    ssh_host: github-hubofwyn
    directory: ~/Development/hubofwyn
```

### SSH Configuration Integration

The ds tool MUST respect the SSH configuration in `~/.ssh/config`:

```
# Personal (default)
Host github.com
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_personal

# Work account
Host github-work
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_work

# Business accounts
Host github-business
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_business
```

When cloning:
- Parse the repository owner from the URL
- Map to the correct SSH host configuration
- Clone to the appropriate directory

## MCP Server Integration

### Repository Status Reporting

The ds tool SHOULD integrate with the MCP server for repository status:

```yaml
# Resource endpoint for repo status
devops://repo_status

# The ds tool should provide repository data in this format:
{
  "repositories": [
    {
      "path": "/Users/verlyn13/Development/personal/ds-go",
      "account": "verlyn13",
      "name": "ds-go",
      "status": "clean|dirty",
      "branch": "main",
      "ahead": 0,
      "behind": 0,
      "uncommitted": 0,
      "untracked": 0,
      "last_fetch": "2025-09-28T08:00:00Z"
    }
  ]
}
```

### Audit Integration

All ds operations that modify the filesystem SHOULD be logged:

```yaml
# Audit log location (per MCP config)
~/Library/Application Support/devops.mcp/audit.jsonl

# Log format for ds operations
{
  "timestamp": "2025-09-28T08:15:00Z",
  "tool": "ds",
  "operation": "clone|organize|fetch",
  "account": "verlyn13",
  "repository": "ds-go",
  "path": "/Users/verlyn13/Development/personal/ds-go",
  "success": true,
  "duration_ms": 1234
}
```

### Telemetry

If telemetry is enabled in MCP config, ds SHOULD emit metrics:

```yaml
# OpenTelemetry endpoint from MCP config
endpoint: "http://127.0.0.1:4318"

# Metrics to emit
- ds.scan.duration
- ds.repositories.total
- ds.repositories.dirty
- ds.fetch.duration
- ds.clone.success_rate
```

## Command Behavior Requirements

### `ds scan`

MUST:
- Scan ONLY directories under `~/Development/`
- Respect the account/directory mapping
- Cache scan results for performance
- Skip non-git directories
- Ignore nested git repositories (submodules handled separately)

### `ds status`

MUST:
- Show account association for each repository
- Indicate if repository is in wrong directory
- Flag SSH configuration mismatches
- Integrate with MCP repo_status if available

### `ds clone <url>`

MUST:
1. Parse repository URL to extract owner
2. Map owner to account configuration
3. Determine correct directory from account mapping
4. Use appropriate SSH host configuration
5. Clone to `~/Development/{account}/repo-name`
6. REFUSE to clone if account mapping is unclear

Example:
```bash
# Cloning personal repo
ds clone git@github.com:verlyn13/ds-go.git
# → Clones to ~/Development/personal/ds-go using default SSH

# Cloning work repo
ds clone git@github.com:jjohnson-47/course-tooling.git
# → Clones to ~/Development/work/course-tooling using github-work SSH
```

### `ds organize`

MUST:
- Scan all repositories
- Detect misplaced repositories (wrong account directory)
- Propose moves to correct directories
- Update git remote URLs to use correct SSH host
- Create backup before moving
- Log all moves to MCP audit log

MUST NOT:
- Move repositories without confirmation
- Delete any data
- Break git history

### `ds fetch`

SHOULD:
- Respect MCP rate limits (2 req/sec default)
- Use parallel fetching with worker pool
- Report fetch status to MCP telemetry
- Handle SSH key authentication gracefully

## Configuration File Schema

The ds configuration file (`~/.config/ds/config.yaml`) MUST follow this schema:

```yaml
# Version for future compatibility
version: 1

# Base directory (MUST be ~/Development)
base_dir: ~/Development

# Account mappings (REQUIRED)
accounts:
  verlyn13:
    type: personal
    ssh_host: github.com
    email: verlyn13@gmail.com

  jjohnson-47:
    type: work
    ssh_host: github-work
    email: jjohnson-47@school.edu

  happy-patterns:
    type: business
    ssh_host: github-business
    email: admin@happy-patterns.com

  happy-patterns-org:
    type: business-org
    ssh_host: github-business-org

  hubofwyn:
    type: business
    ssh_host: github-hubofwyn

# Optional: Organization mappings
orgs:
  ScopeTechGtHb: github-business  # Maps org to SSH host

# Integration settings
integrations:
  mcp:
    enabled: true
    config_path: ~/.config/devops-mcp/config.toml

  telemetry:
    enabled: true
    endpoint: http://127.0.0.1:4318

  audit:
    enabled: true
    path: ~/Library/Application Support/devops.mcp/audit.jsonl

# Performance settings
performance:
  workers: 10
  cache_ttl: 300  # seconds
  fetch_timeout: 30
```

## Error Handling

The ds tool MUST handle these error cases gracefully:

1. **Unknown repository owner**: Prompt user to map account
2. **Missing SSH configuration**: Show setup instructions
3. **Wrong directory**: Offer to move with `ds organize`
4. **MCP server unavailable**: Continue without integration
5. **Permission denied**: Check SSH key and configuration

## Validation Commands

The ds tool SHOULD provide validation commands:

```bash
# Validate system configuration
ds validate

# Check SSH configuration
ds validate ssh

# Verify account mappings
ds validate accounts

# Test MCP integration
ds validate mcp
```

## Migration Path

For existing installations:

1. Detect current repository locations
2. Map to correct account directories
3. Generate migration plan
4. Execute with user confirmation
5. Update git remotes
6. Verify SSH access

## Security Requirements

MUST:
- Never store credentials
- Use SSH keys exclusively
- Respect file permissions (755 for directories)
- Never commit sensitive files (.env, *.key, *.pem)
- Integrate with gopass if configured in MCP

MUST NOT:
- Use HTTPS URLs for private repositories
- Store passwords in configuration
- Log sensitive information

## Testing Requirements

The ds tool MUST be tested against:

1. Multiple account repositories
2. SSH configuration variations
3. MCP server integration
4. Directory permission scenarios
5. Migration from legacy structures

## Compliance Verification

The ds tool MUST provide compliance checking:

```bash
# Full system compliance check
ds compliance

# Output format
✅ Directory structure correct
✅ SSH configuration valid
✅ Account mappings complete
⚠️  3 repositories in wrong directories
✅ MCP integration active
```

## Future Compatibility

The ds tool MUST:
- Version its configuration file
- Support configuration migration
- Maintain backward compatibility
- Integrate with system-setup-update updates

## Implementation Priority

1. **Critical**: Directory structure compliance
2. **Critical**: SSH configuration mapping
3. **High**: MCP server integration
4. **High**: Account-based organization
5. **Medium**: Telemetry and audit logging
6. **Low**: Advanced MCP features

## References

- System Setup Repository: `~/Development/personal/system-setup-update`
- MCP Server: `~/Development/personal/devops-mcp`
- System Documentation: `~/Development/personal/system-setup-update/README.md`
- Policy as Code: `~/Development/personal/system-setup-update/04-policies/policy-as-code.yaml`

---

**Last Updated**: September 28, 2025
**Maintainer**: verlyn13
**Status**: AUTHORITATIVE - This document defines mandatory requirements