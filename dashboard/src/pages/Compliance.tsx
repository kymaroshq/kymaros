import { useState, useMemo } from 'react';
import { FileDown, LoaderCircle, ShieldCheck, TrendingUp, AlertTriangle, Sparkles } from 'lucide-react';
import { useCompliance, useDailyScores, useLicense } from '../hooks/useKymarosData';
import { exportCompliancePDF } from '../api/kymarosApi';
import MetricCard from '../components/ui/MetricCard';
import HeatmapCalendar from '../components/ui/HeatmapCalendar';
import Skeleton from '../components/ui/Skeleton';
import ErrorState from '../components/ui/ErrorState';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type Framework = 'soc2' | 'iso27001' | 'dora';

const frameworkLabels: Record<Framework, string> = {
  soc2: 'SOC 2',
  iso27001: 'ISO 27001',
  dora: 'DORA',
};

function complianceStatus(rate: number): 'pass' | 'fail' | 'partial' | 'neutral' {
  if (rate >= 95) return 'pass';
  if (rate >= 80) return 'partial';
  if (rate > 0) return 'fail';
  return 'neutral';
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

function Compliance() {
  const [framework, setFramework] = useState<Framework>('soc2');
  const [exporting, setExporting] = useState(false);
  const period = '90';

  const license = useLicense();
  const compliance = useCompliance(framework, period);
  const dailyScores = useDailyScores(90);

  // Convert DailySummary[] to heatmap format
  const heatmapData = useMemo(() => {
    if (!dailyScores.data) return [];
    return dailyScores.data.map((d) => ({
      date: d.date,
      score: d.score,
      tests: d.tests,
      failures: d.failures,
    }));
  }, [dailyScores.data]);

  const c = compliance.data;

  const handleExportPDF = async () => {
    setExporting(true);
    try {
      const blob = await exportCompliancePDF(framework, period);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `kymaros-compliance-${framework}-${new Date().toISOString().slice(0, 10)}.pdf`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (err) {
      console.error('PDF export failed:', err);
    } finally {
      setExporting(false);
    }
  };

  if (license.data && license.data.tier === 'community') {
    return (
      <div className="py-16">
        <div className="relative mx-auto max-w-2xl rounded-2xl border border-slate-700/50 bg-slate-800/30 p-12 text-center">
          <div className="mb-6 inline-flex rounded-xl bg-blue-500/10 p-4">
            <Sparkles className="h-8 w-8 text-blue-400" />
          </div>
          <h2 className="text-xl font-semibold text-slate-200">
            Compliance reports are available on Kymaros Team
          </h2>
          <p className="mx-auto mt-3 max-w-md text-sm text-slate-400">
            Generate audit-ready evidence for SOC 2, ISO 27001, DORA, HIPAA, and PCI-DSS.
            365 documented DR tests per year &mdash; automatically.
          </p>
          <div className="mt-8 flex justify-center gap-4">
            <a
              href="mailto:sales@kymaros.io"
              className="rounded-lg bg-blue-600 px-5 py-2.5 text-sm font-semibold text-white hover:bg-blue-500"
            >
              Start Free Trial
            </a>
            <a
              href="https://kymaros.io#pricing"
              className="rounded-lg border border-slate-700 px-5 py-2.5 text-sm font-semibold text-slate-300 hover:border-slate-600"
            >
              Learn more
            </a>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Compliance</h1>
          <p className="text-sm text-slate-400">
            Framework compliance tracking and audit readiness
          </p>
        </div>
        {license.data?.features.pdfExport ? (
          <button
            type="button"
            onClick={handleExportPDF}
            disabled={exporting || !c}
            className="flex items-center gap-2 rounded-lg border border-navy-600 bg-navy-800 px-4 py-2 text-sm font-medium text-slate-200 transition-colors hover:bg-navy-700 disabled:opacity-60 disabled:cursor-not-allowed"
          >
            {exporting ? (
              <LoaderCircle className="h-4 w-4 animate-spin" />
            ) : (
              <FileDown className="h-4 w-4" />
            )}
            {exporting ? 'Generating...' : 'Export PDF'}
          </button>
        ) : (
          <button
            type="button"
            disabled
            className="flex items-center gap-2 rounded-lg border border-navy-600 bg-navy-800 px-4 py-2 text-sm font-medium text-slate-400 opacity-60 cursor-not-allowed"
            title="Available in Enterprise tier"
          >
            <FileDown className="h-4 w-4" />
            Export PDF
            <span className="rounded bg-navy-600 px-1.5 py-0.5 text-[10px] font-semibold uppercase text-slate-500">
              Enterprise
            </span>
          </button>
        )}
      </div>

      {/* Framework tabs */}
      <div className="flex rounded-lg border border-navy-700 p-0.5 w-fit">
        {(Object.entries(frameworkLabels) as [Framework, string][]).map(
          ([key, label]) => (
            <button
              key={key}
              type="button"
              onClick={() => setFramework(key)}
              className={`rounded-md px-4 py-2 text-sm font-medium transition-colors ${
                framework === key
                  ? 'bg-navy-600 text-white'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              {label}
            </button>
          ),
        )}
      </div>

      {/* Loading */}
      {compliance.loading ? (
        <div className="space-y-4">
          <div className="grid grid-cols-4 gap-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-28 w-full rounded-xl" />
            ))}
          </div>
          <Skeleton className="h-40 w-full rounded-xl" />
        </div>
      ) : compliance.error ? (
        <ErrorState message={compliance.error} onRetry={compliance.refetch} />
      ) : c ? (
        <>
          {/* Summary card */}
          <div className="rounded-xl border border-navy-700 bg-navy-800 p-5">
            <div className="flex items-center gap-3 mb-4">
              <ShieldCheck className="h-6 w-6 text-blue-400" />
              <div>
                <h2 className="text-lg font-semibold text-white">
                  {frameworkLabels[framework]} Compliance
                </h2>
                <p className="text-xs text-slate-400">
                  Period: {c.period} | Status: {c.status}
                </p>
              </div>
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
              <MetricCard
                label="Average Score"
                value={`${c.averageScore.toFixed(1)}`}
                status={complianceStatus(c.averageScore)}
                icon={<TrendingUp className="h-5 w-5" />}
              />
              <MetricCard
                label="Test Coverage"
                value={`${c.daysWithTests} / ${c.daysInPeriod}`}
                subtitle={`${((c.daysWithTests / Math.max(c.daysInPeriod, 1)) * 100).toFixed(0)}% days tested`}
              />
              <MetricCard
                label="Tests Executed"
                value={c.testsExecuted}
                subtitle={`${c.issuesDetected} issues detected`}
                status={c.issuesDetected > 0 ? 'fail' : 'neutral'}
                icon={<AlertTriangle className="h-5 w-5" />}
              />
              <MetricCard
                label="RTO"
                value={c.averageRTO}
                subtitle={`Target: ${c.rtoTarget} | ${c.rtoCompliant ? 'Compliant' : 'Non-compliant'}`}
              />
            </div>
          </div>

          {/* Heatmap */}
          <div className="rounded-xl border border-navy-700 bg-navy-800 p-5">
            <h2 className="mb-4 text-lg font-semibold text-white">
              Daily Test Results (90 days)
            </h2>
            {dailyScores.loading ? (
              <Skeleton className="h-24 w-full" />
            ) : dailyScores.error ? (
              <ErrorState
                message={dailyScores.error}
                onRetry={dailyScores.refetch}
              />
            ) : heatmapData.length === 0 ? (
              <p className="py-6 text-center text-sm text-slate-500">
                No daily data available.
              </p>
            ) : (
              <HeatmapCalendar data={heatmapData} days={90} />
            )}
          </div>
        </>
      ) : (
        <p className="py-10 text-center text-sm text-slate-500">
          No compliance data available for {frameworkLabels[framework]}.
        </p>
      )}
    </div>
  );
}

export default Compliance;
