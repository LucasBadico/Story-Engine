import { SECard } from '../SECard';

export interface LoadingStateProps {
  message?: string;
  className?: string;
}

export function LoadingState({
  message = 'Loading...',
  className = '',
}: LoadingStateProps) {
  return (
    <SECard className={`p-[var(--se-space-xl)] text-center ${className}`}>
      <div className="flex flex-col items-center gap-[var(--se-space-md)]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--se-text)]"></div>
        <p className="text-[var(--se-text-muted)]">{message}</p>
      </div>
    </SECard>
  );
}

