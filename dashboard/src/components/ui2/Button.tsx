import { cn } from '@/lib/utils';
import { forwardRef } from 'react';
import type { ButtonHTMLAttributes } from 'react';

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';
type Size = 'sm' | 'md';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
}

const variantClasses: Record<Variant, string> = {
  primary: 'bg-accent text-surface-0 hover:opacity-90',
  secondary: 'bg-surface-2 border border-border-default text-text-primary hover:bg-surface-3',
  ghost: 'text-text-secondary hover:bg-surface-2 hover:text-text-primary',
  danger: 'bg-surface-2 border border-border-default text-text-primary hover:bg-status-danger/10 hover:border-status-danger/30 hover:text-status-danger',
};

const sizeClasses: Record<Size, string> = {
  sm: 'h-7 px-2.5 text-xs',
  md: 'h-9 px-3.5 text-sm',
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'secondary', size = 'md', disabled, children, ...props }, ref) => (
    <button
      ref={ref}
      disabled={disabled}
      className={cn(
        'inline-flex items-center justify-center gap-1.5 rounded-md font-medium',
        'transition-colors duration-150',
        'focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-accent',
        'disabled:cursor-not-allowed disabled:opacity-50',
        variantClasses[variant],
        sizeClasses[size],
        className
      )}
      {...props}
    >
      {children}
    </button>
  )
);
Button.displayName = 'Button';
