import {
  Server,
  HardDrive,
  Bell,
  Shield,
  Trash2,
  CheckCircle,
  XCircle,
  MessageSquare,
  Globe,
  Lock,
} from 'lucide-react';
import { useLicense } from '../hooks/useKymarosData';

// ---------------------------------------------------------------------------
// Section component
// ---------------------------------------------------------------------------

function Section({
  title,
  icon: Icon,
  children,
  danger = false,
}: {
  title: string;
  icon: typeof Server;
  children: React.ReactNode;
  danger?: boolean;
}) {
  return (
    <div
      className={`rounded-xl border p-5 ${
        danger
          ? 'border-red-500/20 bg-red-500/5'
          : 'border-navy-700 bg-navy-800'
      }`}
    >
      <div className="mb-4 flex items-center gap-2.5">
        <Icon
          className={`h-5 w-5 ${danger ? 'text-red-400' : 'text-blue-400'}`}
        />
        <h2
          className={`text-lg font-semibold ${
            danger ? 'text-red-300' : 'text-white'
          }`}
        >
          {title}
        </h2>
      </div>
      {children}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Row component
// ---------------------------------------------------------------------------

function SettingRow({
  label,
  value,
  badge,
}: {
  label: string;
  value: string;
  badge?: 'connected' | 'not_configured';
}) {
  return (
    <div className="flex items-center justify-between border-b border-navy-700/50 py-3 last:border-b-0">
      <span className="text-sm text-slate-400">{label}</span>
      <div className="flex items-center gap-2">
        <span className="text-sm font-medium text-white">{value}</span>
        {badge === 'connected' && (
          <span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/10 px-2 py-0.5 text-xs font-medium text-emerald-400">
            <CheckCircle className="h-3 w-3" /> Connected
          </span>
        )}
        {badge === 'not_configured' && (
          <span className="inline-flex items-center gap-1 rounded-full bg-slate-500/10 px-2 py-0.5 text-xs font-medium text-slate-400">
            <XCircle className="h-3 w-3" /> Not configured
          </span>
        )}
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

function Settings() {
  const license = useLicense();

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-white">Settings</h1>
        <p className="text-sm text-slate-400">
          Cluster configuration (read-only for MVP)
        </p>
      </div>

      {/* General */}
      <Section title="General" icon={Server}>
        <SettingRow label="Cluster Name" value="production" />
        <SettingRow label="Sandbox Namespace Prefix" value="kymaros-sandbox-" />
        <SettingRow label="Default Validation Level" value="standard" />
        <SettingRow label="Default TTL After Finished" value="30m" />
        <SettingRow label="Cleanup Enabled" value="Yes" />
      </Section>

      {/* Backup Providers */}
      <Section title="Backup Providers" icon={HardDrive}>
        <SettingRow label="Velero" value="v1.14.0" badge="connected" />
        <div className="flex items-center justify-between border-b border-navy-700/50 py-3">
          <span className="text-sm text-slate-400">Kasten K10</span>
          {license.data?.features.multiBackup ? (
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-white">--</span>
              <span className="inline-flex items-center gap-1 rounded-full bg-slate-500/10 px-2 py-0.5 text-xs font-medium text-slate-400">
                <XCircle className="h-3 w-3" /> Not configured
              </span>
            </div>
          ) : (
            <div className="flex items-center gap-2 text-xs text-slate-500">
              <Lock className="h-3.5 w-3.5" />
              Available in Team
            </div>
          )}
        </div>
        <div className="flex items-center justify-between py-3">
          <span className="text-sm text-slate-400">Trilio</span>
          {license.data?.features.multiBackup ? (
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-white">--</span>
              <span className="inline-flex items-center gap-1 rounded-full bg-slate-500/10 px-2 py-0.5 text-xs font-medium text-slate-400">
                <XCircle className="h-3 w-3" /> Not configured
              </span>
            </div>
          ) : (
            <div className="flex items-center gap-2 text-xs text-slate-500">
              <Lock className="h-3.5 w-3.5" />
              Available in Team
            </div>
          )}
        </div>
      </Section>

      {/* Notifications */}
      <Section title="Notifications" icon={Bell}>
        <div className="space-y-0">
          <div className="flex items-center justify-between border-b border-navy-700/50 py-3">
            <div className="flex items-center gap-2">
              <MessageSquare className="h-4 w-4 text-slate-500" />
              <span className="text-sm text-slate-400">Slack</span>
            </div>
            <span className="inline-flex items-center gap-1 rounded-full bg-slate-500/10 px-2 py-0.5 text-xs font-medium text-slate-400">
              <XCircle className="h-3 w-3" /> Not configured
            </span>
          </div>
          <div className="flex items-center justify-between border-b border-navy-700/50 py-3">
            <div className="flex items-center gap-2">
              <Bell className="h-4 w-4 text-slate-500" />
              <span className="text-sm text-slate-400">PagerDuty</span>
            </div>
            <span className="inline-flex items-center gap-1 rounded-full bg-slate-500/10 px-2 py-0.5 text-xs font-medium text-slate-400">
              <XCircle className="h-3 w-3" /> Not configured
            </span>
          </div>
          <div className="flex items-center justify-between py-3">
            <div className="flex items-center gap-2">
              <Globe className="h-4 w-4 text-slate-500" />
              <span className="text-sm text-slate-400">Webhook</span>
            </div>
            <span className="inline-flex items-center gap-1 rounded-full bg-slate-500/10 px-2 py-0.5 text-xs font-medium text-slate-400">
              <XCircle className="h-3 w-3" /> Not configured
            </span>
          </div>
        </div>
      </Section>

      {/* SLA Targets */}
      <Section title="SLA Targets" icon={Shield}>
        <SettingRow label="RTO Target" value="15m" />
        <SettingRow label="Alert Score Threshold" value="< 70" />
        <SettingRow label="Consecutive Failure Alert" value="3 failures" />
        <SettingRow label="Compliance Target" value="99.5%" />
      </Section>

      {/* Danger Zone */}
      <Section title="Danger Zone" icon={Trash2} danger>
        <p className="mb-4 text-sm text-slate-400">
          These actions are irreversible. Handle with care.
        </p>
        <div className="flex flex-wrap gap-3">
          <button
            type="button"
            disabled
            className="flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-2 text-sm font-medium text-red-400 opacity-60 cursor-not-allowed"
          >
            <Trash2 className="h-4 w-4" />
            Cleanup All Sandboxes
          </button>
          <button
            type="button"
            disabled
            className="flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-2 text-sm font-medium text-red-400 opacity-60 cursor-not-allowed"
          >
            <Trash2 className="h-4 w-4" />
            Purge Old Reports
          </button>
        </div>
      </Section>
    </div>
  );
}

export default Settings;
