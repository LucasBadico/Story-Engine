/**
 * Interface for navigation operations
 * Implementations: BrowserNavigator (web), ObsidianNavigator (obsidian)
 */
export interface Navigator {
  /**
   * Open external URL
   */
  openExternal(url: string): void;

  /**
   * Open note (Obsidian only)
   */
  openNote?(path: string): void;
}

