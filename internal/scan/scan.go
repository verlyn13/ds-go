package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/verlyn13/ds-go/internal/config"
	"github.com/verlyn13/ds-go/internal/git"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// Repository is an alias for git.Repository with additional scanner metadata
type Repository struct {
    *git.Repository
    ScanTime time.Time `json:"scan_time"`
}

// Scanner handles repository discovery and scanning
type Scanner struct {
	config      *config.Config
	gitClient   *git.Git
	workerCount int
	indexPath   string
	fetchCache  map[string]time.Time
	mu          sync.RWMutex
}

// New creates a new Scanner
func New(cfg *config.Config, workerCount int) *Scanner {
	if workerCount <= 0 {
		workerCount = 10
	}
	
	indexPath := filepath.Join(cfg.BaseDir, ".ds-index.json")
	
	s := &Scanner{
		config:      cfg,
		gitClient:   git.New(),
		workerCount: workerCount,
		indexPath:   indexPath,
		fetchCache:  make(map[string]time.Time),
	}
	
	s.loadFetchCache()
	return s
}

// Scan discovers and analyzes all git repositories
func (s *Scanner) Scan(searchPath string) ([]Repository, error) {
	if searchPath == "" {
		searchPath = s.config.BaseDir
	}

	// Find all .git directories
	repoPaths, err := s.findRepositories(searchPath)
	if err != nil {
		return nil, fmt.Errorf("finding repositories: %w", err)
	}

	// Process repositories concurrently
	repos := make([]Repository, 0, len(repoPaths))
	var mu sync.Mutex
	
	g, ctx := errgroup.WithContext(context.Background())
	sem := semaphore.NewWeighted(int64(s.workerCount))
	
	for _, path := range repoPaths {
		path := path // capture loop variable
		
		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)
			
			gitRepo, err := s.gitClient.GetStatus(path)
			if err != nil {
				// Skip repos that fail to scan
				return nil
			}
			
			// Enhance with organization info
			s.enhanceRepoInfo(gitRepo)
			
			// Add fetch time from cache
			s.mu.RLock()
			if fetchTime, ok := s.fetchCache[path]; ok {
				gitRepo.LastFetch = &fetchTime
			}
			s.mu.RUnlock()
			
			repo := Repository{
				Repository: gitRepo,
				ScanTime:   time.Now(),
			}
			
			mu.Lock()
			repos = append(repos, repo)
			mu.Unlock()
			
			return nil
		})
	}
	
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("scanning repositories: %w", err)
	}
	
	return repos, nil
}

// findRepositories walks the directory tree to find git repositories
func (s *Scanner) findRepositories(root string) ([]string, error) {
	var repos []string
	
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip directories we can't read
		}
		
		// Skip hidden directories (except .git)
		if d.IsDir() && d.Name() != ".git" && len(d.Name()) > 0 && d.Name()[0] == '.' {
			return filepath.SkipDir
		}
		
		// Skip node_modules, vendor, etc.
		if d.IsDir() && (d.Name() == "node_modules" || d.Name() == "vendor" || d.Name() == "target") {
			return filepath.SkipDir
		}
		
		// Found a .git directory
		if d.IsDir() && d.Name() == ".git" {
			repoPath := filepath.Dir(path)
			repos = append(repos, repoPath)
			return filepath.SkipDir // Don't descend into .git
		}
		
		// Limit depth to 4 levels
		relPath, _ := filepath.Rel(root, path)
		depth := len(filepath.SplitList(relPath))
		if depth > 4 {
			return filepath.SkipDir
		}
		
		return nil
	})
	
	return repos, err
}

