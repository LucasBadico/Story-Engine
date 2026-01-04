import { useState, useEffect } from 'react';
import { useServices } from '../../providers/ServicesProvider';
import { useWorlds } from '../../hooks/useWorlds';
import { useSaveState } from '../../hooks/useSaveState';
import { SECard } from '../../components/SECard';
import { SEButton } from '../../components/SEButton';
import { SEInput } from '../../components/SEInput';
import { SESelect, SelectItem } from '../../components/SESelect';
import { SEToggle } from '../../components/SEToggle';
import { InlineEditable } from '../../components/base/InlineEditable';
import { CollapsibleSection } from '../../components/base/CollapsibleSection';
import { SavingIndicator } from '../../components/base/SavingIndicator';
import { LoadingState } from '../../components/base/LoadingState';
import { ErrorState } from '../../components/base/ErrorState';
import type { Story, World } from '@story-engine/shared-ts';

export interface StoryDetailsProps {
  storyId: string;
  onBack?: () => void;
}

export function StoryDetails({ storyId, onBack }: StoryDetailsProps) {
  const { storyEngine } = useServices();
  const { worlds, loading: worldsLoading } = useWorlds();
  const [story, setStory] = useState<Story | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const { state: saveState, save } = useSaveState();

  const [rpgEnabled, setRpgEnabled] = useState(false);

  useEffect(() => {
    const fetchStory = async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await storyEngine.getStory(storyId);
        setStory(data);
        setRpgEnabled(!!data.world_id);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch story'));
      } finally {
        setLoading(false);
      }
    };

    fetchStory();
  }, [storyId, storyEngine]);

  const handleUpdateTitle = async (title: string) => {
    if (!story) return;
    await save(async () => {
      const updated = await storyEngine.updateStory(story.id, { title });
      setStory(updated);
    });
  };

  const handleUpdateWorld = async (worldId: string | null) => {
    if (!story) return;
    await save(async () => {
      const updated = await storyEngine.updateStory(story.id, { world_id: worldId });
      setStory(updated);
      setRpgEnabled(!!worldId);
    });
  };

  if (loading) {
    return <LoadingState message="Loading story..." />;
  }

  if (error) {
    return <ErrorState message={error.message} />;
  }

  if (!story) {
    return <ErrorState message="Story not found" />;
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
            <label className="text-[var(--se-text-muted)] text-sm mb-1 block">Title</label>
            <InlineEditable
              value={story.title}
              onSave={handleUpdateTitle}
              className="text-xl font-semibold"
            />
          </div>

          <div>
            <label className="text-[var(--se-text-muted)] text-sm mb-1 block">Status</label>
            <p className="text-[var(--se-text)]">{story.status}</p>
          </div>
        </div>
      </SECard>

      <CollapsibleSection title="World" defaultOpen>
        <div className="flex flex-col gap-[var(--se-space-md)]">
          <div>
            <label className="text-[var(--se-text-muted)] text-sm mb-1 block">
              Select World
            </label>
            {worldsLoading ? (
              <p className="text-[var(--se-text-muted)]">Loading worlds...</p>
            ) : (
              <SESelect
                selectedKeys={story.world_id ? [story.world_id] : []}
                onSelectionChange={(keys) => {
                  const selected = Array.from(keys)[0] as string | undefined;
                  handleUpdateWorld(selected || null);
                }}
                placeholder="No world selected"
              >
                {worlds.map((world) => (
                  <SelectItem key={world.id} textValue={world.name}>
                    {world.name}
                  </SelectItem>
                ))}
              </SESelect>
            )}
          </div>
        </div>
      </CollapsibleSection>

      <CollapsibleSection title="Features">
        <div className="flex flex-col gap-[var(--se-space-md)]">
          <div className="flex items-center justify-between">
            <label className="text-[var(--se-text)]">RPG Enabled</label>
            <SEToggle
              isSelected={rpgEnabled}
              isDisabled={!story.world_id}
              onValueChange={(enabled) => {
                setRpgEnabled(enabled);
                // TODO: Update story features
              }}
            />
          </div>
        </div>
      </CollapsibleSection>

      {story.world_id && rpgEnabled && (
        <CollapsibleSection title="RPG Settings">
          <div className="flex flex-col gap-[var(--se-space-md)]">
            <p className="text-[var(--se-text-muted)]">
              RPG configuration will appear here when world and RPG are enabled.
            </p>
          </div>
        </CollapsibleSection>
      )}
    </div>
  );
}

