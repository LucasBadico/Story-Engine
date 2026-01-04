import { Card, CardProps } from '@heroui/react';
import { forwardRef } from 'react';

export interface SECardProps extends CardProps {
  children: React.ReactNode;
}

export const SECard = forwardRef<HTMLDivElement, SECardProps>(
  ({ className = '', children, ...props }, ref) => {
    return (
      <Card
        ref={ref}
        className={`bg-[var(--se-surface)] text-[var(--se-text)] rounded-[var(--se-radius-md)] ${className}`}
        {...props}
      >
        {children}
      </Card>
    );
  }
);

SECard.displayName = 'SECard';