// SaveIndex saves the repository index to disk
func (s *Scanner) SaveIndex(repos []Repository) error {
	data := struct {
		LastScan     time.Time    `json:"last_scan"`
		Repositories []Repository `json:"repositories"`
	}{
		LastScan:     time.Now(),
		Repositories: repos,
	}
	
	file, err := os.Create(s.indexPath)
	if err != nil {
		return fmt.Errorf("creating index file: %w", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// LoadIndex loads the repository index from disk
func (s *Scanner) LoadIndex() ([]Repository, error) {
	file, err := os.Open(s.indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Repository{}, nil
		}
		return nil, fmt.Errorf("opening index file: %w", err)
	}
	defer file.Close()
	
	var data struct {
		LastScan     time.Time    `json:"last_scan"`
		Repositories []Repository `json:"repositories"`
	}
	
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding index: %w", err)
	}
	
	return data.Repositories, nil
}

// loadFetchCache loads the fetch cache from disk
func (s *Scanner) loadFetchCache() {
	cachePath := filepath.Join(s.config.BaseDir, ".ds-fetch-cache.json")
	
	file, err := os.Open(cachePath)
	if err != nil {
		return
	}
	defer file.Close()
	
	json.NewDecoder(file).Decode(&s.fetchCache)
}

// saveFetchCache saves the fetch cache to disk
func (s *Scanner) saveFetchCache() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	cachePath := filepath.Join(s.config.BaseDir, ".ds-fetch-cache.json")
	
	file, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(s.fetchCache)
}

// UpdateFetchTime updates the fetch time for a repository
func (s *Scanner) UpdateFetchTime(repoPath string) {
	s.mu.Lock()
	s.fetchCache[repoPath] = time.Now()
	s.mu.Unlock()
	
	s.saveFetchCache()
}

// enhanceRepoInfo adds organization awareness and folder info to repository
func (s *Scanner) enhanceRepoInfo(repo *git.Repository) {
	// Check if account is a known personal account
	if _, isAccount := s.config.Accounts[repo.Account]; isAccount {
		repo.IsOrg = false
		repo.FolderName = repo.Account
		return
	}
	
	// Check if it's a known organization
	for org := range s.config.Orgs {
		if org == repo.Account {
			repo.IsOrg = true
			repo.FolderName = repo.Account
			return
		}
	}
	
	// Try to infer from path structure
	// If repo is in ~/Projects/account/repo-name format
	pathParts := strings.Split(repo.Path, string(filepath.Separator))
	for i, part := range pathParts {
		if part == "Projects" && i+2 < len(pathParts) {
			potentialAccount := pathParts[i+1]
			// If the folder name matches a known account or org, use it
			if _, isAccount := s.config.Accounts[potentialAccount]; isAccount {
				repo.FolderName = potentialAccount
				repo.IsOrg = false
				return
			}
			for org := range s.config.Orgs {
				if org == potentialAccount {
					repo.FolderName = potentialAccount
					repo.IsOrg = true
					return
				}
			}
			// If not found in config but has a folder structure, use the folder name
			if pathParts[i+2] == repo.Name {
				repo.FolderName = potentialAccount
				// Guess if it's an org based on naming patterns
				repo.IsOrg = strings.Contains(potentialAccount, "-org") || 
							strings.Contains(potentialAccount, "org-") ||
							strings.HasSuffix(potentialAccount, "Org")
				return
			}
		}
	}
	
	// Default: use account name as folder
	repo.FolderName = repo.Account
	repo.IsOrg = false
}

