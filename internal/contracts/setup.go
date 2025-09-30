package contracts

import (
    "net/http"
    "os"
)

// SetupEnforcement adds contract enforcement to your HTTP server
func SetupEnforcement(handler http.Handler) http.Handler {
    mode := ModeMonitor // Start with monitoring
    if os.Getenv("CONTRACT_ENFORCE") == "true" {
        mode = ModeEnforce
    }

    enforcer := NewUniversalContractEnforcer(
        WithMode(mode),
        WithServiceName("ds-go"),
    )

    return enforcer.Middleware(handler)
}
