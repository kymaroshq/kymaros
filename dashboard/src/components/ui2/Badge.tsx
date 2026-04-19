import type { ReactNode } from 'react';
import { cn } from '@/lib/utils';

type Variant = 'success' | 'warning' | 'danger' | 'info' | 'neutral';

interface BadgeProps {
  variant?: Variant;
  size?: 'sm' | 'md';
  dot?: boolean;
  children: ReactNode;
}

const variantClasses: Record<Variant, string> = {
  success: 'bg-status-success/10 text-status-success border-status-success/20',
  warning: 'bg-status-warning/10 text-status-warning border-status-warning/20',
  danger: 'bg-status-danger/10 text-status-danger border-status-danger/20',
  info: 'bg-status-info/10 text-status-info border-status-info/20',
  neutral: 'bg-surface-3 text-text-secondary border-border-default',
};

const dotVariantClasses: Record<Variant, string> = {
  success: 'bg-status-success',
  warning: 'bg-status-warning',
  danger: 'bg-status-danger',
  info: 'bg-status-info',
  neutral: 'bg-status-neutral',
};

export function Badge({ variant = 'neutral', size = 'sm', dot, children }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 border rounded font-medium',
        size === 'sm' ? 'px-1.5 py-0.5 text-2xs' : 'px-2 py-1 text-xs',
        variantClasses[variant]
      )}
    >
      {dot && (
        <span className={cn('h-1.5 w-1.5 rounded-full', dotVariantClasses[variant])} />
      )}
      {children}
    </span>
  );
}
