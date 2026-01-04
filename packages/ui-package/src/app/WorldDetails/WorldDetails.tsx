import { useState, useEffect } from 'react';
import { useServices } from '../../providers/ServicesProvider';
import { useSaveState } from '../../hooks/useSaveState';
import { SECard } from '../../components/SECard';
import { SEButton } from '../../components/SEButton';
import { SEInput } from '../../components/SEInput';
import { InlineEditable } from '../../components/base/InlineEditable';
import { SavingIndicator } from '../../components/base/SavingIndicator';
import { LoadingState } from '../../components/base/LoadingState';
import { ErrorState } from '../../components/base/ErrorState';
import type { World } from '@story-engine/shared-ts';

export interface WorldDetailsProps {
  worldId: string;
  onBack?: () => void;
}

export function WorldDetails({ worldId, onBack }: WorldDetailsProps) {
  const { storyEngine } = useServices();
  const [world, setWorld] = useState<World | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const { state: saveState, save } = useSaveState();

  useEffect(() => {
    const fetchWorld = async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await storyEngine.getWorld(worldId);
        setWorld(data);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch world'));
      } finally {
        setLoading(false);
      }
    };

    fetchWorld();
  }, [worldId, storyEngine]);

  const handleUpdateName = async (name: string) => {
    if (!world) return;
    await save(async () => {
      const updated = await storyEngine.updateWorld(world.id, { name });
      setWorld(updated);
    });
  };

  const handleUpdateDescription = async (description: string) => {
    if (!world) return;
    await save(async () => {
      const updated = await storyEngine.updateWorld(world.id, { description });
      setWorld(updated);
    });
  };

  if (loading) {
    return <LoadingState message="Loading world..." />;
  }

  if (error) {
    return <ErrorState message={error.message} />;
  }

  if (!world) {
    return <ErrorState message="World not found" />;
  }

  return (
    <div className="flex flex-col gap-[var(--se-space-lg)]">
      <div className="flex items-center justify-between">
        {onBack && (
          <SEButton variant="light" onPress={onBack}>
            ← Back
          </SEButton>
        )}
        <SavingIndicator state={saveState} />
      </div>

      <SECard className="p-[var(--se-space-lg)]">
        <div className="flex flex-col gap-[var(--se-space-md)]">
          <div>
            <label className="text-[var(--se-text-muted)] text-sm mb-1 block">Name</label>
            <InlineEditable
              value={world.name}
              onSave={handleUpdateName}
              className="text-xl font-semibold"
            />
          </div>

          <div>
            <label className="text-[var(--se-text-muted)] text-sm mb-1 block">Description</label>
            <InlineEditable
              value={world.description || ''}
              onSave={handleUpdateDescription}
              placeholder="No description"
            />
          </div>

          <div>
            <label className="text-[var(--se-text-muted)] text-sm mb-1 block">Genre</label>
            <p className="text-[var(--se-text)]">{world.genre}</p>
          </div>
        </div>
      </SECard>

      <SECard className="p-[var(--se-space-lg)]">
        <h3 className="text-[var(--se-text)] font-semibold mb-[var(--se-space-md)]">
          RPG Schema Editor
        </h3>
        <p className="text-[var(--se-text-muted)]">
          RPG schema editor will appear here. This is a placeholder for future implementation.
        </p>
      </SECard>
    </div>
  );
}

