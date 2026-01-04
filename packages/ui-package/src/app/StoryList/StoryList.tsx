import { useState } from 'react';
import { useStories } from '../../hooks/useStories';
import { SECard } from '../../components/SECard';
import { SEButton } from '../../components/SEButton';
import { SEInput } from '../../components/SEInput';
import { LoadingState } from '../../components/base/LoadingState';
import { ErrorState } from '../../components/base/ErrorState';
import { EmptyState } from '../../components/base/EmptyState';
import type { Story } from '@story-engine/shared-ts';

export interface StoryListProps {
  onSelectStory?: (story: Story) => void;
  onCreateStory?: () => void;
}

export function StoryList({ onSelectStory, onCreateStory }: StoryListProps) {
  const { stories, loading, error, refetch } = useStories();
  const [searchQuery, setSearchQuery] = useState('');

  const filteredStories = stories.filter((story) =>
    story.title.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (loading) {
    return <LoadingState message="Loading stories..." />;
  }

  if (error) {
    return <ErrorState message={error.message} onRetry={refetch} />;
  }

  if (stories.length === 0) {
    return (
      <EmptyState
        title="No stories yet"
        description="Create your first story to get started"
        action={
          onCreateStory && (
            <SEButton onPress={onCreateStory}>Create Story</SEButton>
          )
        }
      />
    );
  }

  return (
    <div className="flex flex-col gap-[var(--se-space-md)]">
      <div className="flex items-center gap-[var(--se-space-md)]">
        <SEInput
          placeholder="Search stories..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="flex-1"
        />
        {onCreateStory && (
          <SEButton onPress={onCreateStory}>Create Story</SEButton>
        )}
      </div>

      <div className="flex flex-col gap-[var(--se-space-sm)]">
        {filteredStories.length === 0 ? (
          <EmptyState title="No stories found" description="Try a different search term" />
        ) : (
          filteredStories.map((story) => (
            <SECard
              key={story.id}
              isPressable
              onPress={() => onSelectStory?.(story)}
              className="p-[var(--se-space-md)] hover:bg-[var(--se-surface-hover)] cursor-pointer"
            >
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-[var(--se-text)] font-semibold">{story.title}</h3>
                  <p className="text-[var(--se-text-muted)] text-sm">
                    Status: {story.status} • Version: {story.version_number}
                  </p>
                </div>
              </div>
            </SECard>
          ))
        )}
      </div>
    </div>
  );
}

