import * as React from 'react';
import { cva } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const buttonVariants = cva(
  'inline-flex items-center justify-center whitespace-nowrap rounded-xl text-sm font-black transition-all duration-200 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-primary/30 disabled:pointer-events-none disabled:opacity-50 active:translate-y-1',
  {
    variants: {
      variant: {
        default:
          'bg-primary text-white hover:bg-primary-hover border-b-4 border-primary-hover active:border-b-0',
        outline:
          'border-4 border-border bg-white text-muted-foreground hover:bg-muted hover:border-primary/30 hover:text-primary active:border-b-0 active:translate-y-1',
        ghost:
          'hover:bg-primary-light hover:text-primary active:bg-primary-light',
        link: 'text-primary underline-offset-4 hover:underline decoration-2',
        destructive:
          'bg-danger text-white hover:bg-red-600 border-b-4 border-red-700 active:border-b-0',
        success:
          'bg-success text-white hover:bg-green-600 border-b-4 border-green-700 active:border-b-0',
        secondary:
          'bg-secondary text-white hover:bg-secondary-hover border-b-4 border-secondary-hover active:border-b-0',
        accent:
          'bg-accent text-white hover:bg-accent-hover border-b-4 border-accent-hover active:border-b-0',
      },
      size: {
        default: 'h-12 px-5 py-2',
        sm: 'h-10 px-4 text-xs',
        lg: 'h-14 px-8 text-lg',
        icon: 'h-12 w-12',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  }
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'default' | 'outline' | 'ghost' | 'link' | 'destructive' | 'success' | 'secondary' | 'accent';
  size?: 'default' | 'sm' | 'lg' | 'icon';
  asChild?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, ...props }, ref) => {
    return (
      <button
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    );
  }
);
Button.displayName = 'Button';

export { Button, buttonVariants };
