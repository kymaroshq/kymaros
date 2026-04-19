import { useState, useEffect, useCallback } from 'react';

export interface Notification {
  id: string;
  type: 'success' | 'warning' | 'error' | 'info';
  title: string;
  description?: string;
  timestamp: string;
  read: boolean;
  href?: string;
}

const STORAGE_KEY = 'kymaros-notifications';
const MAX = 50;

function load(): Notification[] {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]');
  } catch { return []; }
}

function save(items: Notification[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(items.slice(0, MAX)));
}

let items = load();
let listeners: (() => void)[] = [];
function notify() { listeners.forEach((fn) => fn()); }

export function pushNotification(n: Omit<Notification, 'id' | 'timestamp' | 'read'>) {
  items = [{ ...n, id: Math.random().toString(36).slice(2), timestamp: new Date().toISOString(), read: false }, ...items].slice(0, MAX);
  save(items);
  notify();
}

export function useNotifications() {
  const [, setTick] = useState(0);

  useEffect(() => {
    const fn = () => setTick((t) => t + 1);
    listeners.push(fn);
    return () => { listeners = listeners.filter((l) => l !== fn); };
  }, []);

  const markAllRead = useCallback(() => {
    items = items.map((n) => ({ ...n, read: true }));
    save(items);
    notify();
  }, []);

  const clear = useCallback(() => {
    items = [];
    save(items);
    notify();
  }, []);

  return {
    notifications: items,
    unreadCount: items.filter((n) => !n.read).length,
    markAllRead,
    clear,
  };
}
