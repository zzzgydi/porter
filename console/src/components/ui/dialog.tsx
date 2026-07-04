import { cn } from '@/lib/utils'
import { type HTMLAttributes, type ReactNode, createContext, useContext, useEffect, useState } from 'react'
import { X } from 'lucide-react'

interface DialogContextValue {
  open: boolean
  setOpen: (v: boolean) => void
}

const DialogContext = createContext<DialogContextValue>({ open: false, setOpen: () => {} })

export function Dialog({ children, open, onOpenChange }: { children: ReactNode; open?: boolean; onOpenChange?: (v: boolean) => void }) {
  const [internalOpen, setInternalOpen] = useState(false)
  const isOpen = open !== undefined ? open : internalOpen
  const setIsOpen = onOpenChange || setInternalOpen
  return (
    <DialogContext.Provider value={{ open: isOpen, setOpen: setIsOpen }}>
      {children}
    </DialogContext.Provider>
  )
}

export function DialogTrigger({ children, asChild }: { children: ReactNode; asChild?: boolean }) {
  const { setOpen } = useContext(DialogContext)
  if (asChild) {
    return <span onClick={() => setOpen(true)}>{children}</span>
  }
  return (
    <button onClick={() => setOpen(true)} className="inline-flex items-center justify-center">
      {children}
    </button>
  )
}

export function DialogContent({ children, className }: { children: ReactNode; className?: string }) {
  const { open, setOpen } = useContext(DialogContext)
  const [show, setShow] = useState(open)

  useEffect(() => {
    if (open) {
      setShow(true)
    } else {
      const t = setTimeout(() => setShow(false), 150)
      return () => clearTimeout(t)
    }
  }, [open])

  if (!show) return null
  return (
    <div
      className={cn(
        'fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 transition-opacity duration-150',
        open ? 'opacity-100' : 'opacity-0',
      )}
      onClick={() => setOpen(false)}
    >
      <div
        className={cn(
          'relative z-50 grid w-full max-w-lg gap-4 rounded-xl border bg-background p-6 shadow-lg transition-all duration-150',
          open ? 'scale-100 opacity-100' : 'scale-[0.96] opacity-0',
          className,
        )}
        onClick={(e) => e.stopPropagation()}
      >
        {children}
        <button
          onClick={() => setOpen(false)}
          className="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
        >
          <span className="sr-only">Close</span>
          <X className="h-4 w-4" />
        </button>
      </div>
    </div>
  )
}

export function DialogHeader({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('flex flex-col space-y-1.5 text-center sm:text-left', className)} {...props} />
}

export function DialogTitle({ className, ...props }: HTMLAttributes<HTMLHeadingElement>) {
  return <h2 className={cn('text-lg font-semibold leading-none tracking-tight', className)} {...props} />
}

export function DialogDescription({ className, ...props }: HTMLAttributes<HTMLParagraphElement>) {
  return <p className={cn('text-sm text-muted-foreground', className)} {...props} />
}

export function DialogFooter({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2', className)} {...props} />
}
