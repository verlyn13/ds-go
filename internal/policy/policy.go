package policy

import (
    "fmt"
    "os/exec"
    "time"

    "gopkg.in/yaml.v3"
    "os"
)

type Severity string

const (
    SevCritical Severity = "critical"
    SevHigh     Severity = "high"
    SevMedium   Severity = "medium"
    SevLow      Severity = "low"
)

type Config struct {
    Validation struct {
        Checks []struct {
            Name        string   `yaml:"name"`
            Description string   `yaml:"description"`
            Command     string   `yaml:"command"`
            Severity    Severity `yaml:"severity"`
        } `yaml:"checks"`
    } `yaml:"validation"`
}

type CheckResult struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Severity    Severity `json:"severity"`
    Passed      bool     `json:"passed"`
    Error       string   `json:"error,omitempty"`
    DurationMs  int64    `json:"duration_ms"`
}

type Summary struct {
    Total    int `json:"total"`
    Passed   int `json:"passed"`
    Failed   int `json:"failed"`
    Warnings int `json:"warnings"`
}

type Report struct {
    Results []CheckResult `json:"results"`
    Summary Summary       `json:"summary"`
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil { return nil, err }
    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil { return nil, err }
    return &cfg, nil
}

func RunChecks(cfg *Config) (*Report, error) {
    var results []CheckResult
    var passed, failed int
    for _, c := range cfg.Validation.Checks {
        start := time.Now()
        res := CheckResult{Name: c.Name, Description: c.Description, Severity: c.Severity}
        cmd := exec.Command("/bin/sh", "-c", c.Command)
        if err := cmd.Run(); err != nil {
            res.Passed = false
            res.Error = err.Error()
            failed++
        } else {
            res.Passed = true
            passed++
        }
        res.DurationMs = time.Since(start).Milliseconds()
        results = append(results, res)
    }
    r := &Report{Results: results, Summary: Summary{Total: len(results), Passed: passed, Failed: failed}}
    return r, nil
}

// FailIfAboveSeverity returns true if any failed check at or above threshold exists
func FailIfAboveSeverity(r *Report, threshold Severity, cfg *Config) bool {
    sevOrder := map[Severity]int{SevCritical: 3, SevHigh: 2, SevMedium: 1, SevLow: 0}
    th := sevOrder[threshold]
    for i, res := range r.Results {
        if !res.Passed {
            if sevOrder[cfg.Validation.Checks[i].Severity] >= th {
                return true
            }
        }
    }
    return false
}

func SeverityFromString(s string) (Severity, error) {
    switch s {
    case string(SevCritical): return SevCritical, nil
    case string(SevHigh): return SevHigh, nil
    case string(SevMedium): return SevMedium, nil
    case string(SevLow): return SevLow, nil
    default:
        return SevLow, fmt.Errorf("invalid severity: %s", s)
    }
}