// CloneRepo clones a repository with the appropriate SSH configuration
func CloneRepo(repoURL string, cfg *config.Config, targetPath string) error {
	// Parse the repository URL to extract owner and repo name
	// Support formats: 
	// - https://github.com/owner/repo
	// - github.com/owner/repo
	// - owner/repo
	// - git@github.com:owner/repo.git
	
	var owner, repoName string
	
	// Remove .git suffix if present
	repoURL = strings.TrimSuffix(repoURL, ".git")
	
	// Parse different URL formats
	if strings.Contains(repoURL, "github.com") {
		// Handle full URLs
		re := regexp.MustCompile(`github\.com[:/]([^/]+)/([^/]+)`)
		matches := re.FindStringSubmatch(repoURL)
		if len(matches) == 3 {
			owner = matches[1]
			repoName = matches[2]
		}
	} else if strings.Contains(repoURL, "/") {
		// Handle owner/repo format
		parts := strings.Split(repoURL, "/")
		if len(parts) == 2 {
			owner = parts[0]
			repoName = parts[1]
		}
	}
	
	if owner == "" || repoName == "" {
		return fmt.Errorf("invalid repository URL format: %s", repoURL)
	}
	
	// Determine SSH host based on owner
	sshHost := "github.com" // default
	if account, ok := cfg.Accounts[owner]; ok {
		sshHost = account.SSHHost
	} else {
		// Check if it's an organization
		for org, host := range cfg.Orgs {
			if org == owner {
				sshHost = host
				break
			}
		}
	}
	
	// Build the SSH clone URL
	cloneURL := fmt.Sprintf("git@%s:%s/%s.git", sshHost, owner, repoName)
	
	// Determine target directory
	if targetPath == "" {
		// Organize by account/owner name
		// Use the account name if it exists, otherwise use the owner name
		folderName := owner
		
		// Check if this is a known account
		if _, isAccount := cfg.Accounts[owner]; isAccount {
			folderName = owner
		} else {
			// Check if it's an organization we track
			for org := range cfg.Orgs {
				if org == owner {
					folderName = owner
					break
				}
			}
		}
		
		targetPath = filepath.Join(cfg.BaseDir, folderName, repoName)
	}
	
	// Ensure parent directory exists
	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("creating parent directory: %w", err)
	}
	
	// Execute git clone
	fmt.Printf("Cloning %s/%s to %s\n", owner, repoName, targetPath)
	fmt.Printf("Using SSH host: %s\n", sshHost)
	
	cmd := exec.Command("git", "clone", cloneURL, targetPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	
	// Set up git config for the repository if we have email configured
	if account, ok := cfg.Accounts[owner]; ok && account.Email != "" {
		// Set user email for this repository
		emailCmd := exec.Command("git", "config", "user.email", account.Email)
		emailCmd.Dir = targetPath
		if err := emailCmd.Run(); err != nil {
			fmt.Printf("Warning: couldn't set email config: %v\n", err)
		}
	}
	
	fmt.Printf("✓ Successfully cloned %s/%s\n", owner, repoName)
	return nil
}

// OrganizeRepos reorganizes repositories into proper account/org folder structure
func OrganizeRepos(repos []Repository, cfg *config.Config, dryRun, force bool) error {
    toMove := OrganizePlan(repos, cfg)
	
    // toMove already computed
	
	if len(toMove) == 0 {
		fmt.Println("✓ All repositories are already organized")
		return nil
	}
	
	// Display what will be moved
	fmt.Printf("Found %d repositories to organize:\n\n", len(toMove))
	
	for _, m := range toMove {
		relOld, _ := filepath.Rel(cfg.BaseDir, m.oldPath)
		relNew, _ := filepath.Rel(cfg.BaseDir, m.newPath)
		
		if m.repo.IsOrg {
			fmt.Printf("  [ORG] %s → %s\n", relOld, relNew)
		} else {
			fmt.Printf("  [USR] %s → %s\n", relOld, relNew)
		}
	}
	
    if dryRun {
        fmt.Println("\n[DRY RUN] No files were moved. Remove --dry-run to apply changes.")
        return nil
    }
	
	// Confirm before moving
	if !force {
		fmt.Print("\nProceed with reorganization? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted")
			return nil
		}
	}
	
	// Move repositories
	var moved, failed int
	for _, m := range toMove {
		// Create target directory if needed
		targetDir := filepath.Dir(m.newPath)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			fmt.Printf("  ✗ Failed to create directory %s: %v\n", targetDir, err)
			failed++
			continue
		}
		
		// Check if destination exists
		if _, err := os.Stat(m.newPath); err == nil && !force {
			fmt.Printf("  ✗ Destination exists: %s (use --force to overwrite)\n", m.newPath)
			failed++
			continue
		}
		
		// Move the repository
		if err := os.Rename(m.oldPath, m.newPath); err != nil {
			fmt.Printf("  ✗ Failed to move %s: %v\n", m.repo.Name, err)
			failed++
			continue
		}
		
		relPath, _ := filepath.Rel(cfg.BaseDir, m.newPath)
		fmt.Printf("  ✓ Moved to %s\n", relPath)
		moved++
	}
	
	fmt.Printf("\n✓ Reorganization complete: %d moved, %d failed\n", moved, failed)
	
	if moved > 0 {
		fmt.Println("\nRun 'ds scan' to update the repository index")
	}
	
	return nil
}

