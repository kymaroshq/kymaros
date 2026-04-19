import { useState, useEffect, useRef, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { Home, ListChecks, FileText, Settings, BookOpen, ExternalLink, Sun, Moon, Key } from 'lucide-react';
import { useThemeStore } from '@/hooks/useThemeStore';
import { useTests } from '@/hooks/useKymarosData';
import { cn } from '@/lib/utils';

interface PaletteItem {
  id: string;
  label: string;
  icon: React.ReactNode;
  group: string;
  action: () => void;
  keywords?: string;
}

interface CommandPaletteProps {
  open: boolean;
  onClose: () => void;
}

export function CommandPalette({ open, onClose }: CommandPaletteProps) {
  const navigate = useNavigate();
  const { theme, toggle: toggleTheme } = useThemeStore();
  const tests = useTests();
  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const go = (path: string) => { navigate(path); onClose(); };

  const items: PaletteItem[] = useMemo(() => {
    const all: PaletteItem[] = [
      { id: 'nav-dash', label: 'Go to Dashboard', icon: <Home className="h-3.5 w-3.5" />, group: 'Navigation', action: () => go('/'), keywords: 'home overview' },
      { id: 'nav-tests', label: 'Go to Tests', icon: <ListChecks className="h-3.5 w-3.5" />, group: 'Navigation', action: () => go('/tests'), keywords: 'restore test' },
      { id: 'nav-reports', label: 'Go to Reports', icon: <FileText className="h-3.5 w-3.5" />, group: 'Navigation', action: () => go('/reports'), keywords: 'history runs' },
      { id: 'nav-settings', label: 'Go to Settings', icon: <Settings className="h-3.5 w-3.5" />, group: 'Navigation', action: () => go('/settings'), keywords: 'config' },
      { id: 'act-theme', label: `Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`, icon: theme === 'dark' ? <Sun className="h-3.5 w-3.5" /> : <Moon className="h-3.5 w-3.5" />, group: 'Actions', action: () => { toggleTheme(); onClose(); } },
      { id: 'act-token', label: 'Set API token', icon: <Key className="h-3.5 w-3.5" />, group: 'Actions', action: () => go('/settings?tab=api') },
      { id: 'res-docs', label: 'Open documentation', icon: <BookOpen className="h-3.5 w-3.5" />, group: 'Resources', action: () => { window.open('https://docs.kymaros.io', '_blank'); onClose(); } },
      { id: 'res-github', label: 'Open GitHub', icon: <ExternalLink className="h-3.5 w-3.5" />, group: 'Resources', action: () => { window.open('https://github.com/kymaroshq/kymaros', '_blank'); onClose(); } },
    ];

    if (tests.data) {
      tests.data.slice(0, 8).forEach((t) => {
        all.push({
          id: `test-${t.name}`,
          label: t.name,
          icon: <ListChecks className="h-3.5 w-3.5" />,
          group: 'Tests',
          action: () => go(`/reports/${t.name}`),
          keywords: `${t.namespace} ${t.name}`,
        });
      });
    }

    return all;
  }, [theme, tests.data]);

  const filtered = useMemo(() => {
    if (!query) return items;
    const q = query.toLowerCase();
    return items.filter((item) =>
      item.label.toLowerCase().includes(q) || (item.keywords?.toLowerCase().includes(q))
    );
  }, [items, query]);

  // Reset on open
  useEffect(() => {
    if (open) {
      setQuery('');
      setSelected(0);
      setTimeout(() => inputRef.current?.focus(), 50);
    }
  }, [open]);

  // Keyboard nav
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'ArrowDown') { e.preventDefault(); setSelected((s) => Math.min(s + 1, filtered.length - 1)); }
      if (e.key === 'ArrowUp') { e.preventDefault(); setSelected((s) => Math.max(s - 1, 0)); }
      if (e.key === 'Enter' && filtered[selected]) { e.preventDefault(); filtered[selected].action(); }
      if (e.key === 'Escape') { e.preventDefault(); onClose(); }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open, filtered, selected, onClose]);

  if (!open) return null;

  // Group items
  const groups: { name: string; items: (PaletteItem & { globalIdx: number })[] }[] = [];
  let idx = 0;
  filtered.forEach((item) => {
    let group = groups.find((g) => g.name === item.group);
    if (!group) { group = { name: item.group, items: [] }; groups.push(group); }
    group.items.push({ ...item, globalIdx: idx++ });
  });

  return (
    <div className="fixed inset-0 z-[60] flex items-start justify-center pt-[15vh]">
      <div className="absolute inset-0 bg-black/70 animate-fade-in" onClick={onClose} />
      <div className="relative w-full max-w-xl mx-4 bg-surface-2 border border-border-default rounded-lg shadow-lg overflow-hidden animate-slide-up">
        <div className="border-b border-border-subtle px-3 py-2 flex items-center gap-2">
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => { setQuery(e.target.value); setSelected(0); }}
            placeholder="Type a command or search..."
            className="flex-1 bg-transparent text-sm text-text-primary placeholder:text-text-tertiary focus:outline-none py-1"
          />
          <kbd className="text-2xs font-mono text-text-tertiary border border-border-subtle rounded px-1.5 py-0.5">ESC</kbd>
        </div>

        <div className="max-h-80 overflow-y-auto py-1">
          {filtered.length === 0 ? (
            <div className="py-6 text-center text-sm text-text-tertiary">No results found.</div>
          ) : (
            groups.map((group) => (
              <div key={group.name}>
                <div className="px-3 pt-2 pb-1 text-2xs font-medium uppercase tracking-wider text-text-tertiary">{group.name}</div>
                {group.items.map((item) => (
                  <button
                    key={item.id}
                    onClick={item.action}
                    onMouseEnter={() => setSelected(item.globalIdx)}
                    className={cn(
                      'w-full flex items-center gap-2.5 px-3 py-1.5 text-sm text-left transition-colors duration-100',
                      item.globalIdx === selected ? 'bg-surface-4 text-text-primary' : 'text-text-secondary'
                    )}
                  >
                    <span className="text-text-tertiary">{item.icon}</span>
                    <span className="flex-1">{item.label}</span>
                  </button>
                ))}
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
