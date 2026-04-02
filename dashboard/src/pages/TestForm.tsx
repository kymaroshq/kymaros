import { useState, useEffect, useCallback, type FormEvent } from 'react';
import { useNavigate, useParams, Link } from 'react-router-dom';
import { ArrowLeft, Plus, X, Loader2 } from 'lucide-react';
import { kymarosApi } from '../api/kymarosApi';
import type { CreateTestInput, TestResponse } from '../types/kymaros';

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const TIMEZONES = [
  'UTC',
  'Europe/Brussels',
  'Europe/London',
  'Europe/Paris',
  'Europe/Berlin',
  'US/Eastern',
  'US/Central',
  'US/Mountain',
  'US/Pacific',
  'Asia/Tokyo',
  'Asia/Shanghai',
] as const;

const NETWORK_ISOLATION_OPTIONS = ['strict', 'group'] as const;

// ---------------------------------------------------------------------------
// Cron preview helper
// ---------------------------------------------------------------------------

function describeCron(cron: string, timezone: string): string {
  const parts = cron.trim().split(/\s+/);
  if (parts.length !== 5) return 'Invalid cron expression';

  const [minute, hour, dayOfMonth, month, dayOfWeek] = parts;

  const isWildcard = (v: string): boolean => v === '*';
  const isNumber = (v: string): boolean => /^\d+$/.test(v);

  if (
    isNumber(minute) &&
    isNumber(hour) &&
    isWildcard(dayOfMonth) &&
    isWildcard(month) &&
    isWildcard(dayOfWeek)
  ) {
    return `Runs every day at ${hour.padStart(2, '0')}:${minute.padStart(2, '0')} ${timezone}`;
  }

  if (
    isNumber(minute) &&
    isNumber(hour) &&
    isWildcard(dayOfMonth) &&
    isWildcard(month) &&
    isNumber(dayOfWeek)
  ) {
    const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
    const dayName = days[parseInt(dayOfWeek, 10)] ?? dayOfWeek;
    return `Runs every ${dayName} at ${hour.padStart(2, '0')}:${minute.padStart(2, '0')} ${timezone}`;
  }

  return `Schedule: ${cron} (${timezone})`;
}

// ---------------------------------------------------------------------------
// Form state type
// ---------------------------------------------------------------------------

interface FormState {
  name: string;
  provider: string;
  backupName: string;
  namespaces: string[];
  cron: string;
  timezone: string;
  sandboxPrefix: string;
  ttl: string;
  networkIsolation: string;
  quotaCpu: string;
  quotaMemory: string;
  healthCheckRef: string;
  healthCheckTimeout: string;
  maxRTO: string;
  alertOnExceed: boolean;
}

const DEFAULT_FORM: FormState = {
  name: '',
  provider: 'velero',
  backupName: 'latest',
  namespaces: [''],
  cron: '0 3 * * *',
  timezone: 'UTC',
  sandboxPrefix: 'kymaros-test',
  ttl: '30m0s',
  networkIsolation: 'strict',
  quotaCpu: '',
  quotaMemory: '',
  healthCheckRef: '',
  healthCheckTimeout: '10m0s',
  maxRTO: '',
  alertOnExceed: false,
};

// ---------------------------------------------------------------------------
// Helpers to map API response to form state
// ---------------------------------------------------------------------------

function testResponseToForm(t: TestResponse): FormState {
  return {
    name: t.name,
    provider: t.provider,
    backupName: 'latest',
    namespaces: t.sourceNamespaces.length > 0 ? t.sourceNamespaces : [''],
    cron: t.schedule,
    timezone: 'UTC',
    sandboxPrefix: 'kymaros-test',
    ttl: '30m0s',
    networkIsolation: 'strict',
    quotaCpu: '',
    quotaMemory: '',
    healthCheckRef: '',
    healthCheckTimeout: '10m0s',
    maxRTO: t.rtoTarget ?? '',
    alertOnExceed: false,
  };
}

function formToInput(form: FormState): CreateTestInput {
  return {
    name: form.name,
    provider: form.provider,
    backupName: form.backupName,
    namespaces: form.namespaces.filter((ns) => ns.trim() !== ''),
    cron: form.cron,
    timezone: form.timezone,
    sandboxPrefix: form.sandboxPrefix,
    ttl: form.ttl,
    networkIsolation: form.networkIsolation,
    quotaCpu: form.quotaCpu,
    quotaMemory: form.quotaMemory,
    healthCheckRef: form.healthCheckRef,
    healthCheckTimeout: form.healthCheckTimeout,
    maxRTO: form.maxRTO,
    alertOnExceed: form.alertOnExceed,
  };
}

// ---------------------------------------------------------------------------
// Reusable field components
// ---------------------------------------------------------------------------

interface InputFieldProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  required?: boolean;
  disabled?: boolean;
  type?: string;
}

