import { useState, useMemo, useCallback, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  ReferenceArea,
  CartesianGrid,
} from 'recharts';
import {
  Activity,
  TestTube,
  AlertTriangle,
  Timer,
  Clock,
  Search,
  MoreVertical,
  Play,
  Trash2,
  CheckCircle,
  Shield,
} from 'lucide-react';
import { kymarosApi } from '../api/kymarosApi';
import {
  useSummary,
  useTests,
  useLatestReports,
  useDailyScores,
  useAlerts,
  useUpcoming,
  useReportsForTest,
} from '../hooks/useKymarosData';
import type {
  TestResponse,
  RestoreReport,
  DailySummary,
} from '../types/kymaros';
import MetricCard from '../components/ui/MetricCard';
import StatusBadge from '../components/ui/StatusBadge';
import SparklineChart from '../components/ui/SparklineChart';
import Skeleton from '../components/ui/Skeleton';
import ErrorState from '../components/ui/ErrorState';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function scoreColor(score: number): string {
  if (score >= 80) return 'text-emerald-400';
  if (score >= 50) return 'text-amber-400';
  return 'text-red-400';
}

function scoreStatus(result: string): 'pass' | 'fail' | 'partial' | 'not_tested' {
  const lower = result.toLowerCase();
  if (lower === 'pass' || lower === 'passed' || lower === 'success') return 'pass';
  if (lower === 'partial') return 'partial';
  if (lower === 'fail' || lower === 'failed' || lower === 'error') return 'fail';
  return 'not_tested';
}

