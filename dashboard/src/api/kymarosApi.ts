import type {
  SummaryResponse,
  DailySummary,
  TestResponse,
  RestoreReport,
  Alert,
  ComplianceResponse,
  UpcomingTest,
  CreateTestInput,
  ReportLogsResponse,
} from '../types/kymaros';
import type { LicenseResponse } from '../types/license';

const API_BASE = '/api/v1';

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
  const response = await fetch(url);
  if (!response.ok) {
    throw new ApiError(response.status, `API request failed: ${response.statusText}`);
  }
  return response.json() as Promise<T>;
}

async function fetchMutate<T>(method: string, url: string, body?: unknown): Promise<T> {
  const res = await fetch(url, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) throw new ApiError(res.status, await res.text());
  if (res.status === 204) return {} as T;
  return res.json();
}

export const kymarosApi = {
  getSummary: (): Promise<SummaryResponse> =>
    fetchJSON<SummaryResponse>(`${API_BASE}/summary`),

  getDailyScores: (days = 30): Promise<DailySummary[]> =>
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

  getCompliance: (framework: string, period: string): Promise<ComplianceResponse> =>
    fetchJSON<ComplianceResponse>(
      `${API_BASE}/compliance?framework=${encodeURIComponent(framework)}&period=${encodeURIComponent(period)}`,
    ),

  getUpcoming: (): Promise<UpcomingTest[]> =>
    fetchJSON<UpcomingTest[]>(`${API_BASE}/upcoming`),

  getHealth: (): Promise<{ status: string }> =>
    fetchJSON<{ status: string }>(`${API_BASE}/health`),

  createTest: (input: CreateTestInput): Promise<TestResponse> =>
    fetchMutate<TestResponse>('POST', `${API_BASE}/tests`, input),

  updateTest: (name: string, input: CreateTestInput): Promise<TestResponse> =>
    fetchMutate<TestResponse>('PUT', `${API_BASE}/tests/${encodeURIComponent(name)}`, input),

  deleteTest: (name: string): Promise<void> =>
    fetchMutate<void>('DELETE', `${API_BASE}/tests/${encodeURIComponent(name)}`),

  triggerTest: (name: string): Promise<{ message: string }> =>
    fetchMutate<{ message: string }>('POST', `${API_BASE}/tests/${encodeURIComponent(name)}/trigger`),
};

export async function exportCompliancePDF(framework: string, period: string): Promise<Blob> {
  const res = await fetch(`${API_BASE}/compliance/pdf?framework=${encodeURIComponent(framework)}&period=${encodeURIComponent(period)}`);
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || `PDF export failed (${res.status})`);
  }
  return res.blob();
}

export async function getLicense(): Promise<LicenseResponse> {
  const res = await fetch(`${API_BASE}/license`);
  if (!res.ok) {
    // Fallback to community if endpoint unavailable
    return {
      tier: 'community',
      isExpired: false,
      isTrialing: false,
      features: {
        maxHistoryDays: 7,
        compliancePage: false,
        pdfExport: false,
        csvExport: false,
        multiBackup: false,
        multiCluster: false,
        sso: false,
        rtoAnalytics: false,
        regressionAlerts: false,
        timeline: false,
        scoreBreakdown: false,
      },
    };
  }
  return res.json() as Promise<LicenseResponse>;
}
