import type { FC } from 'react';

interface SkeletonProps {
  className?: string;
}

const Skeleton: FC<SkeletonProps> = ({ className = 'h-4 w-full' }) => {
  return (
    <div
      className={`animate-pulse rounded bg-slate-200 dark:bg-slate-700 ${className}`}
      aria-hidden="true"
    />
  );
};

const SkeletonCard: FC = () => {
  return (
    <div className="rounded-lg border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-[#1e293b]">
      <Skeleton className="mb-3 h-3 w-24" />
      <Skeleton className="mb-2 h-7 w-20" />
      <Skeleton className="h-3 w-32" />
    </div>
  );
};

interface SkeletonTableProps {
  rows?: number;
  columns?: number;
}

const SkeletonTable: FC<SkeletonTableProps> = ({ rows = 5, columns = 4 }) => {
  return (
    <div className="overflow-hidden rounded-lg border border-slate-200 bg-white dark:border-slate-700 dark:bg-[#1e293b]">
      {/* Header */}
      <div className="flex gap-4 border-b border-slate-200 bg-slate-50 px-4 py-3 dark:border-slate-700 dark:bg-slate-800/50">
        {Array.from({ length: columns }).map((_, i) => (
          <Skeleton key={`header-${i}`} className="h-3 flex-1" />
        ))}
      </div>
      {/* Rows */}
      {Array.from({ length: rows }).map((_, rowIndex) => (
        <div
          key={`row-${rowIndex}`}
          className={`flex gap-4 border-b border-slate-100 px-4 py-3 last:border-b-0 dark:border-slate-700/50 ${
            rowIndex % 2 === 1 ? 'bg-slate-50/50 dark:bg-slate-800/20' : ''
          }`}
        >
          {Array.from({ length: columns }).map((_, colIndex) => (
            <Skeleton key={`cell-${rowIndex}-${colIndex}`} className="h-4 flex-1" />
          ))}
        </div>
      ))}
    </div>
  );
};

export default Skeleton;
export { SkeletonCard, SkeletonTable };
