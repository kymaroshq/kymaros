import type { FC } from 'react';
import { AreaChart, Area, ResponsiveContainer } from 'recharts';

interface SparklineChartProps {
  data: number[];
  width?: number;
  height?: number;
  color?: string;
}

function autoColor(value: number): string {
  if (value >= 80) return '#10b981';
  if (value >= 50) return '#f59e0b';
  return '#ef4444';
}

const SparklineChart: FC<SparklineChartProps> = ({
  data,
  width = 100,
  height = 30,
  color,
}) => {
  const lastValue = data.length > 0 ? data[data.length - 1] : 0;
  const resolvedColor = color ?? autoColor(lastValue);
  const chartData = data.map((value, index) => ({ index, value }));

  const CustomDot = (props: any) => {
    const { cx, cy, index } = props;
    if (index !== chartData.length - 1) return null;
    return (
      <circle cx={cx} cy={cy} r={3} fill={resolvedColor} stroke="#1e293b" strokeWidth={1.5} />
    );
  };

  return (
    <div style={{ width, height }}>
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={chartData} margin={{ top: 0, right: 0, bottom: 0, left: 0 }}>
          <defs>
            <linearGradient id={`spark-${resolvedColor.replace('#', '')}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={resolvedColor} stopOpacity={0.3} />
              <stop offset="100%" stopColor={resolvedColor} stopOpacity={0} />
            </linearGradient>
          </defs>
          <Area
            type="monotone"
            dataKey="value"
            stroke={resolvedColor}
            strokeWidth={1.5}
            fill={`url(#spark-${resolvedColor.replace('#', '')})`}
            isAnimationActive={false}
            dot={<CustomDot />}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
};

export default SparklineChart;
