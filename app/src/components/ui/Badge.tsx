import type { HTMLAttributes } from 'react'
import { cn } from '../../lib/cn'

type BadgeVariant = 'default' | 'success' | 'warning' | 'info'

export type BadgeProps = HTMLAttributes<HTMLSpanElement> & {
  variant?: BadgeVariant
}

const variantClasses: Record<BadgeVariant, string> = {
  default: 'bg-white/10 text-white',
  success: 'bg-emerald-500/20 text-emerald-300',
  warning: 'bg-amber-500/20 text-amber-200',
  info: 'bg-blue-500/20 text-blue-200',
}

export const Badge = ({ className, variant = 'default', ...rest }: BadgeProps) => {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide',
        variantClasses[variant],
        className,
      )}
      {...rest}
    />
  )
}

