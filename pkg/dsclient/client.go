package dsclient

import (
    "context"
    "encoding/json"
    "fmt"
    "bytes"
    "net/http"
    "net/url"
    "time"
)

// Client is a minimal HTTP client for the ds local API.
type Client struct {
    BaseURL    string
    Token      string
    HTTPClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithToken sets the bearer token.
func WithToken(token string) Option { return func(c *Client) { c.Token = token } }

// WithHTTPClient provides a custom http.Client.
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.HTTPClient = h } }

// New creates a new Client.
func New(base string, opts ...Option) *Client {
    c := &Client{BaseURL: base, HTTPClient: &http.Client{Timeout: 30 * time.Second}}
    for _, opt := range opts { opt(c) }
    return c
}

// Health fetches /v1/health.
func (c *Client) Health(ctx context.Context) (HealthResponse, error) {
    var out HealthResponse
    return out, c.get(ctx, "/v1/health", nil, &out)
}

// SelfStatus fetches /api/self-status.
func (c *Client) SelfStatus(ctx context.Context) (map[string]any, error) {
    var out map[string]any
    if err := c.get(ctx, "/api/self-status", nil, &out); err != nil { return nil, err }
    return out, nil
}

// Discovery fetches /api/discovery/services.
func (c *Client) Discovery(ctx context.Context) (map[string]any, error) {
    var out map[string]any
    if err := c.get(ctx, "/api/discovery/services", nil, &out); err != nil { return nil, err }
    return out, nil
}

// Capabilities fetches /v1/capabilities.
func (c *Client) Capabilities(ctx context.Context) (map[string]any, error) {
    var out map[string]any
    if err := c.get(ctx, "/v1/capabilities", nil, &out); err != nil { return nil, err }
    return out, nil
}

// Status fetches /v1/status with optional filters.
func (c *Client) Status(ctx context.Context, account, path string, dirty bool) (StatusResponse, error) {
    q := url.Values{}
    if account != "" { q.Set("account", account) }
    if path != "" { q.Set("path", path) }
    if dirty { q.Set("dirty", "true") }
    var out StatusResponse
    return out, c.get(ctx, "/v1/status", q, &out)
}

// Scan triggers /v1/scan and returns count.
func (c *Client) Scan(ctx context.Context, path string) (ScanResponse, error) {
    q := url.Values{}
    if path != "" { q.Set("path", path) }
    var out ScanResponse
    return out, c.get(ctx, "/v1/scan", q, &out)
}

// OrganizePlan gets planned moves.
func (c *Client) OrganizePlan(ctx context.Context, requireClean bool, path string) (OrganizePlanResponse, error) {
    q := url.Values{}
    if requireClean { q.Set("require_clean", "true") }
    if path != "" { q.Set("path", path) }
    var out OrganizePlanResponse
    return out, c.get(ctx, "/v1/organize/plan", q, &out)
}

// OrganizeApply applies planned moves.
func (c *Client) OrganizeApply(ctx context.Context, requireClean, force, dryRun bool, path string) (OrganizeApplyResponse, error) {
    q := url.Values{}
    if requireClean { q.Set("require_clean", "true") }
    if force { q.Set("force", "true") }
    if dryRun { q.Set("dry_run", "true") }
    if path != "" { q.Set("path", path) }
    var out OrganizeApplyResponse
    return out, c.post(ctx, "/v1/organize/apply", q, nil, &out)
}

// PolicyCheck runs policy check.
func (c *Client) PolicyCheck(ctx context.Context, file, failOn string) (PolicyResponse, error) {
    if file == "" { file = ".project-compliance.yaml" }
    if failOn == "" { failOn = "critical" }
    q := url.Values{"file": {file}, "fail_on": {failOn}}
    var out PolicyResponse
    return out, c.get(ctx, "/v1/policy/check", q, &out)
}

// Exec runs a command across repositories.
func (c *Client) Exec(ctx context.Context, cmd string, q url.Values) (ExecResponse, error) {
    body := map[string]string{"cmd": cmd}
    var out ExecResponse
    return out, c.post(ctx, "/v1/exec", q, body, &out)
}

// Helpers
func (c *Client) get(ctx context.Context, path string, q url.Values, dst any) error {
    u := c.BaseURL + path
    if q != nil && len(q) > 0 { u += "?" + q.Encode() }
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
    req.Header.Set("Accept", "application/json")
    if c.Token != "" { req.Header.Set("Authorization", "Bearer "+c.Token) }
    resp, err := c.HTTPClient.Do(req)
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode >= 400 { return fmt.Errorf("GET %s: HTTP %d", path, resp.StatusCode) }
    return json.NewDecoder(resp.Body).Decode(dst)
}

func (c *Client) post(ctx context.Context, path string, q url.Values, body any, dst any) error {
    u := c.BaseURL + path
    if q != nil && len(q) > 0 { u += "?" + q.Encode() }
    var req *http.Request
    if body != nil {
        b, _ := json.Marshal(body)
        req, _ = http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
        req.Header.Set("Content-Type", "application/json")
    } else {
        req, _ = http.NewRequestWithContext(ctx, http.MethodPost, u, nil)
    }
    req.Header.Set("Accept", "application/json")
    if c.Token != "" { req.Header.Set("Authorization", "Bearer "+c.Token) }
    resp, err := c.HTTPClient.Do(req)
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode >= 400 { return fmt.Errorf("POST %s: HTTP %d", path, resp.StatusCode) }
    return json.NewDecoder(resp.Body).Decode(dst)
}
