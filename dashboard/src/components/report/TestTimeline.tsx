import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui2/Card';
import { Timeline } from '@/components/ui2/Timeline';
import type { TimelineItem } from '@/components/ui2/Timeline';

interface TestTimelineProps {
  startedAt?: string;
  completedAt?: string;
  result?: string;
}

export function TestTimeline({ startedAt, completedAt, result }: TestTimelineProps) {
  const items: TimelineItem[] = [];

  if (startedAt) {
    // Sandbox creation is the first step — same timestamp as test start
    items.push({ id: 'sandbox', title: 'Sandbox created', timestamp: startedAt, status: 'completed' });
    items.push({ id: 'restore', title: 'Restore triggered', timestamp: startedAt, status: 'completed', description: 'Backup restored into sandbox namespace' });
  }

  if (completedAt) {
    const status = result === 'fail' ? 'failed' as const : 'completed' as const;
    const title = result === 'fail' ? 'Validation failed' : result === 'partial' ? 'Validation partial' : 'Validation passed';
    items.push({ id: 'validated', title, timestamp: completedAt, status });
    items.push({ id: 'cleanup', title: 'Sandbox cleaned up', timestamp: completedAt, status: 'completed' });
  }

  if (items.length === 0) {
    return null;
  }

  return (
    <Card>
      <CardHeader><CardTitle>Test Timeline</CardTitle></CardHeader>
      <CardBody><Timeline items={items} /></CardBody>
    </Card>
  );
}
