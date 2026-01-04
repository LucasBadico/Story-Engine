import type { Navigator } from '@story-engine/shared-ts';
import { App } from 'obsidian';

export class ObsidianNavigator implements Navigator {
  constructor(private app: App) {}

  openExternal(url: string): void {
    // Obsidian doesn't have a direct way to open external URLs
    // This would need to use Electron's shell.openExternal if available
    // For now, we'll just log it
    console.log('Would open external URL:', url);
  }

  openNote(path: string): void {
    this.app.workspace.openLinkText(path, '', false);
  }
}

