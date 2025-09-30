# dsclient (Go) — DS Local API Client

Typed Go client for the DS local API (schema_version "ds.v1").

Install

```go
import "github.com/verlyn13/ds-go/pkg/dsclient"
```

Quick Start

```go
ctx := context.Background()
c := dsclient.New("http://127.0.0.1:7777", dsclient.WithToken(os.Getenv("DS_TOKEN")))

// Health
h, err := c.Health(ctx)
if err != nil { log.Fatal(err) }
fmt.Println("ok:", h.OK, "uptime:", h.UptimeSec)

// Status (dirty=true)
s, err := c.Status(ctx, "", "", true)
if err != nil { log.Fatal(err) }
fmt.Println("schema:", s.SchemaVersion, "dirty repos:", len(s.Data))

// Scan
scan, err := c.Scan(ctx, "")
fmt.Println("scanned repos:", scan.Count)

// Organize (plan only)
plan, err := c.OrganizePlan(ctx, true, "")
fmt.Println("moves:", len(plan.Data))

// Policy
pol, err := c.PolicyCheck(ctx, ".project-compliance.yaml", "critical")
fmt.Println("failed threshold:", pol.FailedThreshold)

// Exec across repos
execRes, err := c.Exec(ctx, "mise run lint", nil)
fmt.Println("results:", len(execRes.Results))
```

Notes
- All responses include `SchemaVersion` with value `"ds.v1"`.
- Array responses use `{ schema_version, data: [...] }` wrappers.
- Set `DS_TOKEN` and pass WithToken() to enable Authorization.

