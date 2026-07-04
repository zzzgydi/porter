import { type ButtonHTMLAttributes, forwardRef, type ReactElement, cloneElement, Children } from 'react'
import { cn } from '@/lib/utils'
import { Loader2 } from 'lucide-react'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link'
  size?: 'default' | 'sm' | 'lg' | 'icon'
  loading?: boolean
  asChild?: boolean
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'default', size = 'default', loading = false, asChild = false, children, disabled, ...props }, ref) => {
    const classes = cn(
      'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 active:scale-[0.98]',
      variant === 'default' && 'bg-primary text-primary-foreground shadow-sm hover:bg-primary/90 hover:shadow',
      variant === 'destructive' && 'bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90',
      variant === 'outline' && 'border border-input bg-background hover:bg-accent hover:text-accent-foreground',
      variant === 'secondary' && 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
      variant === 'ghost' && 'hover:bg-accent hover:text-accent-foreground',
      variant === 'link' && 'text-primary underline-offset-4 hover:underline',
      size === 'default' && 'h-10 px-4 py-2',
      size === 'sm' && 'h-9 rounded-md px-3',
      size === 'lg' && 'h-11 rounded-md px-8',
      size === 'icon' && 'h-10 w-10',
      className,
    )

    if (asChild && Children.count(children) === 1) {
      const child = Children.only(children) as ReactElement<{ className?: string; ref?: React.Ref<unknown> }>
      return cloneElement(child, {
        className: cn(classes, child.props.className),
      })
    }

    return (
      <button
        ref={ref}
        className={classes}
        disabled={disabled || loading}
        {...props}
      >
        {loading && <Loader2 className="h-4 w-4 animate-spin" />}
        {children}
      </button>
    )
  },
)
Button.displayName = 'Button'
