#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';

const repoRoot = process.cwd();
const DS_DEFAULT = 'http://127.0.0.1:7777';
const BAD_PORTS = ['http://127.0.0.1:7171', 'http://127.0.0.1:4319'];
const ALLOWED_OCCURRENCES = [
  // Allow cross-repo context in env docs
  path.join('docs', 'env.md'),
  // Allow dedicated policy doc to mention other ports
  path.join('docs', 'policies', 'ports-and-env.md'),
  // Allow stage scaffold script to reference other repos' defaults in examples
  path.join('scripts', 'create-stage-issue.mjs'),
  // Allow this validator to include those strings
  path.join('scripts', 'validate-conventions.mjs'),
];

const TARGET_FILES = [
  'README.md',
  // Scripts must never default to 7171/4319 in DS repo
  ...listFiles('scripts'),
];

function listFiles(dir) {
  const abs = path.join(repoRoot, dir);
  if (!fs.existsSync(abs)) return [];
  const walk = (d) => fs.readdirSync(d, { withFileTypes: true }).flatMap((ent) => {
    const p = path.join(d, ent.name);
    return ent.isDirectory() ? walk(p) : [path.relative(repoRoot, p)];
  });
  return walk(abs);
}

function fileText(p) { return fs.readFileSync(p, 'utf8'); }

let errors = [];

for (const file of TARGET_FILES) {
  if (!fs.existsSync(file)) continue;
  const txt = fileText(file);
  // Check that README includes DS ports block
  if (file === 'README.md') {
    if (!/DS_BASE_URL/.test(txt) || !/127\.0\.0\.1:7777/.test(txt)) {
      errors.push(`README.md missing DS ports/env conventions (DS_BASE_URL, 127.0.0.1:7777)`);
    }
  }
  // Disallow Bridge/MCP defaults in DS scripts/docs (except allowed files)
  if (!ALLOWED_OCCURRENCES.includes(file)) {
    for (const bad of BAD_PORTS) {
      if (txt.includes(bad)) {
        errors.push(`${file}: contains forbidden default URL ${bad} in DS repo`);
      }
    }
  }
}

if (errors.length) {
  console.error('Ports/Env conventions violations found:\n' + errors.map(e=>` - ${e}`).join('\n'));
  process.exit(1);
}
console.log('Conventions OK: DS defaults and no forbidden ports in scripts/docs.');
