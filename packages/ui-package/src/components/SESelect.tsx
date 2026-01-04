import { Select, SelectProps, SelectItem } from '@heroui/react';
import { forwardRef } from 'react';

export type SESelectProps = SelectProps;

export const SESelect = forwardRef<HTMLSelectElement, SESelectProps>(
  ({ className = '', children, ...props }, ref) => {
    return (
      <Select
        ref={ref}
        className={`rounded-[var(--se-radius-md)] ${className}`}
        classNames={{
          trigger: 'bg-[var(--se-surface)] border-[var(--se-border)] text-[var(--se-text)]',
          value: 'text-[var(--se-text)]',
        }}
        {...props}
      >
        {children}
      </Select>
    );
  }
);

SESelect.displayName = 'SESelect';

export { SelectItem };

