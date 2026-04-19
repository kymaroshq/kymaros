import { Card, CardHeader, CardTitle } from '@/components/ui2/Card';
import { DataTable } from '@/components/ui2/DataTable';
import { Badge } from '@/components/ui2/Badge';
import { EmptyState } from '@/components/ui2/EmptyState';
import { Activity } from 'lucide-react';
import { formatRelativeTime } from '@/lib/utils';
import type { EventLog } from '@/types/kymaros';

export function EventsTable({ events }: { events: EventLog[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Events</CardTitle>
        {events.length > 0 && <span className="text-2xs text-text-tertiary font-mono">{events.length}</span>}
      </CardHeader>
      {events.length === 0 ? (
        <EmptyState icon={<Activity className="h-6 w-6" />} title="No Kubernetes events" />
      ) : (
        <DataTable<EventLog>
          columns={[
            {
              key: 'type',
              label: 'Type',
              render: (row) => (
                <Badge variant={row.type === 'Warning' ? 'warning' : 'info'} size="sm">{row.type}</Badge>
              ),
            },
            {
              key: 'reason',
              label: 'Reason',
              render: (row) => <span className="font-medium text-text-primary text-xs">{row.reason}</span>,
            },
            {
              key: 'message',
              label: 'Message',
              render: (row) => <span className="text-xs text-text-secondary truncate max-w-xs block">{row.message}</span>,
            },
            {
              key: 'count',
              label: 'Count',
              align: 'right',
              render: (row) => <span className="text-xs font-mono text-text-tertiary">&times;{row.count ?? 1}</span>,
            },
            {
              key: 'lastTimestamp',
              label: 'When',
              align: 'right',
              render: (row) => <span className="text-2xs text-text-tertiary font-mono">{formatRelativeTime(row.lastTimestamp)}</span>,
            },
          ]}
          rows={events}
          compact
        />
      )}
    </Card>
  );
}
