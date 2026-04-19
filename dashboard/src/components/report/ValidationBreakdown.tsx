import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui2/Card';
import { ProgressBar } from '@/components/ui2/ProgressBar';
import type { ValidationLevelResult } from '@/types/kymaros';

interface ValidationBreakdownProps {
  levels?: {
    restoreIntegrity?: ValidationLevelResult;
    completeness?: ValidationLevelResult;
    podStartup?: ValidationLevelResult;
    internalHealth?: ValidationLevelResult;
    crossNamespaceDeps?: ValidationLevelResult;
    rtoCompliance?: ValidationLevelResult;
  };
  totalScore: number;
}

const rows = [
  { key: 'restoreIntegrity', label: 'Restore Integrity', max: 25 },
  { key: 'completeness', label: 'Completeness', max: 20 },
  { key: 'podStartup', label: 'Pod Startup', max: 20 },
  { key: 'internalHealth', label: 'Health Checks', max: 20 },
  { key: 'crossNamespaceDeps', label: 'Cross-NS Dependencies', max: 10 },
  { key: 'rtoCompliance', label: 'RTO Compliance', max: 5 },
] as const;

function levelTone(status?: string): 'success' | 'warning' | 'danger' | 'default' {
  if (status === 'pass') return 'success';
  if (status === 'partial') return 'warning';
  if (status === 'fail') return 'danger';
  return 'default';
}

function levelScore(level: ValidationLevelResult | undefined, max: number): number {
  if (!level) return 0;
  if (level.status === 'pass') return max;
  if (level.status === 'partial') return Math.round(max * 0.6);
  return 0;
}

export function ValidationBreakdown({ levels, totalScore }: ValidationBreakdownProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Validation Breakdown</CardTitle>
        <span className="text-2xs text-text-tertiary font-mono">6 levels</span>
      </CardHeader>
      <CardBody className="space-y-2.5">
        {rows.map((row) => {
          const level = levels?.[row.key];
          const value = levelScore(level, row.max);
          return (
            <div key={row.key} className="grid grid-cols-[180px_1fr_60px] items-center gap-4">
              <span className="text-sm text-text-secondary">{row.label}</span>
              <ProgressBar value={value} max={row.max} tone={levelTone(level?.status)} />
              <span className="text-xs font-mono text-text-primary tabular-nums text-right">
                {value}/{row.max}
              </span>
            </div>
          );
        })}

        <div className="grid grid-cols-[180px_1fr_60px] items-center gap-4 pt-2.5 border-t border-border-subtle">
          <span className="text-sm font-medium text-text-primary">Total</span>
          <ProgressBar
            value={totalScore}
            max={100}
            tone={totalScore >= 90 ? 'success' : totalScore >= 70 ? 'default' : totalScore >= 50 ? 'warning' : 'danger'}
          />
          <span className="text-sm font-semibold font-mono tabular-nums text-right text-text-primary">
            {Math.round(totalScore)}/100
          </span>
        </div>
      </CardBody>
    </Card>
  );
}
