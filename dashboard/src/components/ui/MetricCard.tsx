import type { FC, ReactNode } from 'react';
import { TrendingUp, TrendingDown } from 'lucide-react';

type MetricStatus = 'pass' | 'fail' | 'partial' | 'neutral';

interface MetricCardProps {
  label: string;
  value: string | number;
  subtitle?: string;
  trend?: number;
  status?: MetricStatus;
  icon?: ReactNode;
}

const statusBorderColors: Record<MetricStatus, string> = {
  pass: 'border-l-[#10b981]',
  partial: 'border-l-[#f59e0b]',
  fail: 'border-l-[#ef4444]',
  neutral: 'border-l-slate-300 dark:border-l-slate-600',
};

const MetricCard: FC<MetricCardProps> = ({
  label,
  value,
  subtitle,
  trend,
  status = 'neutral',
  icon,
}) => {
  return (
    <div
      className={`rounded-lg border border-slate-200 bg-white p-4 shadow-sm transition-shadow hover:shadow-md dark:border-slate-700 dark:bg-[#1e293b] border-l-4 ${statusBorderColors[status]}`}
    >
      <div className="flex items-start justify-between">
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium text-slate-500 dark:text-slate-400">
            {label}
          </p>
          <p className="mt-1 font-mono text-2xl font-bold text-slate-900 dark:text-white">
            {value}
          </p>
          {subtitle && (
            <p className="mt-0.5 text-xs text-slate-400 dark:text-slate-500">
              {subtitle}
            </p>
          )}
        </div>
        <div className="ml-3 flex flex-col items-end gap-2">
          {icon && (
            <div className="text-slate-400 dark:text-slate-500">{icon}</div>
          )}
          {trend !== undefined && trend !== 0 && (
            <span
              className={`inline-flex items-center gap-0.5 text-xs font-mono font-medium ${
                trend > 0 ? 'text-[#10b981]' : 'text-[#ef4444]'
              }`}
            >
              {trend > 0 ? (
                <TrendingUp className="h-3 w-3" />
              ) : (
                <TrendingDown className="h-3 w-3" />
              )}
              {trend > 0 ? '+' : ''}
              {trend}
            </span>
          )}
        </div>
      </div>
    </div>
  );
};

export default MetricCard;
