import type { RestoreReport } from '@/types/kymaros';

function download(content: string, filename: string, mimeType: string): void {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

export function exportReportsCSV(reports: RestoreReport[]): void {
  const header = ['Name', 'Test', 'Namespace', 'Score', 'Result', 'RTO', 'Started At'];
  const lines = reports.map((r) => [
    r.metadata.name,
    r.spec.testRef,
    r.metadata.namespace,
    String(Math.round(r.status?.score ?? 0)),
    r.status?.result ?? '',
    r.status?.rto?.measured ?? '',
    r.status?.startedAt ?? r.metadata.creationTimestamp,
  ]);

  const csv = [header, ...lines]
    .map((row) => row.map((cell) => `"${String(cell).replace(/"/g, '""')}"`).join(','))
    .join('\n');

  download(csv, `kymaros-reports-${new Date().toISOString().split('T')[0]}.csv`, 'text/csv');
}

export function exportReportsJSON(reports: RestoreReport[]): void {
  const json = JSON.stringify({ exportedAt: new Date().toISOString(), reports }, null, 2);
  download(json, `kymaros-reports-${new Date().toISOString().split('T')[0]}.json`, 'application/json');
}
