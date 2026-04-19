import { useState, useMemo, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card } from '@/components/ui2/Card';
import { DataTable } from '@/components/ui2/DataTable';
import { Badge } from '@/components/ui2/Badge';
import { ScoreIndicator } from '@/components/ui2/ScoreIndicator';
import { Button } from '@/components/ui2/Button';
import { EmptyState } from '@/components/ui2/EmptyState';
import { TableSkeleton } from '@/components/ui2/Skeleton';
import { Alert } from '@/components/ui2/Alert';
import { useTests } from '@/hooks/useKymarosData';
import { kymarosApi } from '@/api/kymarosApi';
import { formatRelativeTime, formatCron, getResultVariant, cn, jsonToYaml } from '@/lib/utils';
import { toast } from '@/lib/toast';
import { pushNotification } from '@/hooks/useNotifications';
import { Search, Plus, ListChecks, X, Play, MoreVertical, FileCode } from 'lucide-react';
import { CodeBlock } from '@/components/ui2/CodeBlock';
import type { TestResponse, CreateTestInput } from '@/types/kymaros';

export default function Tests() {
  const navigate = useNavigate();
  const tests = useTests();
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'pass' | 'partial' | 'fail'>('all');
  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<TestResponse | null>(null);
  const [deleting, setDeleting] = useState<string | null>(null);
  const [triggering, setTriggering] = useState<string | null>(null);
  const [viewYamlOf, setViewYamlOf] = useState<TestResponse | null>(null);
  const [actionMsg, setActionMsg] = useState<{ type: 'success' | 'danger'; text: string } | null>(null);

  const filtered = useMemo(() => {
    if (!tests.data) return [];
    return tests.data.filter((t) => {
      if (search) {
        const q = search.toLowerCase();
        if (!t.name.toLowerCase().includes(q) && !t.namespace.toLowerCase().includes(q)) return false;
      }
      if (statusFilter !== 'all' && t.lastResult !== statusFilter) return false;
      return true;
    });
  }, [tests.data, search, statusFilter]);

  const handleTrigger = useCallback(async (name: string) => {
    setTriggering(name);
    try {
      await kymarosApi.triggerTest(name);
      toast.success('Test triggered', `${name} is now running`);
      pushNotification({ type: 'info', title: 'Test triggered', description: name, href: `/reports/${name}` });
      tests.refetch();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Unknown error';
      toast.error('Failed to trigger test', msg);
      pushNotification({ type: 'error', title: `Failed to trigger ${name}`, description: msg });
    } finally {
      setTriggering(null);
    }
  }, [tests]);

  const handleDelete = useCallback(async (name: string) => {
    try {
      await kymarosApi.deleteTest(name);
      setDeleting(null);
      toast.success('Test deleted', name);
      pushNotification({ type: 'success', title: 'Test deleted', description: name });
      tests.refetch();
    } catch (err: unknown) {
      toast.error('Failed to delete test', err instanceof Error ? err.message : 'Unknown error');
    }
  }, [tests]);

  const hasActiveFilters = search !== '' || statusFilter !== 'all';

  return (
    <div className="px-6 py-6 space-y-6 max-w-[1600px] mx-auto">
      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Tests</h1>
          <p className="text-sm text-text-tertiary mt-0.5">
            {tests.data ? `${tests.data.length} restore test${tests.data.length !== 1 ? 's' : ''} configured` : 'Manage restore test definitions'}
          </p>
        </div>
        <Button variant="primary" onClick={() => setCreating(true)}>
          <Plus className="h-3.5 w-3.5" />
          New test
        </Button>
      </div>

      {actionMsg && <Alert variant={actionMsg.type}>{actionMsg.text}</Alert>}

      {/* Table */}
      <Card>
        {/* Filters */}
        <div className="flex flex-wrap items-center gap-2 px-4 py-3 border-b border-border-subtle">
          <div className="relative flex-1 min-w-48 max-w-xs">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-text-tertiary pointer-events-none" />
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search test name..."
              className="w-full h-9 rounded-md border border-border-default bg-surface-2 pl-8 pr-3 text-sm text-text-primary placeholder:text-text-tertiary focus:outline-none focus:ring-1 focus:ring-accent focus:border-accent transition-colors duration-150"
            />
          </div>
          <div className="flex items-center gap-1 bg-surface-3 rounded-md p-0.5">
            {(['all', 'pass', 'partial', 'fail'] as const).map((s) => (
              <button
                key={s}
                onClick={() => setStatusFilter(s)}
                className={cn(
                  'text-xs font-medium px-2 py-1 rounded transition-colors duration-150',
                  statusFilter === s ? 'bg-surface-1 text-text-primary' : 'text-text-tertiary hover:text-text-primary'
                )}
              >
                {s === 'all' ? 'All' : s}
              </button>
            ))}
          </div>
          {hasActiveFilters && (
            <Button variant="ghost" size="sm" onClick={() => { setSearch(''); setStatusFilter('all'); }}>
              <X className="h-3 w-3" /> Clear
            </Button>
          )}
        </div>

        {/* Content */}
        {tests.loading ? (
          <TableSkeleton rows={6} />
        ) : filtered.length === 0 ? (
          hasActiveFilters ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <Search className="h-6 w-6 text-text-tertiary mb-3" />
              <h3 className="text-sm font-medium text-text-primary">No tests match your filters</h3>
              <Button variant="secondary" onClick={() => { setSearch(''); setStatusFilter('all'); }} className="mt-4">Clear filters</Button>
            </div>
          ) : (
            <EmptyState
              icon={<ListChecks className="h-6 w-6" />}
              title="No tests configured yet"
              description="Create your first RestoreTest to start validating backups."
              action={<Button variant="primary" onClick={() => setCreating(true)}><Plus className="h-3.5 w-3.5" /> Create test</Button>}
            />
          )
        ) : (
          <DataTable<TestResponse>
            columns={[
              {
                key: 'name',
                label: 'Test',
                render: (row) => (
                  <div>
                    <div className="font-medium text-text-primary text-sm">{row.name}</div>
                    <div className="text-2xs text-text-tertiary font-mono">{row.namespace}</div>
                  </div>
                ),
              },
              {
                key: 'schedule',
                label: 'Schedule',
                render: (row) => <span className="text-xs text-text-secondary">{formatCron(row.schedule)}</span>,
              },
              {
                key: 'lastRunAt',
                label: 'Last Run',
                align: 'right',
                render: (row) => (
                  <span className="text-xs text-text-tertiary font-mono">
                    {row.lastRunAt ? formatRelativeTime(row.lastRunAt) : 'Never'}
                  </span>
                ),
              },
              {
                key: 'lastScore',
                label: 'Score',
                align: 'right',
                render: (row) => row.lastScore > 0 ? (
                  <div className="inline-flex items-center gap-2">
                    <ScoreIndicator value={row.lastScore} size="sm" />
                    <Badge variant={getResultVariant(row.lastResult)} dot size="sm">{row.lastResult}</Badge>
                  </div>
                ) : (
                  <span className="text-xs text-text-tertiary">&mdash;</span>
                ),
              },
              {
                key: 'actions',
                label: '',
                align: 'right',
                width: '80px',
                render: (row) => (
                  <div className="flex items-center justify-end gap-1" onClick={(e) => e.stopPropagation()}>
                    <button
                      onClick={() => handleTrigger(row.name)}
                      disabled={triggering === row.name}
                      className="h-7 w-7 flex items-center justify-center rounded-md text-text-tertiary hover:bg-surface-3 hover:text-text-primary transition-colors duration-150 disabled:opacity-50"
                      title="Trigger now"
                    >
                      <Play className={cn('h-3.5 w-3.5', triggering === row.name && 'animate-spin')} />
                    </button>
                    <RowMenu
                      onEdit={() => setEditing(row)}
                      onDelete={() => setDeleting(row.name)}
                      onViewYaml={() => setViewYamlOf(row)}
                      onViewHistory={() => navigate(`/reports/${row.name}`)}
                    />
                  </div>
                ),
              },
            ]}
            rows={filtered}
            keyField="name"
            onRowClick={(row) => navigate(`/reports/${row.name}`)}
            compact
          />
        )}
      </Card>

      {/* Delete confirmation */}
      {deleting && (
        <DeleteDialog name={deleting} onConfirm={() => handleDelete(deleting)} onCancel={() => setDeleting(null)} />
      )}

      {/* View YAML dialog */}
      {viewYamlOf && (
        <div className="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto py-8">
          <div className="absolute inset-0 bg-black/70" onClick={() => setViewYamlOf(null)} />
          <div className="relative bg-surface-2 border border-border-default rounded-lg shadow-lg p-5 w-full max-w-2xl animate-fade-in">
            <h2 className="text-base font-semibold text-text-primary">{viewYamlOf.name}</h2>
            <p className="text-sm text-text-tertiary mt-1">RestoreTest manifest</p>
            <div className="mt-4">
              <CodeBlock maxLines={24}>{jsonToYaml(viewYamlOf)}</CodeBlock>
            </div>
            <div className="mt-5 flex justify-end">
              <Button variant="secondary" onClick={() => setViewYamlOf(null)}>Close</Button>
            </div>
          </div>
        </div>
      )}

      {/* Create/Edit dialog */}
      {(creating || editing) && (
        <TestFormDialog
          test={editing}
          onClose={() => { setCreating(false); setEditing(null); }}
          onSaved={() => { setCreating(false); setEditing(null); tests.refetch(); setActionMsg({ type: 'success', text: editing ? 'Test updated.' : 'Test created.' }); setTimeout(() => setActionMsg(null), 3000); }}
        />
      )}
    </div>
  );
}

