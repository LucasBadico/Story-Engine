import type { Notifier } from '@story-engine/shared-ts';
import { Notice } from 'obsidian';

export class ObsidianNotifier implements Notifier {
  success(message: string, duration?: number): void {
    new Notice(message, duration);
  }

  error(message: string, duration?: number): void {
    new Notice(`Error: ${message}`, duration || 5000);
  }

  info(message: string, duration?: number): void {
    new Notice(message, duration);
  }

  warning(message: string, duration?: number): void {
    new Notice(`Warning: ${message}`, duration || 5000);
  }
}

