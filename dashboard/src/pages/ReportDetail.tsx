import { useMemo, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts';
import {
  ArrowLeft,
  Activity,
  Timer,
  HardDrive,
  Clock,
  CheckCircle,
  XCircle,
  Box,
  Network,
  KeyRound,
  FileText,
  Database,
  ChevronDown,
  AlertTriangle,
  Terminal,
} from 'lucide-react';
import { useReportsForTest, useTests, useReportLogs } from '../hooks/useKymarosData';
import type { RestoreReport, PodLog, EventLog } from '../types/kymaros';
import MetricCard from '../components/ui/MetricCard';
import StatusBadge from '../components/ui/StatusBadge';
import ValidationLevelBar from '../components/ui/ValidationLevelBar';
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

function scoreStatus(result: string | undefined): 'pass' | 'fail' | 'partial' | 'not_tested' {
  if (!result) return 'not_tested';
  const lower = result.toLowerCase();
  if (lower === 'pass' || lower === 'passed' || lower === 'success') return 'pass';
  if (lower === 'partial') return 'partial';
  if (lower === 'fail' || lower === 'failed' || lower === 'error') return 'fail';
  return 'not_tested';
}

function levelStatus(status: string | undefined): 'pass' | 'fail' {
  return status === 'pass' ? 'pass' : 'fail';
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

function parseDurationToSeconds(dur: string | undefined): number {
  if (!dur) return 0;
  let total = 0;
  const hourMatch = dur.match(/(\d+)\s*h/);
  const minMatch = dur.match(/(\d+)\s*m/);
  const secMatch = dur.match(/([\d.]+)\s*s/);
  if (hourMatch) total += parseInt(hourMatch[1], 10) * 3600;
  if (minMatch) total += parseInt(minMatch[1], 10) * 60;
  if (secMatch) total += parseFloat(secMatch[1]);
  return total;
}

function formatCheckDuration(dur: string | undefined): string {
  if (!dur) return '--';
  if (dur === '0s' || dur === '0ms') return '<1s';
  return dur;
}

function metricStatus(score: number): 'pass' | 'fail' | 'partial' | 'neutral' {
  if (score >= 80) return 'pass';
  if (score >= 50) return 'partial';
  return 'fail';
}

// ---------------------------------------------------------------------------
// Completeness item type
// ---------------------------------------------------------------------------

interface CompletenessItem {
  label: string;
  icon: typeof Box;
  present: boolean;
}

// ---------------------------------------------------------------------------
// Timeline event type
// ---------------------------------------------------------------------------

interface TimelineEvent {
  time: string;
  label: string;
  status: 'done' | 'active' | 'pending';
}

// ---------------------------------------------------------------------------
// Pod Logs components
// ---------------------------------------------------------------------------

function PodLogCard({ pod }: { pod: PodLog }) {
  const [expanded, setExpanded] = useState(false);
  const hasErrors = pod.phase === 'Failed' || pod.phase === 'CrashLoopBackOff';

  return (
    <div
      className={`mb-2 rounded-lg border ${
        hasErrors
          ? 'border-red-500/30 bg-red-500/5'
          : 'border-navy-700 bg-navy-800/30'
      }`}
    >
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center justify-between p-3 text-left"
      >
        <div className="flex items-center gap-2">
          <span
            className={`h-2 w-2 rounded-full ${
              hasErrors
                ? 'bg-red-400'
                : pod.phase === 'Running' || pod.phase === 'Succeeded'
                  ? 'bg-emerald-400'
                  : 'bg-amber-400'
            }`}
          />
          <span className="font-mono text-sm text-slate-300">{pod.podName}</span>
          <span className="text-xs text-slate-500">{pod.phase}</span>
        </div>
        <ChevronDown
          className={`h-4 w-4 text-slate-500 transition-transform ${
            expanded ? 'rotate-180' : ''
          }`}
        />
      </button>

      {expanded &&
        pod.containers.map((container) => (
          <div
            key={`${container.name}-${container.type}`}
            className="border-t border-navy-700/50 p-3"
          >
            <div className="mb-2 flex items-center gap-2">
              <span className="text-xs text-slate-400">{container.name}</span>
              {container.type !== 'container' && (
                <span className="rounded bg-navy-700 px-1.5 py-0.5 text-[10px] text-slate-400">
                  {container.type}
                </span>
              )}
              {container.truncated && (
                <span className="text-[10px] text-slate-500">
                  (last {container.totalLines} lines)
                </span>
              )}
            </div>
            <pre className="max-h-64 overflow-x-auto overflow-y-auto rounded bg-navy-900 p-3 font-mono text-xs leading-relaxed text-slate-300">
              {container.log || '(no logs)'}
            </pre>
          </div>
        ))}
    </div>
  );
}

function PodLogsSection({ reportName }: { reportName: string }) {
  const { data, loading } = useReportLogs(reportName);

  if (loading) return <Skeleton className="h-32 w-full rounded-xl" />;
  if (!data?.podLogs?.length && !data?.events?.length) return null;

  const warningEvents = data.events?.filter((e: EventLog) => e.type === 'Warning') ?? [];

  return (
    <div className="rounded-xl border border-navy-700 bg-navy-800 p-5">
      <h2 className="mb-4 flex items-center gap-2 text-lg font-semibold text-white">
        <Terminal className="h-5 w-5 text-slate-400" />
        Pod Logs
      </h2>

      {warningEvents.length > 0 && (
        <div className="mb-4 rounded-lg border border-amber-500/20 bg-amber-500/10 p-3">
          <p className="mb-2 flex items-center gap-1.5 text-xs font-medium text-amber-400">
            <AlertTriangle className="h-3.5 w-3.5" />
            Warning events
          </p>
          {warningEvents.map((event: EventLog, i: number) => (
            <div key={i} className="mb-1 text-xs text-slate-300">
              <span className="text-amber-400">{event.reason}</span>
              {' \u2014 '}
              {event.involvedObject}: {event.message}
              {(event.count ?? 0) > 1 && (
                <span className="text-slate-500"> (x{event.count})</span>
              )}
            </div>
          ))}
        </div>
      )}

      {data.podLogs?.map((pod: PodLog) => (
        <PodLogCard key={pod.podName} pod={pod} />
      ))}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

function ReportDetail() {
  const { testName } = useParams<{ testName: string }>();
  const navigate = useNavigate();
  const decodedName = testName ? decodeURIComponent(testName) : '';

  const reportsState = useReportsForTest(decodedName);
  const testsState = useTests();

  // Find the matching test CRD for extra metadata
  const testCrd = useMemo(() => {
    if (!testsState.data || !decodedName) return null;
    return testsState.data.find((t) => t.name === decodedName) ?? null;
  }, [testsState.data, decodedName]);

  // Sort reports chronologically (latest first for display, oldest first for chart)
  const sortedReports = useMemo<RestoreReport[]>(() => {
    if (!reportsState.data) return [];
    return reportsState.data
      .slice()
      .sort(
        (a, b) =>
          new Date(b.metadata.creationTimestamp).getTime() -
          new Date(a.metadata.creationTimestamp).getTime(),
      );
  }, [reportsState.data]);

  const latestReport: RestoreReport | undefined = sortedReports[0];

  // Score trend data for chart (oldest first)
  const chartData = useMemo(() => {
    return sortedReports
      .slice()
      .reverse()
      .map((r) => ({
        date: r.metadata?.creationTimestamp?.slice(0, 10) ?? '',
        score: r.status?.score ?? 0,
      }));
  }, [sortedReports]);

  // Score trend (difference between last two reports)
  const scoreTrend = useMemo(() => {
    if (sortedReports.length < 2) return 0;
    return (sortedReports[0].status?.score ?? 0) - (sortedReports[1].status?.score ?? 0);
  }, [sortedReports]);

  // Validation levels from latest report
  const validations = useMemo(() => {
    const vl = latestReport?.status?.validationLevels;
    if (!vl) return [];
    return [
      { name: 'Restore Integrity', ...vl.restoreIntegrity },
      { name: 'Completeness', ...vl.completeness },
      { name: 'Pod Startup', ...vl.podStartup },
      { name: 'Internal Health', ...vl.internalHealth },
      { name: 'Cross-NS Deps', ...vl.crossNamespaceDeps },
      { name: 'RTO Compliance', ...vl.rtoCompliance },
    ].filter((v) => v.status);
  }, [latestReport]);

  // Score breakdown with weighted components
  const scoreBreakdown = useMemo(() => {
    const vl = latestReport?.status?.validationLevels;
    if (!vl) return [];

    const weights = [
      { name: 'Restore Integrity', max: 25, key: 'restoreIntegrity' },
      { name: 'Completeness', max: 20, key: 'completeness' },
      { name: 'Pod Startup', max: 20, key: 'podStartup' },
      { name: 'Internal Health', max: 20, key: 'internalHealth' },
      { name: 'Cross-NS Deps', max: 10, key: 'crossNamespaceDeps' },
      { name: 'RTO Compliance', max: 5, key: 'rtoCompliance' },
    ];

    return weights.map((w) => {
      const level = vl[w.key as keyof typeof vl];
      const earned = level?.status === 'pass' ? w.max : 0;
      return { name: w.name, max: w.max, earned };
    });
  }, [latestReport]);

  // Enrich cross-ns deps with explanatory message when not configured
  const enrichedValidations = useMemo(() => {
    return validations.map((v) => {
      if (v.name === 'Cross-NS Deps' && v.status !== 'pass' && !v.detail) {
        return { ...v, detail: 'No cross-namespace dependencies configured. Add namespaces to your RestoreTest spec.' };
      }
      return v;
    });
  }, [validations]);

  // Completeness grid — derive from validation names
  const completenessItems = useMemo<CompletenessItem[]>(() => {
    const c = latestReport?.status?.completeness;
    return [
      { label: `Deployments ${c?.deployments ?? ''}`, icon: Box, present: !c?.deployments?.includes('0/') },
      { label: `Services ${c?.services ?? ''}`, icon: Network, present: !c?.services?.includes('0/') },
      { label: `Secrets ${c?.secrets ?? ''}`, icon: KeyRound, present: !c?.secrets?.includes('0/') },
      { label: `ConfigMaps ${c?.configMaps ?? ''}`, icon: FileText, present: !c?.configMaps?.includes('0/') },
      { label: `PVCs ${c?.pvcs ?? ''}`, icon: Database, present: !c?.pvcs?.includes('0/') },
    ];
  }, [latestReport]);

  // Timeline events
  const timelineEvents = useMemo<TimelineEvent[]>(() => {
    if (!latestReport) return [];
    const st = latestReport.status;
    if (!st) return [];
    const events: TimelineEvent[] = [];

    if (st.startedAt) {
      events.push({ time: st.startedAt, label: 'Sandbox created', status: 'done' });
    }
    if (st.rto?.measured) {
      events.push({
        time: st.startedAt ?? st.completedAt ?? '',
        label: `Restore completed (${st.rto.measured})`,
        status: 'done',
      });
    }
    if (st.completedAt) {
      events.push({ time: st.completedAt, label: 'Cleanup & completed', status: 'done' });
    }

    return events;
  }, [latestReport]);

  // ---------------------------------------------------------------------------
  // Loading
  // ---------------------------------------------------------------------------

  if (reportsState.loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-64" />
        <div className="grid grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-28 w-full rounded-xl" />
          ))}
        </div>
        <Skeleton className="h-40 w-full rounded-xl" />
      </div>
    );
  }

  // ---------------------------------------------------------------------------
  // Error
  // ---------------------------------------------------------------------------

  if (reportsState.error) {
    return (
      <ErrorState
        message={reportsState.error}
        onRetry={reportsState.refetch}
      />
    );
  }

  // ---------------------------------------------------------------------------
  // No report
  // ---------------------------------------------------------------------------

  if (!latestReport) {
    return (
      <div className="space-y-6">
        <button
          type="button"
          onClick={() => navigate(-1)}
          className="flex items-center gap-1.5 text-sm text-slate-400 transition hover:text-white"
        >
          <ArrowLeft className="h-4 w-4" />
          Back
        </button>
        <div className="flex flex-col items-center justify-center rounded-xl border border-navy-700 bg-navy-800 px-6 py-16 text-center">
          <HardDrive className="mb-3 h-10 w-10 text-slate-600" />
          <h2 className="text-lg font-semibold text-white">{decodedName}</h2>
          <p className="mt-2 text-sm text-slate-400">
            No report available for this test yet.
          </p>
        </div>
      </div>
    );
  }

  // ---------------------------------------------------------------------------
  // Render
  // ---------------------------------------------------------------------------

  const st = latestReport.status;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <nav className="flex items-center gap-1.5 text-sm text-slate-400">
          <button type="button" onClick={() => navigate('/')} className="transition hover:text-white">
            Dashboard
          </button>
          <span>/</span>
          <span className="text-white font-medium">{decodedName}</span>
        </nav>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-white">{decodedName}</h1>
            <StatusBadge status={scoreStatus(st?.result)} />
          </div>
          <p className="text-sm text-slate-400">
            {testCrd?.sourceNamespaces.join(', ') ?? latestReport.spec.testRef}
          </p>
        </div>
        <span className={`font-mono text-4xl font-bold ${scoreColor(st?.score ?? 0)}`}>
          {st?.score ?? '--'}
        </span>
      </div>

      {/* Metric cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          label="Score"
          value={st?.score ?? 0}
          trend={scoreTrend}
          status={metricStatus(st?.score ?? 0)}
          icon={<Activity className="h-5 w-5" />}
        />
        <MetricCard
          label="RTO"
          value={formatDuration(parseDurationToSeconds(st?.rto?.measured))}
          subtitle={testCrd ? `Target: ${testCrd.rtoTarget}` : undefined}
          icon={<Timer className="h-5 w-5" />}
        />
        <MetricCard
          label="Backup Age"
          value={formatTimeAgo(latestReport.metadata.creationTimestamp)}
          icon={<HardDrive className="h-5 w-5" />}
        />
        <MetricCard
          label="Last Test"
          value={st?.completedAt ? formatTimeAgo(st.completedAt) : '--'}
          subtitle={st?.completedAt?.slice(0, 16).replace('T', ' ')}
          icon={<Clock className="h-5 w-5" />}
        />
      </div>

      {/* Validation level bars */}
      {validations.length > 0 && (
        <div className="rounded-xl border border-navy-700 bg-navy-800 p-5">
          <h2 className="mb-4 text-lg font-semibold text-white">
            Validation Results
          </h2>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {enrichedValidations.map((v) => (
              <ValidationLevelBar
                key={v.name}
                name={v.name}
                status={levelStatus(v.status)}
                percentage={v.status === 'pass' ? 100 : 0}
                detail={v.detail ?? ''}
              />
            ))}
          </div>
        </div>
      )}

      {/* Score Breakdown */}
      {validations.length > 0 && (
        <div className="rounded-xl border border-navy-700 bg-navy-800 p-5">
          <h2 className="mb-4 text-lg font-semibold text-white">Score Breakdown</h2>
          <div className="space-y-2">
            {scoreBreakdown.map((item) => (
              <div key={item.name} className="flex items-center gap-3">
                <span className="w-40 text-sm text-slate-400">{item.name}</span>
                <div className="flex-1 h-2 rounded-full bg-navy-700 overflow-hidden">
                  <div
                    className="h-full rounded-full transition-all"
                    style={{
                      width: `${(item.earned / item.max) * 100}%`,
                      backgroundColor: item.earned === item.max ? '#10b981' : item.earned > 0 ? '#f59e0b' : '#ef4444',
                    }}
                  />
                </div>
                <span className={`w-16 text-right font-mono text-sm ${
                  item.earned === item.max ? 'text-emerald-400' : item.earned > 0 ? 'text-amber-400' : 'text-red-400'
                }`}>
                  {item.earned}/{item.max}
                </span>
              </div>
            ))}
            <div className="flex items-center gap-3 border-t border-navy-700 pt-2 mt-2">
              <span className="w-40 text-sm font-semibold text-white">Total</span>
              <div className="flex-1" />
              <span className={`w-16 text-right font-mono text-sm font-bold ${scoreColor(st?.score ?? 0)}`}>
                {st?.score ?? 0}/100
              </span>
            </div>
          </div>
        </div>
      )}

      {/* Health checks table */}
      {(st?.checks?.length ?? 0) > 0 && (
        <div className="rounded-xl border border-navy-700 bg-navy-800 p-5">
          <h2 className="mb-4 text-lg font-semibold text-white">
            Health Checks
          </h2>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-navy-700 text-left text-xs font-medium uppercase tracking-wider text-slate-500">
                  <th className="pb-3 pr-4">Check</th>
                  <th className="pb-3 pr-4">Status</th>
                  <th className="pb-3 pr-4">Duration</th>
                  <th className="pb-3">Message</th>
                </tr>
              </thead>
              <tbody>
                {st?.checks?.map((check) => (
                  <tr
                    key={check.name}
                    className={`border-b border-navy-700/50 ${
                      check.status !== 'pass' ? 'bg-red-500/5' : ''
                    }`}
                  >
                    <td className="py-2.5 pr-4 font-medium text-white">
                      {check.name}
                    </td>
                    <td className="py-2.5 pr-4">
                      {check.status === 'pass' ? (
                        <span className="flex items-center gap-1 text-emerald-400">
                          <CheckCircle className="h-4 w-4" /> Pass
                        </span>
                      ) : (
                        <span className="flex items-center gap-1 text-red-400">
                          <XCircle className="h-4 w-4" /> Fail
                        </span>
                      )}
                    </td>
                    <td className="py-2.5 pr-4 font-mono text-slate-400">
                      {formatCheckDuration(check.duration)}
                    </td>
                    <td className="py-2.5 text-slate-400">{check.message}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Pod logs */}
      {latestReport.metadata?.name && (
        <PodLogsSection reportName={latestReport.metadata.name} />
      )}

      {/* Completeness grid */}
      <div className="rounded-xl border border-navy-700 bg-navy-800 p-5">
        <h2 className="mb-4 text-lg font-semibold text-white">
          Restore Completeness
        </h2>
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
          {completenessItems.map((item) => {
            const Icon = item.icon;
            return (
              <div
                key={item.label}
                className={`flex flex-col items-center gap-2 rounded-lg border p-4 ${
                  item.present
                    ? 'border-emerald-500/20 bg-emerald-500/5'
                    : 'border-navy-700 bg-navy-900/50'
                }`}
              >
                <Icon
                  className={`h-6 w-6 ${
                    item.present ? 'text-emerald-400' : 'text-slate-600'
                  }`}
                />
                <span
                  className={`text-xs font-medium ${
                    item.present ? 'text-emerald-300' : 'text-slate-500'
                  }`}
                >
                  {item.label}
                </span>
                <span
                  className={`text-xs ${
                    item.present ? 'text-emerald-400' : 'text-slate-600'
                  }`}
                >
                  {item.present ? 'Verified' : 'N/A'}
                </span>
              </div>
            );
          })}
        </div>
      </div>

      {/* Score history chart */}
      {chartData.length > 1 && (
        <div className="rounded-xl border border-navy-700 bg-navy-800 p-5">
          <h2 className="mb-4 text-lg font-semibold text-white">
            Score History (30 days)
          </h2>
          <ResponsiveContainer width="100%" height={240}>
            <LineChart
              data={chartData}
              margin={{ top: 5, right: 5, bottom: 5, left: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis
                dataKey="date"
                tick={{ fill: '#94a3b8', fontSize: 11 }}
                tickFormatter={(d: string) => d.slice(5)}
                stroke="#334155"
              />
              <YAxis
                domain={[0, 100]}
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
              />
              <Line
                type="monotone"
                dataKey="score"
                stroke="#3b82f6"
                strokeWidth={2}
                dot={{ fill: '#3b82f6', r: 3 }}
                isAnimationActive={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Timeline */}
      {timelineEvents.length > 0 && (
        <div className="rounded-xl border border-navy-700 bg-navy-800 p-5">
          <h2 className="mb-4 text-lg font-semibold text-white">
            Test Timeline
          </h2>
          <div className="relative ml-4 border-l-2 border-navy-600 pl-6">
            {timelineEvents.map((event, idx) => (
              <div key={idx} className="relative mb-6 last:mb-0">
                <div className="absolute -left-[31px] top-0.5 h-3 w-3 rounded-full border-2 border-navy-800 bg-emerald-400" />
                <p className="text-sm font-medium text-white">{event.label}</p>
                <p className="text-xs text-slate-500">
                  {new Date(event.time).toLocaleString()}
                </p>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

export default ReportDetail;
