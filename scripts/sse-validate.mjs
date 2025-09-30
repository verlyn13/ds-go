#!/usr/bin/env node
// Minimal SSE validator for DS endpoints (status/fetch)
// Usage: DS_BASE_URL=http://127.0.0.1:7777 DS_TOKEN=... node scripts/sse-validate.mjs

const BASE = process.env.DS_BASE_URL || 'http://127.0.0.1:7777';
const TOKEN = process.env.DS_TOKEN || '';
const TIMEOUT_MS = Number(process.env.SSE_TIMEOUT_MS || 5000);
const MAX_EVENTS = Number(process.env.SSE_MAX_EVENTS || 1);

async function* sseStream(url) {
  const headers = { Accept: 'text/event-stream' };
  if (TOKEN) headers['Authorization'] = `Bearer ${TOKEN}`;
  const res = await fetch(url, { headers });
  if (!res.ok || !res.body) throw new Error(`HTTP ${res.status}`);
  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buf = '';
  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    buf += decoder.decode(value, { stream: true });
    let idx;
    while ((idx = buf.indexOf('\n\n')) >= 0) {
      const chunk = buf.slice(0, idx).trim();
      buf = buf.slice(idx + 2);
      yield chunk;
    }
  }
}

function parseEvent(block) {
  const ev = { event: '', data: '' };
  for (const line of block.split('\n')) {
    if (line.startsWith('event:')) ev.event = line.slice(6).trim();
    if (line.startsWith('data:')) ev.data = line.slice(5).trim();
  }
  return ev;
}

function validateRepo(obj) {
  const required = ['Path','Name','Account','IsClean','Ahead','Behind','HasUpstream'];
  return required.every((k) => Object.prototype.hasOwnProperty.call(obj, k));
}

function validateFetch(obj) {
  return typeof obj.RepoName === 'string' && typeof obj.Success === 'boolean' && typeof obj.Duration === 'string';
}

async function run() {
  const statusURL = `${BASE}/v1/status/sse`;
  const fetchURL = `${BASE}/v1/fetch/sse`;
  const deadline = Date.now() + TIMEOUT_MS;
  let statusOK = false;
  let fetchOK = false;

  // Status SSE
  try {
    let seen = 0;
    for await (const block of sseStream(statusURL)) {
      const ev = parseEvent(block);
      if (ev.event && ev.data) {
        const obj = JSON.parse(ev.data);
        if (validateRepo(obj)) { statusOK = true; }
        if (++seen >= MAX_EVENTS) break;
      }
      if (Date.now() > deadline) break;
    }
  } catch (e) {
    console.error('Status SSE error:', e.message);
  }

  // Fetch SSE (will require repos; still validate event shape if any)
  try {
    let seen = 0;
    for await (const block of sseStream(fetchURL)) {
      const ev = parseEvent(block);
      if (ev.event && ev.data) {
        const obj = JSON.parse(ev.data);
        if (validateFetch(obj)) { fetchOK = true; }
        if (++seen >= MAX_EVENTS) break;
      }
      if (Date.now() > deadline) break;
    }
  } catch (e) {
    console.error('Fetch SSE error:', e.message);
  }

  if (!statusOK) {
    console.error('SSE validation failed: no valid repo events observed');
    process.exit(1);
  }
  console.log('SSE validation OK:', { statusOK, fetchOK });
}

run();

