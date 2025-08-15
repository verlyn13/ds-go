package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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