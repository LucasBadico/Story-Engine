import { useState, useEffect } from 'react';
import { useServices } from '../providers/ServicesProvider';
import type { Story } from '@story-engine/shared-ts';

export function useStories() {
  const { storyEngine } = useServices();
  const [stories, setStories] = useState<Story[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchStories = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await storyEngine.listStories();
      setStories(data);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch stories'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStories();
  }, []);

  return {
    stories,
    loading,
    error,
    refetch: fetchStories,
  };
}

