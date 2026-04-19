import { Moon, Sun } from 'lucide-react';
import { useThemeStore } from '@/hooks/useThemeStore';

export function ThemeToggle() {
  const { theme, toggle } = useThemeStore();
  return (
    <button
      onClick={toggle}
      className="h-8 w-8 flex items-center justify-center rounded-md text-text-tertiary hover:bg-surface-2 hover:text-text-primary transition-colors duration-150"
      aria-label={theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'}
    >
      {theme === 'dark' ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
    </button>
  );
}
