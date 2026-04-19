import { type FC, useEffect, useState } from 'react';
import { Outlet } from 'react-router-dom';
import Header from './Header';
import { kymarosApi } from '../../api/kymarosApi';

const Layout: FC = () => {
  const [version, setVersion] = useState<string>('');

  useEffect(() => {
    kymarosApi.getHealth().then((h: { version: string }) => setVersion(h.version)).catch(() => {});
  }, []);

  return (
    <div className="flex min-h-screen flex-col bg-slate-50 dark:bg-[#0f172a]/95">
      <Header />
      <div className="flex-1 pt-14">
        <main className="mx-auto max-w-7xl px-4 pb-6 lg:px-8">
          <div className="py-6">
            <Outlet />
          </div>
        </main>
      </div>
      <footer className="border-t border-slate-200 px-4 py-3 text-center text-xs text-slate-400 dark:border-slate-700 dark:text-slate-500">
        <span>Kymaros{version ? ` v${version}` : ''}</span>
        <span className="mx-2">·</span>
        <a
          href="https://github.com/kymaroshq/kymaros"
          target="_blank"
          rel="noopener noreferrer"
          className="hover:text-slate-600 dark:hover:text-slate-300"
        >
          GitHub
        </a>
        <span className="mx-2">·</span>
        <a
          href="https://docs.kymaros.io"
          target="_blank"
          rel="noopener noreferrer"
          className="hover:text-slate-600 dark:hover:text-slate-300"
        >
          Docs
        </a>
        <span className="mx-2">·</span>
        <span>Apache 2.0</span>
      </footer>
    </div>
  );
};

export default Layout;
