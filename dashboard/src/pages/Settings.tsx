import { useState, useEffect, useCallback, type ReactNode } from 'react';
import {
  Settings as SettingsIcon,
  Server,
  HardDrive,
  Bell,
  Shield,
  BarChart3,
  Download,
  Trash2,
  Send,
} from 'lucide-react';
import { kymarosApi } from '../api/kymarosApi';
import type {
  HealthResponse,
  ProviderConfig,
  NotificationConfig,
  SandboxConfig,
  SummaryResponse,
} from '../types/kymaros';

export default function SettingsPage() {
  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <h1 className="text-xl font-semibold text-white flex items-center gap-2">
        <SettingsIcon className="w-5 h-5" />
        Settings
      </h1>
      <ClusterInfoSection />
      <BackupProviderSection />
      <NotificationsSection />
      <SandboxDefaultsSection />
      <MetricsSection />
    </div>
  );
}

// ── Cluster Info ──

function ClusterInfoSection() {
  const [data, setData] = useState<HealthResponse | null>(null);
  useEffect(() => { kymarosApi.getHealth().then(setData).catch(() => {}); }, []);
  if (!data) return <SectionSkeleton />;
  return (
    <Section title="Cluster info" icon={<Server className="w-4 h-4" />}>
      <InfoRow label="Kymaros version" value={data.version} />
      <InfoRow label="Kubernetes version" value={data.kubernetesVersion} />
      <InfoRow label="Namespace" value={data.namespace} mono />
      <InfoRow label="Uptime" value={data.uptime} />
      <InfoRow label="Pod" value={data.pod} mono />
    </Section>
  );
}

// ── Backup Provider ──

function BackupProviderSection() {
  const [data, setData] = useState<ProviderConfig | null>(null);
  useEffect(() => { kymarosApi.getProviderConfig().then(setData).catch(() => {}); }, []);
  if (!data) return <SectionSkeleton />;
  return (
    <Section title="Backup provider" icon={<HardDrive className="w-4 h-4" />}>
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-white capitalize">{data.provider}</span>
          {data.version && <span className="text-xs text-slate-500">{data.version}</span>}
        </div>
        <StatusBadge status={data.status} />
      </div>
      <InfoRow label="Namespace" value={data.namespace} mono />
      <InfoRow label="Default BSL" value={data.defaultBSL || '—'} />
      <InfoRow label="Backups" value={data.backupCount} />
      <InfoRow label="Schedules" value={data.scheduleCount} />
      <InfoRow label="Last backup" value={data.lastBackupAge ? `${data.lastBackupAge} ago` : '—'} />
    </Section>
  );
}

// ── Notifications ──

function NotificationsSection() {
  const [data, setData] = useState<NotificationConfig | null>(null);
  const [testing, setTesting] = useState<string | null>(null);

  useEffect(() => { kymarosApi.getNotificationConfig().then(setData).catch(() => {}); }, []);

  const handleTest = useCallback(async (ch: { type: string; secretRef?: string }) => {
    setTesting(ch.type);
    try { await kymarosApi.testNotification(ch); } finally { setTesting(null); }
  }, []);

  if (!data) return <SectionSkeleton />;
  return (
    <Section title="Notifications" icon={<Bell className="w-4 h-4" />}>
      {data.channels.length === 0 ? (
        <p className="text-sm text-slate-500">No notifications configured. Add notifications to your RestoreTest CRDs.</p>
      ) : (
        <div className="space-y-3">
          {data.channels.map((ch, i) => (
            <div key={i} className="p-3 rounded-lg bg-slate-800/50 border border-slate-700/50">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-white capitalize">{ch.type}</span>
                  <StatusBadge status="configured" />
                </div>
                <button
                  onClick={() => handleTest({ type: ch.type, secretRef: ch.secretRef })}
                  disabled={testing === ch.type}
                  className="text-xs text-slate-400 hover:text-white flex items-center gap-1 transition-colors disabled:opacity-50"
                >
                  <Send className="w-3 h-3" />
                  {testing === ch.type ? 'Sending...' : 'Send test'}
                </button>
              </div>
              {ch.channel && <InfoRow label="Channel" value={ch.channel} />}
              {ch.secretRef && (
                <InfoRow label="Secret" value={`${ch.secretRef} ${ch.secretExists ? '✓' : '✗ missing'}`} />
              )}
              <InfoRow label="Used by" value={ch.usedByTests?.join(', ') || '—'} />
            </div>
          ))}
        </div>
      )}
    </Section>
  );
}

// ── Sandbox Defaults ──

