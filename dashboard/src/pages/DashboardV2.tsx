import { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui2/Card';
import { Stat } from '@/components/ui2/Stat';
import { DataTable } from '@/components/ui2/DataTable';
import { Badge } from '@/components/ui2/Badge';
import { ScoreIndicator } from '@/components/ui2/ScoreIndicator';
import { StatSkeleton, TableSkeleton } from '@/components/ui2/Skeleton';
import { EmptyState } from '@/components/ui2/EmptyState';
import { ScoreTrendChart } from '@/components/charts/ScoreTrendChart';
import { TestDistribution } from '@/components/charts/TestDistribution';
import { useSummary, useTests, useDailyScores, useAlerts, useUpcoming } from '@/hooks/useKymarosData';
import { formatRelativeTime, getResultVariant, formatDuration, formatRTO, formatCron } from '@/lib/utils';
import { AlertTriangle, CheckCircle, Clock, ListChecks, RefreshCw } from 'lucide-react';
import type { TestResponse, Alert, UpcomingTest } from '@/types/kymaros';

function scoreTone(score: number): 'success' | 'warning' | 'danger' | 'default' {
  if (score >= 90) return 'success';
  if (score >= 70) return 'default';
  if (score >= 50) return 'warning';
  return 'danger';
}

export default function DashboardV2() {
  const navigate = useNavigate();
  const summary = useSummary();
  const tests = useTests();
  const dailyScores = useDailyScores(7);
  const alerts = useAlerts();
  const upcoming = useUpcoming();
  const [statusFilter, setStatusFilter] = useState<'all' | 'pass' | 'partial' | 'fail'>('all');

  const [refreshing, setRefreshing] = useState(false);

  function handleRefresh() {
    setRefreshing(true);
    Promise.all([
      summary.refetch(),
      tests.refetch(),
      dailyScores.refetch(),
      alerts.refetch(),
      upcoming.refetch(),
    ]).finally(() => setRefreshing(false));
  }

  const passCount = tests.data?.filter((t) => t.lastResult === 'pass').length ?? 0;
  const partialCount = tests.data?.filter((t) => t.lastResult === 'partial').length ?? 0;
  const failCount = tests.data?.filter((t) => t.lastResult === 'fail').length ?? 0;

  const filteredTests = useMemo(() => {
    if (!tests.data) return [];
    if (statusFilter === 'all') return tests.data;
    return tests.data.filter((t) => t.lastResult === statusFilter);
  }, [tests.data, statusFilter]);

  return (
    <div className="px-6 py-6 space-y-6 max-w-[1600px] mx-auto">
      {/* Page header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Dashboard</h1>
          <p className="text-sm text-text-tertiary mt-0.5">
            Monitoring {summary.data?.totalTests ?? '\u2014'} restore tests across {summary.data?.namespacesCovered ?? '\u2014'} namespaces
          </p>
        </div>
        <button
          onClick={handleRefresh}
          disabled={refreshing}
          className="flex items-center gap-2 px-3 py-1.5 rounded-md text-sm text-text-secondary bg-surface-2 border border-border-default hover:bg-surface-3 hover:text-text-primary transition-colors duration-150 disabled:opacity-50"
        >
          <RefreshCw className={`h-3.5 w-3.5 ${refreshing ? 'animate-spin' : ''}`} />
          Refresh
        </button>
      </div>

      {/* Hero stats */}
      <div className="grid grid-cols-4 gap-px bg-border-subtle rounded-lg overflow-hidden">
        {summary.loading ? (
          Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="bg-surface-2 p-4"><StatSkeleton /></div>
          ))
        ) : (
          <>
            <div className="bg-surface-2 p-4">
              <Stat
                label="Average Score"
                value={summary.data?.averageScore ?? '\u2014'}
                unit="/100"
                tone={scoreTone(summary.data?.averageScore ?? 0)}
                size="lg"
              />
            </div>
            <div className="bg-surface-2 p-4">
              <Stat
                label="Tests Last 24h"
                value={summary.data?.testsLastNight ?? '\u2014'}
                sublabel={`${passCount} passed \u00b7 ${failCount} failed`}
                size="lg"
              />
            </div>
            <div className="bg-surface-2 p-4">
              <Stat
                label="Issues"
                value={(summary.data?.totalFailures ?? 0) + (summary.data?.totalPartial ?? 0)}
                sublabel={`${summary.data?.totalFailures ?? 0} critical \u00b7 ${summary.data?.totalPartial ?? 0} warnings`}
                tone={(summary.data?.totalFailures ?? 0) > 0 ? 'danger' : (summary.data?.totalPartial ?? 0) > 0 ? 'warning' : 'default'}
                size="lg"
              />
            </div>
            <div className="bg-surface-2 p-4">
              <Stat
                label="Avg RTO"
                value={formatRTO(summary.data?.averageRTO)}
                sublabel={tests.data?.[0]?.rtoTarget ? `Target ${formatDuration(tests.data[0].rtoTarget)}` : undefined}
                size="lg"
              />
            </div>
          </>
        )}
      </div>

      {/* Charts row */}
      <div className="grid grid-cols-3 gap-4">
        <Card className="col-span-2">
          <CardHeader>
            <CardTitle>Score Trend</CardTitle>
            <span className="text-2xs text-text-tertiary font-mono">7 days</span>
          </CardHeader>
          <CardBody>
            {dailyScores.loading ? (
              <div className="h-48 bg-surface-3 rounded animate-pulse-subtle" />
            ) : (
              <ScoreTrendChart data={dailyScores.data ?? []} />
            )}
          </CardBody>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Test Distribution</CardTitle>
          </CardHeader>
          <CardBody>
            {tests.loading ? (
              <div className="h-48 bg-surface-3 rounded animate-pulse-subtle" />
            ) : (
              <TestDistribution pass={passCount} partial={partialCount} fail={failCount} />
            )}
          </CardBody>
        </Card>
      </div>

      {/* Tests table */}
      <Card>
        <CardHeader>
          <CardTitle>Restore Tests</CardTitle>
          <div className="flex items-center gap-1 bg-surface-3 rounded-md p-0.5">
            {(['all', 'pass', 'partial', 'fail'] as const).map((f) => (
              <button
                key={f}
                onClick={() => setStatusFilter(f)}
                className={`text-xs font-medium px-2 py-1 rounded transition-colors duration-150 ${
                  statusFilter === f
                    ? 'bg-surface-1 text-text-primary'
                    : 'text-text-tertiary hover:text-text-primary'
                }`}
              >
                {f === 'all' ? `All (${tests.data?.length ?? 0})` : f === 'pass' ? `Pass (${passCount})` : f === 'partial' ? `Partial (${partialCount})` : `Fail (${failCount})`}
              </button>
            ))}
          </div>
        </CardHeader>
        {tests.loading ? (
          <TableSkeleton />
        ) : filteredTests.length === 0 ? (
          <EmptyState
            icon={<ListChecks className="h-6 w-6" />}
            title={statusFilter === 'all' ? 'No restore tests configured' : `No ${statusFilter} tests`}
            description={statusFilter === 'all' ? 'Create a RestoreTest resource to start validating your backups.' : undefined}
          />
        ) : (
          <DataTable<TestResponse>
            columns={[
              {
                key: 'name',
                label: 'Name',
                render: (row) => (
                  <div>
                    <div className="font-medium text-text-primary">{row.name}</div>
                    <div className="text-2xs text-text-tertiary">{row.namespace}</div>
                  </div>
                ),
              },
              {
                key: 'lastScore',
                label: 'Score',
                align: 'right',
                render: (row) => <ScoreIndicator value={row.lastScore} size="sm" />,
              },
              {
                key: 'lastResult',
                label: 'Status',
                render: (row) => (
                  <Badge variant={getResultVariant(row.lastResult)} dot>
                    {row.lastResult}
                  </Badge>
                ),
              },
              {
                key: 'schedule',
                label: 'Schedule',
                render: (row) => (
                  <span className="text-data text-text-secondary text-xs">
                    {formatCron(row.schedule)}
                  </span>
                ),
              },
              {
                key: 'lastRunAt',
                label: 'Last Run',
                align: 'right',
                render: (row) => (
                  <span className="text-data text-text-secondary">
                    {row.lastRunAt ? formatRelativeTime(row.lastRunAt) : '\u2014'}
                  </span>
                ),
              },
            ]}
            rows={filteredTests}
            keyField="name"
            onRowClick={(row) => navigate(`/reports/${row.name}`)}
          />
        )}
      </Card>

      {/* Alerts + Upcoming */}
      <div className="grid grid-cols-2 gap-4">
        {/* Alerts */}
        <Card>
          <CardHeader>
            <CardTitle>Recent Alerts</CardTitle>
            <span className="text-2xs text-text-tertiary font-mono">{alerts.data?.length ?? 0}</span>
          </CardHeader>
          {alerts.loading ? (
            <TableSkeleton rows={3} />
          ) : !alerts.data || alerts.data.length === 0 ? (
            <EmptyState
              icon={<CheckCircle className="h-5 w-5" />}
              title="All clear"
              description="No alerts in the last 48 hours."
              className="py-8"
            />
          ) : (
            <div className="divide-y divide-border-subtle">
              {alerts.data.slice(0, 5).map((alert: Alert, idx: number) => (
                <button
                  key={`${alert.testName}-${idx}`}
                  type="button"
                  className="w-full px-4 py-3 flex items-start justify-between gap-3 hover:bg-surface-3 transition-colors duration-150 text-left"
                  onClick={() => navigate(`/reports/${alert.testName}`)}
                >
                  <div className="flex items-start gap-2.5 min-w-0">
                    <AlertTriangle className="h-3.5 w-3.5 mt-0.5 shrink-0 text-status-warning" />
                    <div className="min-w-0">
                      <div className="text-sm text-text-primary font-medium truncate">{alert.testName}</div>
                      <div className="text-2xs text-text-tertiary truncate mt-0.5">{alert.message}</div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 shrink-0">
                    <ScoreIndicator value={alert.score} size="sm" />
                    <span className="text-2xs text-text-tertiary font-mono">
                      {formatRelativeTime(alert.timestamp)}
                    </span>
                  </div>
                </button>
              ))}
            </div>
          )}
        </Card>

        {/* Upcoming */}
        <Card>
          <CardHeader>
            <CardTitle>Upcoming Tests</CardTitle>
            <span className="text-2xs text-text-tertiary font-mono">{upcoming.data?.length ?? 0}</span>
          </CardHeader>
          {upcoming.loading ? (
            <TableSkeleton rows={3} />
          ) : !upcoming.data || upcoming.data.length === 0 ? (
            <EmptyState
              icon={<Clock className="h-5 w-5" />}
              title="No tests scheduled"
              description="Create a RestoreTest to get started."
              className="py-8"
            />
          ) : (
            <div className="divide-y divide-border-subtle">
              {upcoming.data.map((test: UpcomingTest) => (
                <button
                  key={test.name}
                  type="button"
                  className="w-full px-4 py-3 flex items-center justify-between gap-3 hover:bg-surface-3 transition-colors duration-150 text-left"
                  onClick={() => navigate(`/reports/${test.name}`)}
                >
                  <div className="min-w-0">
                    <div className="text-sm text-text-primary font-medium truncate">{test.name}</div>
                    <div className="text-2xs text-text-tertiary font-mono truncate mt-0.5">{test.namespace}</div>
                  </div>
                  <div className="flex items-center gap-3 shrink-0">
                    <span className="text-2xs text-text-tertiary">last score:</span>
                    <ScoreIndicator value={test.lastScore} size="sm" />
                    <span className="text-2xs text-text-tertiary font-mono">
                      {formatRelativeTime(test.nextRunAt)}
                    </span>
                  </div>
                </button>
              ))}
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
