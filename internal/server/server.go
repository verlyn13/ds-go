package server

import (
    "bufio"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "time"

    "github.com/verlyn13/ds-go/internal/config"
    "github.com/verlyn13/ds-go/internal/policy"
    "github.com/verlyn13/ds-go/internal/runner"
    "github.com/verlyn13/ds-go/internal/scan"
    "os"
)

type Server struct {
    cfg         *config.Config
    workerCount int
    token       string
    started     time.Time
    corsEnabled bool
}

func New(cfg *config.Config, workers int) *Server {
    if workers <= 0 { workers = 10 }
    return &Server{cfg: cfg, workerCount: workers}
}

func (s *Server) Start(addr string) error {
    if s.started.IsZero() { s.started = time.Now() }
    mux := http.NewServeMux()

    mux.HandleFunc("/v1/capabilities", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        s.writeJSON(w, http.StatusOK, map[string]interface{}{
            "version": 1,
            "schema": "v1",
            "endpoints": []string{
                "/v1/capabilities",
                "/v1/health",
                "/v1/status",
                "/v1/status/stream",
                "/v1/status/sse",
                "/v1/scan",
                "/v1/organize/plan",
                "/v1/organize/apply",
                "/v1/fetch",
                "/v1/fetch/sse",
                "/v1/policy/check",
                "/v1/exec",
            },
            "timestamp": time.Now().UTC(),
            "openapi_url": "/openapi.yaml",
            "schema_version": "ds.v1",
        })
    }))

    // Health endpoint
    mux.HandleFunc("/v1/health", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        up := time.Since(s.started).Seconds()
        s.writeJSON(w, http.StatusOK, map[string]interface{}{
            "ok": true,
            "version": 1,
            "uptime_sec": int(up),
            "workers": s.workerCount,
            "auth": s.token != "",
            "timestamp": time.Now().UTC(),
            "schema_version": "ds.v1",
        })
    }))

    // OpenAPI exposure
    mux.HandleFunc("/openapi.yaml", s.wrapAuth(serveOpenAPI))
    mux.HandleFunc("/api/discovery/openapi", s.wrapAuth(serveOpenAPI))

    // Discovery metadata (minimal)
    mux.HandleFunc("/api/discovery/capabilities", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        s.writeJSON(w, http.StatusOK, map[string]interface{}{
            "contractVersion": 1,
            "schemaVersion": "v1",
            "openapi_url": "/openapi.yaml",
            "endpoints": []string{"/v1/status", "/v1/scan", "/v1/fetch", "/v1/organize/plan", "/v1/policy/check", "/v1/exec"},
        })
    }))

    // Services descriptor (single-service self description for convenience)
    mux.HandleFunc("/api/discovery/services", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        base := "http://" + r.Host
        resp := map[string]interface{}{
            "ds": map[string]string{
                "url":          base,
                "well_known":   base + "/.well-known/obs-bridge.json",
                "openapi":      base + "/openapi.yaml",
                "capabilities": base + "/v1/capabilities",
                "health":       base + "/v1/health",
                "self_status":  base + "/api/self-status",
            },
            "ds_token_present": (s.token != ""),
            "ts": time.Now().UnixMilli(),
        }
        s.writeJSON(w, http.StatusOK, resp)
    }))

    // Well-known bridge descriptor
    mux.HandleFunc("/.well-known/obs-bridge.json", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        s.writeJSON(w, http.StatusOK, map[string]interface{}{
            "contractVersion": 1,
            "schemaVersion": "v1",
            "openapi_url": "/openapi.yaml",
            "capabilities_url": "/api/discovery/capabilities",
            "endpoints": map[string]string{
                "openapi": "/openapi.yaml",
                "capabilities": "/v1/capabilities",
                "health": "/v1/health",
            },
            "all": []string{
                "/v1/health",
                "/v1/status",
                "/v1/status/stream",
                "/v1/status/sse",
                "/v1/scan",
                "/v1/fetch",
                "/v1/fetch/sse",
                "/v1/organize/plan",
                "/v1/organize/apply",
                "/v1/policy/check",
                "/v1/exec",
            },
        })
    }))

    // Self-status for MCP-style probes
    mux.HandleFunc("/api/self-status", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        now := time.Now()
        s.writeJSON(w, http.StatusOK, map[string]interface{}{
            "service": "ds",
            "ok": true,
            "nowMs": now.UnixMilli(),
            "auth": map[string]any{
                "tokenRequired": s.token != "",
                "corsEnabled": s.corsEnabled,
            },
            "endpoints": map[string]string{
                "openapi": "/openapi.yaml",
                "well_known": "/.well-known/obs-bridge.json",
                "capabilities": "/v1/capabilities",
                "health": "/v1/health",
                "status": "/v1/status",
                "fetch": "/v1/fetch",
                "organizePlan": "/v1/organize/plan",
                "organizeApply": "/v1/organize/apply",
                "policyCheck": "/v1/policy/check",
                "exec": "/v1/exec",
            },
            "schema_version": "ds.v1",
        })
    }))

    mux.HandleFunc("/v1/status", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        scanner := scan.New(s.cfg, s.workerCount)
        path := r.URL.Query().Get("path")
        account := r.URL.Query().Get("account")
        dirty := r.URL.Query().Get("dirty") == "true"
        repos, err := scanner.Scan(path)
        if err != nil { s.writeErr(w, err); return }
        if dirty {
            repos = filterDirty(repos)
        }
        if account != "" {
            repos = filterByAccount(repos, account)
        }
        s.writeJSONVersioned(w, r, http.StatusOK, repos)
    }))

    mux.HandleFunc("/v1/status/stream", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        scanner := scan.New(s.cfg, s.workerCount)
        path := r.URL.Query().Get("path")
        account := r.URL.Query().Get("account")
        dirty := r.URL.Query().Get("dirty") == "true"
        repos, err := scanner.Scan(path)
        if err != nil { s.writeErr(w, err); return }
        if dirty { repos = filterDirty(repos) }
        if account != "" { repos = filterByAccount(repos, account) }
        w.Header().Set("Content-Type", "application/x-ndjson")
        bw := bufio.NewWriter(w)
        enc := json.NewEncoder(bw)
        for _, repo := range repos {
            if err := enc.Encode(repo); err != nil {
                return
            }
            bw.Flush()
        }
    }))

    mux.HandleFunc("/v1/status/sse", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        scanner := scan.New(s.cfg, s.workerCount)
        path := r.URL.Query().Get("path")
        account := r.URL.Query().Get("account")
        dirty := r.URL.Query().Get("dirty") == "true"
        repos, err := scanner.Scan(path)
        if err != nil { s.writeErr(w, err); return }
        if dirty { repos = filterDirty(repos) }
        if account != "" { repos = filterByAccount(repos, account) }
        sseStart(w)
        for _, repo := range repos {
            if err := sseData(w, repo, "repo"); err != nil { return }
        }
    }))

    mux.HandleFunc("/v1/scan", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        scanner := scan.New(s.cfg, s.workerCount)
        path := r.URL.Query().Get("path")
        repos, err := scanner.Scan(path)
        if err != nil { s.writeErr(w, err); return }
        if err := scanner.SaveIndex(repos); err != nil { s.writeErr(w, err); return }
        s.writeJSONVersioned(w, r, http.StatusOK, map[string]int{"count": len(repos)})
    }))

    mux.HandleFunc("/v1/organize/plan", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        scanner := scan.New(s.cfg, s.workerCount)
        path := r.URL.Query().Get("path")
        requireClean := r.URL.Query().Get("require_clean") == "true"
        repos, err := scanner.Scan(path)
        if err != nil { s.writeErr(w, err); return }
        if requireClean {
            for _, r := range repos {
                if !r.IsClean { s.writeErr(w, fmt.Errorf("require-clean: '%s' has uncommitted changes", r.Name)); return }
            }
        }
        plan := scan.OrganizePlanJSON(repos, s.cfg)
        s.writeJSONVersioned(w, r, http.StatusOK, plan)
    }))

    mux.HandleFunc("/v1/organize/apply", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        scanner := scan.New(s.cfg, s.workerCount)
        path := r.URL.Query().Get("path")
        requireClean := r.URL.Query().Get("require_clean") == "true"
        force := r.URL.Query().Get("force") == "true"
        dryRun := r.URL.Query().Get("dry_run") == "true"
        repos, err := scanner.Scan(path)
        if err != nil { s.writeErr(w, err); return }
        if requireClean {
            for _, r := range repos {
                if !r.IsClean { s.writeErr(w, fmt.Errorf("require-clean: '%s' has uncommitted changes", r.Name)); return }
            }
        }
        results, moved, failed := scan.ApplyOrganizePlan(repos, s.cfg, dryRun, force)
        s.writeJSONVersioned(w, r, http.StatusOK, map[string]interface{}{
            "moved": moved,
            "failed": failed,
            "results": results,
        })
    }))

    mux.HandleFunc("/v1/fetch", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        scanner := scan.New(s.cfg, s.workerCount)
        path := r.URL.Query().Get("path")
        account := r.URL.Query().Get("account")
        dirty := r.URL.Query().Get("dirty") == "true"
        repos, err := scanner.Scan(path)
        if err != nil { s.writeErr(w, err); return }
        if dirty { repos = filterDirty(repos) }
        if account != "" { repos = filterByAccount(repos, account) }
        fetcher := scan.NewFetcher(s.workerCount)
        results := fetcher.FetchAll(repos, false)
        s.writeJSONVersioned(w, r, http.StatusOK, map[string]interface{}{"results": results})
    }))

    mux.HandleFunc("/v1/fetch/sse", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        scanner := scan.New(s.cfg, s.workerCount)
        path := r.URL.Query().Get("path")
        account := r.URL.Query().Get("account")
        dirty := r.URL.Query().Get("dirty") == "true"
        repos, err := scanner.Scan(path)
        if err != nil { s.writeErr(w, err); return }
        if dirty { repos = filterDirty(repos) }
        if account != "" { repos = filterByAccount(repos, account) }
        fetcher := scan.NewFetcher(s.workerCount)
        sseStart(w)
        ctx := r.Context()
        stream := fetcher.FetchAllStream(ctx, repos)
        for res := range stream {
            if err := sseData(w, res, "fetch"); err != nil { return }
        }
    }))

    mux.HandleFunc("/v1/policy/check", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        file := r.URL.Query().Get("file")
        if file == "" { file = ".project-compliance.yaml" }
        failOn := r.URL.Query().Get("fail_on")
        if failOn == "" { failOn = "critical" }
        cfg, err := policy.Load(file)
        if err != nil { s.writeErr(w, err); return }
        report, err := policy.RunChecks(cfg)
        if err != nil { s.writeErr(w, err); return }
        // Include a fail flag in response
        th, err := policy.SeverityFromString(failOn)
        if err != nil { s.writeErr(w, err); return }
        shouldFail := policy.FailIfAboveSeverity(report, th, cfg)
        s.writeJSONVersioned(w, r, http.StatusOK, map[string]interface{}{
            "report": report,
            "failed_threshold": shouldFail,
        })
    }))

    mux.HandleFunc("/v1/exec", s.wrapAuth(func(w http.ResponseWriter, r *http.Request) {
        // Accept command via query param or POST JSON body {"cmd":"..."}
        cmdStr := r.URL.Query().Get("cmd")
        if cmdStr == "" && r.Method == http.MethodPost {
            var body struct{ Cmd string `json:"cmd"` }
            if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
                cmdStr = body.Cmd
            }
        }
        if cmdStr == "" { s.writeErr(w, fmt.Errorf("missing cmd")); return }

        scanner := scan.New(s.cfg, s.workerCount)
        path := r.URL.Query().Get("path")
        account := r.URL.Query().Get("account")
        dirty := r.URL.Query().Get("dirty") == "true"
        timeoutSec, _ := strconv.Atoi(r.URL.Query().Get("timeout"))
        repos, err := scanner.Scan(path)
        if err != nil { s.writeErr(w, err); return }
        if dirty { repos = filterDirty(repos) }
        if account != "" { repos = filterByAccount(repos, account) }
        results := runner.ExecInRepos(repos, cmdStr, time.Duration(timeoutSec)*time.Second)
        s.writeJSONVersioned(w, r, http.StatusOK, map[string]interface{}{"results": results})
    }))

    // Optional CORS support for dashboard dev
    handler := http.Handler(mux)
    if v := getenv("DS_CORS"); v == "1" || v == "true" { s.corsEnabled = true }
    if s.corsEnabled {
        handler = s.wrapCORS(handler)
    }
    srv := &http.Server{Addr: addr, Handler: handler}
    log.Printf("ds serve listening on %s", addr)
    return srv.ListenAndServe()
}

