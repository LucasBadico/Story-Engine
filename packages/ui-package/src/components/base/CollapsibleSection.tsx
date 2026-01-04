import { useState } from 'react';
import { SECard } from '../SECard';

export interface CollapsibleSectionProps {
  title: string;
  children: React.ReactNode;
  defaultOpen?: boolean;
  className?: string;
}

export function CollapsibleSection({
  title,
  children,
  defaultOpen = false,
  className = '',
}: CollapsibleSectionProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen);

  return (
    <SECard className={className}>
      <div className="p-[var(--se-space-md)]">
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="w-full flex items-center justify-between text-left"
        >
          <h3 className="text-[var(--se-text)] font-semibold">{title}</h3>
          <span className="text-[var(--se-text-muted)]">
            {isOpen ? '▼' : '▶'}
          </span>
        </button>
        {isOpen && <div className="mt-[var(--se-space-md)]">{children}</div>}
      </div>
    </SECard>
  );
}

