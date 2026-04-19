interface TestDistributionProps {
  pass: number;
  partial: number;
  fail: number;
}

export function TestDistribution({ pass, partial, fail }: TestDistributionProps) {
  const total = pass + partial + fail;
  if (total === 0) {
    return <div className="text-sm text-text-tertiary text-center py-8">No tests yet</div>;
  }

  const items = [
    { label: 'Pass', value: pass, color: 'bg-status-success', textColor: 'text-status-success' },
    { label: 'Partial', value: partial, color: 'bg-status-warning', textColor: 'text-status-warning' },
    { label: 'Fail', value: fail, color: 'bg-status-danger', textColor: 'text-status-danger' },
  ];

  return (
    <div className="space-y-4">
      {/* Stacked bar */}
      <div className="h-2 bg-surface-4 rounded-full overflow-hidden flex">
        {items.map((item) =>
          item.value > 0 ? (
            <div
              key={item.label}
              className={item.color}
              style={{ width: `${(item.value / total) * 100}%` }}
            />
          ) : null
        )}
      </div>

      {/* Legend */}
      <div className="space-y-2.5">
        {items.map((item) => (
          <div key={item.label} className="flex items-center justify-between text-sm">
            <div className="flex items-center gap-2">
              <span className={`h-2 w-2 rounded-full ${item.color}`} />
              <span className="text-text-secondary">{item.label}</span>
            </div>
            <div className="flex items-baseline gap-1.5">
              <span className="text-metric text-text-primary">{item.value}</span>
              <span className="text-2xs text-text-tertiary font-mono">
                {Math.round((item.value / total) * 100)}%
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
