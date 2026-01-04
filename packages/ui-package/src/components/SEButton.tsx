import { Button, ButtonProps } from '@heroui/react';
import { forwardRef } from 'react';

export interface SEButtonProps extends ButtonProps {
  children: React.ReactNode;
}

export const SEButton = forwardRef<HTMLButtonElement, SEButtonProps>(
  ({ className = '', children, ...props }, ref) => {
    return (
      <Button
        ref={ref}
        className={`rounded-[var(--se-radius-md)] ${className}`}
        {...props}
      >
        {children}
      </Button>
    );
  }
);

SEButton.displayName = 'SEButton';

