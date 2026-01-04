import { useState, useEffect } from 'react';
import { useServices } from '../providers/ServicesProvider';
import type { World } from '@story-engine/shared-ts';

export function useWorlds() {
  const { storyEngine } = useServices();
  const [worlds, setWorlds] = useState<World[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchWorlds = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await storyEngine.listWorlds();
      setWorlds(data);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch worlds'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchWorlds();
  }, []);

  return {
    worlds,
    loading,
    error,
    refetch: fetchWorlds,
  };
}

