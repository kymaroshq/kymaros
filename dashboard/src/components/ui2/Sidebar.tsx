import { NavLink } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { Home, ListChecks, FileText, Settings, BookOpen, ExternalLink } from 'lucide-react';
import Logo from '@/components/ui/Logo';
import { useEffect, useState } from 'react';
import { kymarosApi } from '@/api/kymarosApi';

const navigation = [
  { name: 'Dashboard', href: '/', icon: Home },
  { name: 'Tests', href: '/tests', icon: ListChecks },
  { name: 'Reports', href: '/reports', icon: FileText },
  { name: 'Settings', href: '/settings', icon: Settings },
];

export function Sidebar() {
  const [version, setVersion] = useState('');

  useEffect(() => {
    kymarosApi.getHealth()
      .then((h: { version: string }) => setVersion(h.version))
      .catch(() => {});
  }, []);

  return (
    <aside className="w-52 bg-surface-1 border-r border-border-subtle flex flex-col h-screen sticky top-0 shrink-0">
      {/* Logo */}
      <div className="px-3 h-14 flex items-center gap-2 border-b border-border-subtle">
        <Logo size="sm" showText={true} />
        {version && (
          <span className="text-2xs text-text-tertiary font-mono ml-auto">v{version}</span>
        )}
      </div>

      {/* Nav */}
      <nav className="flex-1 p-2 space-y-0.5">
        {navigation.map((item) => (
          <NavLink
            key={item.name}
            to={item.href}
            end={item.href === '/'}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-2.5 px-2.5 py-1.5 rounded-md text-sm transition-colors duration-150',
                isActive
                  ? 'bg-surface-3 text-text-primary font-medium'
                  : 'text-text-secondary hover:bg-surface-2 hover:text-text-primary'
              )
            }
          >
            <item.icon className="h-4 w-4" />
            {item.name}
          </NavLink>
        ))}
      </nav>

      {/* Footer */}
      <div className="p-2 border-t border-border-subtle space-y-0.5">
        <a
          href="https://docs.kymaros.io"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-2.5 px-2.5 py-1.5 rounded-md text-sm text-text-tertiary hover:text-text-primary hover:bg-surface-2 transition-colors duration-150"
        >
          <BookOpen className="h-4 w-4" />
          Documentation
        </a>
        <a
          href="https://github.com/kymaroshq/kymaros"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-2.5 px-2.5 py-1.5 rounded-md text-sm text-text-tertiary hover:text-text-primary hover:bg-surface-2 transition-colors duration-150"
        >
          <ExternalLink className="h-4 w-4" />
          GitHub
        </a>
        <div className="px-2.5 py-1 text-2xs text-text-disabled">
          Apache 2.0
        </div>
      </div>
    </aside>
  );
}
