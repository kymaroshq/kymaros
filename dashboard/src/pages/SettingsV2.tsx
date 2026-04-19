import { useState, useEffect, useCallback } from 'react';
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui2/Card';
import { Badge } from '@/components/ui2/Badge';
import { Button } from '@/components/ui2/Button';
import { Alert } from '@/components/ui2/Alert';
import { EmptyState } from '@/components/ui2/EmptyState';
import { cn } from '@/lib/utils';
import { kymarosApi, getAuthToken, setAuthToken } from '@/api/kymarosApi';
import { CodeBlock } from '@/components/ui2/CodeBlock';
import type {
  HealthResponse,
  ProviderConfig,
  NotificationConfig,
  SandboxConfig,
} from '@/types/kymaros';
import {
  Server,
  HardDrive,
  Bell,
  Shield,
  Info,
  Send,
  Trash2,
  ExternalLink,
  BookOpen,
  CheckCircle,
  XCircle,
  Eye,
  EyeOff,
  Save,
  Key,
} from 'lucide-react';

type Tab = 'provider' | 'sandbox' | 'notifications' | 'api' | 'about';

const tabs: { id: Tab; label: string }[] = [
  { id: 'provider', label: 'Provider' },
  { id: 'sandbox', label: 'Sandbox' },
  { id: 'notifications', label: 'Notifications' },
  { id: 'api', label: 'API' },
  { id: 'about', label: 'About' },
];

export default function SettingsV2() {
  const [activeTab, setActiveTab] = useState<Tab>('provider');

  return (
    <div className="px-6 py-6 space-y-6 max-w-4xl mx-auto">
      <div>
        <h1 className="text-xl font-semibold text-text-primary">Settings</h1>
        <p className="text-sm text-text-tertiary mt-0.5">Manage your Kymaros configuration</p>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-border-subtle">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={cn(
              'px-3 py-2 text-sm font-medium border-b-2 -mb-px transition-colors duration-150',
              activeTab === tab.id
                ? 'text-text-primary border-accent'
                : 'text-text-tertiary border-transparent hover:text-text-primary'
            )}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {activeTab === 'provider' && <ProviderTab />}
      {activeTab === 'sandbox' && <SandboxTab />}
      {activeTab === 'notifications' && <NotificationsTab />}
      {activeTab === 'api' && <ApiTab />}
      {activeTab === 'about' && <AboutTab />}
    </div>
  );
}

// ── InfoRow ──

function InfoRow({ label, value, mono }: { label: string; value: unknown; mono?: boolean }) {
  return (
    <div className="flex items-center justify-between py-1.5">
      <span className="text-sm text-text-tertiary">{label}</span>
      <span className={cn('text-sm text-text-primary', mono && 'font-mono')}>{String(value ?? '\u2014')}</span>
    </div>
  );
}

function SectionSkeleton() {
  return (
    <Card>
      <CardBody>
        <div className="space-y-3 animate-pulse-subtle">
          <div className="h-4 w-32 bg-surface-3 rounded" />
          <div className="h-3 w-full bg-surface-3 rounded" />
          <div className="h-3 w-3/4 bg-surface-3 rounded" />
          <div className="h-3 w-1/2 bg-surface-3 rounded" />
        </div>
      </CardBody>
    </Card>
  );
}

// ── Provider Tab ──

