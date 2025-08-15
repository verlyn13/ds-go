# ds - Dead Simple Repository Manager (Go Edition)

A blazing-fast Git repository scanner built with Go 1.25, designed for managing multiple GitHub accounts and repositories with native performance.

## ⚡ Performance Features

- **Native Go concurrency** with goroutines and channels
- **Container-aware GOMAXPROCS** (Go 1.25) - automatically respects cgroup CPU limits
- **Lock-free atomic counters** for progress tracking
- **Minimal allocations** with string builders and buffer reuse
- **Parallel git operations** with semaphore-based rate limiting
- **Swiss-table maps** (Go 1.24+) for faster lookups

## 🚀 Installation

### From Source (Recommended)

```bash
# Requires Go 1.25+
git clone https://github.com/verlyn13/ds-go.git
cd ds-go
make install
```

### Quick Build

```bash
go build -o ds ./cmd/ds
sudo mv ds /usr/local/bin/
```

### Using go install

```bash
go install github.com/verlyn13/ds-go/cmd/ds@latest
```

## 📊 Usage

### Basic Commands

```bash
# Show repository status (instant, no fetching)
ds status
ds s              # short alias

# Fetch all repositories (updates remote tracking)
ds fetch
ds f              # short alias
ds fetch -q       # quiet mode

# Show only dirty repositories
ds status -d
ds status --dirty

# Filter by account
ds status -a verlyn13
ds status --account verlyn13

# JSON output for scripting
ds status --json

# Scan and rebuild index
ds scan
ds scan --fetch   # fetch before scanning
```

### Output Example

```
📊 Repository Status: 23 total | 15 clean | 5 changes | 2 ahead | 1 behind

verlyn13 (12)
  ✓ ds-go              clean         synced      2m ago: Initial commit
  ● journal            3 files       ↑5          1h ago: Add daily notes
  ↓ dotfiles           clean         ↓2          3d ago: Update vim config
  
jjohnson-47 (8)
  ✓ course-work        clean         synced      1d ago: Submit assignment
  ● latex-dev          1 file        no upstream 5m ago: Draft thesis chapter
```

### Status Icons

- `✓` Green check - Clean and synced
- `●` Yellow dot - Has uncommitted changes
- `↑` Blue arrow - Ahead of remote (need to push)
- `↓` Cyan arrow - Behind remote (need to pull)
- `🟣` Purple - Has stashed changes

## ⚙️ Configuration

Configuration is stored in `~/.config/ds/config.yaml`:

```yaml
base_dir: ~/Projects

accounts:
  verlyn13:
    type: personal
    ssh_host: github-personal
    email: personal@example.com
  jjohnson-47:
    type: school
    ssh_host: github-work
    email: school@university.edu

organizations:
  ScopeTechGtHb: github-scope
  The-Nash-Group: github.com

folder_structure:
  personal: [verlyn13]
  school: [jjohnson-47]
  orgs: [ScopeTechGtHb, The-Nash-Group]
```

## 🏗️ Architecture

### Core Components

- **Scanner** (`internal/scan/`) - Concurrent repository discovery using filepath.WalkDir
- **Git wrapper** (`internal/git/`) - Direct git command execution with context timeouts
- **UI** (`internal/ui/`) - Table rendering with lipgloss styling and native ANSI codes
- **Config** (`internal/config/`) - XDG-compliant configuration with YAML/JSON support

### Concurrency Model

```go
// Native Go concurrency with errgroup and semaphore
g, ctx := errgroup.WithContext(context.Background())
sem := semaphore.NewWeighted(int64(workerCount))

for _, repo := range repos {
    g.Go(func() error {
        sem.Acquire(ctx, 1)
        defer sem.Release(1)
        // Process repository
        return nil
    })
}
```

## 🔧 Development

### Building

```bash
make build          # Build binary
make test           # Run tests
make bench          # Run benchmarks
make lint           # Run golangci-lint
make race           # Run with race detector
```

### Profiling

```bash
make profile-cpu    # CPU profiling
make profile-mem    # Memory profiling
```

### Release

```bash
make release        # Create release with GoReleaser
make snapshot       # Test release locally
```

## 🎯 Performance Optimizations

1. **No automatic fetching** - Status checks use cached remote info
2. **Parallel git operations** - Concurrent execution with rate limiting
3. **Minimal allocations** - String builders and buffer reuse
4. **Native ANSI codes** - Direct terminal control for faster rendering
5. **Lock-free progress** - Atomic counters instead of mutexes
6. **Efficient walking** - Skip hidden dirs, node_modules, vendor
7. **Container-aware** - Automatic CPU limit detection (Go 1.25)

## 📦 Dependencies

- `spf13/cobra` - CLI framework
- `charmbracelet/lipgloss` v1 - Terminal styling (stable)
- `jedib0t/go-pretty/v6` - Table rendering
- `adrg/xdg` - XDG directory support
- `golang.org/x/sync` - Errgroup and semaphore

## 🚦 Go 1.25 Features Used

- Container-aware GOMAXPROCS
- Swiss-table maps for better performance
- DWARF v5 debug info (smaller binaries)
- Flight-recorder tracing (optional)

## 📝 License

MIT License - See LICENSE file

## 🤝 Contributing

Pull requests welcome! Please ensure:
- Code passes `golangci-lint`
- Tests pass with race detector
- Benchmarks show no regression
- Follow existing code style

---

Built for speed with Go 1.25 🚀