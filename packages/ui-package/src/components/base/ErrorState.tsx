import { SECard } from '../SECard';
import { SEButton } from '../SEButton';

export interface ErrorStateProps {
  title?: string;
  message: string;
  onRetry?: () => void;
  className?: string;
}

export function ErrorState({
  title = 'Error',
  message,
  onRetry,
  className = '',
}: ErrorStateProps) {
  return (
    <SECard className={`p-[var(--se-space-xl)] text-center ${className}`}>
      <div className="flex flex-col items-center gap-[var(--se-space-md)]">
        <h3 className="text-red-500 text-lg font-semibold">{title}</h3>
        <p className="text-[var(--se-text-muted)] max-w-md">{message}</p>
        {onRetry && (
          <SEButton onPress={onRetry} className="mt-[var(--se-space-md)]">
            Retry
          </SEButton>
        )}
      </div>
    </SECard>
  );
}

