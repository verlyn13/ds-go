package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Repository represents git repository status
type Repository struct {
	Path         string
	Name         string
	Account      string   // The owner (user or org) from remote URL
	FolderName   string   // The folder name it should be organized under
	IsOrg        bool     // Whether this is an organization repo
	RemoteURL    string
	Branch       string
	IsClean      bool
	Uncommitted  int
	Ahead        int
	Behind       int
	LastCommit   string
	LastFetch    *time.Time
	HasStash     bool
	HasUpstream  bool
}

// Git wraps git command execution
type Git struct {
	timeout time.Duration
}

// New creates a new Git wrapper
func New() *Git {
	return &Git{
		timeout: 5 * time.Second,
	}
}

// GetStatus returns the status of a git repository
func (g *Git) GetStatus(repoPath string) (*Repository, error) {
	repo := &Repository{
		Path: repoPath,
		Name: filepath.Base(repoPath),
	}

	// Check if it's a git repo
	if !g.isGitRepo(repoPath) {
		return nil, fmt.Errorf("not a git repository")
	}

	// Get current branch
	branch, err := g.runCommand(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		repo.Branch = "unknown"
	} else {
		repo.Branch = strings.TrimSpace(branch)
	}

	// Get remote URL
	remoteURL, err := g.runCommand(repoPath, "remote", "get-url", "origin")
	if err != nil {
		repo.RemoteURL = "no remote"
	} else {
		repo.RemoteURL = strings.TrimSpace(remoteURL)
		repo.Account = g.extractAccount(repo.RemoteURL)
	}

	// Check for upstream
	upstream, err := g.runCommand(repoPath, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	repo.HasUpstream = err == nil && upstream != ""

	// Get uncommitted changes count
	status, err := g.runCommand(repoPath, "status", "--porcelain")
	if err == nil {
		lines := strings.Split(strings.TrimSpace(status), "\n")
		if len(lines) == 1 && lines[0] == "" {
			repo.Uncommitted = 0
			repo.IsClean = true
		} else {
			repo.Uncommitted = len(lines)
			repo.IsClean = false
		}
	}

	// Get ahead/behind counts (only if we have upstream)
	if repo.HasUpstream {
		revList, err := g.runCommand(repoPath, "rev-list", "--left-right", "--count", "@{u}...HEAD")
		if err == nil {
			parts := strings.Fields(revList)
			if len(parts) == 2 {
				repo.Behind, _ = strconv.Atoi(parts[0])
				repo.Ahead, _ = strconv.Atoi(parts[1])
			}
		}
	}

	// Get last commit info
	lastCommit, err := g.runCommand(repoPath, "log", "-1", "--pretty=%cr: %s")
	if err == nil {
		repo.LastCommit = strings.TrimSpace(lastCommit)
		if len(repo.LastCommit) > 60 {
			repo.LastCommit = repo.LastCommit[:57] + "..."
		}
	} else {
		repo.LastCommit = "No commits"
	}

	// Check for stashes
	stashList, err := g.runCommand(repoPath, "stash", "list")
	repo.HasStash = err == nil && stashList != ""

	return repo, nil
}

// Fetch runs git fetch on a repository
func (g *Git) Fetch(repoPath string) error {
	_, err := g.runCommand(repoPath, "fetch", "--all", "--quiet")
	return err
}

// Pull runs git pull on a repository
func (g *Git) Pull(repoPath string) error {
	_, err := g.runCommand(repoPath, "pull", "--ff-only")
	return err
}

// Push runs git push on a repository
func (g *Git) Push(repoPath string) error {
	_, err := g.runCommand(repoPath, "push")
	return err
}

// runCommand executes a git command with timeout
func (g *Git) runCommand(repoPath string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", repoPath}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timed out")
		}
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// isGitRepo checks if a directory is a git repository
func (g *Git) isGitRepo(path string) bool {
	_, err := g.runCommand(path, "rev-parse", "--git-dir")
	return err == nil
}

// extractAccount extracts account name from remote URL
func (g *Git) extractAccount(remoteURL string) string {
	// Handle custom SSH hosts: happy-patterns:owner/repo.git
	// or standard SSH: git@github.com:owner/repo.git
	if strings.Contains(remoteURL, ":") && !strings.HasPrefix(remoteURL, "http") {
		parts := strings.Split(remoteURL, ":")
		if len(parts) == 2 {
			pathParts := strings.Split(parts[1], "/")
			if len(pathParts) >= 1 {
				// Remove .git suffix if present
				account := strings.TrimSuffix(pathParts[0], ".git")
				return account
			}
		}
	}
	
	// Handle HTTPS URLs: https://github.com/user/repo.git
	if strings.Contains(remoteURL, "github.com/") {
		parts := strings.Split(remoteURL, "github.com/")
		if len(parts) == 2 {
			pathParts := strings.Split(parts[1], "/")
			if len(pathParts) >= 1 {
				return pathParts[0]
			}
		}
	}
	
	return "unknown"
}