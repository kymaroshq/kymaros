import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';

import type { DailySummary } from '@/types/kymaros';

export function ScoreTrendChart({ data }: { data: DailySummary[] }) {
  if (data.length === 0) {
    return <div className="h-48 flex items-center justify-center text-sm text-text-tertiary">No data yet</div>;
  }

  return (
    <div className="h-48 -mx-2">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data} margin={{ top: 8, right: 8, bottom: 4, left: 4 }}>
          <CartesianGrid
            strokeDasharray="3 3"
            stroke="var(--color-border-subtle)"
            vertical={false}
          />
          <XAxis
            dataKey="date"
            stroke="var(--color-text-tertiary)"
            fontSize={10}
            tickLine={false}
            axisLine={false}
            tickFormatter={(d: string) => d.slice(5)}
          />
          <YAxis
            stroke="var(--color-text-tertiary)"
            fontSize={10}
            tickLine={false}
            axisLine={false}
            domain={[0, 100]}
            ticks={[0, 25, 50, 75, 100]}
            width={28}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: 'var(--color-surface-3)',
              border: '1px solid var(--color-border-default)',
              borderRadius: 6,
              fontSize: 12,
              fontFamily: 'var(--font-mono)',
              padding: '8px 12px',
              color: 'var(--color-text-primary)',
            }}
            labelStyle={{ color: 'var(--color-text-secondary)', marginBottom: 4 }}
            labelFormatter={(label) => String(label)}
            formatter={(value) => [Math.round(Number(value)), 'Score']}
          />
          <Line
            type="monotone"
            dataKey="score"
            name="Score"
            stroke="var(--color-accent)"
            strokeWidth={1.5}
            dot={false}
            activeDot={{ r: 4, fill: 'var(--color-accent)', strokeWidth: 0 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
