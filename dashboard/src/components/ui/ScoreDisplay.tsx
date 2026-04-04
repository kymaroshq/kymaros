import type { FC } from 'react';
import { TrendingUp, TrendingDown } from 'lucide-react';

interface ScoreDisplayProps {
  score: number;
  size?: 'sm' | 'md' | 'lg' | 'xl';
  trend?: number;
  showBar?: boolean;
}

const sizeClasses = {
  sm: 'text-lg',
  md: 'text-2xl',
  lg: 'text-4xl',
  xl: 'text-6xl',
} as const;

const trendSizeClasses = {
  sm: 'text-xs',
  md: 'text-sm',
  lg: 'text-base',
  xl: 'text-lg',
} as const;

const barHeightClasses = {
  sm: 'h-1',
  md: 'h-1.5',
  lg: 'h-2',
  xl: 'h-2.5',
} as const;

function getScoreColor(score: number): string {
  if (score >= 80) return 'text-[#10b981]';
  if (score >= 50) return 'text-[#f59e0b]';
  return 'text-[#ef4444]';
}

function getBarBgColor(score: number): string {
  if (score >= 80) return 'bg-[#10b981]';
  if (score >= 50) return 'bg-[#f59e0b]';
  return 'bg-[#ef4444]';
}

const ScoreDisplay: FC<ScoreDisplayProps> = ({
  score,
  size = 'md',
  trend,
  showBar = false,
}) => {
  const colorClass = getScoreColor(score);
  const clampedScore = Math.max(0, Math.min(100, score));

  return (
    <div className="inline-flex flex-col items-center gap-1">
      <div className="flex items-baseline gap-1.5">
        <span
          className={`font-mono font-bold leading-none ${sizeClasses[size]} ${colorClass}`}
        >
          {score}
        </span>
        {trend !== undefined && trend !== 0 && (
          <span
            className={`inline-flex items-center gap-0.5 font-mono ${trendSizeClasses[size]} ${
              trend > 0 ? 'text-[#10b981]' : 'text-[#ef4444]'
            }`}
          >
            {trend > 0 ? (
              <TrendingUp className="h-3.5 w-3.5" />
            ) : (
              <TrendingDown className="h-3.5 w-3.5" />
            )}
            {trend > 0 ? '+' : ''}
            {trend}
          </span>
        )}
      </div>
      {showBar && (
        <div
          className={`w-full rounded-full bg-slate-200 dark:bg-slate-700 ${barHeightClasses[size]}`}
        >
          <div
            className={`${barHeightClasses[size]} rounded-full transition-all duration-300 ${getBarBgColor(score)}`}
            style={{ width: `${clampedScore}%` }}
          />
        </div>
      )}
    </div>
  );
};

export default ScoreDisplay;
