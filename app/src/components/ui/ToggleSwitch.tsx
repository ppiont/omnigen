import type { ButtonHTMLAttributes } from 'react'
import { motion } from 'framer-motion'
import { cn } from '../../lib/cn'

export type ToggleSwitchProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  checked: boolean
}

export const ToggleSwitch = ({ checked, className, ...rest }: ToggleSwitchProps) => {
  return (
    <motion.button
      type="button"
      className={cn(
        'relative inline-flex h-7 w-12 items-center rounded-full border border-bg-highlight px-1',
        checked ? 'bg-primary/70' : 'bg-bg-highlight',
        className,
      )}
      animate={{ backgroundColor: checked ? '#7cff00' : 'rgba(26,31,51,0.8)' }}
      transition={{ duration: 0.2 }}
      aria-pressed={checked}
      {...rest}
    >
      <motion.span
        className="h-5 w-5 rounded-full bg-white shadow"
        layout
        transition={{ type: 'spring', stiffness: 500, damping: 30 }}
        style={{ marginLeft: checked ? 'auto' : 0 }}
      />
    </motion.button>
  )
}

