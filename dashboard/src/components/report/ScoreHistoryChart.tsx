import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid, ReferenceLine } from 'recharts';
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui2/Card';
import { EmptyState } from '@/components/ui2/EmptyState';
import { TrendingUp } from 'lucide-react';
import type { RestoreReport } from '@/types/kymaros';

interface ScoreHistoryChartProps {
  reports: RestoreReport[];
  currentDate?: string;
}

export function ScoreHistoryChart({ reports, currentDate }: ScoreHistoryChartProps) {
  const data = reports.map((r) => ({
    date: r.status?.startedAt?.split('T')[0] ?? r.metadata.creationTimestamp.split('T')[0],
    score: Math.round(r.status?.score ?? 0),
  }));

  return (
    <Card>
      <CardHeader>
        <CardTitle>Score History</CardTitle>
        <span className="text-2xs text-text-tertiary font-mono">{reports.length} runs</span>
      </CardHeader>
      <CardBody>
        {data.length === 0 ? (
          <EmptyState icon={<TrendingUp className="h-6 w-6" />} title="No historical data yet" />
        ) : (
          <div className="h-48 -mx-2">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={data} margin={{ top: 8, right: 8, bottom: 4, left: 4 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border-subtle)" vertical={false} />
                <XAxis dataKey="date" stroke="var(--color-text-tertiary)" fontSize={10} tickLine={false} axisLine={false} tickFormatter={(d: string) => d.slice(5)} />
                <YAxis stroke="var(--color-text-tertiary)" fontSize={10} tickLine={false} axisLine={false} domain={[0, 100]} ticks={[0, 50, 100]} width={28} />
                <Tooltip
                  contentStyle={{ backgroundColor: 'var(--color-surface-3)', border: '1px solid var(--color-border-default)', borderRadius: 6, fontSize: 12, fontFamily: 'var(--font-mono)', padding: '8px 12px', color: 'var(--color-text-primary)' }}
                  formatter={(value) => [Math.round(Number(value)), 'Score']}
                />
                <Line type="monotone" dataKey="score" stroke="var(--color-accent)" strokeWidth={1.5} dot={false} activeDot={{ r: 4, strokeWidth: 0 }} />
                {currentDate && <ReferenceLine x={currentDate} stroke="var(--color-text-tertiary)" strokeDasharray="2 2" />}
              </LineChart>
            </ResponsiveContainer>
          </div>
        )}
      </CardBody>
    </Card>
  );
}
