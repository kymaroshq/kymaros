import type { ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface StatProps {
  label: string;
  value: string | number;
  unit?: string;
  change?: { value: number; direction: 'up' | 'down' };
  icon?: ReactNode;
  sublabel?: string;
  tone?: 'default' | 'success' | 'warning' | 'danger';
  size?: 'sm' | 'md' | 'lg';
}

const toneClasses = {
  default: 'text-text-primary',
  success: 'text-score-excellent',
  warning: 'text-score-warning',
  danger: 'text-score-critical',
};

const sizeClasses = {
  sm: 'text-xl',
  md: 'text-2xl',
  lg: 'text-3xl',
};

export function Stat({ label, value, unit, change, icon, sublabel, tone = 'default', size = 'md' }: StatProps) {
  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center justify-between text-2xs font-medium text-text-tertiary uppercase tracking-wider">
        <span>{label}</span>
        {icon && <span className="text-text-tertiary">{icon}</span>}
      </div>
      <div className="flex items-baseline gap-2">
        <span className={cn('text-metric-large', sizeClasses[size], toneClasses[tone])}>
          {typeof value === 'number' ? Math.round(value) : value}
        </span>
        {unit && (
          <span className="text-sm text-text-secondary font-mono">{unit}</span>
        )}
        {change && (
          <span
            className={cn(
              'text-2xs font-mono font-medium px-1.5 py-0.5 rounded',
              change.direction === 'up' ? 'text-status-success bg-status-success/10' : 'text-status-danger bg-status-danger/10'
            )}
          >
            {change.direction === 'up' ? '\u2191' : '\u2193'} {Math.abs(change.value)}
          </span>
        )}
      </div>
      {sublabel && (
        <div className="text-2xs text-text-tertiary font-mono">{sublabel}</div>
      )}
    </div>
  );
}
