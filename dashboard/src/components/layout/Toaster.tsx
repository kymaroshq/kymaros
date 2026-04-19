import { useState, useEffect } from 'react';
import { toast, type ToastItem } from '@/lib/toast';
import { CheckCircle, AlertCircle, AlertTriangle, Info } from 'lucide-react';
import { cn } from '@/lib/utils';

const icons = {
  success: CheckCircle,
  error: AlertCircle,
  warning: AlertTriangle,
  info: Info,
};

const colors = {
  success: 'text-status-success border-status-success/20',
  error: 'text-status-danger border-status-danger/20',
  warning: 'text-status-warning border-status-warning/20',
  info: 'text-status-info border-status-info/20',
};

export function Toaster() {
  const [items, setItems] = useState<ToastItem[]>([]);

  useEffect(() => toast.subscribe(setItems), []);

  if (items.length === 0) return null;

  return (
    <div className="fixed bottom-4 right-4 z-[100] space-y-2 w-80">
      {items.map((t) => {
        const Icon = icons[t.type];
        return (
          <div
            key={t.id}
            className={cn(
              'bg-surface-3 border rounded-md shadow-lg px-3 py-2.5 flex items-start gap-2.5 animate-slide-up',
              colors[t.type]
            )}
          >
            <Icon className="h-4 w-4 shrink-0 mt-0.5" />
            <div className="flex-1 min-w-0">
              <div className="text-sm font-medium text-text-primary">{t.message}</div>
              {t.description && <div className="text-xs text-text-tertiary mt-0.5">{t.description}</div>}
            </div>
          </div>
        );
      })}
    </div>
  );
}
