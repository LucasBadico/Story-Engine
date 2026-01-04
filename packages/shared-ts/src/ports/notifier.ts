/**
 * Interface for notifications
 * Implementations: ToastNotifier (web), ObsidianNotifier (obsidian)
 */
export interface Notifier {
  success(message: string, duration?: number): void;
  error(message: string, duration?: number): void;
  info(message: string, duration?: number): void;
  warning(message: string, duration?: number): void;
}

