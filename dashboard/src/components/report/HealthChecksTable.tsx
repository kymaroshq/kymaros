import { Card, CardHeader, CardTitle } from '@/components/ui2/Card';
import { DataTable } from '@/components/ui2/DataTable';
import { Badge } from '@/components/ui2/Badge';
import { EmptyState } from '@/components/ui2/EmptyState';
import { CheckCircle } from 'lucide-react';

interface Check {
  name: string;
  status: string;
  duration?: string;
  message?: string;
}

export function HealthChecksTable({ checks }: { checks: Check[] }) {
  const failed = checks.filter((c) => c.status === 'fail').length;

  return (
    <Card className="flex-1">
      <CardHeader>
        <CardTitle>Health Checks</CardTitle>
        {checks.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-2xs text-text-tertiary font-mono">{checks.length} total</span>
            {failed > 0 && <Badge variant="danger" size="sm">{failed} failed</Badge>}
          </div>
        )}
      </CardHeader>
      {checks.length === 0 ? (
        <EmptyState icon={<CheckCircle className="h-6 w-6" />} title="No health checks configured" />
      ) : (
        <DataTable<Check>
          keyField="name"
          columns={[
            {
              key: 'name',
              label: 'Name',
              render: (row) => <span className="font-medium text-text-primary text-xs font-mono">{row.name}</span>,
            },
            {
              key: 'status',
              label: 'Status',
              render: (row) => (
                <Badge variant={row.status === 'pass' ? 'success' : 'danger'} dot>{row.status}</Badge>
              ),
            },
            { key: 'duration', label: 'Duration', align: 'right', mono: true },
            {
              key: 'message',
              label: 'Message',
              render: (row) => (
                <span className="text-xs text-text-secondary whitespace-pre-wrap break-words">{row.message ?? '\u2014'}</span>
              ),
            },
          ]}
          rows={checks}
          compact
        />
      )}
    </Card>
  );
}
