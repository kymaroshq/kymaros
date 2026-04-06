import { useState, useEffect, useCallback, useRef } from 'react';
import { kymarosApi } from '../api/kymarosApi';
import type {
  SummaryResponse,
  DailySummary,
  TestResponse,
  RestoreReport,
  Alert,
  UpcomingTest,
  ReportLogsResponse,
} from '../types/kymaros';

interface ApiState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  lastFetchedAt: number | null;
  isStale: boolean;
  refetch: () => void;
}

const STALE_THRESHOLD_MS = 120_000; // 2 minutes

export function useApiData<T>(
  fetcher: () => Promise<T>,
  intervalMs = 60_000,
  deps: ReadonlyArray<unknown> = [],
): ApiState<T> {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastFetchedAt, setLastFetchedAt] = useState<number | null>(null);
  const [isStale, setIsStale] = useState(false);
  const mountedRef = useRef(true);
  const fetcherRef = useRef(fetcher);

  // Keep fetcher ref current without triggering re-renders
  fetcherRef.current = fetcher;

  const doFetch = useCallback(async () => {
    try {
      setLoading((prev) => prev || data === null);
      setError(null);
      const result = await fetcherRef.current();
      if (mountedRef.current) {
        setData(result);
        setLastFetchedAt(Date.now());
        setIsStale(false);
      }
    } catch (err) {
      if (mountedRef.current) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      }
    } finally {
      if (mountedRef.current) {
        setLoading(false);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);

  // Initial fetch and optional polling interval
  useEffect(() => {
    mountedRef.current = true;
    void doFetch();

    let intervalId: ReturnType<typeof setInterval> | undefined;
    if (intervalMs > 0) {
      intervalId = setInterval(() => {
        void doFetch();
      }, intervalMs);
    }

    return () => {
      mountedRef.current = false;
      if (intervalId) clearInterval(intervalId);
    };
  }, [doFetch, intervalMs]);

  // Staleness check
  useEffect(() => {
    if (lastFetchedAt === null) return;

    const checkStale = () => {
      const age = Date.now() - (lastFetchedAt ?? 0);
      setIsStale(age > STALE_THRESHOLD_MS);
    };

    const staleCheckId = setInterval(checkStale, 10_000);
    return () => clearInterval(staleCheckId);
  }, [lastFetchedAt]);

  return { data, loading, error, lastFetchedAt, isStale, refetch: doFetch };
}

// Specific hooks

export function useSummary(): ApiState<SummaryResponse> {
  return useApiData(() => kymarosApi.getSummary());
}

export function useDailyScores(days = 30): ApiState<DailySummary[]> {
  return useApiData(() => kymarosApi.getDailyScores(days), 300_000, [days]);
}

export function useTests(): ApiState<TestResponse[]> {
  return useApiData(() => kymarosApi.getTests());
}

export function useLatestReports(): ApiState<RestoreReport[]> {
  return useApiData(() => kymarosApi.getLatestReports());
}

export function useReportsForTest(name: string): ApiState<RestoreReport[]> {
  return useApiData(() => kymarosApi.getReportsForTest(name), 60_000, [name]);
}

export function useAlerts(hours = 48): ApiState<Alert[]> {
  return useApiData(() => kymarosApi.getAlerts(hours), 30_000, [hours]);
}

export function useUpcoming(): ApiState<UpcomingTest[]> {
  return useApiData(() => kymarosApi.getUpcoming(), 60_000);
}

export function useReportLogs(reportName: string): ApiState<ReportLogsResponse> {
  return useApiData(() => kymarosApi.getReportLogs(reportName), 0, [reportName]);
}

