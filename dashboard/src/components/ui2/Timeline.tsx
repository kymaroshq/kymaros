import { cn } from '@/lib/utils';

export interface TimelineItem {
  id: string;
  title: string;
  timestamp: string;
  description?: string;
  status?: 'completed' | 'active' | 'pending' | 'failed';
}

const statusClasses = {
  completed: 'bg-status-success border-status-success',
  active: 'bg-accent border-accent animate-pulse-subtle',
  pending: 'bg-surface-3 border-border-default',
  failed: 'bg-status-danger border-status-danger',
};

export function Timeline({ items }: { items: TimelineItem[] }) {
  return (
    <ol className="space-y-4">
      {items.map((item, index) => (
        <li key={item.id} className="relative flex gap-3">
          {index < items.length - 1 && (
            <div aria-hidden className="absolute left-[5px] top-3 bottom-[-16px] w-px bg-border-subtle" />
          )}
          <div className={cn('mt-1 h-2.5 w-2.5 rounded-full border-2 shrink-0', statusClasses[item.status ?? 'completed'])} />
          <div className="min-w-0 flex-1">
            <div className="text-sm font-medium text-text-primary">{item.title}</div>
            <div className="text-2xs text-text-tertiary font-mono mt-0.5">
              {new Date(item.timestamp).toLocaleTimeString('en-US', { hour12: false })}
            </div>
            {item.description && (
              <div className="text-xs text-text-secondary mt-1">{item.description}</div>
            )}
          </div>
        </li>
      ))}
    </ol>
  );
}
