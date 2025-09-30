# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`ds-go` is a fast Git repository manager that scans and tracks repositories across multiple GitHub accounts. It's designed for users with multiple GitHub identities (personal, work, business) and integrates with verlyn13's system organization structure.

## Key Architecture Concepts

The codebase follows a clean architecture pattern with clear separation of concerns:

- **cmd/ds/main.go**: CLI entry point using Cobra for command handling
- **internal/config**: Configuration management with XDG compliance and account mappings
- **internal/scan**: Repository discovery, status checking, and organization logic
- **internal/git**: Git operations wrapper for status, branch info, and remote URL parsing
- **internal/ui**: Terminal UI components using lipgloss for styled tables and output

The system is organization-aware, mapping repositories to specific GitHub accounts and SSH configurations based on the repository owner.

## Development Commands

This project uses **mise** for task management and **direnv** for environment isolation, following system standards.

### Initial Setup
```bash
# Allow direnv for this project (one-time)
direnv allow

# Install project-specific Go version and tools
mise install

# View all available tasks
mise tasks
```

### Build and Install
```bash
mise run build         # Build binary locally
mise run install       # Build and install to /usr/local/bin
mise run clean         # Clean build artifacts

# Legacy Makefile still works
make build            # Delegates to mise
```

### Development and Testing
```bash
mise run dev           # Run in development mode
mise run test          # Run tests with coverage
mise run bench         # Run benchmarks
mise run race          # Run with race detector
mise run ci           # Run full CI pipeline locally
```

### Code Quality
```bash
mise run lint          # Run golangci-lint with project config
mise run fmt           # Format code with go fmt and gofumpt
```

### System Compliance
```bash
mise run validate      # Check system integration compliance
./scripts/validate-compliance.sh  # Detailed compliance report
```

### Profiling
```bash
mise run profile-cpu   # Profile CPU usage
mise run profile-mem   # Profile memory usage
```

### Dependencies
```bash
mise run deps          # Update all dependencies
go mod tidy            # Clean up go.mod
```

### Environment Variables
The `.envrc` file automatically sets:
- `GOMOD`: Project module name
- `DS_CONFIG_PATH`: ds configuration location
- `DS_DEV_MODE`: Development mode flag
- Project bin added to PATH

## System Integration Requirements

This tool MUST respect verlyn13's system organization:

### Directory Structure
- All repos MUST be under `~/Development/{account}/`
- Accounts: personal (verlyn13), work (jjohnson-47), business (happy-patterns), business-org (happy-patterns-org), hubofwyn
- Each account has its own SSH configuration in ~/.ssh/config

### Account Mapping
The tool automatically maps GitHub owners to accounts and determines:
- Which directory to clone into
- Which SSH host configuration to use
- How to organize existing repositories

### Configuration
- Config location: `~/.config/ds/config.yaml`
- Index cache: `~/Development/.ds-index.json`
- Fetch cache: `~/Development/.ds-fetch-cache.json`

## Common Development Tasks

### Adding New Commands
1. Create command struct in `cmd/ds/main.go`
2. Register with rootCmd in init()
3. Implement logic using internal packages

### Modifying Repository Scanning
- Core logic in `internal/scan/scan.go`
- Repository status from `internal/git/git.go`
- Concurrent processing using goroutines and semaphores

### Updating Configuration Schema
1. Modify structs in `internal/config/config.go`
2. Update YAML tags for serialization
3. Handle backward compatibility in Load()

### Working with Git Operations
- All Git commands wrapped in `internal/git/git.go`
- Use exec.Command with proper error handling
- Parse output carefully for cross-platform compatibility

## Performance Considerations

- Default worker count: 10 concurrent operations
- Caching: Index and fetch times cached to disk
- Semaphore-based concurrency control
- Parallel fetching with progress reporting

## Important Files

- `/cmd/ds/main.go`: CLI commands and flags
- `/internal/config/config.go`: Configuration management
- `/internal/scan/scan.go`: Repository discovery and scanning
- `/internal/git/git.go`: Git operations wrapper
- `/SYSTEM_INTEGRATION.md`: Detailed system requirements

## API Server

ds-go includes a built-in HTTP API server for automation and agent integration:

### Starting the Server
```bash
ds serve --addr 127.0.0.1:7777
```

### Key Endpoints
- `GET /v1/capabilities` - API discovery
- `GET /v1/status` - Repository status (supports filters)
- `GET /v1/scan` - Scan for repositories
- `GET /v1/fetch` - Update remote info
- `GET /v1/organize/plan` - Preview organization
- `POST /v1/organize/apply` - Apply organization
- `GET /v1/policy/check` - Compliance checks
- `POST /v1/exec` - Execute commands across repos

### Streaming Support
- `/v1/status/sse` - Real-time status via Server-Sent Events
- `/v1/fetch/sse` - Live fetch progress
- `/v1/status/stream` - NDJSON streaming

### Documentation
- **API Reference**: `/API.md`
- **OpenAPI Spec**: `/openapi.yaml`
- **Examples**: `/examples/agents/`
- **MCP Config**: `/examples/agents/mcp-config.yaml`
- **AI Discovery**: `/.well-known/ai-discovery.json`

### Integration for AI Agents
```python
# Example: Check repository health
import requests

api = "http://127.0.0.1:7777/v1"
status = requests.get(f"{api}/status?dirty=true").json()
print(f"Dirty repos: {status['data']['summary']['dirty']}")
```

## Dependencies

- cobra: CLI framework
- lipgloss: Terminal styling
- go-pretty: Table formatting
- adrg/xdg: XDG directory compliance
- golang.org/x/sync: Concurrency utilities