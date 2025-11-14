import type { InputHTMLAttributes, ReactNode } from 'react'
import { forwardRef } from 'react'
import { cn } from '../../lib/cn'

export type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: string
  icon?: ReactNode
  error?: string
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, icon, type = 'text', error, className, id, ...rest }, ref) => {
    const inputId = id ?? rest.name
    const describedBy = error ? `${inputId}-error` : undefined

    return (
      <div className="space-y-2">
        {label && (
          <label htmlFor={inputId} className="text-sm font-medium text-white/80">
            {label}
          </label>
        )}
        <div className="relative">
          {icon && (
            <span className="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-white/50" aria-hidden>
              {icon}
            </span>
          )}
          <input
            id={inputId}
            ref={ref}
            type={type}
            className={cn(
              'w-full rounded-2xl border bg-black/30 px-4 py-3 text-sm text-white placeholder:text-white/40 outline-none transition focus:ring-2 focus:ring-primary/40',
              icon ? 'pl-11' : '',
              error ? 'border-secondary' : 'border-white/15 focus:border-primary',
              className,
            )}
            aria-invalid={!!error}
            aria-describedby={describedBy}
            {...rest}
          />
        </div>
        {error && (
          <p id={describedBy} className="text-xs text-secondary">
            {error}
          </p>
        )}
      </div>
    )
  },
)

Input.displayName = 'Input'

