package dsclient

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

func TestHealthAndToken(t *testing.T) {
    var gotAuth string
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/v1/health" {
            gotAuth = r.Header.Get("Authorization")
            _ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "schema_version": "ds.v1"})
            return
        }
        w.WriteHeader(http.StatusNotFound)
    }))
    defer srv.Close()

    c := New(srv.URL, WithToken("t0ken"))
    out, err := c.Health(context.Background())
    if err != nil { t.Fatalf("health: %v", err) }
    if out.SchemaVersion != "ds.v1" { t.Fatalf("missing schema_version") }
    if gotAuth != "Bearer t0ken" { t.Fatalf("auth header not set: %q", gotAuth) }
}

func TestSelfStatus(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/api/self-status" {
            _ = json.NewEncoder(w).Encode(map[string]any{
                "service": "ds",
                "ok": true,
                "nowMs": time.Now().UnixMilli(),
                "schema_version": "ds.v1",
            })
            return
        }
        w.WriteHeader(http.StatusNotFound)
    }))
    defer srv.Close()

    c := New(srv.URL)
    out, err := c.SelfStatus(context.Background())
    if err != nil { t.Fatalf("self-status: %v", err) }
    if out["service"] != "ds" { t.Fatalf("missing service") }
    if out["schema_version"] != "ds.v1" { t.Fatalf("missing schema_version") }
    if _, ok := out["nowMs"].(float64); !ok { t.Fatalf("nowMs not a number") }
}

func TestDiscovery(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/api/discovery/services" {
            _ = json.NewEncoder(w).Encode(map[string]any{
                "ds": map[string]string{
                    "url": "http://127.0.0.1:7777",
                    "well_known": "http://127.0.0.1:7777/.well-known/obs-bridge.json",
                    "openapi": "http://127.0.0.1:7777/openapi.yaml",
                    "health": "http://127.0.0.1:7777/v1/health",
                    "self_status": "http://127.0.0.1:7777/api/self-status",
                },
                "ds_token_present": false,
                "ts": time.Now().UnixMilli(),
            })
            return
        }
        w.WriteHeader(http.StatusNotFound)
    }))
    defer srv.Close()

    c := New(srv.URL)
    out, err := c.Discovery(context.Background())
    if err != nil { t.Fatalf("discovery: %v", err) }
    if _, ok := out["ds"].(map[string]interface{}); !ok { t.Fatalf("missing ds service info") }
    if _, ok := out["ts"].(float64); !ok { t.Fatalf("ts not a number") }
}

func TestCapabilities(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/v1/capabilities" {
            _ = json.NewEncoder(w).Encode(map[string]any{
                "version": 1,
                "schema": "v1",
                "schema_version": "ds.v1",
                "endpoints": []string{
                    "/v1/health",
                    "/v1/status",
                    "/v1/scan",
                },
            })
            return
        }
        w.WriteHeader(http.StatusNotFound)
    }))
    defer srv.Close()

    c := New(srv.URL)
    out, err := c.Capabilities(context.Background())
    if err != nil { t.Fatalf("capabilities: %v", err) }
    if out["schema_version"] != "ds.v1" { t.Fatalf("missing schema_version") }
    if endpoints, ok := out["endpoints"].([]interface{}); !ok || len(endpoints) < 3 {
        t.Fatalf("missing endpoints")
    }
}

func TestStatusTyped(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/v1/status" {
            _ = json.NewEncoder(w).Encode(StatusResponse{SchemaVersion: "ds.v1", Data: []Repository{}})
            return
        }
        w.WriteHeader(http.StatusNotFound)
    }))
    defer srv.Close()

    c := New(srv.URL)
    out, err := c.Status(context.Background(), "", "", false)
    if err != nil { t.Fatalf("status: %v", err) }
    if out.SchemaVersion != "ds.v1" { t.Fatalf("missing schema_version") }
}
