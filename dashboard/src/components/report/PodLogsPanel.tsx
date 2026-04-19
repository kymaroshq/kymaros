import { useState } from 'react';
import { Card, CardHeader, CardTitle } from '@/components/ui2/Card';
import { Badge } from '@/components/ui2/Badge';
import { CodeBlock } from '@/components/ui2/CodeBlock';
import { EmptyState } from '@/components/ui2/EmptyState';
import { FileText, ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { PodLog } from '@/types/kymaros';

function phaseVariant(phase: string): 'success' | 'warning' | 'danger' | 'neutral' {
  if (phase === 'Running' || phase === 'Succeeded') return 'success';
  if (phase === 'Pending') return 'warning';
  if (phase === 'Failed' || phase === 'CrashLoopBackOff') return 'danger';
  return 'neutral';
}

export function PodLogsPanel({ pods }: { pods: PodLog[] }) {
  const [openPods, setOpenPods] = useState<Set<string>>(() => {
    const initial = new Set<string>();
    if (pods.length > 0) initial.add(pods[0].podName);
    return initial;
  });

  const toggle = (podName: string) => {
    setOpenPods((prev) => {
      const next = new Set(prev);
      if (next.has(podName)) next.delete(podName);
      else next.add(podName);
      return next;
    });
  };

  if (pods.length === 0) {
    return (
      <Card>
        <CardHeader><CardTitle>Pod Logs</CardTitle></CardHeader>
        <EmptyState icon={<FileText className="h-6 w-6" />} title="No pod logs captured" />
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Pod Logs</CardTitle>
        <span className="text-2xs text-text-tertiary font-mono">{pods.length} pods</span>
      </CardHeader>
      <div className="divide-y divide-border-subtle">
        {pods.map((pod) => {
          const isOpen = openPods.has(pod.podName);
          return (
            <div key={pod.podName}>
              <button
                type="button"
                onClick={() => toggle(pod.podName)}
                className="w-full flex items-center gap-2 px-4 py-2.5 text-left hover:bg-surface-3 transition-colors duration-150"
              >
                <ChevronRight className={cn('h-3.5 w-3.5 text-text-tertiary transition-transform duration-150', isOpen && 'rotate-90')} />
                <span className="font-mono text-xs text-text-primary truncate flex-1">{pod.podName}</span>
                <Badge variant={phaseVariant(pod.phase)} size="sm">{pod.phase}</Badge>
              </button>
              {isOpen && (
                <div className="px-4 pb-3 space-y-3">
                  {pod.containers.map((container) => (
                    <div key={container.name}>
                      <div className="flex items-center gap-2 mb-1.5">
                        <span className="text-2xs uppercase tracking-wider text-text-tertiary">
                          {container.type === 'init' ? 'init' : 'container'}
                        </span>
                        <span className="text-xs font-mono text-text-secondary">{container.name}</span>
                        {container.truncated && <Badge variant="neutral" size="sm">truncated</Badge>}
                      </div>
                      <CodeBlock maxLines={15}>{container.log || 'No logs available'}</CodeBlock>
                    </div>
                  ))}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </Card>
  );
}
