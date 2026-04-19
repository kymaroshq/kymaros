import { cn } from '@/lib/utils';

interface ProgressBarProps {
  value: number;
  max: number;
  tone?: 'success' | 'warning' | 'danger' | 'default';
  size?: 'sm' | 'md';
  showLabel?: boolean;
}

const toneClasses = {
  success: 'bg-status-success',
  warning: 'bg-status-warning',
  danger: 'bg-status-danger',
  default: 'bg-accent',
};

export function ProgressBar({ value, max, tone = 'default', size = 'md', showLabel }: ProgressBarProps) {
  const percent = max > 0 ? Math.min(100, (value / max) * 100) : 0;
  const height = size === 'sm' ? 'h-1' : 'h-1.5';

  return (
    <div className="flex items-center gap-3">
      <div className={cn('flex-1 bg-surface-4 rounded-full overflow-hidden', height)}>
        <div
          className={cn('h-full rounded-full transition-all duration-150', toneClasses[tone])}
          style={{ width: `${percent}%` }}
        />
      </div>
      {showLabel && (
        <span className="text-2xs text-text-tertiary font-mono tabular-nums whitespace-nowrap">
          {value}/{max}
        </span>
      )}
    </div>
  );
}
