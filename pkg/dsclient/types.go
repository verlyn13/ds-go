package dsclient

import "time"

// HealthResponse is returned by /v1/health
type HealthResponse struct {
    OK            bool      `json:"ok"`
    Version       int       `json:"version"`
    UptimeSec     int       `json:"uptime_sec"`
    Workers       int       `json:"workers"`
    Auth          bool      `json:"auth"`
    Timestamp     time.Time `json:"timestamp"`
    SchemaVersion string    `json:"schema_version"`
}

// Repository represents a scanned git repo entry
type Repository struct {
    Path        string     `json:"Path"`
    Name        string     `json:"Name"`
    Account     string     `json:"Account"`
    FolderName  string     `json:"FolderName"`
    IsOrg       bool       `json:"IsOrg"`
    RemoteURL   string     `json:"RemoteURL"`
    Branch      string     `json:"Branch"`
    IsClean     bool       `json:"IsClean"`
    Uncommitted int        `json:"Uncommitted"`
    Ahead       int        `json:"Ahead"`
    Behind      int        `json:"Behind"`
    LastCommit  string     `json:"LastCommit"`
    LastFetch   *time.Time `json:"LastFetch"`
    HasStash    bool       `json:"HasStash"`
    HasUpstream bool       `json:"HasUpstream"`
    ScanTime    time.Time  `json:"scan_time"`
}

// StatusResponse is returned by /v1/status
type StatusResponse struct {
    SchemaVersion string        `json:"schema_version"`
    Data          []Repository  `json:"data"`
}

// ScanResponse is returned by /v1/scan
type ScanResponse struct {
    SchemaVersion string `json:"schema_version"`
    Count         int    `json:"count"`
}

// MovePlan is returned by /v1/organize/plan
type MovePlan struct {
    Name    string `json:"name"`
    Account string `json:"account"`
    IsOrg   bool   `json:"is_org"`
    OldPath string `json:"old_path"`
    NewPath string `json:"new_path"`
}

// OrganizePlanResponse wraps move plans
type OrganizePlanResponse struct {
    SchemaVersion string     `json:"schema_version"`
    Data          []MovePlan `json:"data"`
}

// OrganizeApplyResponse returned by /v1/organize/apply
type OrganizeApplyResponse struct {
    SchemaVersion string     `json:"schema_version"`
    Moved         int        `json:"moved"`
    Failed        int        `json:"failed"`
    Results       []MovePlan `json:"results"`
}

// FetchResult from /v1/fetch
type FetchResult struct {
    RepoName string `json:"RepoName"`
    Success  bool   `json:"Success"`
    Error    *string `json:"Error"`
    Duration string `json:"Duration"`
}

// FetchResponse wraps fetch results
type FetchResponse struct {
    SchemaVersion string        `json:"schema_version"`
    Results       []FetchResult `json:"results"`
}

// PolicyCheckResult is one check result
type PolicyCheckResult struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Severity    string `json:"severity"`
    Passed      bool   `json:"passed"`
    Error       string `json:"error,omitempty"`
    DurationMs  int64  `json:"duration_ms"`
}

// PolicySummary aggregates results
type PolicySummary struct {
    Total    int `json:"total"`
    Passed   int `json:"passed"`
    Failed   int `json:"failed"`
    Warnings int `json:"warnings"`
}

// PolicyReport wraps policy execution
type PolicyReport struct {
    Results []PolicyCheckResult `json:"results"`
    Summary PolicySummary       `json:"summary"`
}

// PolicyResponse returned by /v1/policy/check
type PolicyResponse struct {
    SchemaVersion   string       `json:"schema_version"`
    Report          PolicyReport `json:"report"`
    FailedThreshold bool         `json:"failed_threshold"`
}

// ExecResult holds per‑repo execution result
type ExecResult struct {
    Repo       string `json:"repo"`
    Path       string `json:"path"`
    Success    bool   `json:"success"`
    Error      string `json:"error,omitempty"`
    DurationMs int64  `json:"duration_ms"`
}

// ExecResponse wraps exec results
type ExecResponse struct {
    SchemaVersion string       `json:"schema_version"`
    Results       []ExecResult `json:"results"`
}

