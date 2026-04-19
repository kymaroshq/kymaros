import { useState, useEffect, useCallback } from 'react';

type Theme = 'dark' | 'light';

let globalTheme: Theme = (localStorage.getItem('kymaros-theme') as Theme) ?? 'dark';
let listeners: (() => void)[] = [];

// Apply immediately to prevent flash
document.documentElement.setAttribute('data-theme', globalTheme);

function setGlobalTheme(t: Theme) {
  globalTheme = t;
  localStorage.setItem('kymaros-theme', t);
  document.documentElement.setAttribute('data-theme', t);
  listeners.forEach((fn) => fn());
}

export function useThemeStore() {
  const [theme, setLocal] = useState(globalTheme);

  useEffect(() => {
    const fn = () => setLocal(globalTheme);
    listeners.push(fn);
    return () => { listeners = listeners.filter((l) => l !== fn); };
  }, []);

  const setTheme = useCallback((t: Theme) => setGlobalTheme(t), []);
  const toggle = useCallback(() => setGlobalTheme(globalTheme === 'dark' ? 'light' : 'dark'), []);

  return { theme, setTheme, toggle };
}
