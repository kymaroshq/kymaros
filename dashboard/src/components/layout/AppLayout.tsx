import { useState, useEffect } from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from '@/components/ui2/Sidebar';
import { TopBar } from './TopBar';
import { Toaster } from './Toaster';
import { CommandPalette } from './CommandPalette';
import { NotificationsDrawer } from './NotificationsDrawer';
import { useNotifications } from '@/hooks/useNotifications';

export function AppLayout() {
  const [paletteOpen, setPaletteOpen] = useState(false);
  const [notifsOpen, setNotifsOpen] = useState(false);
  const { unreadCount } = useNotifications();

  // Global keyboard shortcuts
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      const inInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable;

      // Cmd+K always works
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setPaletteOpen((v) => !v);
        return;
      }

      if (inInput) return;

      // Shift+N → notifications
      if (e.shiftKey && e.key === 'N') {
        e.preventDefault();
        setNotifsOpen((v) => !v);
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  return (
    <div className="flex min-h-screen bg-surface-0">
      <Sidebar />
      <div className="flex-1 flex flex-col min-w-0">
        <TopBar
          onOpenPalette={() => setPaletteOpen(true)}
          onOpenNotifs={() => setNotifsOpen(true)}
          unreadCount={unreadCount}
        />
        <main className="flex-1 overflow-x-hidden">
          <Outlet />
        </main>
      </div>
      <CommandPalette open={paletteOpen} onClose={() => setPaletteOpen(false)} />
      <NotificationsDrawer open={notifsOpen} onClose={() => setNotifsOpen(false)} />
      <Toaster />
    </div>
  );
}