// MovePlan represents a proposed move of a repository
type MovePlan struct {
    Name      string `json:"name"`
    Account   string `json:"account"`
    IsOrg     bool   `json:"is_org"`
    OldPath   string `json:"old_path"`
    NewPath   string `json:"new_path"`
}

// OrganizePlan computes which repositories would be moved and where
func OrganizePlan(repos []Repository, cfg *config.Config) []struct{ repo Repository; oldPath, newPath string } {
    var toMove []struct{ repo Repository; oldPath, newPath string }
    for _, repo := range repos {
        if repo.FolderName == "" || repo.FolderName == "unknown" {
            continue
        }
        expectedPath := filepath.Join(cfg.BaseDir, repo.FolderName, repo.Name)
        if repo.Path == expectedPath {
            continue
        }
        if filepath.Dir(repo.Path) == cfg.BaseDir {
            toMove = append(toMove, struct{ repo Repository; oldPath, newPath string }{repo, repo.Path, expectedPath})
        }
    }
    return toMove
}

// OrganizePlanJSON returns a JSON-friendly plan
func OrganizePlanJSON(repos []Repository, cfg *config.Config) []MovePlan {
    var plans []MovePlan
    for _, m := range OrganizePlan(repos, cfg) {
        plans = append(plans, MovePlan{
            Name:    m.repo.Name,
            Account: m.repo.FolderName,
            IsOrg:   m.repo.IsOrg,
            OldPath: m.oldPath,
            NewPath: m.newPath,
        })
    }
    return plans
}

// OrganizeResult describes the outcome of applying an organize operation
type OrganizeResult struct {
    Name    string `json:"name"`
    OldPath string `json:"old_path"`
    NewPath string `json:"new_path"`
    Applied bool   `json:"applied"`
    Error   string `json:"error,omitempty"`
    DryRun  bool   `json:"dry_run"`
}

// ApplyOrganizePlan applies the organize plan and returns structured results
func ApplyOrganizePlan(repos []Repository, cfg *config.Config, dryRun, force bool) ([]OrganizeResult, int, int) {
    plan := OrganizePlan(repos, cfg)
    results := make([]OrganizeResult, 0, len(plan))
    moved := 0
    failed := 0
    for _, m := range plan {
        res := OrganizeResult{
            Name:    m.repo.Name,
            OldPath: m.oldPath,
            NewPath: m.newPath,
            DryRun:  dryRun,
        }
        if dryRun {
            results = append(results, res)
            continue
        }
        // Create target directory
        targetDir := filepath.Dir(m.newPath)
        if err := os.MkdirAll(targetDir, 0755); err != nil {
            res.Error = err.Error()
            failed++
            results = append(results, res)
            continue
        }
        // Destination exists
        if _, err := os.Stat(m.newPath); err == nil && !force {
            res.Error = fmt.Sprintf("destination exists: %s", m.newPath)
            failed++
            results = append(results, res)
            continue
        }
        if err := os.Rename(m.oldPath, m.newPath); err != nil {
            res.Error = err.Error()
            failed++
            results = append(results, res)
            continue
        }
        res.Applied = true
        moved++
        results = append(results, res)
    }
    return results, moved, failed
}
