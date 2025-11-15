import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { forwardRef } from 'react'
import { Loader2 } from 'lucide-react'
import { motion } from 'framer-motion'
import { cn } from '../../lib/cn'

type ButtonVariant = 'primary' | 'secondary' | 'outline'
type ButtonSize = 'sm' | 'md' | 'lg'

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant
  size?: ButtonSize
  loading?: boolean
  icon?: ReactNode
}

const variantClasses: Record<ButtonVariant, string> = {
  primary:
    'bg-gradient-to-r from-primary to-secondary text-white shadow-lg shadow-primary/30 hover:shadow-primary/40 focus-visible:outline focus-visible:outline-secondary/60',
  secondary:
    'bg-white/10 text-white hover:bg-white/20 focus-visible:outline focus-visible:outline-white/60 border border-white/20',
  outline:
    'border border-white/20 text-white hover:bg-white/10 focus-visible:outline focus-visible:outline-white/60 bg-transparent',
}

const sizeClasses: Record<ButtonSize, string> = {
  sm: 'text-sm px-3 py-1.5 rounded-lg',
  md: 'text-sm px-4 py-2 rounded-xl',
  lg: 'text-base px-5 py-3 rounded-2xl',
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', loading, disabled, icon, children, ...rest }, ref) => {
    const isDisabled = disabled || loading

    return (
      <motion.button
        ref={ref}
        whileHover={{ scale: isDisabled ? 1 : 1.01 }}
        whileTap={{ scale: isDisabled ? 1 : 0.99 }}
        className={cn(
          'inline-flex items-center justify-center gap-2 font-semibold transition disabled:cursor-not-allowed disabled:opacity-50',
          variantClasses[variant],
          sizeClasses[size],
          className,
        )}
        disabled={isDisabled}
        aria-disabled={isDisabled}
        aria-busy={loading}
        {...rest}
      >
        {loading ? (
          <>
            <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
            <span>Processing...</span>
          </>
        ) : (
          <>
            {icon}
            {children}
          </>
        )}
      </motion.button>
    )
  },
)

Button.displayName = 'Button'

