import type { ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface Column<T> {
  key: string;
  label: string;
  render?: (row: T) => ReactNode;
  align?: 'left' | 'right' | 'center';
  width?: string;
  mono?: boolean;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  rows: T[];
  onRowClick?: (row: T) => void;
  empty?: ReactNode;
  compact?: boolean;
  keyField?: keyof T;
}

export function DataTable<T>({
  columns,
  rows,
  onRowClick,
  empty,
  compact,
  keyField,
}: DataTableProps<T>) {
  if (rows.length === 0 && empty) {
    return <div className="py-8 text-center">{empty}</div>;
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-border-subtle">
            {columns.map((col) => (
              <th
                key={col.key}
                className={cn(
                  'text-2xs font-medium text-text-tertiary uppercase tracking-wider',
                  compact ? 'px-3 py-2' : 'px-4 py-3',
                  col.align === 'right' && 'text-right',
                  col.align === 'center' && 'text-center',
                  !col.align && 'text-left'
                )}
                style={col.width ? { width: col.width } : undefined}
              >
                {col.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr
              key={keyField ? String(row[keyField]) : i}
              onClick={onRowClick ? () => onRowClick(row) : undefined}
              className={cn(
                'border-b border-border-subtle last:border-b-0',
                onRowClick && 'hover:bg-surface-3 cursor-pointer transition-colors duration-150'
              )}
            >
              {columns.map((col) => (
                <td
                  key={col.key}
                  className={cn(
                    'text-sm text-text-primary',
                    compact ? 'px-3 py-2' : 'px-4 py-3',
                    col.align === 'right' && 'text-right',
                    col.align === 'center' && 'text-center',
                    col.mono && 'text-data'
                  )}
                >
                  {col.render ? col.render(row) : String((row as Record<string, unknown>)[col.key] ?? '')}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
