import { useState, useMemo, useCallback, type ReactNode } from 'react';
import { Search, ChevronUp, ChevronDown } from 'lucide-react';

interface Column<T> {
  key: string;
  label: string;
  render?: (row: T) => ReactNode;
  sortable?: boolean;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  defaultSort?: string;
  onRowClick?: (row: T) => void;
  searchable?: boolean;
  filters?: ReactNode;
  emptyMessage?: string;
}

type SortDirection = 'asc' | 'desc';

function getNestedValue<T>(obj: T, key: string): unknown {
  return key.split('.').reduce<unknown>((acc, part) => {
    if (acc !== null && acc !== undefined && typeof acc === 'object') {
      return (acc as Record<string, unknown>)[part];
    }
    return undefined;
  }, obj);
}

function DataTable<T extends Record<string, unknown>>({
  columns,
  data,
  defaultSort,
  onRowClick,
  searchable = false,
  filters,
  emptyMessage = 'No data available',
}: DataTableProps<T>): ReactNode {
  const [searchQuery, setSearchQuery] = useState('');
  const [sortKey, setSortKey] = useState<string | undefined>(defaultSort);
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');

  const handleSort = useCallback(
    (key: string) => {
      if (sortKey === key) {
        setSortDirection((prev) => (prev === 'asc' ? 'desc' : 'asc'));
      } else {
        setSortKey(key);
        setSortDirection('asc');
      }
    },
    [sortKey],
  );

  const filteredData = useMemo(() => {
    if (!searchQuery.trim()) return data;
    const query = searchQuery.toLowerCase();
    return data.filter((row) =>
      columns.some((col) => {
        const value = getNestedValue(row, col.key);
        return value !== undefined && value !== null && String(value).toLowerCase().includes(query);
      }),
    );
  }, [data, searchQuery, columns]);

  const sortedData = useMemo(() => {
    if (!sortKey) return filteredData;
    return [...filteredData].sort((a, b) => {
      const aVal = getNestedValue(a, sortKey);
      const bVal = getNestedValue(b, sortKey);

      if (aVal === bVal) return 0;
      if (aVal === null || aVal === undefined) return 1;
      if (bVal === null || bVal === undefined) return -1;

      let comparison = 0;
      if (typeof aVal === 'number' && typeof bVal === 'number') {
        comparison = aVal - bVal;
      } else {
        comparison = String(aVal).localeCompare(String(bVal));
      }

      return sortDirection === 'desc' ? -comparison : comparison;
    });
  }, [filteredData, sortKey, sortDirection]);

  return (
    <div className="overflow-hidden rounded-lg border border-slate-200 bg-white dark:border-slate-700 dark:bg-[#1e293b]">
      {/* Search + filters bar */}
      {(searchable || filters) && (
        <div className="flex flex-wrap items-center gap-3 border-b border-slate-200 p-3 dark:border-slate-700">
          {searchable && (
            <div className="relative flex-1">
              <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <input
                type="text"
                placeholder="Search..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full rounded-md border border-slate-300 bg-transparent py-1.5 pl-8 pr-3 text-sm text-slate-900 placeholder:text-slate-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-slate-600 dark:text-white dark:placeholder:text-slate-500"
              />
            </div>
          )}
          {filters}
        </div>
      )}

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b border-slate-200 bg-slate-50 dark:border-slate-700 dark:bg-slate-800/50">
              {columns.map((col) => (
                <th
                  key={col.key}
                  className={`px-4 py-2.5 text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 ${
                    col.sortable !== false ? 'cursor-pointer select-none hover:text-slate-700 dark:hover:text-slate-200' : ''
                  }`}
                  onClick={() => col.sortable !== false && handleSort(col.key)}
                >
                  <span className="inline-flex items-center gap-1">
                    {col.label}
                    {col.sortable !== false && sortKey === col.key && (
                      sortDirection === 'asc' ? (
                        <ChevronUp className="h-3.5 w-3.5" />
                      ) : (
                        <ChevronDown className="h-3.5 w-3.5" />
                      )
                    )}
                  </span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {sortedData.length === 0 ? (
              <tr>
                <td
                  colSpan={columns.length}
                  className="px-4 py-8 text-center text-sm text-slate-500 dark:text-slate-400"
                >
                  {emptyMessage}
                </td>
              </tr>
            ) : (
              sortedData.map((row, rowIndex) => (
                <tr
                  key={rowIndex}
                  className={`border-b border-slate-100 transition-colors last:border-b-0 dark:border-slate-700/50 ${
                    rowIndex % 2 === 1
                      ? 'bg-slate-50/50 dark:bg-slate-800/20'
                      : ''
                  } ${
                    onRowClick
                      ? 'cursor-pointer hover:bg-blue-50 dark:hover:bg-blue-900/10'
                      : 'hover:bg-slate-50 dark:hover:bg-slate-800/40'
                  }`}
                  onClick={() => onRowClick?.(row)}
                >
                  {columns.map((col) => (
                    <td
                      key={col.key}
                      className="px-4 py-2.5 text-slate-700 dark:text-slate-300"
                    >
                      {col.render
                        ? col.render(row)
                        : (String(getNestedValue(row, col.key) ?? ''))}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// Export the component with explicit typing for the default export
export default DataTable as <T extends Record<string, unknown>>(
  props: DataTableProps<T>,
) => ReactNode;

// Re-export the Column type for consumers
export type { Column, DataTableProps };
