import type {
  SummaryResponse,
  DailySummary,
  TestResponse,
  RestoreReport,
  Alert,
  UpcomingTest,
  CreateTestInput,
  ReportLogsResponse,
  HealthResponse,
  ProviderConfig,
  NotificationConfig,
  SandboxConfig,
} from '../types/kymaros';

const API_BASE = '/api/v1';

export function getAuthToken(): string | null {
  return localStorage.getItem('kymaros_api_token');
}

export function setAuthToken(token: string | null): void {
  if (token) {
    localStorage.setItem('kymaros_api_token', token);
  } else {
    localStorage.removeItem('kymaros_api_token');
  }
}

function authHeaders(): Record<string, string> {
  const token = getAuthToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

class ApiError extends Error {
  readonly status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
    this.name = 'ApiError';
  }
}

function qs(params?: Record<string, string | number | boolean | undefined>): string {
  if (!params) return '';
  const entries = Object.entries(params).filter(
    (entry): entry is [string, string | number | boolean] => entry[1] !== undefined,
  );
  return new URLSearchParams(
    entries.map(([k, v]) => [k, String(v)]),
  ).toString();
}

async function fetchJSON<T>(url: string): Promise<T> {
  const response = await fetch(url, { headers: authHeaders() });
  if (!response.ok) {
    throw new ApiError(response.status, `API request failed: ${response.statusText}`);
  }
  return response.json() as Promise<T>;
}

async function fetchMutate<T>(method: string, url: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { ...authHeaders() };
  if (body) headers['Content-Type'] = 'application/json';
  const res = await fetch(url, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    let detail = `HTTP ${res.status}`;
    try {
      const body = await res.json();
      detail = body.message ?? body.error ?? JSON.stringify(body);
    } catch {
      try { const t = await res.text(); if (t) detail = t; } catch { /* ignore */ }
    }
    throw new ApiError(res.status, detail);
  }
  if (res.status === 204) return {} as T;
  return res.json();
}

export const kymarosApi = {
  getSummary: (): Promise<SummaryResponse> =>
    fetchJSON<SummaryResponse>(`${API_BASE}/summary`),

  getDailyScores: (days = 7): Promise<DailySummary[]> =>
    fetchJSON<DailySummary[]>(`${API_BASE}/summary/daily?days=${days}`),

  getTests: (): Promise<TestResponse[]> =>
    fetchJSON<TestResponse[]>(`${API_BASE}/tests`),

  getReports: (params?: Record<string, string | number | boolean | undefined>): Promise<RestoreReport[]> =>
    fetchJSON<RestoreReport[]>(`${API_BASE}/reports?${qs(params)}`),

  getLatestReports: (): Promise<RestoreReport[]> =>
    fetchJSON<RestoreReport[]>(`${API_BASE}/reports?latest=true`),

  getReportsForTest: (name: string, days = 30): Promise<RestoreReport[]> =>
    fetchJSON<RestoreReport[]>(`${API_BASE}/reports?test=${encodeURIComponent(name)}&days=${days}`),

  getReportLogs: (name: string): Promise<ReportLogsResponse> =>
    fetchJSON<ReportLogsResponse>(`${API_BASE}/reports/${encodeURIComponent(name)}/logs`),

  getAlerts: (hours = 48): Promise<Alert[]> =>
    fetchJSON<Alert[]>(`${API_BASE}/alerts?hours=${hours}`),

  getUpcoming: (): Promise<UpcomingTest[]> =>
    fetchJSON<UpcomingTest[]>(`${API_BASE}/upcoming`),

  getHealth: (): Promise<HealthResponse> =>
    fetchJSON<HealthResponse>(`${API_BASE}/health`),

  getProviderConfig: (): Promise<ProviderConfig> =>
    fetchJSON<ProviderConfig>(`${API_BASE}/config/provider`),

  getNotificationConfig: (): Promise<NotificationConfig> =>
    fetchJSON<NotificationConfig>(`${API_BASE}/config/notifications`),

  testNotification: (channel: { type: string; secretRef?: string }): Promise<{ message: string }> =>
    fetchMutate<{ message: string }>('POST', `${API_BASE}/config/notifications/test`, channel),

  getSandboxConfig: (): Promise<SandboxConfig> =>
    fetchJSON<SandboxConfig>(`${API_BASE}/config/sandbox`),

  cleanupOrphanSandboxes: (): Promise<{ deleted: number }> =>
    fetchMutate<{ deleted: number }>('POST', `${API_BASE}/config/sandbox/cleanup`),

  createTest: (input: CreateTestInput): Promise<TestResponse> =>
    fetchMutate<TestResponse>('POST', `${API_BASE}/tests`, input),

  updateTest: (name: string, input: CreateTestInput): Promise<TestResponse> =>
    fetchMutate<TestResponse>('PUT', `${API_BASE}/tests/${encodeURIComponent(name)}`, input),

  deleteTest: (name: string): Promise<void> =>
    fetchMutate<void>('DELETE', `${API_BASE}/tests/${encodeURIComponent(name)}`),

  triggerTest: (name: string): Promise<{ message: string }> =>
    fetchMutate<{ message: string }>('POST', `${API_BASE}/tests/${encodeURIComponent(name)}/trigger`),
};