function InputField({ label, value, onChange, placeholder, required, disabled, type = 'text' }: InputFieldProps) {
  return (
    <label className="block">
      <span className="mb-1 block text-sm font-medium text-slate-300">
        {label}
        {required && <span className="ml-0.5 text-red-400">*</span>}
      </span>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        required={required}
        disabled={disabled}
        className="w-full rounded-lg border border-navy-600 bg-navy-700 px-3 py-2 text-sm text-white placeholder:text-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:cursor-not-allowed disabled:opacity-50"
      />
    </label>
  );
}

interface SelectFieldProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
  options: ReadonlyArray<{ value: string; label: string; disabled?: boolean }>;
  required?: boolean;
}

function SelectField({ label, value, onChange, options, required }: SelectFieldProps) {
  return (
    <label className="block">
      <span className="mb-1 block text-sm font-medium text-slate-300">
        {label}
        {required && <span className="ml-0.5 text-red-400">*</span>}
      </span>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        required={required}
        className="w-full rounded-lg border border-navy-600 bg-navy-700 px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value} disabled={opt.disabled}>
            {opt.label}
          </option>
        ))}
      </select>
    </label>
  );
}

function SectionCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-navy-700 bg-navy-800 p-5">
      <h3 className="mb-4 text-base font-semibold text-white">{title}</h3>
      {children}
    </div>
  );
}

// ---------------------------------------------------------------------------
// TestForm Page
// ---------------------------------------------------------------------------

