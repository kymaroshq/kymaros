import { X, Bell, CheckCheck, Trash2 } from 'lucide-react';
import { Link } from 'react-router-dom';
import { Button } from '@/components/ui2/Button';
import { Badge } from '@/components/ui2/Badge';
import { EmptyState } from '@/components/ui2/EmptyState';
import { useNotifications } from '@/hooks/useNotifications';
import { formatRelativeTime, cn } from '@/lib/utils';

const typeVariant = { success: 'success', warning: 'warning', error: 'danger', info: 'info' } as const;

interface Props {
  open: boolean;
  onClose: () => void;
}

export function NotificationsDrawer({ open, onClose }: Props) {
  const { notifications, markAllRead, clear } = useNotifications();

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50">
      <div className="absolute inset-0 bg-black/50 animate-fade-in" onClick={onClose} />
      <div className="absolute top-0 right-0 bottom-0 w-96 bg-surface-1 border-l border-border-default shadow-lg flex flex-col animate-slide-up">
        <header className="h-14 px-4 border-b border-border-subtle flex items-center justify-between shrink-0">
          <span className="text-base font-semibold text-text-primary">Notifications</span>
          <button onClick={onClose} className="h-7 w-7 flex items-center justify-center rounded-md text-text-tertiary hover:bg-surface-3 hover:text-text-primary transition-colors duration-150">
            <X className="h-3.5 w-3.5" />
          </button>
        </header>

        {notifications.length === 0 ? (
          <EmptyState icon={<Bell className="h-6 w-6" />} title="No notifications" description="Activity will appear here." />
        ) : (
          <>
            <div className="flex items-center justify-between px-4 py-2 border-b border-border-subtle shrink-0">
              <span className="text-2xs font-mono text-text-tertiary">{notifications.length}</span>
              <div className="flex gap-1">
                <Button variant="ghost" size="sm" onClick={markAllRead}><CheckCheck className="h-3 w-3" /> Read all</Button>
                <Button variant="ghost" size="sm" onClick={clear}><Trash2 className="h-3 w-3" /> Clear</Button>
              </div>
            </div>
            <div className="flex-1 overflow-y-auto">
              {notifications.map((n) => {
                const content = (
                  <div className={cn('px-4 py-3 border-b border-border-subtle flex items-start gap-3 hover:bg-surface-3 transition-colors duration-150', !n.read && 'bg-surface-2')}>
                    <Badge variant={typeVariant[n.type]} dot size="sm">{n.type}</Badge>
                    <div className="flex-1 min-w-0">
                      <div className="text-sm font-medium text-text-primary">{n.title}</div>
                      {n.description && <div className="text-xs text-text-tertiary mt-0.5 line-clamp-2">{n.description}</div>}
                      <div className="text-2xs text-text-tertiary font-mono mt-1">{formatRelativeTime(n.timestamp)}</div>
                    </div>
                  </div>
                );
                return n.href ? <Link key={n.id} to={n.href} onClick={onClose}>{content}</Link> : <div key={n.id}>{content}</div>;
              })}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
