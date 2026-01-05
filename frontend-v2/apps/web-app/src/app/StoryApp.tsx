'use client';

'use client';

import { useState } from 'react';
import { ServicesProvider, StoryList, StoryDetails, WorldDetails } from '@story-engine/ui-package';
import type { Story, World, Services } from '@story-engine/shared-ts';
import { HttpStoryEngineClient } from '../adapters/httpStoryEngineClient';
import { BrowserStorage } from '../adapters/browserStorage';
import { ToastNotifier } from '../adapters/toastNotifier';
import { BrowserNavigator } from '../adapters/browserNavigator';

type View = 'list' | 'story' | 'world';

export function StoryApp() {
  const [view, setView] = useState<View>('list');
  const [selectedStoryId, setSelectedStoryId] = useState<string | null>(null);
  const [selectedWorldId, setSelectedWorldId] = useState<string | null>(null);

  // TODO: Get these from environment or settings
  const baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  const apiKey = process.env.NEXT_PUBLIC_API_KEY || '';
  const tenantId = process.env.NEXT_PUBLIC_TENANT_ID || '';

  const services: Services = {
    storyEngine: new HttpStoryEngineClient(baseUrl, apiKey, tenantId),
    storage: new BrowserStorage(),
    notify: new ToastNotifier(),
    navigator: new BrowserNavigator(),
  };

  const handleSelectStory = (story: Story) => {
    setSelectedStoryId(story.id);
    setView('story');
  };

  const handleCreateStory = async () => {
    // TODO: Implement create story flow
    services.notify.info('Create story feature coming soon');
  };

  const handleBack = () => {
    setView('list');
    setSelectedStoryId(null);
    setSelectedWorldId(null);
  };

  return (
    <ServicesProvider services={services}>
      <div className="min-h-screen bg-[var(--se-bg)] text-[var(--se-text)] p-[var(--se-space-lg)]">
        {view === 'list' && (
          <StoryList
            onSelectStory={handleSelectStory}
            onCreateStory={handleCreateStory}
          />
        )}
        {view === 'story' && selectedStoryId && (
          <StoryDetails storyId={selectedStoryId} onBack={handleBack} />
        )}
        {view === 'world' && selectedWorldId && (
          <WorldDetails worldId={selectedWorldId} onBack={handleBack} />
        )}
      </div>
    </ServicesProvider>
  );
}

