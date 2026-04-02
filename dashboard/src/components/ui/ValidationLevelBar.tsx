import { useState, type FC } from 'react';
import {
  CheckCircle,
  XCircle,
  AlertTriangle,
  MinusCircle,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';
import type { ReactNode } from 'react';

type BarStatus = 'pass' | 'fail' | 'partial' | 'not_tested';

interface ValidationLevelBarProps {
  name: string;
  status: BarStatus;
  percentage: number;
  detail: string;
  expandable?: boolean;
  children?: ReactNode;
}

const statusConfig: Record<
  BarStatus,
  { Icon: FC<{ className?: string }>; barColor: string; iconColor: string }
> = {
  pass: {
    Icon: CheckCircle,
    barColor: 'bg-[#10b981]',
    iconColor: 'text-[#10b981]',
  },
  fail: {
    Icon: XCircle,
    barColor: 'bg-[#ef4444]',
    iconColor: 'text-[#ef4444]',
  },
  partial: {
    Icon: AlertTriangle,
    barColor: 'bg-[#f59e0b]',
    iconColor: 'text-[#f59e0b]',
  },
  not_tested: {
    Icon: MinusCircle,
    barColor: 'bg-slate-400',
    iconColor: 'text-slate-400',
  },
};

const ValidationLevelBar: FC<ValidationLevelBarProps> = ({
  name,
  status,
  percentage,
  detail,
  expandable = false,
  children,
}) => {
  const [expanded, setExpanded] = useState(false);
  const config = statusConfig[status];
  const { Icon } = config;
  const clampedPercentage = Math.max(0, Math.min(100, percentage));

  return (
    <div className="rounded-lg border border-slate-200 bg-white dark:border-slate-700 dark:bg-[#1e293b]">
      <button
        type="button"
        className={`flex w-full items-center gap-3 p-3 text-left ${
          expandable ? 'cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-700/50' : 'cursor-default'
        }`}
        onClick={() => expandable && setExpanded(!expanded)}
        disabled={!expandable}
      >
        <Icon className={`h-5 w-5 shrink-0 ${config.iconColor}`} />
        <div className="min-w-0 flex-1">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-slate-900 dark:text-white">
              {name}
            </span>
            <span className="ml-2 font-mono text-sm font-semibold text-slate-700 dark:text-slate-300">
              {percentage}%
            </span>
          </div>
          <div className="mt-1.5 h-1.5 w-full rounded-full bg-slate-200 dark:bg-slate-600">
            <div
              className={`h-1.5 rounded-full transition-all duration-300 ${config.barColor}`}
              style={{ width: `${clampedPercentage}%` }}
            />
          </div>
          <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">
            {detail}
          </p>
        </div>
        {expandable && (
          <div className="shrink-0 text-slate-400">
            {expanded ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </div>
        )}
      </button>
      {expandable && expanded && children && (
        <div className="border-t border-slate-200 p-3 dark:border-slate-700">
          {children}
        </div>
      )}
    </div>
  );
};

export default ValidationLevelBar;
