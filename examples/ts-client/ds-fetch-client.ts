// Minimal fetch-based client for DS API (schema_version: "ds.v1")

export interface HealthResponse {
  ok: boolean;
  version: number;
  uptime_sec: number;
  workers: number;
  auth: boolean;
  timestamp: string; // RFC3339
  schema_version: string; // "ds.v1"
}

export interface Repository {
  Path: string;
  Name: string;
  Account: string;
  FolderName: string;
  IsOrg: boolean;
  RemoteURL: string;
  Branch: string;
  IsClean: boolean;
  Uncommitted: number;
  Ahead: number;
  Behind: number;
  LastCommit: string;
  LastFetch?: string | null;
  HasStash: boolean;
  HasUpstream: boolean;
  scan_time: string;
}

export interface StatusResponse {
  schema_version: string; // "ds.v1"
  data: Repository[];
}

export interface ScanResponse {
  schema_version: string;
  count: number;
}

export interface MovePlan {
  name: string;
  account: string;
  is_org: boolean;
  old_path: string;
  new_path: string;
}

export interface OrganizePlanResponse {
  schema_version: string;
  data: MovePlan[];
}

export interface OrganizeApplyResponse {
  schema_version: string;
  moved: number;
  failed: number;
  results: MovePlan[];
}

export interface FetchResult { RepoName: string; Success: boolean; Error?: string | null; Duration: string; }
export interface FetchResponse { schema_version: string; results: FetchResult[] }

export interface PolicyCheckResult { name: string; description: string; severity: string; passed: boolean; error?: string; duration_ms: number; }
export interface PolicySummary { total: number; passed: number; failed: number; warnings: number; }
export interface PolicyReport { results: PolicyCheckResult[]; summary: PolicySummary; }
export interface PolicyResponse { schema_version: string; report: PolicyReport; failed_threshold: boolean }

export interface ExecResult { repo: string; path: string; success: boolean; error?: string; duration_ms: number; }
export interface ExecResponse { schema_version: string; results: ExecResult[] }

export class DsFetchClient {
  constructor(private base = 'http://127.0.0.1:7777', private token?: string) {}

  private headers(extra?: Record<string, string>) {
    const h: Record<string, string> = { Accept: 'application/json' };
    if (this.token) h['Authorization'] = `Bearer ${this.token}`;
    return { ...h, ...(extra || {}) };
  }

  async health(): Promise<HealthResponse> {
    const r = await fetch(`${this.base}/v1/health`, { headers: this.headers() });
    return r.json();
  }

  async status(params?: { account?: string; dirty?: boolean; path?: string }): Promise<StatusResponse> {
    const q = new URLSearchParams();
    if (params?.account) q.set('account', params.account);
    if (params?.dirty) q.set('dirty', 'true');
    if (params?.path) q.set('path', params.path);
    const r = await fetch(`${this.base}/v1/status?${q.toString()}`, { headers: this.headers() });
    return r.json();
  }

  async scan(params?: { path?: string }): Promise<ScanResponse> {
    const q = new URLSearchParams();
    if (params?.path) q.set('path', params.path);
    const r = await fetch(`${this.base}/v1/scan?${q.toString()}`, { headers: this.headers() });
    return r.json();
  }

  async organizePlan(requireClean = false, path?: string): Promise<OrganizePlanResponse> {
    const q = new URLSearchParams();
    if (requireClean) q.set('require_clean', 'true');
    if (path) q.set('path', path);
    const r = await fetch(`${this.base}/v1/organize/plan?${q.toString()}`, { headers: this.headers() });
    return r.json();
  }

  async organizeApply(opts: { requireClean?: boolean; force?: boolean; dryRun?: boolean; path?: string }): Promise<OrganizeApplyResponse> {
    const q = new URLSearchParams();
    if (opts.requireClean) q.set('require_clean', 'true');
    if (opts.force) q.set('force', 'true');
    if (opts.dryRun) q.set('dry_run', 'true');
    if (opts.path) q.set('path', opts.path);
    const r = await fetch(`${this.base}/v1/organize/apply?${q.toString()}`, { method: 'POST', headers: this.headers() });
    return r.json();
  }

  async fetchRepos(params?: { account?: string; dirty?: boolean; path?: string }): Promise<FetchResponse> {
    const q = new URLSearchParams();
    if (params?.account) q.set('account', params.account);
    if (params?.dirty) q.set('dirty', 'true');
    if (params?.path) q.set('path', params.path);
    const r = await fetch(`${this.base}/v1/fetch?${q.toString()}`, { headers: this.headers() });
    return r.json();
  }

  async policyCheck(file = '.project-compliance.yaml', failOn: 'critical'|'high'|'medium'|'low' = 'critical'): Promise<PolicyResponse> {
    const q = new URLSearchParams({ file, fail_on: failOn });
    const r = await fetch(`${this.base}/v1/policy/check?${q.toString()}`, { headers: this.headers() });
    return r.json();
  }

  async exec(cmd: string, opts?: { account?: string; dirty?: boolean; timeout?: number; path?: string }): Promise<ExecResponse> {
    const q = new URLSearchParams();
    if (opts?.account) q.set('account', opts.account);
    if (opts?.dirty) q.set('dirty', 'true');
    if (opts?.timeout) q.set('timeout', String(opts.timeout));
    if (opts?.path) q.set('path', opts.path);
    const r = await fetch(`${this.base}/v1/exec?${q.toString()}`, {
      method: 'POST',
      headers: this.headers({ 'Content-Type': 'application/json' }),
      body: JSON.stringify({ cmd })
    });
    return r.json();
  }
}