// ── Row kebab menu (native) ──

function RowMenu({ onEdit, onDelete, onViewYaml, onViewHistory }: { onEdit: () => void; onDelete: () => void; onViewYaml: () => void; onViewHistory: () => void }) {
  const [open, setOpen] = useState(false);
  return (
    <div className="relative">
      <button
        onClick={() => setOpen((v) => !v)}
        className="h-7 w-7 flex items-center justify-center rounded-md text-text-tertiary hover:bg-surface-3 hover:text-text-primary transition-colors duration-150"
      >
        <MoreVertical className="h-3.5 w-3.5" />
      </button>
      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div className="absolute right-0 top-full mt-1 z-50 w-40 bg-surface-3 border border-border-default rounded-md shadow-lg py-1 animate-fade-in">
            <button onClick={() => { onViewHistory(); setOpen(false); }} className="w-full px-3 py-1.5 text-sm text-text-primary hover:bg-surface-4 text-left">View history</button>
            <button onClick={() => { onViewYaml(); setOpen(false); }} className="w-full px-3 py-1.5 text-sm text-text-primary hover:bg-surface-4 text-left flex items-center gap-2"><FileCode className="h-3.5 w-3.5 text-text-tertiary" />View YAML</button>
            <div className="my-1 h-px bg-border-subtle" />
            <button onClick={() => { onEdit(); setOpen(false); }} className="w-full px-3 py-1.5 text-sm text-text-primary hover:bg-surface-4 text-left">Edit</button>
            <div className="my-1 h-px bg-border-subtle" />
            <button onClick={() => { onDelete(); setOpen(false); }} className="w-full px-3 py-1.5 text-sm text-status-danger hover:bg-status-danger/10 text-left">Delete</button>
          </div>
        </>
      )}
    </div>
  );
}

