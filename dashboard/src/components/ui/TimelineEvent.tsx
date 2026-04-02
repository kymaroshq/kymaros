import type { FC } from 'react';

type TimelineStatus = 'success' | 'error' | 'neutral';

interface TimelineEventProps {
  time: string;
  status: TimelineStatus;
  title: string;
  duration?: string;
  isLast?: boolean;
}

const dotColors: Record<TimelineStatus, string> = {
  success: 'bg-[#10b981]',
  error: 'bg-[#ef4444]',
  neutral: 'bg-slate-400',
};

const ringColors: Record<TimelineStatus, string> = {
  success: 'ring-[#10b981]/20',
  error: 'ring-[#ef4444]/20',
  neutral: 'ring-slate-400/20',
};

const TimelineEvent: FC<TimelineEventProps> = ({
  time,
  status,
  title,
  duration,
  isLast = false,
}) => {
  return (
    <div className="flex gap-3">
      {/* Dot + vertical line column */}
      <div className="flex flex-col items-center">
        <div
          className={`mt-1 h-3 w-3 shrink-0 rounded-full ring-4 ${dotColors[status]} ${ringColors[status]}`}
        />
        {!isLast && (
          <div className="w-px flex-1 bg-slate-200 dark:bg-slate-700" />
        )}
      </div>

      {/* Content */}
      <div className={`pb-6 ${isLast ? 'pb-0' : ''}`}>
        <p className="text-sm font-medium text-slate-900 dark:text-white">
          {title}
        </p>
        <div className="mt-0.5 flex items-center gap-2">
          <span className="font-mono text-xs text-slate-500 dark:text-slate-400">
            {time}
          </span>
          {duration && (
            <>
              <span className="text-slate-300 dark:text-slate-600">&middot;</span>
              <span className="font-mono text-xs text-slate-400 dark:text-slate-500">
                {duration}
              </span>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default TimelineEvent;
