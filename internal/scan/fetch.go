package scan

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/verlyn13/ds-go/internal/git"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// FetchResult represents the result of a fetch operation
type FetchResult struct {
	RepoName string
	Success  bool
	Error    error
	Duration time.Duration
}

// Fetcher handles concurrent fetching of repositories
type Fetcher struct {
    gitClient   *git.Git
    workerCount int
}

// NewFetcher creates a new Fetcher with native Go concurrency
func NewFetcher(workerCount int) *Fetcher {
	if workerCount <= 0 {
		workerCount = 10
	}
	return &Fetcher{
		gitClient:   git.New(),
		workerCount: workerCount,
	}
}

// FetchAll fetches all repositories concurrently using native Go primitives
func (f *Fetcher) FetchAll(repos []Repository, showProgress bool) []FetchResult {
	results := make([]FetchResult, len(repos))
	
	// Skip repos without remotes
	var toFetch []int
	for i, repo := range repos {
		if repo.RemoteURL != "no remote" {
			toFetch = append(toFetch, i)
		}
	}
	
	if len(toFetch) == 0 {
		return results
	}
	
	// Progress tracking with atomic counters (lock-free)
	var completed atomic.Int32
	var succeeded atomic.Int32
	
	if showProgress {
		fmt.Printf("\nFetching %d repositories...\n", len(toFetch))
	}
	
	// Use errgroup for structured concurrency
	g, ctx := errgroup.WithContext(context.Background())
	
	// Semaphore for rate limiting (native Go primitive)
	sem := semaphore.NewWeighted(int64(f.workerCount))
	
	for _, idx := range toFetch {
		idx := idx // capture
		repo := repos[idx]
		
		g.Go(func() error {
			// Acquire semaphore (will block if at limit)
			if err := sem.Acquire(ctx, 1); err != nil {
				return nil // Context cancelled
			}
			defer sem.Release(1)
			
			start := time.Now()
			err := f.gitClient.Fetch(repo.Path)
			duration := time.Since(start)
			
			results[idx] = FetchResult{
				RepoName: repo.Name,
				Success:  err == nil,
				Error:    err,
				Duration: duration,
			}
			
			// Update counters atomically
			current := completed.Add(1)
			if err == nil {
				succeeded.Add(1)
			}
			
			if showProgress {
				status := "✓"
				if err != nil {
					status = "✗"
				}
				fmt.Printf("[%d/%d] %s %s (%.1fs)\n", 
					current, len(toFetch), status, repo.Name, duration.Seconds())
			}
			
			return nil
		})
	}
	
	// Wait for all goroutines
	g.Wait()
	
	if showProgress {
		fmt.Printf("\nCompleted: %d/%d successful\n", succeeded.Load(), len(toFetch))
	}
	
	return results
}

// FetchSingle fetches a single repository
func (f *Fetcher) FetchSingle(repo Repository) FetchResult {
	start := time.Now()
	err := f.gitClient.Fetch(repo.Path)
	
	return FetchResult{
		RepoName: repo.Name,
		Success:  err == nil,
		Error:    err,
		Duration: time.Since(start),
	}
}

// FetchAllStream fetches repositories and streams results as they complete.
// Respects context cancelation and limits concurrency via semaphore.
func (f *Fetcher) FetchAllStream(ctx context.Context, repos []Repository) <-chan FetchResult {
    out := make(chan FetchResult)
    // Filter indices to fetch
    var toFetch []int
    for i, repo := range repos {
        if repo.RemoteURL != "no remote" {
            toFetch = append(toFetch, i)
        }
    }
    if len(toFetch) == 0 {
        close(out)
        return out
    }
    sem := semaphore.NewWeighted(int64(f.workerCount))
    g, ctx := errgroup.WithContext(ctx)
    for _, idx := range toFetch {
        idx := idx
        repo := repos[idx]
        g.Go(func() error {
            if err := sem.Acquire(ctx, 1); err != nil { return nil }
            defer sem.Release(1)
            start := time.Now()
            err := f.gitClient.Fetch(repo.Path)
            res := FetchResult{RepoName: repo.Name, Success: err == nil, Error: err, Duration: time.Since(start)}
            select {
            case out <- res:
            case <-ctx.Done():
            }
            return nil
        })
    }
    go func() {
        _ = g.Wait()
        close(out)
    }()
    return out
}
