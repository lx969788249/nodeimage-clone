import { writable } from 'svelte/store';

type Theme = 'light' | 'dark';

function createThemeStore() {
  const key = 'nodeimage-theme';
  const initial = (typeof localStorage !== 'undefined' && (localStorage.getItem(key) as Theme)) || 'light';
  const { subscribe, set, update } = writable<Theme>(initial);

  const apply = (theme: Theme) => {
    if (typeof document !== 'undefined') {
      document.documentElement.classList.toggle('dark', theme === 'dark');
    }
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(key, theme);
    }
  };

  return {
    subscribe,
    set: (theme: Theme) => {
      apply(theme);
      set(theme);
    },
    toggle: () => update((current) => {
      const next = current === 'dark' ? 'light' : 'dark';
      apply(next);
      return next;
    }),
    syncWithSystem: () => {
      if (typeof window === 'undefined') return;
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      apply(prefersDark ? 'dark' : 'light');
      set(prefersDark ? 'dark' : 'light');
    },
    get theme() {
      let value: Theme = 'light';
      const unsubscribe = subscribe((current) => {
        value = current;
      });
      unsubscribe();
      return value;
    }
  };
}

export const themeStore = createThemeStore();