func (s *Server) writeJSON(w http.ResponseWriter, code int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    if err := enc.Encode(v); err != nil {
        fmt.Fprintf(w, `{"ok":false,"error":"%s"}`, err.Error())
    }
}

func (s *Server) writeErr(w http.ResponseWriter, err error) {
    s.writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": err.Error()})
}

// local filter helpers (avoid importing from cmd)
func filterDirty(repos []scan.Repository) []scan.Repository {
    var out []scan.Repository
    for _, r := range repos { if !r.IsClean { out = append(out, r) } }
    return out
}
func filterByAccount(repos []scan.Repository, account string) []scan.Repository {
    var out []scan.Repository
    for _, r := range repos { if r.Account == account { out = append(out, r) } }
    return out
}

// SSE helpers
func sseStart(w http.ResponseWriter) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
}

func sseData(w http.ResponseWriter, v interface{}, event string) error {
    if event != "" {
        if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil { return err }
    }
    b, err := json.Marshal(v)
    if err != nil { return err }
    if _, err := fmt.Fprintf(w, "data: %s\n\n", string(b)); err != nil { return err }
    if f, ok := w.(http.Flusher); ok { f.Flush() }
    return nil
}

// WithToken sets an optional bearer token; when set, all endpoints require Authorization: Bearer <token>
func (s *Server) WithToken(token string) *Server { s.token = token; return s }

