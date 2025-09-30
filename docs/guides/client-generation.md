# DS Client Generation Guide

This guide explains how to generate TypeScript clients for the DS (ds-go) API.

## Prerequisites

You need one of the following tools installed:
- **openapi-typescript** - For TypeScript type definitions only
- **openapi-generator-cli** - For full client generation (Axios, Fetch, etc.)
- **Docker** - Alternative for containerized generation

Install via npm:
```bash
# For types only
npm install -D openapi-typescript

# For full client
npm install -g @openapitools/openapi-generator-cli
```

## Quick Start

### 1. Start the DS API Server
```bash
# Build the latest DS binary
make build

# Start the API server
./ds serve --addr 127.0.0.1:7777
```

The OpenAPI spec will be available at:
- `http://127.0.0.1:7777/openapi.yaml`
- `http://127.0.0.1:7777/v1/capabilities` (includes openapi_url)

### 2. Generate Client

#### TypeScript Types Only
```bash
# Generate type definitions
./scripts/generate-openapi-client-ds.sh \
  examples/dashboard/generated/ds-types \
  openapi.yaml \
  types
```

#### Full Axios Client
```bash
# Generate axios-based client
./scripts/generate-openapi-client-ds.sh \
  examples/dashboard/generated/ds-client \
  openapi.yaml \
  typescript-axios
```

#### Full Fetch Client
```bash
# Generate fetch-based client
./scripts/generate-openapi-client-ds.sh \
  examples/dashboard/generated/ds-client \
  openapi.yaml \
  typescript-fetch
```

## Usage Examples

### Using Generated Types
```typescript
import type { paths } from './generated/ds-types/types';

// Type-safe response types
type StatusResponse = paths['/status']['get']['responses']['200']['content']['application/json'];
type FetchResponse = paths['/fetch']['get']['responses']['200']['content']['application/json'];

// Use with your preferred HTTP client
const response = await fetch('http://127.0.0.1:7777/v1/status?dirty=true');
const data: StatusResponse = await response.json();
```

### Using Axios Client
```typescript
import { Configuration, DefaultApi } from './generated/ds-client';

// Configure the client
const config = new Configuration({
  basePath: 'http://127.0.0.1:7777/v1',
  // Optional: Add auth token if required
  accessToken: process.env.DS_TOKEN,
});

// Create API instance
const api = new DefaultApi(config);

// Make API calls
const status = await api.getStatus({ dirty: true });
const capabilities = await api.getCapabilities();
const scanResult = await api.scanRepositories({ path: '~/Development' });
```

### Using with DS Adapter Pattern
```typescript
// dsAdapter.ts
import { Configuration, DefaultApi } from './generated/ds-client';

export class DSAdapter {
  private api: DefaultApi;

  constructor(basePath = 'http://127.0.0.1:7777/v1') {
    const config = new Configuration({ basePath });
    this.api = new DefaultApi(config);
  }

  async getDirtyRepos() {
    const response = await this.api.getStatus({ dirty: true });
    return response.data.repositories;
  }

  async fetchAll() {
    const response = await this.api.fetchRepositories();
    return response.data;
  }

  async checkCompliance() {
    const response = await this.api.checkPolicy();
    return response.data;
  }
}
```

## Authentication

If the DS server requires authentication (via `--token` flag or `DS_TOKEN` environment variable):

```typescript
const config = new Configuration({
  basePath: 'http://127.0.0.1:7777/v1',
  accessToken: 'your-token-here',
  // Or use a function for dynamic tokens
  accessToken: async () => {
    return await getTokenFromSomewhere();
  },
});
```

## Streaming Endpoints

DS provides streaming endpoints for real-time updates:

### Server-Sent Events (SSE)
```typescript
const evtSource = new EventSource('http://127.0.0.1:7777/v1/status/sse');
evtSource.onmessage = (event) => {
  const repo = JSON.parse(event.data);
  console.log('Repository update:', repo);
};
```

### NDJSON Stream
```typescript
const response = await fetch('http://127.0.0.1:7777/v1/status/stream');
const reader = response.body.getReader();
const decoder = new TextDecoder();

while (true) {
  const { done, value } = await reader.read();
  if (done) break;

  const lines = decoder.decode(value).split('\n');
  for (const line of lines) {
    if (line) {
      const repo = JSON.parse(line);
      console.log('Repository:', repo);
    }
  }
}
```

## Integration with Dashboard

The dashboard can use the generated DS client alongside the Bridge client:

```typescript
// services/api.ts
import { BridgeAdapter } from './bridgeAdapter';
import { DSAdapter } from './dsAdapter';

export const bridge = new BridgeAdapter();
export const ds = new DSAdapter();

// Use in components
const repos = await ds.getDirtyRepos();
const services = await bridge.getServices();
```

## Regeneration

When the DS API changes:

1. Rebuild DS with the latest changes
2. Restart the server
3. Regenerate the client:
   ```bash
   ./scripts/generate-openapi-client-ds.sh \
     path/to/output \
     openapi.yaml \
     typescript-axios
   ```

## Available Generators

The `generate-openapi-client-ds.sh` script supports various generators:

- `types` or `openapi-typescript` - TypeScript types only
- `typescript-axios` - Axios-based client (recommended)
- `typescript-fetch` - Fetch-based client
- `typescript-node` - Node.js client
- `typescript-rxjs` - RxJS observables client

## Troubleshooting

### Cannot find OpenAPI spec
- Ensure DS server is running: `./ds serve`
- Check the server is accessible: `curl http://127.0.0.1:7777/openapi.yaml`
- Verify no auth token is required or provide one

### Generation fails
- Install required tools: `npm install -g @openapitools/openapi-generator-cli`
- Try Docker alternative if npm tools fail
- Check the OpenAPI spec is valid: `npx @apidevtools/swagger-cli validate openapi.yaml`

### Type mismatches
- Regenerate the client after API changes
- Ensure you're using the latest OpenAPI spec
- Check version compatibility between client and server