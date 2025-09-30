# Apply DS Status + Contracts Patch

This patch adds DS discovery/health proxies on the dashboard server, a DsStatusCard, and a Contracts browser route.

Prerequisites
- Your dashboard repo should have:
  - Node server under `server/index.js` (Express) that mounts routes
  - React app with `src/components/Documentation.jsx` and `src/App.jsx`

How to apply
1) From your dashboard repo root:
   git apply ../system-setup-update/examples/dashboard/patches/ds-status-card.patch

2) Set environment for the dashboard server (if proxying through the bridge):
   export OBS_BRIDGE_URL=http://127.0.0.1:7171
   # Optional
   export BRIDGE_TOKEN=...
   export DS_TOKEN=...

3) Start the dashboard server and app. Visit:
   - /docs → DsStatusCard should render with links to OpenAPI/Capabilities/Well-known
   - /contracts → Contract list and schema viewer

Notes
- If your server/app file paths differ, adjust the patch paths accordingly.
- The server routes call the bridge’s `/api/discovery/services` endpoint, and then call DS with `DS_TOKEN` when set.
- For local development against DS directly, you can adjust `/api/ds/*` handlers to point to DS_BASE_URL instead of the bridge discovery.

