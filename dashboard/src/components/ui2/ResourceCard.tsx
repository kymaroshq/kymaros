import { cn } from '@/lib/utils';
import type { ComponentType } from 'react';

interface ResourceCardProps {
  label: string;
  present: number;
  expected: number;
  icon?: ComponentType<{ className?: string }>;
}

export function ResourceCard({ label, present, expected, icon: Icon }: ResourceCardProps) {
  const isComplete = present === expected && expected > 0;
  const isEmpty = expected === 0;

  return (
    <div
      className={cn(
        'border rounded-lg px-3 py-3 flex flex-col gap-1',
        isEmpty
          ? 'bg-surface-2 border-border-subtle opacity-50'
          : isComplete
            ? 'bg-status-success/5 border-status-success/20'
            : 'bg-status-warning/5 border-status-warning/20'
      )}
    >
      <div className="flex items-center gap-2">
        {Icon && (
          <Icon
            className={cn(
              'h-3.5 w-3.5',
              isEmpty ? 'text-text-disabled' : isComplete ? 'text-status-success' : 'text-status-warning'
            )}
          />
        )}
        <span className="text-2xs font-medium uppercase tracking-wider text-text-tertiary truncate">
          {label}
        </span>
      </div>
      <div className="flex items-baseline gap-1">
        {isEmpty ? (
          <span className="text-metric text-lg text-text-disabled">&mdash;</span>
        ) : (
          <>
            <span
              className={cn(
                'text-metric text-lg',
                isComplete ? 'text-status-success' : 'text-status-warning'
              )}
            >
              {present}
            </span>
            <span className="text-2xs text-text-tertiary font-mono">/{expected}</span>
          </>
        )}
      </div>
    </div>
  );
}
