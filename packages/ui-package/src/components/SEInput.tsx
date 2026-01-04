import { Input, InputProps } from '@heroui/react';
import { forwardRef } from 'react';

export interface SEInputProps extends InputProps {}

export const SEInput = forwardRef<HTMLInputElement, SEInputProps>(
  ({ className = '', ...props }, ref) => {
    return (
      <Input
        ref={ref}
        className={`rounded-[var(--se-radius-md)] ${className}`}
        classNames={{
          input: 'text-[var(--se-text)]',
          inputWrapper: 'bg-[var(--se-surface)] border-[var(--se-border)]',
        }}
        {...props}
      />
    );
  }
);

SEInput.displayName = 'SEInput';