// wrapAuth enforces bearer token when configured
func (s *Server) wrapAuth(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if s.token != "" {
            auth := r.Header.Get("Authorization")
            expected := "Bearer " + s.token
            if auth != expected {
                w.Header().Set("WWW-Authenticate", "Bearer")
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
        }
        h(w, r)
    }
}

// wrapCORS enables permissive CORS for local dashboards
func (s *Server) wrapCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,Accept")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// getenv is a tiny wrapper to avoid importing os repeatedly
func getenv(k string) string {
    if v := getEnv(k); v != "" { return v }
    return ""
}

// indirection for testing
var getEnv = func(k string) string { return os.Getenv(k) }

// writeJSONMaybeEnvelope writes either raw JSON or an envelope with schema_version and data
func (s *Server) writeJSONMaybeEnvelope(w http.ResponseWriter, r *http.Request, code int, v interface{}) {
    // Deprecated behavior retained for compatibility; now always include schema_version
    s.writeJSONVersioned(w, r, code, v)
}

// writeJSONVersioned ensures a top-level schema_version. Arrays are wrapped as {schema_version, data}.
func (s *Server) writeJSONVersioned(w http.ResponseWriter, r *http.Request, code int, v interface{}) {
    // If already a map, attach schema_version if missing
    if m, ok := v.(map[string]interface{}); ok {
        if _, exists := m["schema_version"]; !exists {
            m["schema_version"] = "ds.v1"
        }
        s.writeJSON(w, code, m)
        return
    }
    // Default: wrap
    s.writeJSON(w, code, map[string]interface{}{
        "schema_version": "ds.v1",
        "data":           v,
    })
}
