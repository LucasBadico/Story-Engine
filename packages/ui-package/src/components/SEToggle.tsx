import { Switch, SwitchProps } from '@heroui/react';
import { forwardRef } from 'react';

export interface SEToggleProps extends SwitchProps {
  children?: React.ReactNode;
}

export const SEToggle = forwardRef<HTMLInputElement, SEToggleProps>(
  ({ className = '', children, ...props }, ref) => {
    return (
      <Switch
        ref={ref}
        className={className}
        classNames={{
          wrapper: 'bg-[var(--se-surface)]',
        }}
        {...props}
      >
        {children}
      </Switch>
    );
  }
);

SEToggle.displayName = 'SEToggle';

