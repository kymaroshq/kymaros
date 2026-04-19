type ToastType = 'success' | 'error' | 'warning' | 'info';

interface ToastItem {
  id: string;
  type: ToastType;
  message: string;
  description?: string;
}

let listeners: ((toasts: ToastItem[]) => void)[] = [];
let toasts: ToastItem[] = [];

function notify() {
  listeners.forEach((fn) => fn([...toasts]));
}

function push(type: ToastType, message: string, description?: string) {
  const id = Math.random().toString(36).slice(2);
  toasts = [{ id, type, message, description }, ...toasts].slice(0, 5);
  notify();
  setTimeout(() => {
    toasts = toasts.filter((t) => t.id !== id);
    notify();
  }, 4000);
}

export const toast = {
  success: (message: string, description?: string) => push('success', message, description),
  error: (message: string, description?: string) => push('error', message, description),
  warning: (message: string, description?: string) => push('warning', message, description),
  info: (message: string, description?: string) => push('info', message, description),
  subscribe: (fn: (toasts: ToastItem[]) => void) => {
    listeners.push(fn);
    return () => { listeners = listeners.filter((l) => l !== fn); };
  },
  getToasts: () => toasts,
};

export type { ToastItem };
