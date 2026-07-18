import { cn } from '@/lib/utils'
import { AlertCircle } from 'lucide-react'

interface ErrorStateProps {
  title?: string
  message?: string
  className?: string
}

export function ErrorState({
  title = 'Failed to load data',
  message,
  className,
}: ErrorStateProps) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center rounded-xl border border-dashed border-destructive/30 bg-destructive/5 p-10 text-center',
        className,
      )}
    >
      <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10">
        <AlertCircle className="h-6 w-6 text-destructive" />
      </div>
      <h3 className="text-sm font-semibold text-foreground">{title}</h3>
      {message && <p className="mt-1 max-w-xs text-sm text-muted-foreground">{message}</p>}
    </div>
  )
}
