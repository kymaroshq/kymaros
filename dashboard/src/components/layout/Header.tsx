import type { FC } from 'react';
import { NavLink } from 'react-router-dom';
import { Sun, Moon } from 'lucide-react';
import Logo from '../ui/Logo';
import { useTheme } from '../../hooks/useTheme';

interface NavItem {
  label: string;
  to: string;
}

const navItems: NavItem[] = [
  { label: 'Dashboard', to: '/' },
  { label: 'Settings', to: '/settings' },
];

const Header: FC = () => {
  const { theme, toggleTheme } = useTheme();

  return (
    <header className="fixed inset-x-0 top-0 z-50 flex h-14 items-center border-b border-slate-700/50 bg-[#0f172a] px-4 lg:px-8">
      {/* Logo — left */}
      <div className="flex shrink-0 items-center text-white">
        <Logo size="sm" showText />
      </div>

      {/* Nav — center */}
      <nav className="flex flex-1 items-center justify-center gap-1">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              `rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-slate-700/60 text-white'
                  : 'text-slate-400 hover:bg-slate-700/30 hover:text-white'
              }`
            }
          >
            {item.label}
          </NavLink>
        ))}
      </nav>

      {/* Dark mode toggle — right */}
      <button
        type="button"
        onClick={toggleTheme}
        className="rounded-md p-2 text-slate-400 transition-colors hover:bg-slate-700/30 hover:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-[#0f172a]"
        aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
      >
        {theme === 'dark' ? (
          <Sun className="h-5 w-5" />
        ) : (
          <Moon className="h-5 w-5" />
        )}
      </button>
    </header>
  );
};

export default Header;
