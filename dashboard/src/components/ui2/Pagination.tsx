import { cn } from '@/lib/utils';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface PaginationProps {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

function getPageNumbers(current: number, total: number): (number | 'gap')[] {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);
  if (current <= 4) return [1, 2, 3, 4, 5, 'gap', total];
  if (current >= total - 3) return [1, 'gap', total - 4, total - 3, total - 2, total - 1, total];
  return [1, 'gap', current - 1, current, current + 1, 'gap', total];
}

export function Pagination({ page, totalPages, onPageChange }: PaginationProps) {
  if (totalPages <= 1) return null;
  const pages = getPageNumbers(page, totalPages);

  return (
    <nav className="flex items-center gap-1" aria-label="Pagination">
      <button
        className="h-7 w-7 flex items-center justify-center rounded-md text-text-tertiary hover:bg-surface-3 hover:text-text-primary transition-colors duration-150 disabled:opacity-40"
        disabled={page <= 1}
        onClick={() => onPageChange(page - 1)}
      >
        <ChevronLeft className="h-3.5 w-3.5" />
      </button>
      {pages.map((p, i) =>
        p === 'gap' ? (
          <span key={`gap-${i}`} className="px-1 text-text-tertiary text-xs">&hellip;</span>
        ) : (
          <button
            key={p}
            onClick={() => onPageChange(p)}
            className={cn(
              'h-7 min-w-7 px-2 flex items-center justify-center rounded-md text-xs font-mono transition-colors duration-150',
              p === page ? 'bg-surface-3 text-text-primary' : 'text-text-tertiary hover:bg-surface-3 hover:text-text-primary'
            )}
          >
            {p}
          </button>
        )
      )}
      <button
        className="h-7 w-7 flex items-center justify-center rounded-md text-text-tertiary hover:bg-surface-3 hover:text-text-primary transition-colors duration-150 disabled:opacity-40"
        disabled={page >= totalPages}
        onClick={() => onPageChange(page + 1)}
      >
        <ChevronRight className="h-3.5 w-3.5" />
      </button>
    </nav>
  );
}
