# ds-go Quick Start Guide

Get up and running with ds-go in 5 minutes!

## Installation

### From Source
```bash
git clone https://github.com/verlyn13/ds-go.git
cd ds-go
mise run build
mise run install  # or: cp ds ~/.local/bin/
```

### Add to PATH
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

## Initial Setup

### 1. Configure Your Accounts
```bash
# Create configuration
ds init

# Or manually create ~/.config/ds/config.yaml:
cat > ~/.config/ds/config.yaml << 'EOF'
base_dir: /Users/$USER/Development
accounts:
  your-github-username:
    type: personal
    ssh_host: github.com
    email: your-email@example.com
EOF
```

### 2. Scan Your Repositories
```bash
# Scan for all Git repositories
ds scan

# Check status
ds status
```

## Basic Usage

### Check Repository Status
```bash
# All repositories
ds status

# Only dirty repositories
ds status -d

# Specific account
ds status -a verlyn13
```

### Update Remote Information
```bash
# Fetch all
ds fetch

# Watch progress
ds fetch  # Progress shown automatically
```

### Organize Repositories
```bash
# Preview organization plan
ds organize --plan

# Apply organization (safe mode)
ds organize --require-clean
```

## API Server

### Start the Server
```bash
ds serve --addr 127.0.0.1:7777
```

### Quick API Examples
```bash
# Check status via API
curl http://127.0.0.1:7777/v1/status

# Get dirty repos
curl "http://127.0.0.1:7777/v1/status?dirty=true"

# Run command across repos
curl -X POST http://127.0.0.1:7777/v1/exec \
  -H "Content-Type: application/json" \
  -d '{"cmd": "git status"}'
```

## Automation Examples

### Daily Maintenance Script
```bash
#!/bin/bash
# Daily repository maintenance

# Start server if not running
ds serve --addr 127.0.0.1:7777 &
sleep 2

# Run maintenance
curl -s http://127.0.0.1:7777/v1/scan
curl -s http://127.0.0.1:7777/v1/fetch
STATUS=$(curl -s http://127.0.0.1:7777/v1/status)

# Report
echo "$STATUS" | jq '.data.summary'
```

### Python Integration
```python
import requests

# Get repository status
api = "http://127.0.0.1:7777/v1"
status = requests.get(f"{api}/status?dirty=true").json()

# Show dirty repos
for repo in status['data']['repositories']:
    print(f"{repo['name']}: {repo['uncommitted_files']} files")
```

### Shell One-Liners
```bash
# Count dirty repos
ds status -d --json | jq '.data.summary.dirty'

# List repos behind remote
ds status --json | jq -r '.data.repositories[] | select(.behind > 0) | .name'

# Run tests on all projects
ds exec -- 'mise run test'
```

## Advanced Features

### Policy Compliance
```bash
# Check project compliance
ds policy check

# Fail on critical issues
ds policy check --fail-on critical
```

### Git Hooks
```bash
# Install pre-commit and pre-push hooks
ds hooks install
```

### Batch Operations
```bash
# Run command on specific account
ds exec -a verlyn13 -- 'mise run lint'

# Run on dirty repos only
ds exec -d -- 'git status'
```

## Troubleshooting

### Server Won't Start
```bash
# Check if port is in use
lsof -i:7777

# Use different port
ds serve --addr 127.0.0.1:8888
```

### Repositories Not Found
```bash
# Rescan with specific path
ds scan --path ~/Projects

# Check configuration
ds config view
```

### SSH Issues
```bash
# Verify SSH config
cat ~/.ssh/config | grep Host

# Test SSH connection
ssh -T git@github.com
```

## Integration with Tools

### With mise (Version Management)
```bash
# In any repository
mise run build  # Uses project-specific tools
mise run test   # Configured in .mise.toml
```

### With direnv (Environment)
```bash
# Auto-load environment
direnv allow
# Project environment loaded automatically
```

### With AI Assistants
```bash
# Start server for AI access
ds serve

# AI can now use API at http://127.0.0.1:7777/v1
# See /examples/agents/ for integration guides
```

## Common Workflows

### Morning Routine
```bash
ds fetch          # Update all remotes
ds status -d      # Check dirty repos
ds organize --plan # Check organization
```

### Before Pushing Code
```bash
ds policy check   # Verify compliance
ds exec -- 'mise run test'  # Run tests
ds exec -- 'mise run lint'  # Check code quality
```

### Weekly Cleanup
```bash
ds status --json | jq '.data.repositories[] | select(.is_clean == false)'
ds organize --require-clean
```

## Configuration Tips

### Multiple GitHub Accounts
```yaml
# ~/.config/ds/config.yaml
accounts:
  personal-github:
    type: personal
    ssh_host: github.com-personal
  work-github:
    type: work
    ssh_host: github.com-work
```

### SSH Configuration
```
# ~/.ssh/config
Host github.com-personal
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_personal

Host github.com-work
  HostName github.com
  User git
  IdentityFile ~/.ssh/id_ed25519_work
```

## Next Steps

- Read full documentation: [README.md](README.md)
- Explore API: [API.md](API.md)
- Check examples: [/examples/](examples/)
- Configure automation: [/examples/agents/](examples/agents/)

## Getting Help

```bash
# CLI help
ds --help
ds status --help

# Check version
ds version  # If implemented

# View configuration
ds config view
```

## Quick Reference

| Command | Description |
|---------|-------------|
| `ds status` | Show all repos |
| `ds status -d` | Show dirty repos |
| `ds fetch` | Update remotes |
| `ds scan` | Find new repos |
| `ds organize` | Organize by account |
| `ds serve` | Start API server |
| `ds exec -- cmd` | Run command |
| `ds policy check` | Check compliance |

Ready to manage your repositories efficiently! 🚀