import { useParams } from 'react-router-dom';
import { ReportHeader } from '@/components/report/ReportHeader';
import { ValidationBreakdown } from '@/components/report/ValidationBreakdown';
import { HealthChecksTable } from '@/components/report/HealthChecksTable';
import { RestoreCompleteness } from '@/components/report/RestoreCompleteness';
import { PodLogsPanel } from '@/components/report/PodLogsPanel';
import { EventsTable } from '@/components/report/EventsTable';
import { ScoreHistoryChart } from '@/components/report/ScoreHistoryChart';
import { TestTimeline } from '@/components/report/TestTimeline';
import { Stat } from '@/components/ui2/Stat';
import { StatSkeleton } from '@/components/ui2/Skeleton';
import { EmptyState } from '@/components/ui2/EmptyState';
import { useReportsForTest, useReportLogs } from '@/hooks/useKymarosData';
import { formatRelativeTime, formatDuration, formatRTO } from '@/lib/utils';
import { FileX } from 'lucide-react';

export default function ReportDetailV2() {
  const { testName } = useParams<{ testName: string }>();
  const reports = useReportsForTest(testName!);
  const latestReport = reports.data?.[0];
  const logs = useReportLogs(latestReport?.metadata.name ?? '');

  const st = latestReport?.status;

  if (reports.loading) {
    return (
      <div className="px-6 py-6 space-y-6 max-w-[1600px] mx-auto">
        <div className="h-8 w-64 bg-surface-3 rounded animate-pulse-subtle" />
        <div className="grid grid-cols-4 gap-px bg-border-subtle rounded-lg overflow-hidden">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="bg-surface-2 p-4"><StatSkeleton /></div>
          ))}
        </div>
        <div className="h-64 bg-surface-3 rounded animate-pulse-subtle" />
      </div>
    );
  }

  if (!latestReport || !st) {
    return (
      <div className="px-6 py-16">
        <EmptyState
          icon={<FileX className="h-8 w-8" />}
          title="No reports found"
          description={`No restore reports exist yet for test "${testName}". Wait for the next scheduled run or trigger one manually.`}
        />
      </div>
    );
  }

  const score = Math.round(st.score ?? 0);
  const result = st.result ?? 'fail';

  function scoreTone(s: number): 'success' | 'warning' | 'danger' | 'default' {
    if (s >= 90) return 'success';
    if (s >= 70) return 'default';
    if (s >= 50) return 'warning';
    return 'danger';
  }

  return (
    <div className="px-6 py-6 space-y-6 max-w-[1600px] mx-auto">
      <ReportHeader
        testName={testName!}
        namespace={latestReport.metadata.namespace}
        score={score}
        status={result}
      />

      {/* Hero stats */}
      <div className="grid grid-cols-4 gap-px bg-border-subtle rounded-lg overflow-hidden">
        <div className="bg-surface-2 p-4">
          <Stat label="Score" value={score} unit="/100" tone={scoreTone(score)} size="lg" />
        </div>
        <div className="bg-surface-2 p-4">
          <Stat
            label="RTO"
            value={formatRTO(st.rto?.measured)}
            sublabel={st.rto?.target ? `Target ${formatDuration(st.rto.target)} ${st.rto.withinSLA ? '\u2713' : '\u2717'}` : undefined}
            tone={st.rto?.withinSLA ? 'success' : st.rto?.withinSLA === false ? 'danger' : 'default'}
            size="lg"
          />
        </div>
        <div className="bg-surface-2 p-4">
          <Stat
            label="Backup Age"
            value={st.backup?.age ? formatDuration(st.backup.age) : '\u2014'}
            sublabel={st.backup?.name}
            size="lg"
          />
        </div>
        <div className="bg-surface-2 p-4">
          <Stat
            label="Completed"
            value={st.completedAt ? formatRelativeTime(st.completedAt) : '\u2014'}
            size="lg"
          />
        </div>
      </div>

      {/* Validation Breakdown */}
      <ValidationBreakdown levels={st.validationLevels} totalScore={score} />

      {/* Health checks + Completeness */}
      <div className="grid grid-cols-3 gap-4 items-stretch">
        <div className="col-span-2 flex flex-col">
          <HealthChecksTable checks={st.checks ?? []} />
        </div>
        <div className="flex flex-col">
          <RestoreCompleteness completeness={st.completeness} />
        </div>
      </div>

      {/* Pod Logs */}
      {logs.data?.podLogs && logs.data.podLogs.length > 0 && (
        <PodLogsPanel pods={logs.data.podLogs} />
      )}

      {/* Events */}
      {logs.data?.events && logs.data.events.length > 0 && (
        <EventsTable events={logs.data.events} />
      )}

      {/* History + Timeline */}
      <div className="grid grid-cols-3 gap-4">
        <div className="col-span-2">
          <ScoreHistoryChart
            reports={reports.data ?? []}
            currentDate={st.startedAt?.split('T')[0]}
          />
        </div>
        <TestTimeline
          startedAt={st.startedAt}
          completedAt={st.completedAt}
          result={result}
        />
      </div>
    </div>
  );
}
