import type { FC } from 'react';
import {
  CheckCircle,
  XCircle,
  AlertTriangle,
  MinusCircle,
  Loader2,
} from 'lucide-react';

type Status = 'pass' | 'fail' | 'partial' | 'not_tested' | 'running';

interface StatusBadgeProps {
  status: Status;
  size?: 'sm' | 'md';
}

interface StatusConfig {
  label: string;
  bgClass: string;
  textClass: string;
  Icon: FC<{ className?: string }>;
}

const statusMap: Record<Status, StatusConfig> = {
  pass: {
    label: 'Pass',
    bgClass: 'bg-[#10b981]/10',
    textClass: 'text-[#10b981]',
    Icon: CheckCircle,
  },
  fail: {
    label: 'Fail',
    bgClass: 'bg-[#ef4444]/10',
    textClass: 'text-[#ef4444]',
    Icon: XCircle,
  },
  partial: {
    label: 'Partial',
    bgClass: 'bg-[#f59e0b]/10',
    textClass: 'text-[#f59e0b]',
    Icon: AlertTriangle,
  },
  not_tested: {
    label: 'Not Tested',
    bgClass: 'bg-slate-500/10',
    textClass: 'text-slate-500',
    Icon: MinusCircle,
  },
  running: {
    label: 'Running',
    bgClass: 'bg-blue-500/10',
    textClass: 'text-blue-500',
    Icon: Loader2,
  },
};

const sizeClasses = {
  sm: {
    pill: 'px-2 py-0.5 text-xs gap-1',
    icon: 'h-3 w-3',
  },
  md: {
    pill: 'px-2.5 py-1 text-sm gap-1.5',
    icon: 'h-4 w-4',
  },
} as const;

const StatusBadge: FC<StatusBadgeProps> = ({ status, size = 'md' }) => {
  const config = statusMap[status];
  const sizeConfig = sizeClasses[size];
  const { Icon } = config;

  return (
    <span
      className={`inline-flex items-center rounded-full font-medium ${config.bgClass} ${config.textClass} ${sizeConfig.pill}`}
    >
      <Icon
        className={`${sizeConfig.icon} ${status === 'running' ? 'animate-spin' : ''}`}
      />
      {config.label}
    </span>
  );
};

export default StatusBadge;