function ProviderTab() {
  const [data, setData] = useState<ProviderConfig | null>(null);
  useEffect(() => { kymarosApi.getProviderConfig().then(setData).catch(() => {}); }, []);

  if (!data) return <SectionSkeleton />;

  const connected = data.status === 'connected';

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <HardDrive className="h-4 w-4 text-text-tertiary" />
          <CardTitle>Backup Provider</CardTitle>
        </div>
        <Badge variant={connected ? 'success' : 'danger'} dot>
          {connected ? 'Connected' : 'Disconnected'}
        </Badge>
      </CardHeader>
      <CardBody className="space-y-1">
        <Alert variant="info" className="mb-3">
          Provider configuration is managed via Helm values. To change it, update your <code className="font-mono text-xs">values.yaml</code> and run <code className="font-mono text-xs">helm upgrade</code>.
        </Alert>
        {!connected && (
          <Alert variant="danger" title="Cannot reach Velero" className="mb-3">
            Verify that Velero is running in namespace <span className="font-mono">{data.namespace}</span> and
            that the Kymaros service account has the required RBAC permissions.
          </Alert>
        )}
        <InfoRow label="Provider" value={`${data.provider}${data.version ? ` ${data.version}` : ''}`} />
        <InfoRow label="Namespace" value={data.namespace} mono />
        <InfoRow label="Default BSL" value={data.defaultBSL} mono />
        <InfoRow label="Backups available" value={data.backupCount} />
        <InfoRow label="Schedules" value={data.scheduleCount} />
        <InfoRow label="Last backup" value={data.lastBackupAge ? `${data.lastBackupAge} ago` : '\u2014'} />
        {data.lastBackupName && <InfoRow label="Last backup name" value={data.lastBackupName} mono />}
      </CardBody>
    </Card>
  );
}

// ── Sandbox Tab ──

function SandboxTab() {
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

  const hasOrphans = data.orphanedSandboxes > 0;

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Shield className="h-4 w-4 text-text-tertiary" />
            <CardTitle>Sandbox Defaults</CardTitle>
          </div>
          <span className="text-2xs text-text-tertiary font-mono">
            {data.activeSandboxes} active &middot; {data.orphanedSandboxes} orphaned
          </span>
        </CardHeader>
        <CardBody className="space-y-1">
          <Alert variant="info" className="mb-3">
            Sandbox defaults are managed via Helm values. To change them, update your <code className="font-mono text-xs">values.yaml</code> and run <code className="font-mono text-xs">helm upgrade</code>.
          </Alert>
          <InfoRow label="Namespace prefix" value={data.defaults.namespacePrefix} mono />
          <InfoRow label="TTL" value={data.defaults.ttl} mono />
          <InfoRow label="Network isolation" value={data.defaults.networkIsolation === 'strict' ? 'Strict (deny-all)' : data.defaults.networkIsolation} />

          <div className="pt-3 mt-3 border-t border-border-subtle">
            <span className="text-2xs font-medium uppercase tracking-wider text-text-tertiary">Resource Quota</span>
            <div className="grid grid-cols-4 gap-4 mt-2">
              <div>
                <div className="text-2xs text-text-tertiary">CPU</div>
                <div className="text-sm font-mono text-text-primary">{data.defaults.resourceQuota.cpu}</div>
              </div>
              <div>
                <div className="text-2xs text-text-tertiary">Memory</div>
                <div className="text-sm font-mono text-text-primary">{data.defaults.resourceQuota.memory}</div>
              </div>
              <div>
                <div className="text-2xs text-text-tertiary">Storage</div>
                <div className="text-sm font-mono text-text-primary">{data.defaults.resourceQuota.storage}</div>
              </div>
              <div>
                <div className="text-2xs text-text-tertiary">Pods</div>
                <div className="text-sm font-mono text-text-primary">{data.defaults.resourceQuota.pods}</div>
              </div>
            </div>
          </div>
        </CardBody>
      </Card>

      {hasOrphans && (
        <Card>
          <CardBody className="flex items-center justify-between">
            <Alert variant="warning" title={`${data.orphanedSandboxes} orphaned sandboxes detected`} className="flex-1">
              These namespaces should have been cleaned up but are still present.
            </Alert>
            <Button
              variant="danger"
              size="sm"
              onClick={handleCleanup}
              disabled={cleaning}
              className="ml-4 shrink-0"
            >
              <Trash2 className="h-3 w-3" />
              {cleaning ? 'Cleaning...' : 'Cleanup'}
            </Button>
          </CardBody>
        </Card>
      )}
    </div>
  );
}

// ── Notifications Tab ──