function TestForm() {
  const navigate = useNavigate();
  const { name: editName } = useParams<{ name: string }>();
  const isEditMode = editName !== undefined;

  const [form, setForm] = useState<FormState>(DEFAULT_FORM);
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [loadingTest, setLoadingTest] = useState(isEditMode);
  const [loadError, setLoadError] = useState<string | null>(null);

  // Load existing test data in edit mode
  useEffect(() => {
    if (!isEditMode) return;

    let cancelled = false;

    async function loadTest() {
      try {
        setLoadingTest(true);
        setLoadError(null);
        const tests = await kymarosApi.getTests();
        const match = tests.find((t) => t.name === editName);
        if (!match) {
          if (!cancelled) setLoadError(`Test "${editName}" not found`);
          return;
        }
        if (!cancelled) setForm(testResponseToForm(match));
      } catch (err) {
        if (!cancelled) {
          setLoadError(err instanceof Error ? err.message : 'Failed to load test');
        }
      } finally {
        if (!cancelled) setLoadingTest(false);
      }
    }

    void loadTest();
    return () => { cancelled = true; };
  }, [isEditMode, editName]);

  const updateField = useCallback(<K extends keyof FormState>(key: K, value: FormState[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  }, []);

  const addNamespace = useCallback(() => {
    setForm((prev) => ({ ...prev, namespaces: [...prev.namespaces, ''] }));
  }, []);

  const removeNamespace = useCallback((index: number) => {
    setForm((prev) => {
      const updated = prev.namespaces.filter((_, i) => i !== index);
      return { ...prev, namespaces: updated.length > 0 ? updated : [''] };
    });
  }, []);

  const updateNamespace = useCallback((index: number, value: string) => {
    setForm((prev) => {
      const updated = [...prev.namespaces];
      updated[index] = value;
      return { ...prev, namespaces: updated };
    });
  }, []);

  const handleSubmit = useCallback(async (e: FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setSubmitError(null);

    try {
      const input = formToInput(form);
      if (isEditMode && editName) {
        await kymarosApi.updateTest(editName, input);
      } else {
        await kymarosApi.createTest(input);
      }
      navigate('/');
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : 'Submission failed');
    } finally {
      setSubmitting(false);
    }
  }, [form, isEditMode, editName, navigate]);

  // Loading state for edit mode
  if (loadingTest) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-blue-400" />
        <span className="ml-2 text-sm text-slate-400">Loading test...</span>
      </div>
    );
  }

  // Error loading test
  if (loadError) {
    return (
      <div className="space-y-4 py-10 text-center">
        <p className="text-sm text-red-400">{loadError}</p>
        <Link
          to="/"
          className="inline-flex items-center gap-1.5 text-sm text-blue-400 hover:text-blue-300"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Dashboard
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Link
          to="/"
          className="rounded-lg border border-navy-600 p-2 text-slate-400 transition-colors hover:border-navy-500 hover:text-white"
        >
          <ArrowLeft className="h-4 w-4" />
        </Link>
        <h1 className="text-2xl font-bold text-white">
          {isEditMode ? `Edit ${editName}` : 'Create Restore Test'}
        </h1>
      </div>

      <form onSubmit={handleSubmit} className="space-y-5">
        {/* Section 1: Basic Info */}
        <SectionCard title="Basic Info">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <InputField
              label="Name"
              value={form.name}
              onChange={(v) => updateField('name', v)}
              placeholder="e.g. restore-production"
              required
              disabled={isEditMode}
            />
            <SelectField
              label="Backup Provider"
              value={form.provider}
              onChange={(v) => updateField('provider', v)}
              required
              options={[
                { value: 'velero', label: 'Velero' },
                { value: 'restic', label: 'Restic', disabled: true },
                { value: 'kopia', label: 'Kopia', disabled: true },
              ]}
            />
            <InputField
              label="Backup Name"
              value={form.backupName}
              onChange={(v) => updateField('backupName', v)}
              placeholder="latest"
            />
          </div>
        </SectionCard>

        {/* Section 2: Source Namespaces */}
        <SectionCard title="Source Namespaces">
          <div className="space-y-2">
            {form.namespaces.map((ns, idx) => (
              <div key={idx} className="flex items-center gap-2">
                <input
                  type="text"
                  value={ns}
                  onChange={(e) => updateNamespace(idx, e.target.value)}
                  placeholder="e.g. production"
                  required
                  className="flex-1 rounded-lg border border-navy-600 bg-navy-700 px-3 py-2 text-sm text-white placeholder:text-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
                {form.namespaces.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removeNamespace(idx)}
                    className="rounded-lg border border-navy-600 p-2 text-slate-400 transition-colors hover:border-red-500 hover:text-red-400"
                  >
                    <X className="h-4 w-4" />
                  </button>
                )}
              </div>
            ))}
            <button
              type="button"
              onClick={addNamespace}
              className="inline-flex items-center gap-1.5 rounded-lg border border-dashed border-navy-600 px-3 py-1.5 text-xs text-slate-400 transition-colors hover:border-blue-500 hover:text-blue-400"
            >
              <Plus className="h-3.5 w-3.5" />
              Add namespace
            </button>
          </div>
        </SectionCard>

        {/* Section 3: Schedule */}
        <SectionCard title="Schedule">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <InputField
              label="Cron Expression"
              value={form.cron}
              onChange={(v) => updateField('cron', v)}
              placeholder="0 3 * * *"
              required
            />
            <SelectField
              label="Timezone"
              value={form.timezone}
              onChange={(v) => updateField('timezone', v)}
              options={TIMEZONES.map((tz) => ({ value: tz, label: tz }))}
            />
          </div>
          <p className="mt-3 text-xs text-slate-400">
            {describeCron(form.cron, form.timezone)}
          </p>
        </SectionCard>

        {/* Section 4: Sandbox Configuration */}
        <SectionCard title="Sandbox Configuration">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <InputField
              label="Prefix"
              value={form.sandboxPrefix}
              onChange={(v) => updateField('sandboxPrefix', v)}
              placeholder="kymaros-test"
            />
            <InputField
              label="TTL"
              value={form.ttl}
              onChange={(v) => updateField('ttl', v)}
              placeholder="30m0s"
            />
            <SelectField
              label="Network Isolation"
              value={form.networkIsolation}
              onChange={(v) => updateField('networkIsolation', v)}
              options={NETWORK_ISOLATION_OPTIONS.map((opt) => ({
                value: opt,
                label: opt.charAt(0).toUpperCase() + opt.slice(1),
              }))}
            />
            <div />
            <InputField
              label="CPU Quota"
              value={form.quotaCpu}
              onChange={(v) => updateField('quotaCpu', v)}
              placeholder="e.g. 2"
            />
            <InputField
              label="Memory Quota"
              value={form.quotaMemory}
              onChange={(v) => updateField('quotaMemory', v)}
              placeholder="e.g. 4Gi"
            />
          </div>
        </SectionCard>

        {/* Section 5: Health Checks */}
        <SectionCard title="Health Checks">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <InputField
              label="Policy Reference"
              value={form.healthCheckRef}
              onChange={(v) => updateField('healthCheckRef', v)}
              placeholder="Name of a HealthCheckPolicy"
            />
            <InputField
              label="Timeout"
              value={form.healthCheckTimeout}
              onChange={(v) => updateField('healthCheckTimeout', v)}
              placeholder="10m0s"
            />
          </div>
        </SectionCard>

        {/* Section 6: SLA */}
        <SectionCard title="SLA">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <InputField
              label="Max RTO"
              value={form.maxRTO}
              onChange={(v) => updateField('maxRTO', v)}
              placeholder="e.g. 15m"
            />
            <label className="flex items-center gap-3 self-end pb-1">
              <input
                type="checkbox"
                checked={form.alertOnExceed}
                onChange={(e) => updateField('alertOnExceed', e.target.checked)}
                className="h-4 w-4 rounded border-navy-600 bg-navy-700 text-blue-500 focus:ring-blue-500 focus:ring-offset-0"
              />
              <span className="text-sm text-slate-300">Alert on exceed</span>
            </label>
          </div>
        </SectionCard>

        {/* Submit error */}
        {submitError && (
          <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
            {submitError}
          </div>
        )}

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 pt-2">
          <button
            type="button"
            onClick={() => navigate('/')}
            className="rounded-lg border border-navy-600 px-4 py-2 text-sm font-medium text-slate-300 transition-colors hover:border-navy-500 hover:text-white"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={submitting}
            className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
            {isEditMode ? 'Save Changes' : 'Create'}
          </button>
        </div>
      </form>
    </div>
  );
}

export default TestForm;
