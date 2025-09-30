package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "time"

    "github.com/verlyn13/ds-go/pkg/dsclient"
)

func main() {
    base := getenv("DS_BASE_URL", "http://127.0.0.1:7777")
    token := os.Getenv("DS_TOKEN")
    c := dsclient.New(base, dsclient.WithToken(token))

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // 1. Check capabilities
    fmt.Println("=== Capabilities ===")
    caps, err := c.Capabilities(ctx)
    if err != nil {
        fmt.Fprintf(os.Stderr, "capabilities error: %v\n", err)
        os.Exit(1)
    }
    if schemaVersion, ok := caps["schema_version"].(string); ok {
        fmt.Printf("Schema Version: %s\n", schemaVersion)
    }
    if endpoints, ok := caps["endpoints"].([]interface{}); ok {
        fmt.Printf("Available endpoints: %d\n", len(endpoints))
    }

    // 2. Check health
    fmt.Println("\n=== Health ===")
    health, err := c.Health(ctx)
    if err != nil {
        fmt.Fprintf(os.Stderr, "health error: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Server OK: %v\n", health.OK)
    fmt.Printf("Uptime: %d seconds\n", health.UptimeSec)

    // If token is required but missing, advise and exit
    if health.Auth && token == "" {
        fmt.Fprintln(os.Stderr, "DS requires a token. Set DS_TOKEN and retry.")
        os.Exit(2)
    }

    // 3. Self-status
    fmt.Println("\n=== Self Status ===")
    selfStatus, err := c.SelfStatus(ctx)
    if err != nil {
        fmt.Fprintf(os.Stderr, "self-status error: %v\n", err)
        os.Exit(1)
    }
    if nowMs, ok := selfStatus["nowMs"].(float64); ok {
        t := time.Unix(0, int64(nowMs)*int64(time.Millisecond))
        fmt.Printf("Server Time: %s\n", t.Format(time.RFC3339))
    }
    if schemaVersion, ok := selfStatus["schema_version"].(string); ok {
        fmt.Printf("Schema Version: %s\n", schemaVersion)
    }

    // 4. Discovery
    fmt.Println("\n=== Discovery ===")
    discovery, err := c.Discovery(ctx)
    if err != nil {
        fmt.Fprintf(os.Stderr, "discovery error: %v\n", err)
        os.Exit(1)
    }
    if ds, ok := discovery["ds"].(map[string]interface{}); ok {
        fmt.Printf("DS Service URLs:\n")
        for key, val := range ds {
            fmt.Printf("  %s: %v\n", key, val)
        }
    }
    if ts, ok := discovery["ts"].(float64); ok {
        t := time.Unix(0, int64(ts)*int64(time.Millisecond))
        fmt.Printf("Discovery Timestamp: %s\n", t.Format(time.RFC3339))
    }

    // 5. Repository status
    fmt.Println("\n=== Repository Status ===")
    repos, err := c.Status(ctx, "", "", true)
    if err != nil {
        fmt.Fprintf(os.Stderr, "status error: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Dirty repos: %d\n", len(repos.Data))

    // Show first dirty repo if any
    if len(repos.Data) > 0 {
        fmt.Println("\nFirst dirty repo:")
        b, _ := json.MarshalIndent(repos.Data[0], "  ", "  ")
        fmt.Printf("  %s\n", string(b))
    }
}

func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" { return v }
    return def
}
