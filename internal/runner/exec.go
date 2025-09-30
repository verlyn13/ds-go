package runner

import (
    "context"
    "os/exec"
    "time"

    "github.com/verlyn13/ds-go/internal/scan"
)

type ExecResult struct {
    Repo    string `json:"repo"`
    Path    string `json:"path"`
    Success bool   `json:"success"`
    Error   string `json:"error,omitempty"`
    DurationMs int64 `json:"duration_ms"`
}

func ExecInRepos(repos []scan.Repository, command string, timeout time.Duration) []ExecResult {
    results := make([]ExecResult, 0, len(repos))
    for _, r := range repos {
        start := time.Now()
        res := ExecResult{Repo: r.Name, Path: r.Path}
        ctx := context.Background()
        if timeout > 0 {
            var cancel context.CancelFunc
            ctx, cancel = context.WithTimeout(ctx, timeout)
            defer cancel()
        }
        cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
        cmd.Dir = r.Path
        if err := cmd.Run(); err != nil {
            res.Success = false
            res.Error = err.Error()
        } else {
            res.Success = true
        }
        res.DurationMs = time.Since(start).Milliseconds()
        results = append(results, res)
    }
    return results
}

