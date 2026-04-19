import { cn } from '@/lib/utils';
import { getScoreTone } from '@/lib/utils';

interface ScoreIndicatorProps {
  value: number;
  size?: 'sm' | 'md' | 'lg';
  showLabel?: boolean;
}

const toneClasses = {
  excellent: 'text-score-excellent',
  good: 'text-score-good',
  warning: 'text-score-warning',
  critical: 'text-score-critical',
};

const sizeClasses = {
  sm: 'text-base',
  md: 'text-xl',
  lg: 'text-3xl',
};

export function ScoreIndicator({ value, size = 'md', showLabel }: ScoreIndicatorProps) {
  const tone = getScoreTone(value);
  return (
    <span className="inline-flex items-baseline gap-1">
      <span className={cn('text-metric-large', sizeClasses[size], toneClasses[tone])}>
        {Math.round(value)}
      </span>
      {showLabel && (
        <span className="text-2xs text-text-tertiary font-mono">/100</span>
      )}
    </span>
  );
}
