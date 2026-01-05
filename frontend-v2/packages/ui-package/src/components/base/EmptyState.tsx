import { SECard } from '../SECard';

export interface EmptyStateProps {
  title: string;
  description?: string;
  action?: React.ReactNode;
  className?: string;
}

export function EmptyState({
  title,
  description,
  action,
  className = '',
}: EmptyStateProps) {
  return (
    <SECard className={`p-[var(--se-space-xl)] text-center ${className}`}>
      <div className="flex flex-col items-center gap-[var(--se-space-md)]">
        <h3 className="text-[var(--se-text)] text-lg font-semibold">{title}</h3>
        {description && (
          <p className="text-[var(--se-text-muted)] max-w-md">{description}</p>
        )}
        {action && <div className="mt-[var(--se-space-md)]">{action}</div>}
      </div>
    </SECard>
  );
}

