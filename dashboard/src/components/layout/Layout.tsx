import type { FC } from 'react';
import { Outlet } from 'react-router-dom';
import Header from './Header';

const Layout: FC = () => {
  return (
    <div className="min-h-screen bg-slate-50 dark:bg-[#0f172a]/95">
      <Header />
      <div className="pt-14">
        <main className="mx-auto max-w-7xl px-4 pb-6 lg:px-8">
          <div className="py-6">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
};

export default Layout;
