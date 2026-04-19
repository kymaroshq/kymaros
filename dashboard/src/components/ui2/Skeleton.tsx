import { cn } from '@/lib/utils';

export function Skeleton({ className }: { className?: string }) {
  return (
    <div className={cn('bg-surface-3 rounded animate-pulse-subtle', className)} />
  );
}

export function StatSkeleton() {
  return (
    <div className="flex flex-col gap-1">
      <Skeleton className="h-3 w-16" />
      <Skeleton className="h-8 w-20 mt-1" />
      <Skeleton className="h-3 w-24 mt-1" />
    </div>
  );
}

export function TableSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="space-y-0">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex items-center gap-4 px-4 py-3 border-b border-border-subtle last:border-b-0">
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-4 w-12" />
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-4 w-20 ml-auto" />
        </div>
      ))}
    </div>
  );
}
