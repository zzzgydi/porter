import { createContext, useContext, useState, useCallback, useRef, useEffect, type ReactNode } from 'react'
import { cn } from '@/lib/utils'
import { X, CheckCircle2, AlertCircle, Info } from 'lucide-react'

type ToastType = 'success' | 'error' | 'info'

interface Toast {
  id: string
  message: string
  type: ToastType
}

interface ToastContextValue {
  toast: (message: string, type?: ToastType) => void
  success: (message: string) => void
  error: (message: string) => void
  info: (message: string) => void
}

const ToastContext = createContext<ToastContextValue | null>(null)

export function useToast() {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used within ToastProvider')
  return ctx
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])
  const timersRef = useRef<Map<string, number>>(new Map())

  const remove = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
    const timer = timersRef.current.get(id)
    if (timer) {
      window.clearTimeout(timer)
      timersRef.current.delete(id)
    }
  }, [])

  const add = useCallback((message: string, type: ToastType = 'info') => {
    const id = Math.random().toString(36).slice(2)
    setToasts((prev) => [...prev, { id, message, type }])
    const timer = window.setTimeout(() => remove(id), 4000)
    timersRef.current.set(id, timer)
  }, [remove])

  useEffect(() => {
    return () => {
      timersRef.current.forEach((timer) => window.clearTimeout(timer))
    }
  }, [])

  return (
    <ToastContext.Provider
      value={{
        toast: add,
        success: (m) => add(m, 'success'),
        error: (m) => add(m, 'error'),
        info: (m) => add(m, 'info'),
      }}
    >
      {children}
      <div className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2">
        {toasts.map((t) => (
          <ToastItem key={t.id} toast={t} onClose={() => remove(t.id)} />
        ))}
      </div>
    </ToastContext.Provider>
  )
}

function ToastItem({ toast, onClose }: { toast: Toast; onClose: () => void }) {
  const Icon = toast.type === 'success' ? CheckCircle2 : toast.type === 'error' ? AlertCircle : Info
  return (
    <div
      className={cn(
        'animate-fade-in flex w-80 items-start gap-3 rounded-xl border bg-background p-4 shadow-lg',
        toast.type === 'success' && 'border-success/20 bg-success/5 text-success-foreground',
        toast.type === 'error' && 'border-destructive/20 bg-destructive/5 text-destructive-foreground',
        toast.type === 'info' && 'border-primary/20 bg-primary/5 text-primary-foreground',
      )}
    >
      <Icon
        className={cn(
          'mt-0.5 h-5 w-5 shrink-0',
          toast.type === 'success' && 'text-success',
          toast.type === 'error' && 'text-destructive',
          toast.type === 'info' && 'text-primary',
        )}
      />
      <p className="flex-1 text-sm font-medium text-foreground">{toast.message}</p>
      <button
        onClick={onClose}
        className="rounded-md p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  )
}
