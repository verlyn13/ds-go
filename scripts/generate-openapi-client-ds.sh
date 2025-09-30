#!/usr/bin/env bash
set -euo pipefail

# Generate a TypeScript client for DS (ds-go) OpenAPI.
# Requires either:
#  - npx openapi-typescript
#  - or openapi-generator-cli (npm pkg or Docker)

OUT_DIR=${1:-examples/dashboard/generated/ds-client}
SPEC=${2:-openapi.yaml}
GENERATOR=${3:-typescript-axios}

echo "Generating DS TypeScript client..."
echo "  Output: $OUT_DIR"
echo "  Spec: $SPEC"
echo "  Generator: $GENERATOR"

mkdir -p "$OUT_DIR"

# Try openapi-typescript first (generates types only)
if [[ "$GENERATOR" == "types" ]] || [[ "$GENERATOR" == "openapi-typescript" ]]; then
  if command -v npx >/dev/null 2>&1; then
    echo "Using openapi-typescript via npx..."
    npx openapi-typescript "$SPEC" --output "$OUT_DIR/types.ts"

    cat >"$OUT_DIR/README.md" <<EOF
# DS (ds-go) TypeScript Types

This folder contains TypeScript types generated from the DS OpenAPI specification.

## Generated Files
- \`types.ts\` - TypeScript type definitions for all DS API endpoints

## Regeneration
To regenerate these types:
\`\`\`bash
../../scripts/generate-openapi-client-ds.sh "$OUT_DIR" "$SPEC" types
\`\`\`

## Usage Example
\`\`\`typescript
import type { paths } from './types';

type StatusResponse = paths['/status']['get']['responses']['200']['content']['application/json'];
\`\`\`
EOF
    echo "✓ Generated DS types at $OUT_DIR/types.ts"
    exit 0
  fi
fi

# Try openapi-generator-cli for full client generation
if command -v openapi-generator-cli >/dev/null 2>&1; then
  echo "Using openapi-generator-cli for $GENERATOR..."
  openapi-generator-cli generate \
    -i "$SPEC" \
    -g "$GENERATOR" \
    -o "$OUT_DIR" \
    --additional-properties=npmName=ds-api-client,npmVersion=1.0.0,withInterfaces=true

  cat >"$OUT_DIR/README.md" <<EOF
# DS (ds-go) API Client

This folder contains a TypeScript client generated from the DS OpenAPI specification.

## Generated Client
- Full $GENERATOR client implementation
- Includes models, API classes, and configuration

## Regeneration
To regenerate this client:
\`\`\`bash
../../scripts/generate-openapi-client-ds.sh "$OUT_DIR" "$SPEC" $GENERATOR
\`\`\`

## Usage Example
\`\`\`typescript
import { Configuration, DefaultApi } from './';

const config = new Configuration({
  basePath: 'http://127.0.0.1:7777/v1'
});

const api = new DefaultApi(config);
const status = await api.getStatus({ dirty: true });
\`\`\`
EOF
  echo "✓ Generated DS $GENERATOR client at $OUT_DIR"
  exit 0
fi

# Try Docker as fallback
if command -v docker >/dev/null 2>&1; then
  echo "Using openapi-generator via Docker..."
  docker run --rm \
    -v "${PWD}:/local" \
    openapitools/openapi-generator-cli generate \
    -i "/local/$SPEC" \
    -g "$GENERATOR" \
    -o "/local/$OUT_DIR" \
    --additional-properties=npmName=ds-api-client,npmVersion=1.0.0,withInterfaces=true

  echo "✓ Generated DS $GENERATOR client at $OUT_DIR via Docker"
  exit 0
fi

echo "Error: No generator available. Install one of:"
echo "  - npm install -D openapi-typescript (for types only)"
echo "  - npm install -g @openapitools/openapi-generator-cli (for full client)"
echo "  - Docker (for containerized generation)"
exit 1