function NotificationsTab() {
  const [data, setData] = useState<NotificationConfig | null>(null);
  const [testing, setTesting] = useState<string | null>(null);

  useEffect(() => { kymarosApi.getNotificationConfig().then(setData).catch(() => {}); }, []);

  const handleTest = useCallback(async (ch: { type: string; secretRef?: string }) => {
    setTesting(ch.type);
    try { await kymarosApi.testNotification(ch); } finally { setTesting(null); }
  }, []);

  if (!data) return <SectionSkeleton />;

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Bell className="h-4 w-4 text-text-tertiary" />
          <CardTitle>Notification Channels</CardTitle>
        </div>
        <span className="text-2xs text-text-tertiary font-mono">{data.channels.length} configured</span>
      </CardHeader>
      <CardBody>
        <Alert variant="info" className="mb-3">
          Notification channels are configured per-test in RestoreTest CRDs. This view aggregates all channels across your tests.
        </Alert>
        {data.channels.length === 0 ? (
          <EmptyState
            icon={<Bell className="h-6 w-6" />}
            title="No notification channels"
            description="Add notifications to your RestoreTest CRDs to get alerted on failures."
          />
        ) : (
          <div className="space-y-3">
            {data.channels.map((ch, i) => (
              <div key={i} className="border border-border-subtle rounded-lg p-3 space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Badge variant="neutral" size="sm">{ch.type}</Badge>
                    {ch.channel && <span className="text-sm font-mono text-text-primary">{ch.channel}</span>}
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleTest({ type: ch.type, secretRef: ch.secretRef })}
                    disabled={testing === ch.type}
                  >
                    <Send className="h-3 w-3" />
                    {testing === ch.type ? 'Sending...' : 'Test'}
                  </Button>
                </div>
                {ch.secretRef && (
                  <div className="flex items-center gap-1.5 text-xs">
                    {ch.secretExists ? (
                      <><CheckCircle className="h-3 w-3 text-status-success" /><span className="text-text-secondary font-mono">{ch.secretRef}</span></>
                    ) : (
                      <><XCircle className="h-3 w-3 text-status-danger" /><span className="text-status-danger font-mono">{ch.secretRef} (missing)</span></>
                    )}
                  </div>
                )}
                {ch.usedByTests && ch.usedByTests.length > 0 && (
                  <div className="text-2xs text-text-tertiary">
                    Used by: <span className="font-mono">{ch.usedByTests.join(', ')}</span>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </CardBody>
    </Card>
  );
}

// ── API Tab ──

function ApiTab() {
  const [token, setToken] = useState('');
  const [showToken, setShowToken] = useState(false);
  const [storedToken, setStoredToken] = useState<string | null>(null);
  const [justSaved, setJustSaved] = useState(false);

  useEffect(() => { setStoredToken(getAuthToken()); }, []);

  const hasToken = !!storedToken;

  const onSave = () => {
    if (!token.trim()) return;
    setAuthToken(token.trim());
    setStoredToken(token.trim());
    setToken('');
    setJustSaved(true);
    setTimeout(() => setJustSaved(false), 3000);
  };

  const onClear = () => {
    setAuthToken(null);
    setStoredToken(null);
    setToken('');
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Key className="h-4 w-4 text-text-tertiary" />
          <CardTitle>API Authentication</CardTitle>
        </div>
        <Badge variant={hasToken ? 'success' : 'warning'} dot>
          {hasToken ? 'Token configured' : 'No token set'}
        </Badge>
      </CardHeader>
      <CardBody className="space-y-4">
        {justSaved && (
          <Alert variant="success">Token saved. Write actions will now work.</Alert>
        )}

        {!hasToken && (
          <Alert variant="warning" title="Write actions are disabled">
            Without an API token, actions like Trigger, Create, Edit and Delete will return HTTP 401.
          </Alert>
        )}

        <div>
          <label className="text-sm font-medium text-text-primary block mb-1.5">Bearer token</label>
          <div className="flex gap-2">
            <input
              type={showToken ? 'text' : 'password'}
              value={token}
              onChange={(e) => setToken(e.target.value)}
              placeholder={hasToken ? '\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022 (token set)' : 'Paste your token here'}
              className="form-input font-mono flex-1"
            />
            <Button variant="secondary" onClick={() => setShowToken((v) => !v)} disabled={!token}>
              {showToken ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
            </Button>
            <Button variant="primary" onClick={onSave} disabled={!token.trim()}>
              <Save className="h-3.5 w-3.5" /> Save
            </Button>
            {hasToken && (
              <Button variant="danger" onClick={onClear}>
                <Trash2 className="h-3.5 w-3.5" /> Clear
              </Button>
            )}
          </div>
          <p className="text-2xs text-text-tertiary mt-1.5">Stored in your browser only. Never sent anywhere except the Kymaros API.</p>
        </div>

        <div>
          <div className="text-2xs uppercase tracking-wider text-text-tertiary mb-2">How to retrieve your token</div>
          <CodeBlock>{`kubectl get secret kymaros-api-token -n kymaros-system -o jsonpath='{.data.token}' | base64 -d`}</CodeBlock>
        </div>
      </CardBody>
    </Card>
  );
}

// ── About Tab ──

function AboutTab() {
  const [health, setHealth] = useState<HealthResponse | null>(null);
  useEffect(() => { kymarosApi.getHealth().then(setHealth).catch(() => {}); }, []);

  if (!health) return <SectionSkeleton />;

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Server className="h-4 w-4 text-text-tertiary" />
            <CardTitle>System</CardTitle>
          </div>
          <Badge variant={health.status === 'ok' ? 'success' : 'danger'} dot>{health.status}</Badge>
        </CardHeader>
        <CardBody>
          <dl className="grid grid-cols-2 gap-x-6 gap-y-3">
            <div>
              <dt className="text-2xs uppercase tracking-wider text-text-tertiary">Kymaros version</dt>
              <dd className="text-sm text-text-primary font-mono mt-0.5">{health.version}</dd>
            </div>
            <div>
              <dt className="text-2xs uppercase tracking-wider text-text-tertiary">Kubernetes version</dt>
              <dd className="text-sm text-text-primary font-mono mt-0.5">{health.kubernetesVersion}</dd>
            </div>
            <div>
              <dt className="text-2xs uppercase tracking-wider text-text-tertiary">Namespace</dt>
              <dd className="text-sm text-text-primary font-mono mt-0.5">{health.namespace}</dd>
            </div>
            <div>
              <dt className="text-2xs uppercase tracking-wider text-text-tertiary">Uptime</dt>
              <dd className="text-sm text-text-primary font-mono mt-0.5">{health.uptime}</dd>
            </div>
            <div>
              <dt className="text-2xs uppercase tracking-wider text-text-tertiary">Pod</dt>
              <dd className="text-sm text-text-primary font-mono mt-0.5">{health.pod}</dd>
            </div>
            <div>
              <dt className="text-2xs uppercase tracking-wider text-text-tertiary">Prometheus</dt>
              <dd className="text-sm text-text-primary font-mono mt-0.5">:8443/metrics</dd>
            </div>
          </dl>
        </CardBody>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Info className="h-4 w-4 text-text-tertiary" />
            <CardTitle>Resources</CardTitle>
          </div>
        </CardHeader>
        <CardBody className="space-y-0.5">
          <ResourceLink icon={<BookOpen className="h-4 w-4" />} href="https://docs.kymaros.io" title="Documentation" />
          <ResourceLink icon={<ExternalLink className="h-4 w-4" />} href="https://github.com/kymaroshq/kymaros" title="GitHub" />
          <ResourceLink icon={<ExternalLink className="h-4 w-4" />} href="https://charts.kymaros.io" title="Helm Chart Repository" />
        </CardBody>
      </Card>

      <div className="text-center py-4">
        <p className="text-xs text-text-tertiary">Kymaros is open source software released under the Apache 2.0 license.</p>
      </div>
    </div>
  );
}

function ResourceLink({ icon, href, title }: { icon: React.ReactNode; href: string; title: string }) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className="flex items-center gap-3 px-3 py-2 rounded-md hover:bg-surface-3 transition-colors duration-150 group"
    >
      <div className="text-text-tertiary">{icon}</div>
      <span className="text-sm text-text-primary flex-1">{title}</span>
      <ExternalLink className="h-3 w-3 text-text-tertiary opacity-0 group-hover:opacity-100 transition-opacity duration-150" />
    </a>
  );
}