function SandboxDefaultsSection() {
  const [data, setData] = useState<SandboxConfig | null>(null);
  const [cleaning, setCleaning] = useState(false);

  useEffect(() => { kymarosApi.getSandboxConfig().then(setData).catch(() => {}); }, []);

  const handleCleanup = useCallback(async () => {
    setCleaning(true);
    try {
      await kymarosApi.cleanupOrphanSandboxes();
      const updated = await kymarosApi.getSandboxConfig();
      setData(updated);
    } finally { setCleaning(false); }
  }, []);

  if (!data) return <SectionSkeleton />;
  return (
    <Section title="Sandbox defaults" icon={<Shield className="w-4 h-4" />}>
      <p className="text-xs text-slate-600 mb-3">
        Configured via Helm values. Change with <code className="text-xs bg-slate-800 px-1 rounded">helm upgrade</code>.
      </p>
      <InfoRow label="Namespace prefix" value={data.defaults.namespacePrefix} mono />
      <InfoRow label="TTL" value={data.defaults.ttl} />
      <InfoRow label="Network isolation" value={data.defaults.networkIsolation === 'strict' ? 'strict (deny-all)' : data.defaults.networkIsolation} />
      <div className="mt-2 mb-1"><span className="text-xs text-slate-500">Resource quota</span></div>
      <div className="pl-3">
        <InfoRow label="CPU" value={data.defaults.resourceQuota.cpu} />
        <InfoRow label="Memory" value={data.defaults.resourceQuota.memory} />
        <InfoRow label="Storage" value={data.defaults.resourceQuota.storage} />
        <InfoRow label="Max pods" value={data.defaults.resourceQuota.pods} />
      </div>
      <div className="border-t border-slate-700/50 mt-3 pt-3 flex items-center justify-between">
        <div>
          <InfoRow label="Active sandboxes" value={data.activeSandboxes} />
          <InfoRow label="Orphaned sandboxes" value={data.orphanedSandboxes} />
        </div>
        {data.orphanedSandboxes > 0 && (
          <button
            onClick={handleCleanup}
            disabled={cleaning}
            className="text-xs px-3 py-1.5 rounded-md border border-red-500/30 text-red-400 hover:bg-red-500/10 flex items-center gap-1 transition-colors disabled:opacity-50"
          >
            <Trash2 className="w-3 h-3" />
            {cleaning ? 'Cleaning...' : 'Cleanup orphans'}
          </button>
        )}
      </div>
    </Section>
  );
}

// ── Metrics ──

function MetricsSection() {
  const [data, setData] = useState<SummaryResponse | null>(null);
  useEffect(() => { kymarosApi.getSummary().then(setData).catch(() => {}); }, []);
  if (!data) return <SectionSkeleton />;
  return (
    <Section title="Metrics" icon={<BarChart3 className="w-4 h-4" />}>
      <InfoRow label="Prometheus endpoint" value=":8443/metrics" mono />
      <InfoRow label="Metrics exposed" value="5" />
      <div className="border-t border-slate-700/50 mt-3 pt-3 space-y-1">
        <InfoRow label="Total tests" value={data.totalTests} />
        <InfoRow label="Average score" value={Math.round(data.averageScore)} />
        <InfoRow label="Total failures" value={data.totalFailures} />
      </div>
      <div className="mt-3">
        <a
          href="/grafana/kymaros-dashboard.json"
          download="kymaros-grafana-dashboard.json"
          className="text-xs px-3 py-1.5 rounded-md border border-slate-600 text-slate-400 hover:text-white hover:border-slate-500 inline-flex items-center gap-1 transition-colors"
        >
          <Download className="w-3 h-3" />
          Download Grafana dashboard
        </a>
      </div>
    </Section>
  );
}

// ── Shared components ──

function Section({ title, icon, children }: { title: string; icon: ReactNode; children: ReactNode }) {
  return (
    <div className="bg-slate-800/30 border border-slate-700/50 rounded-lg p-5">
      <h2 className="text-sm font-medium text-slate-400 uppercase tracking-wider flex items-center gap-2 mb-4">
        {icon}{title}
      </h2>
      {children}
    </div>
  );
}

function InfoRow({ label, value, mono }: { label: string; value: unknown; mono?: boolean }) {
  return (
    <div className="flex items-center justify-between py-1">
      <span className="text-sm text-slate-500">{label}</span>
      <span className={`text-sm text-slate-200 ${mono ? 'font-mono' : ''}`}>{String(value ?? '—')}</span>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    connected: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
    configured: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
    'not found': 'bg-red-500/10 text-red-400 border-red-500/20',
    error: 'bg-red-500/10 text-red-400 border-red-500/20',
  };
  return (
    <span className={`text-xs px-2 py-0.5 rounded border ${colors[status] || 'bg-slate-700 text-slate-400 border-slate-600'}`}>
      {status}
    </span>
  );
}

function SectionSkeleton() {
  return (
    <div className="bg-slate-800/30 border border-slate-700/50 rounded-lg p-5 animate-pulse">
      <div className="h-4 w-32 bg-slate-700 rounded mb-4" />
      <div className="space-y-2">
        <div className="h-3 w-full bg-slate-700/50 rounded" />
        <div className="h-3 w-3/4 bg-slate-700/50 rounded" />
        <div className="h-3 w-1/2 bg-slate-700/50 rounded" />
      </div>
    </div>
  );
}
