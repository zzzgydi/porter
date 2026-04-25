import { cn } from '@/lib/utils'
import { type HTMLAttributes } from 'react'

export function Separator({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('shrink-0 bg-border h-[1px] w-full', className)} {...props} />
}
