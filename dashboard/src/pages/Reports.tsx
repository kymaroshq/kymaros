import { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card } from '@/components/ui2/Card';
import { Stat } from '@/components/ui2/Stat';
import { DataTable } from '@/components/ui2/DataTable';
import { Badge } from '@/components/ui2/Badge';
import { ScoreIndicator } from '@/components/ui2/ScoreIndicator';
import { Button } from '@/components/ui2/Button';
import { EmptyState } from '@/components/ui2/EmptyState';
import { Pagination } from '@/components/ui2/Pagination';
import { TableSkeleton } from '@/components/ui2/Skeleton';
import { useApiData } from '@/hooks/useKymarosData';
import { kymarosApi } from '@/api/kymarosApi';
import { formatRelativeTime, formatRTO, getResultVariant } from '@/lib/utils';
import { Search, FileText, X } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { RestoreReport } from '@/types/kymaros';

const PAGE_SIZE = 25;
const DAYS_OPTIONS = [
  { label: '7 days', value: 7 },
  { label: '30 days', value: 30 },
  { label: '90 days', value: 90 },
];

export default function Reports() {
  const navigate = useNavigate();
  const [days, setDays] = useState(30);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'pass' | 'partial' | 'fail'>('all');
  const [page, setPage] = useState(1);

  const reports = useApiData<RestoreReport[]>(
    () => kymarosApi.getReports({ days }),
    0,
    [days]
  );

  const filtered = useMemo(() => {
    if (!reports.data) return [];
    return reports.data.filter((r) => {
      if (search && !r.spec.testRef.toLowerCase().includes(search.toLowerCase()) &&
          !r.metadata.namespace.toLowerCase().includes(search.toLowerCase())) return false;
      if (statusFilter !== 'all' && r.status?.result !== statusFilter) return false;
      return true;
    });
  }, [reports.data, search, statusFilter]);

  const totalPages = Math.ceil(filtered.length / PAGE_SIZE);
  const paged = filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  // Stats
  const stats = useMemo(() => {
    if (!reports.data || reports.data.length === 0) return null;
    const total = reports.data.length;
    const tests = new Set(reports.data.map((r) => r.spec.testRef)).size;
    const pass = reports.data.filter((r) => r.status?.result === 'pass').length;
    const fail = reports.data.filter((r) => r.status?.result === 'fail').length;
    return {
      tests,
      runs: total,
      successRate: (pass / total) * 100,
      failureRate: (fail / total) * 100,
    };
  }, [reports.data]);

  const hasActiveFilters = search !== '' || statusFilter !== 'all';

  const resetFilters = () => {
    setSearch('');
    setStatusFilter('all');
    setPage(1);
  };

  return (
    <div className="px-6 py-6 space-y-6 max-w-[1600px] mx-auto">
      {/* Header */}
      <div>
        <h1 className="text-xl font-semibold text-text-primary">Reports</h1>
        <p className="text-sm text-text-tertiary mt-0.5">
          {reports.data
            ? `${filtered.length} runs across ${stats?.tests ?? 0} tests`
            : 'Historical records of restore validations'}
        </p>
      </div>

      {/* Stats */}
      {stats && (
        <div className="grid grid-cols-4 gap-px bg-border-subtle rounded-lg overflow-hidden">
          <div className="bg-surface-2 p-4">
            <Stat label="Tests" value={stats.tests} size="md" />
          </div>
          <div className="bg-surface-2 p-4">
            <Stat label="Total Runs" value={stats.runs} size="md" />
          </div>
          <div className="bg-surface-2 p-4">
            <Stat
              label="Success Rate"
              value={`${Math.round(stats.successRate)}%`}
              tone={stats.successRate >= 90 ? 'success' : stats.successRate >= 70 ? 'default' : 'warning'}
              size="md"
            />
          </div>
          <div className="bg-surface-2 p-4">
            <Stat
              label="Failure Rate"
              value={`${Math.round(stats.failureRate)}%`}
              tone={stats.failureRate > 10 ? 'danger' : 'default'}
              size="md"
            />
          </div>
        </div>
      )}

      {/* Table */}
      <Card>
        {/* Filters bar */}
        <div className="flex flex-wrap items-center gap-2 px-4 py-3 border-b border-border-subtle">
          <div className="relative flex-1 min-w-48 max-w-xs">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-text-tertiary pointer-events-none" />
            <input
              value={search}
              onChange={(e) => { setSearch(e.target.value); setPage(1); }}
              placeholder="Search test name..."
              className="w-full h-9 rounded-md border border-border-default bg-surface-2 pl-8 pr-3 text-sm text-text-primary placeholder:text-text-tertiary focus:outline-none focus:ring-1 focus:ring-accent focus:border-accent transition-colors duration-150"
            />
          </div>

          <div className="flex items-center gap-1 bg-surface-3 rounded-md p-0.5">
            {DAYS_OPTIONS.map((opt) => (
              <button
                key={opt.value}
                onClick={() => { setDays(opt.value); setPage(1); }}
                className={cn(
                  'text-xs font-medium px-2 py-1 rounded transition-colors duration-150',
                  days === opt.value ? 'bg-surface-1 text-text-primary' : 'text-text-tertiary hover:text-text-primary'
                )}
              >
                {opt.label}
              </button>
            ))}
          </div>

          <div className="flex items-center gap-1 bg-surface-3 rounded-md p-0.5">
            {(['all', 'pass', 'partial', 'fail'] as const).map((s) => (
              <button
                key={s}
                onClick={() => { setStatusFilter(s); setPage(1); }}
                className={cn(
                  'text-xs font-medium px-2 py-1 rounded transition-colors duration-150',
                  statusFilter === s ? 'bg-surface-1 text-text-primary' : 'text-text-tertiary hover:text-text-primary'
                )}
              >
                {s === 'all' ? 'All' : s}
              </button>
            ))}
          </div>

          {hasActiveFilters && (
            <Button variant="ghost" size="sm" onClick={resetFilters}>
              <X className="h-3 w-3" /> Clear
            </Button>
          )}
        </div>

        {/* Content */}
        {reports.loading ? (
          <TableSkeleton rows={8} />
        ) : filtered.length === 0 ? (
          hasActiveFilters ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <Search className="h-6 w-6 text-text-tertiary mb-3" />
              <h3 className="text-sm font-medium text-text-primary">No reports match your filters</h3>
              <p className="text-xs text-text-tertiary mt-1">Try adjusting the date range or removing some filters.</p>
              <Button variant="secondary" onClick={resetFilters} className="mt-4">Clear all filters</Button>
            </div>
          ) : (
            <EmptyState
              icon={<FileText className="h-6 w-6" />}
              title="No reports yet"
              description="Reports will appear here after your first restore test runs."
            />
          )
        ) : (
          <DataTable<RestoreReport>
            columns={[
              {
                key: 'testRef',
                label: 'Test',
                render: (r) => (
                  <div>
                    <div className="font-medium text-text-primary text-sm">{r.spec.testRef}</div>
                    <div className="text-2xs text-text-tertiary font-mono">{r.metadata.namespace}</div>
                  </div>
                ),
              },
              {
                key: 'score',
                label: 'Score',
                align: 'right',
                render: (r) => <ScoreIndicator value={r.status?.score ?? 0} size="sm" />,
              },
              {
                key: 'result',
                label: 'Status',
                render: (r) => (
                  <Badge variant={getResultVariant(r.status?.result ?? 'fail')} dot>
                    {r.status?.result ?? 'unknown'}
                  </Badge>
                ),
              },
              {
                key: 'rto',
                label: 'RTO',
                align: 'right',
                mono: true,
                render: (r) => (
                  <span className="text-data text-text-secondary">
                    {formatRTO(r.status?.rto?.measured)}
                  </span>
                ),
              },
              {
                key: 'ran',
                label: 'Ran',
                align: 'right',
                render: (r) => (
                  <span className="text-2xs text-text-tertiary font-mono">
                    {formatRelativeTime(r.status?.startedAt ?? r.metadata.creationTimestamp)}
                  </span>
                ),
              },
            ]}
            rows={paged}
            onRowClick={(r) => navigate(`/reports/${r.spec.testRef}`)}
            compact
          />
        )}

        {/* Pagination */}
        {filtered.length > PAGE_SIZE && (
          <div className="flex items-center justify-between gap-4 px-4 py-3 border-t border-border-subtle">
            <span className="text-xs text-text-tertiary font-mono">
              {(page - 1) * PAGE_SIZE + 1}&ndash;{Math.min(page * PAGE_SIZE, filtered.length)} of {filtered.length}
            </span>
            <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
          </div>
        )}
      </Card>
    </div>
  );
}