// ── Delete dialog ──

function DeleteDialog({ name, onConfirm, onCancel }: { name: string; onConfirm: () => void; onCancel: () => void }) {
  const [confirm, setConfirm] = useState('');
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/70" onClick={onCancel} />
      <div className="relative bg-surface-2 border border-border-default rounded-lg shadow-lg p-5 w-full max-w-md animate-fade-in">
        <h2 className="text-base font-semibold text-text-primary">Delete test?</h2>
        <p className="text-sm text-text-tertiary mt-1">This permanently removes <span className="font-mono text-text-primary">{name}</span> and stops all scheduled runs.</p>
        <div className="mt-4 space-y-1.5">
          <label className="text-sm font-medium text-text-primary">Type <span className="font-mono">{name}</span> to confirm</label>
          <input
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            placeholder={name}
            className="w-full h-9 rounded-md border border-border-default bg-surface-2 px-3 text-sm font-mono text-text-primary placeholder:text-text-tertiary focus:outline-none focus:ring-1 focus:ring-accent"
          />
        </div>
        <div className="mt-5 flex justify-end gap-2">
          <Button variant="secondary" onClick={onCancel}>Cancel</Button>
          <Button variant="danger" onClick={onConfirm} disabled={confirm !== name}>Delete</Button>
        </div>
      </div>
    </div>
  );
}

// ── Create/Edit form dialog ──

