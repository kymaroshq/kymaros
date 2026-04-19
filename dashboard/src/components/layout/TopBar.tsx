import { useLocation, Link } from 'react-router-dom';
import { Search, Bell } from 'lucide-react';
import { Badge } from '@/components/ui2/Badge';
import { ThemeToggle } from './ThemeToggle';

interface TopBarProps {
  onOpenPalette: () => void;
  onOpenNotifs: () => void;
  unreadCount: number;
}

function useBreadcrumbs() {
  const { pathname } = useLocation();
  const segments = pathname.split('/').filter(Boolean);
  const crumbs: { label: string; href: string }[] = [{ label: 'Dashboard', href: '/' }];
  segments.forEach((seg, i) => {
    crumbs.push({
      label: seg.charAt(0).toUpperCase() + seg.slice(1).replace(/-/g, ' '),
      href: '/' + segments.slice(0, i + 1).join('/'),
    });
  });
  return crumbs;
}

export function TopBar({ onOpenPalette, onOpenNotifs, unreadCount }: TopBarProps) {
  const crumbs = useBreadcrumbs();

  return (
    <header className="h-12 bg-surface-0 border-b border-border-subtle px-4 flex items-center justify-between gap-4 sticky top-0 z-20">
      <nav className="flex items-center gap-1.5 text-sm min-w-0">
        {crumbs.map((c, i) => (
          <div key={c.href} className="flex items-center gap-1.5">
            {i > 0 && <span className="text-text-tertiary">/</span>}
            {i === crumbs.length - 1 ? (
              <span className="text-text-primary font-medium truncate">{c.label}</span>
            ) : (
              <Link to={c.href} className="text-text-tertiary hover:text-text-primary transition-colors duration-150 truncate">{c.label}</Link>
            )}
          </div>
        ))}
        <span className="ml-2"><Badge variant="success" dot size="sm">Live</Badge></span>
      </nav>

      <div className="flex items-center gap-1">
        <button
          onClick={onOpenPalette}
          className="flex items-center gap-2 px-2.5 py-1 bg-surface-2 border border-border-default rounded-md text-xs text-text-tertiary hover:bg-surface-3 transition-colors duration-150"
        >
          <Search className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">Search</span>
          <kbd className="text-2xs font-mono border border-border-subtle rounded px-1 py-0.5 ml-2">⌘K</kbd>
        </button>

        <button
          onClick={onOpenNotifs}
          className="relative h-8 w-8 flex items-center justify-center rounded-md text-text-tertiary hover:bg-surface-2 hover:text-text-primary transition-colors duration-150"
        >
          <Bell className="h-4 w-4" />
          {unreadCount > 0 && <span className="absolute top-1.5 right-1.5 h-1.5 w-1.5 rounded-full bg-status-danger" />}
        </button>

        <ThemeToggle />
      </div>
    </header>
  );
}
