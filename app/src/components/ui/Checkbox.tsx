import type { InputHTMLAttributes } from 'react'
import { forwardRef } from 'react'
import { motion } from 'framer-motion'
import { cn } from '../../lib/cn'

export type CheckboxProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: string
}

export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(({ label, className, ...rest }, ref) => {
  return (
    <label className={cn('flex cursor-pointer items-center gap-2 text-sm text-text-secondary', className)}>
      <input ref={ref} type="checkbox" className="sr-only" {...rest} />
      <motion.span
        className="flex h-5 w-5 items-center justify-center rounded-md border border-bg-highlight bg-bg-highlight"
        initial={false}
        animate={rest.checked ? { backgroundColor: '#7cff00', borderColor: '#7cff00' } : {}}
        transition={{ type: 'spring', stiffness: 400, damping: 25 }}
      >
        <motion.span
          className="h-2.5 w-2.5 rounded-sm bg-white"
          initial={false}
          animate={{ scale: rest.checked ? 1 : 0.2, opacity: rest.checked ? 1 : 0 }}
          transition={{ duration: 0.15 }}
        />
      </motion.span>
      {label && <span>{label}</span>}
    </label>
  )
})

Checkbox.displayName = 'Checkbox'

