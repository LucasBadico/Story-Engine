import type { Navigator } from '@story-engine/shared-ts';

export class BrowserNavigator implements Navigator {
  openExternal(url: string): void {
    window.open(url, '_blank', 'noopener,noreferrer');
  }

  openNote?(): void {
    // Not applicable for web
  }
}

