import { useState, useCallback } from 'react';

export type SaveState = 'idle' | 'saving' | 'saved' | 'error';

export function useSaveState() {
  const [state, setState] = useState<SaveState>('idle');
  const [error, setError] = useState<Error | null>(null);

  const save = useCallback(async <T,>(saveFn: () => Promise<T>) => {
    setState('saving');
    setError(null);
    try {
      const result = await saveFn();
      setState('saved');
      setTimeout(() => setState('idle'), 2000);
      return result;
    } catch (err) {
      setState('error');
      setError(err instanceof Error ? err : new Error('Save failed'));
      throw err;
    }
  }, []);

  const reset = useCallback(() => {
    setState('idle');
    setError(null);
  }, []);

  return {
    state,
    error,
    save,
    reset,
  };
}