function metricStatus(score: number): 'pass' | 'fail' | 'partial' | 'neutral' {
  if (score >= 80) return 'pass';
  if (score >= 50) return 'partial';
  return 'fail';
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${Math.round(seconds)}s`;
  const m = Math.floor(seconds / 60);
  const s = Math.round(seconds % 60);
  return s > 0 ? `${m}m ${s}s` : `${m}m`;
}

function formatTimeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  if (diff < 0) return 'just now';
  const seconds = Math.floor(diff / 1000);
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

function parseDurationToSeconds(dur: string): number {
  let total = 0;
  const hourMatch = dur.match(/(\d+)\s*h/);
  const minMatch = dur.match(/(\d+)\s*m/);
  const secMatch = dur.match(/([\d.]+)\s*s/);
  if (hourMatch) total += parseInt(hourMatch[1], 10) * 3600;
  if (minMatch) total += parseInt(minMatch[1], 10) * 60;
  if (secMatch) total += parseFloat(secMatch[1]);
  return total;
}

function backupAgeColor(ageStr: string): string {
  if (!ageStr || ageStr === '--' || ageStr === 'just now') return 'text-slate-400';
  const daysMatch = ageStr.match(/^(\d+)d ago$/);
  const hoursMatch = ageStr.match(/^(\d+)h ago$/);
  const minsMatch = ageStr.match(/^(\d+)m ago$/);
  const secsMatch = ageStr.match(/^(\d+)s ago$/);
  let hours = 0;
  if (daysMatch) hours = parseInt(daysMatch[1], 10) * 24;
  else if (hoursMatch) hours = parseInt(hoursMatch[1], 10);
  else if (minsMatch) hours = 0;
  else if (secsMatch) hours = 0;
  else return 'text-slate-400';
  if (hours < 24) return 'text-emerald-400';
  if (hours < 48) return 'text-amber-400';
  return 'text-red-400';
}

// ---------------------------------------------------------------------------
// Merged row type
// ---------------------------------------------------------------------------

interface MergedRow {
  name: string;
  namespace: string;
  score: number;
  result: string;
  rtoTarget: string;
  lastRunAt: string;
  phase: string;
  backupAge: string;
}

type FilterKey = 'all' | 'pass' | 'fail' | 'partial';

// ---------------------------------------------------------------------------
// SparklineCell — fetches report history for a single test
// ---------------------------------------------------------------------------

function SparklineCell({ testName }: { testName: string }) {
  const { data: reports } = useReportsForTest(testName);

  const scores = useMemo(() => {
    if (!reports || reports.length === 0) return [];
    return reports
      .slice()
      .sort(
        (a, b) =>
          new Date(a.metadata.creationTimestamp).getTime() -
          new Date(b.metadata.creationTimestamp).getTime(),
      )
      .map((r) => r.status?.score ?? 0);
  }, [reports]);

  if (scores.length < 2) return <span className="text-xs text-slate-500">--</span>;

  return <SparklineChart data={scores} width={100} height={30} />;
}

// ---------------------------------------------------------------------------
// Toast notification
// ---------------------------------------------------------------------------

interface ToastState {
  message: string;
  variant: 'success' | 'error';
}

function Toast({ toast, onDismiss }: { toast: ToastState; onDismiss: () => void }) {
  useEffect(() => {
    const timer = setTimeout(onDismiss, 3000);
    return () => clearTimeout(timer);
  }, [onDismiss]);

  return (
    <div className="fixed right-4 top-4 z-50 animate-in slide-in-from-top-2">
      <div
        className={`rounded-lg border px-4 py-3 text-sm font-medium shadow-lg ${
          toast.variant === 'success'
            ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-400'
            : 'border-red-500/30 bg-red-500/10 text-red-400'
        }`}
      >
        {toast.message}
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Confirm dialog
// ---------------------------------------------------------------------------

function ConfirmDialog({
  title,
  message,
  confirmLabel,
  onConfirm,
  onCancel,
}: {
  title: string;
  message: string;
  confirmLabel: string;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      <div className="w-full max-w-sm rounded-xl border border-navy-700 bg-navy-800 p-6 shadow-xl">
        <h3 className="text-lg font-semibold text-white">{title}</h3>
        <p className="mt-2 text-sm text-slate-400">{message}</p>
        <div className="mt-5 flex items-center justify-end gap-3">
          <button
            type="button"
            onClick={onCancel}
            className="rounded-lg border border-navy-600 px-4 py-2 text-sm font-medium text-slate-300 transition-colors hover:border-navy-500 hover:text-white"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onConfirm}
            className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-red-500"
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Row action menu
// ---------------------------------------------------------------------------

function RowActionMenu({
  testName,
  onTrigger,
  onDelete,
}: {
  testName: string;
  onTrigger: (name: string) => void;
  onDelete: (name: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    function handleClickOutside(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [open]);

  return (
    <div ref={menuRef} className="relative">
      <button
        type="button"
        onClick={(e) => {
          e.stopPropagation();
          setOpen((prev) => !prev);
        }}
        className="rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-navy-700 hover:text-white"
      >
        <MoreVertical className="h-4 w-4" />
      </button>
      {open && (
        <div className="absolute right-0 top-full z-20 mt-1 w-40 overflow-hidden rounded-lg border border-navy-600 bg-navy-800 shadow-xl">
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              setOpen(false);
              onTrigger(testName);
            }}
            className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-slate-300 transition-colors hover:bg-navy-700 hover:text-white"
          >
            <Play className="h-3.5 w-3.5" />
            Trigger Now
          </button>
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              setOpen(false);
              onDelete(testName);
            }}
            className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-red-400 transition-colors hover:bg-navy-700 hover:text-red-300"
          >
            <Trash2 className="h-3.5 w-3.5" />
            Delete
          </button>
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Dashboard Page
// ---------------------------------------------------------------------------

function Dashboard() {
  const navigate = useNavigate();
  const summary = useSummary();
  const tests = useTests();
  const latestReports = useLatestReports();
  const dailyScores = useDailyScores(30);
  const alerts = useAlerts(48);
  const upcoming = useUpcoming();

  const [filter, setFilter] = useState<FilterKey>('all');
  const [search, setSearch] = useState('');
  const [toast, setToast] = useState<ToastState | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const dismissToast = useCallback(() => setToast(null), []);

  const handleTrigger = useCallback(async (name: string) => {
    try {
      const result = await kymarosApi.triggerTest(name);
      setToast({ message: result.message || `Triggered ${name}`, variant: 'success' });
    } catch (err) {
      setToast({
        message: err instanceof Error ? err.message : `Failed to trigger ${name}`,
        variant: 'error',
      });
    }
  }, []);

  const handleDeleteConfirm = useCallback(async () => {
    if (!deleteTarget) return;
    try {
      await kymarosApi.deleteTest(deleteTarget);
      setToast({ message: `Deleted ${deleteTarget}`, variant: 'success' });
      setDeleteTarget(null);
      tests.refetch();
    } catch (err) {
      setToast({
        message: err instanceof Error ? err.message : `Failed to delete ${deleteTarget}`,
        variant: 'error',
      });
      setDeleteTarget(null);
    }
  }, [deleteTarget, tests]);

  // Merge tests + latest reports
  const mergedRows = useMemo<MergedRow[]>(() => {
    if (!tests.data) return [];
    const reportMap = new Map<string, RestoreReport>();
    if (latestReports.data) {
      for (const r of latestReports.data) {
        reportMap.set(r.spec.testRef, r);
      }
    }
    return tests.data.map((t: TestResponse) => {
      const report = reportMap.get(t.name);
      return {
        name: t.name,
        namespace: t.sourceNamespaces.join(', '),
        score: report?.status?.score ?? t.lastScore,
        result: report?.status?.result ?? t.lastResult,
        rtoTarget: t.rtoTarget,
        lastRunAt: report?.status?.completedAt ?? t.lastRunAt,
        phase: t.phase,
        backupAge: report ? formatTimeAgo(report.metadata.creationTimestamp) : '--',
      };
    });
  }, [tests.data, latestReports.data]);

  // Filter counts
  const filterCounts = useMemo(() => {
    const counts = { all: 0, pass: 0, fail: 0, partial: 0 };
    for (const row of mergedRows) {
      counts.all++;
      const s = scoreStatus(row.result);
      if (s === 'pass') counts.pass++;
      else if (s === 'fail') counts.fail++;
      else if (s === 'partial') counts.partial++;
    }
    return counts;
  }, [mergedRows]);

  // Filtered + searched + sorted rows
  const filteredRows = useMemo(() => {
    let rows = mergedRows;

    if (filter !== 'all') {
      rows = rows.filter((r) => scoreStatus(r.result) === filter);
    }

    if (search.trim()) {
      const q = search.toLowerCase();
      rows = rows.filter(
        (r) =>
          r.name.toLowerCase().includes(q) ||
          r.namespace.toLowerCase().includes(q),
      );
    }

    return rows.slice().sort((a, b) => a.score - b.score);
  }, [mergedRows, filter, search]);

  // Chart data
  const chartData = useMemo<DailySummary[]>(() => {
    if (!dailyScores.data) return [];
    return dailyScores.data
      .slice()
      .sort((a, b) => a.date.localeCompare(b.date));
  }, [dailyScores.data]);

  // Score trend vs yesterday (positive = improved, negative = declined)
  const scoreTrendVsYesterday = useMemo(() => {
    if (!dailyScores.data || dailyScores.data.length < 2) return 0;
    const sorted = dailyScores.data.slice().sort((a, b) => b.date.localeCompare(a.date));
    return Math.round((sorted[0]?.score ?? 0) - (sorted[1]?.score ?? 0));
  }, [dailyScores.data]);

  // Score min/max for chart reference areas
  const chartYDomain: [number, number] = [0, 100];

  // ---------------------------------------------------------------------------
  // Loading state
  // ---------------------------------------------------------------------------

  if (summary.loading && tests.loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <div className="grid grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-28 w-full rounded-xl" />
          ))}
        </div>
        <Skeleton className="h-64 w-full rounded-xl" />
      </div>
    );
  }

  // ---------------------------------------------------------------------------
  // Error state
  // ---------------------------------------------------------------------------

  if (summary.error && tests.error) {
    return (
      <ErrorState
        message={summary.error ?? tests.error ?? 'Failed to load dashboard data'}
        onRetry={() => {
          summary.refetch();
          tests.refetch();
        }}
      />
    );
  }

  // ---------------------------------------------------------------------------
  // Render
  // ---------------------------------------------------------------------------

  const s = summary.data;

  return (
    <div className="space-y-6">
      {/* Toast notification */}
      {toast && <Toast toast={toast} onDismiss={dismissToast} />}

      {/* Delete confirmation dialog */}
      {deleteTarget && (
        <ConfirmDialog
          title="Delete Test"
          message={`Are you sure you want to delete "${deleteTarget}"? This action cannot be undone.`}
          confirmLabel="Delete"
          onConfirm={handleDeleteConfirm}
          onCancel={() => setDeleteTarget(null)}
        />
      )}

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Dashboard</h1>
          <p className="text-sm text-slate-400">
            {s ? `${s.namespacesCovered} namespaces monitored` : 'Loading...'}
          </p>
        </div>
        <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-500/10 px-3 py-1 text-xs font-medium text-emerald-400">
          <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" />
          Live
        </span>
      </div>

      {/* Metric cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          label="Average Score"
          value={s ? Math.round(s.averageScore) : '--'}
          status={s ? metricStatus(s.averageScore) : 'neutral'}
          trend={scoreTrendVsYesterday !== 0 ? scoreTrendVsYesterday : undefined}
          subtitle={
            scoreTrendVsYesterday !== 0
              ? `${scoreTrendVsYesterday > 0 ? '↑ +' : '↓ '}${scoreTrendVsYesterday} vs yesterday`
              : undefined
          }
          icon={<Activity className="h-5 w-5" />}
        />
        <MetricCard
          label="Tests Last Night"
          value={s?.testsLastNight ?? '--'}
          subtitle={s ? `${filterCounts.pass} passed, ${filterCounts.fail} failed` : undefined}
          status={
            s
              ? filterCounts.fail > 0
                ? 'fail'
                : filterCounts.partial > 0
                  ? 'partial'
                  : 'pass'
              : 'neutral'
          }
          icon={<TestTube className="h-5 w-5" />}
        />
        <MetricCard
          label="Issues Found"
          value={s ? s.totalFailures + s.totalPartial : '--'}
          subtitle={
            s
              ? s.totalFailures + s.totalPartial === 0
                ? 'All clear — no issues detected'
                : `${s.totalFailures} failed, ${s.totalPartial} partial`
              : undefined
          }
          status={
            s
              ? s.totalFailures + s.totalPartial === 0
                ? 'pass'
                : s.totalFailures > 0
                  ? 'fail'
                  : 'partial'
              : 'neutral'
          }
          icon={<AlertTriangle className="h-5 w-5" />}
        />
        <MetricCard
          label="Avg RTO"
          value={
            s
              ? formatDuration(parseDurationToSeconds(s.averageRTO))
              : '--'
          }
          subtitle={(() => {
            if (!s || !tests.data || tests.data.length === 0) return undefined;
            const firstTarget = tests.data[0]?.rtoTarget;
            if (!firstTarget) return undefined;
            const targetSec = parseDurationToSeconds(firstTarget);
            const avgSec = parseDurationToSeconds(s.averageRTO);
            const within = avgSec <= targetSec;
            return `Target: ${firstTarget} ${within ? '✓' : '✗'}`;
          })()}
          status={(() => {
            if (!s || !tests.data || tests.data.length === 0) return 'neutral';
            const firstTarget = tests.data[0]?.rtoTarget;
            if (!firstTarget) return 'neutral';
            const targetSec = parseDurationToSeconds(firstTarget);
            const avgSec = parseDurationToSeconds(s.averageRTO);
            return avgSec <= targetSec ? 'pass' : 'fail';
          })()}
          icon={<Timer className="h-5 w-5" />}
        />
      </div>

      {/* Tests table */}
      <div className="rounded-xl border border-navy-700 bg-navy-800 p-5">
        <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <h2 className="text-lg font-semibold text-white">Restore Tests</h2>
          <div className="flex items-center gap-3">
            {/* Filter buttons */}
            <div className="flex rounded-lg border border-navy-600 p-0.5">
              {(
                [
                  ['all', 'All'],
                  ['pass', 'Pass'],
                  ['fail', 'Fail'],
                  ['partial', 'Partial'],
                ] as const
              ).map(([key, label]) => (
                <button
                  key={key}
                  type="button"
                  onClick={() => setFilter(key)}
                  className={`rounded-md px-3 py-1 text-xs font-medium transition-colors ${
                    filter === key
                      ? 'bg-navy-600 text-white'
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  {label} ({filterCounts[key]})
                </button>
              ))}
            </div>
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-500" />
              <input
                type="text"
                placeholder="Search namespace..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="rounded-lg border border-navy-600 bg-navy-900 py-1.5 pl-8 pr-3 text-xs text-white placeholder:text-slate-500 focus:border-blue-500 focus:outline-none"
              />
            </div>
          </div>
        </div>

        {mergedRows.length === 0 ? (
          <div className="flex flex-col items-center gap-3 py-12 text-center">
            <Shield className="h-10 w-10 text-slate-600" />
            <p className="text-sm font-medium text-slate-300">No restore tests configured yet</p>
            <p className="text-xs text-slate-500">
              Create a RestoreTest CRD with <code className="text-slate-400">kubectl apply</code> to start validating your backups.
            </p>
          </div>
        ) : filteredRows.length === 0 ? (
          <p className="py-8 text-center text-sm text-slate-500">
            No tests match your filters.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-navy-700 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                  <th className="pb-3 pr-4">Namespace</th>
                  <th className="pb-3 pr-4">Score</th>
                  <th className="pb-3 pr-4">Status</th>
                  <th className="pb-3 pr-4">RTO</th>
                  <th className="pb-3 pr-4">Backup Age</th>
                  <th className="pb-3 pr-4">Last Run</th>
                  <th className="pb-3 pr-4">Trend</th>
                  <th className="pb-3 w-10"><span className="sr-only">Actions</span></th>
                </tr>
              </thead>
              <tbody>
                {filteredRows.map((row) => (
                  <tr
                    key={row.name}
                    onClick={() => navigate(`/reports/${encodeURIComponent(row.name)}`)}
                    className="cursor-pointer border-b border-navy-700/50 transition-colors hover:bg-navy-700/30"
                  >
                    <td className="py-3 pr-4 font-medium text-white">
                      {row.namespace || row.name}
                    </td>
                    <td className={`py-3 pr-4 font-mono font-bold ${scoreColor(row.score)}`}>
                      {row.score}
                    </td>
                    <td className="py-3 pr-4">
                      <StatusBadge status={scoreStatus(row.result)} size="sm" />
                    </td>
                    <td className="py-3 pr-4 font-mono text-slate-300">
                      {row.rtoTarget}
                    </td>
                    <td className={`py-3 pr-4 ${backupAgeColor(row.backupAge)}`}>{row.backupAge}</td>
                    <td className="py-3 pr-4 text-slate-400">
                      {row.lastRunAt ? formatTimeAgo(row.lastRunAt) : '--'}
                    </td>
                    <td className="py-3 pr-4">
                      <SparklineCell testName={row.name} />
                    </td>
                    <td className="py-3">
                      <RowActionMenu
                        testName={row.name}
                        onTrigger={handleTrigger}
                        onDelete={setDeleteTarget}
                      />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Score trend chart */}
      {dailyScores.loading ? (
        <Skeleton className="h-72 w-full rounded-xl" />
      ) : dailyScores.error ? (
        <ErrorState message={dailyScores.error} onRetry={dailyScores.refetch} />
      ) : chartData.length > 0 ? (
        <div className="rounded-xl border border-navy-700 bg-navy-800 p-5">
          <h2 className="mb-4 text-lg font-semibold text-white">
            Score Trend (30 days)
          </h2>
          <ResponsiveContainer width="100%" height={260}>
            <AreaChart
              data={chartData}
              margin={{ top: 5, right: 5, bottom: 5, left: 5 }}
            >
              <defs>
                <linearGradient id="scoreFill" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#3b82f6" stopOpacity={0.3} />
                  <stop offset="100%" stopColor="#3b82f6" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis
                dataKey="date"
                tick={{ fill: '#94a3b8', fontSize: 11 }}
                tickFormatter={(d: string) => d.slice(5)}
                stroke="#334155"
              />
              <YAxis
                domain={chartYDomain}
                tick={{ fill: '#94a3b8', fontSize: 11 }}
                stroke="#334155"
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: '#1e293b',
                  border: '1px solid #334155',
                  borderRadius: 8,
                  color: '#e2e8f0',
                  fontSize: 12,
                }}
                formatter={(value) => [Math.round(Number(value)), 'score']}
              />
              {/* Reference areas for score zones */}
              <ReferenceArea y1={80} y2={100} fill="#10b981" fillOpacity={0.06} />
              <ReferenceArea y1={50} y2={80} fill="#f59e0b" fillOpacity={0.06} />
              <ReferenceArea y1={0} y2={50} fill="#ef4444" fillOpacity={0.06} />
              <Area
                type="monotone"
                dataKey="score"
                stroke="#3b82f6"
                strokeWidth={2}
                fill="url(#scoreFill)"
                isAnimationActive={false}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      ) : null}

      {/* Two-column: Alerts + Upcoming */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {/* Alerts */}
        <div className="rounded-xl border border-navy-700 bg-navy-800 p-5">
          <h2 className="mb-3 text-lg font-semibold text-white">
            Recent Alerts
          </h2>
          {alerts.loading ? (
            <div className="space-y-2">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full rounded-lg" />
              ))}
            </div>
          ) : alerts.error ? (
            <ErrorState message={alerts.error} onRetry={alerts.refetch} />
          ) : !alerts.data || alerts.data.length === 0 ? (
            <div className="flex flex-col items-center gap-2 py-6 text-center">
              <CheckCircle className="h-5 w-5 text-emerald-400" />
              <p className="text-sm text-slate-500">
                All clear — no alerts in the last 48 hours.
              </p>
            </div>
          ) : (
            <div className="space-y-2 max-h-72 overflow-y-auto">
              {alerts.data.slice(0, 10).map((alert, idx) => (
                <div
                  key={`${alert.testName}-${alert.timestamp}-${idx}`}
                  className="flex items-start gap-3 rounded-lg border border-navy-700/50 bg-navy-900/50 p-3"
                >
                  <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-amber-400" />
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium text-white">
                      {alert.testName}
                    </p>
                    <p className="text-xs text-slate-400">{alert.message}</p>
                    <p className="mt-1 text-xs text-slate-500">
                      {formatTimeAgo(alert.timestamp)}
                    </p>
                  </div>
                  <span
                    className={`font-mono text-sm font-bold ${scoreColor(alert.score)}`}
                  >
                    {alert.score}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Upcoming */}
        <div className="rounded-xl border border-navy-700 bg-navy-800 p-5">
          <h2 className="mb-3 text-lg font-semibold text-white">
            Upcoming Tests
          </h2>
          {upcoming.loading ? (
            <div className="space-y-2">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full rounded-lg" />
              ))}
            </div>
          ) : upcoming.error ? (
            <ErrorState message={upcoming.error} onRetry={upcoming.refetch} />
          ) : !upcoming.data || upcoming.data.length === 0 ? (
            <p className="py-6 text-center text-sm text-slate-500">
              No tests scheduled. Create a RestoreTest to get started.
            </p>
          ) : (
            <div className="space-y-2 max-h-72 overflow-y-auto">
              {upcoming.data.map((test) => (
                <div
                  key={test.name}
                  className="flex items-center justify-between rounded-lg border border-navy-700/50 bg-navy-900/50 p-3"
                >
                  <div>
                    <p className="text-sm font-medium text-white">
                      {test.name}
                    </p>
                    <p className="text-xs text-slate-400">{test.namespace}</p>
                  </div>
                  <div className="text-right">
                    <p className="flex items-center gap-1 text-xs text-slate-400">
                      <Clock className="h-3 w-3" />
                      {formatTimeAgo(test.nextRunAt)}
                    </p>
                    <p
                      className={`mt-0.5 font-mono text-sm font-bold ${scoreColor(test.lastScore)}`}
                    >
                      {test.lastScore}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default Dashboard;
