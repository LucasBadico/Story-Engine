export type SaveState = 'idle' | 'saving' | 'saved' | 'error';

export interface SavingIndicatorProps {
  state: SaveState;
  className?: string;
}

export function SavingIndicator({ state, className = '' }: SavingIndicatorProps) {
  const getContent = () => {
    switch (state) {
      case 'saving':
        return <span className="text-[var(--se-text-muted)]">Saving...</span>;
      case 'saved':
        return <span className="text-green-500">Saved</span>;
      case 'error':
        return <span className="text-red-500">Error</span>;
      default:
        return null;
    }
  };

  if (state === 'idle') return null;

  return <div className={`text-sm ${className}`}>{getContent()}</div>;
}

