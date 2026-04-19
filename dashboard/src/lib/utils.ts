import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatRelativeTime(date: string | Date): string {
  const d = new Date(date);
  const now = Date.now();
  const diff = now - d.getTime();

  if (diff < 0) {
    const absDiff = -diff;
    const seconds = Math.floor(absDiff / 1000);
    if (seconds < 60) return `in ${seconds}s`;
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `in ${minutes}m`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `in ${hours}h`;
    const days = Math.floor(hours / 24);
    return `in ${days}d`;
  }

  const seconds = Math.floor(diff / 1000);
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  return d.toISOString().split('T')[0];
}

export function getScoreTone(score: number): 'excellent' | 'good' | 'warning' | 'critical' {
  if (score >= 90) return 'excellent';
  if (score >= 70) return 'good';
  if (score >= 50) return 'warning';
  return 'critical';
}

export function getResultVariant(result: string): 'success' | 'warning' | 'danger' {
  if (result === 'pass') return 'success';
  if (result === 'partial') return 'warning';
  return 'danger';
}

export function parseDurationToSeconds(dur: string): number {
  let total = 0;
  const hourMatch = dur.match(/(\d+)\s*h/);
  const minMatch = dur.match(/(\d+)\s*m/);
  const secMatch = dur.match(/([\d.]+)\s*s/);
  if (hourMatch) total += parseInt(hourMatch[1]) * 3600;
  if (minMatch) total += parseInt(minMatch[1]) * 60;
  if (secMatch) total += parseFloat(secMatch[1]);
  return total;
}

export function formatDuration(dur: string): string {
  const s = parseDurationToSeconds(dur);
  if (s < 60) return `${Math.round(s)}s`;
  if (s < 3600) return `${Math.floor(s / 60)}m ${Math.round(s % 60)}s`;
  return `${Math.floor(s / 3600)}h ${Math.floor((s % 3600) / 60)}m`;
}

export function jsonToYaml(obj: unknown, indent = 0): string {
  const pad = '  '.repeat(indent);
  if (obj === null || obj === undefined) return `${pad}null`;
  if (typeof obj === 'boolean' || typeof obj === 'number') return `${pad}${obj}`;
  if (typeof obj === 'string') {
    if (obj.includes('\n') || obj.includes(': ') || obj.includes('#') || /^[{[]/.test(obj)) {
      return `${pad}"${obj.replace(/"/g, '\\"')}"`;
    }
    return `${pad}${obj}`;
  }
  if (Array.isArray(obj)) {
    if (obj.length === 0) return `${pad}[]`;
    return obj.map((item) => {
      if (typeof item === 'object' && item !== null) {
        const inner = jsonToYaml(item, indent + 1).trimStart();
        return `${pad}- ${inner}`;
      }
      return `${pad}- ${typeof item === 'string' ? item : JSON.stringify(item)}`;
    }).join('\n');
  }
  if (typeof obj === 'object') {
    const entries = Object.entries(obj as Record<string, unknown>);
    if (entries.length === 0) return `${pad}{}`;
    return entries.map(([key, val]) => {
      if (val === null || val === undefined) return `${pad}${key}: null`;
      if (typeof val === 'string' || typeof val === 'number' || typeof val === 'boolean') {
        return `${pad}${key}: ${jsonToYaml(val, 0).trim()}`;
      }
      return `${pad}${key}:\n${jsonToYaml(val, indent + 1)}`;
    }).join('\n');
  }
  return `${pad}${JSON.stringify(obj)}`;
}

export function formatCron(cron: string): string {
  const parts = cron.trim().split(/\s+/);
  if (parts.length < 5) return cron;
  const [min, hour, dom, mon, dow] = parts;
  if (dom === '*' && mon === '*' && dow === '*') return `Daily at ${hour}:${min.padStart(2, '0')} UTC`;
  if (dom === '*' && mon === '*' && dow !== '*') {
    const days: Record<string, string> = { '0': 'Sun', '1': 'Mon', '2': 'Tue', '3': 'Wed', '4': 'Thu', '5': 'Fri', '6': 'Sat' };
    return `${days[dow] ?? dow} at ${hour}:${min.padStart(2, '0')} UTC`;
  }
  return cron;
}

export function formatRTO(rto: string | undefined | null): string {
  if (!rto) return '\u2014';
  const s = parseDurationToSeconds(rto);
  if (s <= 0 || s >= 86400) return '\u2014';
  return formatDuration(rto);
}
