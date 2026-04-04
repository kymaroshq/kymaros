import { useMemo, type FC } from 'react';

interface HeatmapDay {
  date: string;
  score: number;
  tests: number;
  failures: number;
}

interface HeatmapCalendarProps {
  data: HeatmapDay[];
  days?: number;
}

function getCellColor(day: HeatmapDay | undefined): string {
  if (!day) return 'bg-slate-100 dark:bg-slate-800';
  if (day.failures > 0) return 'bg-[#ef4444]';
  if (day.score >= 80) return 'bg-[#10b981]';
  if (day.score >= 50) return 'bg-[#f59e0b]';
  return 'bg-[#ef4444]';
}

function getCellOpacity(day: HeatmapDay | undefined): string {
  if (!day) return 'opacity-30';
  if (day.tests === 0) return 'opacity-20';
  return 'opacity-100';
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

const HeatmapCalendar: FC<HeatmapCalendarProps> = ({ data, days = 90 }) => {
  const grid = useMemo(() => {
    const dataMap = new Map(data.map((d) => [d.date, d]));
    const cells: (HeatmapDay | undefined)[] = [];
    const today = new Date();

    for (let i = days - 1; i >= 0; i--) {
      const date = new Date(today);
      date.setDate(today.getDate() - i);
      const key = date.toISOString().split('T')[0];
      cells.push(dataMap.get(key));
    }

    return cells;
  }, [data, days]);

  // Organize into columns of 7 rows (weeks)
  const weeks = useMemo(() => {
    const result: (HeatmapDay | undefined)[][] = [];
    for (let i = 0; i < grid.length; i += 7) {
      result.push(grid.slice(i, i + 7));
    }
    return result;
  }, [grid]);

  return (
    <div className="inline-block">
      <div className="flex gap-0.5">
        {weeks.map((week, weekIndex) => (
          <div key={weekIndex} className="flex flex-col gap-0.5">
            {week.map((day, dayIndex) => {
              const cellKey = day?.date ?? `empty-${weekIndex}-${dayIndex}`;
              return (
                <div
                  key={cellKey}
                  className={`h-3 w-3 rounded-[2px] transition-colors ${getCellColor(day)} ${getCellOpacity(day)}`}
                  title={
                    day
                      ? `${formatDate(day.date)}: Score ${day.score}, ${day.tests} tests, ${day.failures} failures`
                      : 'No data'
                  }
                />
              );
            })}
          </div>
        ))}
      </div>
      {/* Legend */}
      <div className="mt-2 flex items-center gap-1.5 text-xs text-slate-500 dark:text-slate-400">
        <span>Less</span>
        <div className="h-3 w-3 rounded-[2px] bg-slate-100 opacity-30 dark:bg-slate-800" />
        <div className="h-3 w-3 rounded-[2px] bg-[#ef4444]" />
        <div className="h-3 w-3 rounded-[2px] bg-[#f59e0b]" />
        <div className="h-3 w-3 rounded-[2px] bg-[#10b981]" />
        <span>More</span>
      </div>
    </div>
  );
};

export default HeatmapCalendar;
