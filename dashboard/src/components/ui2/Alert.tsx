import type { ReactNode } from 'react';
import { cn } from '@/lib/utils';
import { AlertCircle, CheckCircle, Info, AlertTriangle } from 'lucide-react';

type Variant = 'info' | 'success' | 'warning' | 'danger';

interface AlertProps {
  variant?: Variant;
  title?: string;
  children?: ReactNode;
  className?: string;
}

const config: Record<Variant, { bg: string; Icon: typeof Info }> = {
  info: { bg: 'bg-status-info/5 border-status-info/20 text-status-info', Icon: Info },
  success: { bg: 'bg-status-success/5 border-status-success/20 text-status-success', Icon: CheckCircle },
  warning: { bg: 'bg-status-warning/5 border-status-warning/20 text-status-warning', Icon: AlertTriangle },
  danger: { bg: 'bg-status-danger/5 border-status-danger/20 text-status-danger', Icon: AlertCircle },
};

export function Alert({ variant = 'info', title, children, className }: AlertProps) {
  const { bg, Icon } = config[variant];
  return (
    <div className={cn('flex items-start gap-2.5 rounded-md border px-3 py-2.5', bg, className)}>
      <Icon className="h-4 w-4 shrink-0 mt-0.5" />
      <div className="flex-1 min-w-0">
        {title && <div className="text-sm font-medium">{title}</div>}
        {children && <div className="text-xs text-text-secondary mt-0.5">{children}</div>}
      </div>
    </div>
  );
}
