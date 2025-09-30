package server

import (
    "crypto/sha256"
    "encoding/hex"
    "net/http"
    _ "embed"
)

//go:embed openapi.yaml
var openAPISpec []byte

var openAPIETag = func() string {
    sum := sha256.Sum256(openAPISpec)
    return `W/"` + hex.EncodeToString(sum[:]) + `"`
}()

func serveOpenAPI(w http.ResponseWriter, r *http.Request) {
    // ETag handling
    if tag := r.Header.Get("If-None-Match"); tag != "" && tag == openAPIETag {
        w.WriteHeader(http.StatusNotModified)
        return
    }
    w.Header().Set("Content-Type", "application/yaml")
    w.Header().Set("ETag", openAPIETag)
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write(openAPISpec)
}
