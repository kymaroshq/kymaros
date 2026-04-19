import type { ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface CardProps {
  children: ReactNode;
  className?: string;
  interactive?: boolean;
  onClick?: () => void;
}

export function Card({ children, className, interactive, onClick }: CardProps) {
  return (
    <div
      className={cn(
        'bg-surface-2 border border-border-default rounded-lg overflow-hidden',
        interactive && 'hover:bg-surface-3 hover:border-border-strong transition-colors duration-150 cursor-pointer',
        className
      )}
      onClick={onClick}
      role={onClick ? 'button' : undefined}
      tabIndex={onClick ? 0 : undefined}
      onKeyDown={onClick ? (e) => { if (e.key === 'Enter') onClick(); } : undefined}
    >
      {children}
    </div>
  );
}

export function CardHeader({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div className={cn('px-4 py-3 border-b border-border-subtle flex items-center justify-between', className)}>
      {children}
    </div>
  );
}

export function CardTitle({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <h3 className={cn('text-sm font-medium text-text-primary', className)}>
      {children}
    </h3>
  );
}

export function CardBody({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div className={cn('px-4 py-3', className)}>
      {children}
    </div>
  );
}
