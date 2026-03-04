import * as React from 'react';
import { cn } from '@/lib/utils';

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> { }

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          'flex h-14 w-full rounded-2xl border-4 border-border bg-white px-5 py-3 text-lg font-bold transition-all',
          'placeholder:text-muted-foreground placeholder:font-semibold',
          'focus-visible:outline-none focus-visible:border-primary focus-visible:ring-4 focus-visible:ring-primary/20',
          'disabled:cursor-not-allowed disabled:opacity-50 disabled:bg-muted',
          'hover:border-primary/40',
          className
        )}
        ref={ref}
        {...props}
      />
    );
  }
);
Input.displayName = 'Input';

export { Input };
