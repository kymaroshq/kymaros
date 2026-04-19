import { ChevronRight } from 'lucide-react';
import { Link } from 'react-router-dom';
import { Badge } from '@/components/ui2/Badge';
import { ScoreIndicator } from '@/components/ui2/ScoreIndicator';

interface ReportHeaderProps {
  testName: string;
  namespace: string;
  score: number;
  status: string;
}

export function ReportHeader({ testName, namespace, score, status }: ReportHeaderProps) {
  const variant = status === 'pass' ? 'success' as const : status === 'partial' ? 'warning' as const : 'danger' as const;

  return (
    <div className="space-y-3">
      <nav className="flex items-center gap-1 text-xs text-text-tertiary">
        <Link to="/" className="hover:text-text-primary transition-colors duration-150">Dashboard</Link>
        <ChevronRight className="h-3 w-3" />
        <span className="text-text-secondary">{testName}</span>
      </nav>

      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-semibold text-text-primary truncate">{testName}</h1>
            <ScoreIndicator value={score} size="lg" showLabel />
          </div>
          <div className="flex items-center gap-3 mt-1.5">
            <span className="text-sm text-text-tertiary font-mono">{namespace}</span>
            <span className="text-text-tertiary">&middot;</span>
            <Badge variant={variant} dot size="md">{status}</Badge>
          </div>
        </div>

      </div>
    </div>
  );
}