function TestFormDialog({ test, onClose, onSaved }: { test: TestResponse | null; onClose: () => void; onSaved: () => void }) {
  const isEdit = !!test;
  const [form, setForm] = useState<CreateTestInput>({
    name: test?.name ?? '',
    provider: 'velero',
    backupName: 'latest',
    namespaces: test?.sourceNamespaces ?? [''],
    cron: test?.schedule ?? '0 3 * * *',
    timezone: 'UTC',
    sandboxPrefix: '',
    ttl: '30m',
    networkIsolation: 'strict',
    quotaCpu: '2',
    quotaMemory: '4Gi',
    healthCheckRef: '',
    healthCheckTimeout: '10m',
    maxRTO: '15m',
    alertOnExceed: true,
  });
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const set = (key: keyof CreateTestInput) => (value: string | boolean | string[]) => {
    setForm((f) => ({ ...f, [key]: value }));
  };

  const handleSubmit = async () => {
    setError(null);
    setSaving(true);
    try {
      if (isEdit) {
        await kymarosApi.updateTest(test!.name, form);
      } else {
        await kymarosApi.createTest(form);
      }
      onSaved();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto py-8">
      <div className="absolute inset-0 bg-black/70" onClick={onClose} />
      <div className="relative bg-surface-2 border border-border-default rounded-lg shadow-lg p-5 w-full max-w-lg animate-fade-in">
        <h2 className="text-base font-semibold text-text-primary">{isEdit ? `Edit ${test!.name}` : 'Create test'}</h2>
        <p className="text-sm text-text-tertiary mt-1">Define the restore validation parameters.</p>

        {error && <Alert variant="danger" className="mt-3">{error}</Alert>}

        <div className="mt-4 space-y-3 max-h-[60vh] overflow-y-auto">
          {!isEdit && (
            <Field label="Name" required>
              <input value={form.name} onChange={(e) => set('name')(e.target.value)} placeholder="my-app-test" className="form-input" />
            </Field>
          )}
          <Field label="Source namespaces" description="Comma-separated" required>
            <input value={form.namespaces.join(',')} onChange={(e) => set('namespaces')(e.target.value.split(',').map((s) => s.trim()).filter(Boolean))} placeholder="production,cache" className="form-input" />
          </Field>
          <div className="grid grid-cols-2 gap-3">
            <Field label="Schedule (cron)" required>
              <input value={form.cron} onChange={(e) => set('cron')(e.target.value)} placeholder="0 3 * * *" className="form-input font-mono" />
            </Field>
            <Field label="Timezone">
              <input value={form.timezone} onChange={(e) => set('timezone')(e.target.value)} placeholder="UTC" className="form-input" />
            </Field>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <Field label="Sandbox TTL">
              <input value={form.ttl} onChange={(e) => set('ttl')(e.target.value)} placeholder="30m" className="form-input font-mono" />
            </Field>
            <Field label="Network isolation">
              <select value={form.networkIsolation} onChange={(e) => set('networkIsolation')(e.target.value)} className="form-input">
                <option value="strict">Strict</option>
                <option value="group">Group</option>
              </select>
            </Field>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <Field label="CPU quota">
              <input value={form.quotaCpu} onChange={(e) => set('quotaCpu')(e.target.value)} placeholder="2" className="form-input font-mono" />
            </Field>
            <Field label="Memory quota">
              <input value={form.quotaMemory} onChange={(e) => set('quotaMemory')(e.target.value)} placeholder="4Gi" className="form-input font-mono" />
            </Field>
          </div>
          <Field label="Health check policy ref">
            <input value={form.healthCheckRef} onChange={(e) => set('healthCheckRef')(e.target.value)} placeholder="my-app-checks" className="form-input font-mono" />
          </Field>
          <div className="grid grid-cols-2 gap-3">
            <Field label="Max RTO">
              <input value={form.maxRTO} onChange={(e) => set('maxRTO')(e.target.value)} placeholder="15m" className="form-input font-mono" />
            </Field>
            <Field label="Health check timeout">
              <input value={form.healthCheckTimeout} onChange={(e) => set('healthCheckTimeout')(e.target.value)} placeholder="10m" className="form-input font-mono" />
            </Field>
          </div>
        </div>

        <div className="mt-5 flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button variant="primary" onClick={handleSubmit} disabled={saving || !form.name || form.namespaces.length === 0}>
            {saving ? 'Saving...' : isEdit ? 'Save changes' : 'Create'}
          </Button>
        </div>
      </div>
    </div>
  );
}

function Field({ label, description, required, children }: { label: string; description?: string; required?: boolean; children: React.ReactNode }) {
  return (
    <div className="space-y-1">
      <label className="text-sm font-medium text-text-primary">
        {label}{required && <span className="text-status-danger ml-0.5">*</span>}
      </label>
      {children}
      {description && <p className="text-2xs text-text-tertiary">{description}</p>}
    </div>
  );
}
