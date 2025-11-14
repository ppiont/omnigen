import type { ButtonHTMLAttributes, MouseEvent, ReactNode } from 'react'
import { motion } from 'framer-motion'
import { useEffect, useMemo, useState } from 'react'
import { cn } from '../lib/cn'

type PrimaryButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  loading?: boolean
  variant?: 'solid' | 'gradient'
  loadingContent?: ReactNode
}

const variantClassNames: Record<NonNullable<PrimaryButtonProps['variant']>, string> = {
  solid: 'bg-primary hover:bg-primary/90 text-[#05120a] focus-visible:outline-primary/60',
  gradient:
    'bg-gradient-to-r from-primary to-aurora-teal text-[#05120a] shadow-[0_10px_30px_rgba(124,255,0,0.4)] hover:shadow-[0_15px_40px_rgba(124,255,0,0.5)] focus-visible:outline-primary/60',
}

type Ripple = {
  id: number
  x: number
  y: number
}

export const PrimaryButton = ({
  className,
  children,
  loading,
  variant = 'solid',
  loadingContent,
  disabled,
  onClick,
  ...rest
}: PrimaryButtonProps) => {
  const isDisabled = loading || disabled
  const buttonContent = loading ? loadingContent ?? 'Please wait...' : children
  const [ripples, setRipples] = useState<Ripple[]>([])

  useEffect(() => {
    if (!ripples.length) return
    const timeout = setTimeout(() => setRipples((current) => current.slice(1)), 400)
    return () => clearTimeout(timeout)
  }, [ripples])

  const handleClick = (event: MouseEvent<HTMLButtonElement>) => {
    if (variant === 'gradient' && !isDisabled) {
      const rect = event.currentTarget.getBoundingClientRect()
      setRipples((current) => [
        ...current,
        { id: Date.now(), x: event.clientX - rect.left, y: event.clientY - rect.top },
      ])
    }
    onClick?.(event)
  }

  const rippleElements = useMemo(
    () =>
      ripples.map((ripple) => (
        <span
          key={ripple.id}
          className="pointer-events-none absolute h-1 w-1 rounded-full bg-white/40 ripple-anim"
          style={{ left: ripple.x, top: ripple.y }}
        />
      )),
    [ripples],
  )

  return (
    <motion.button
      whileHover={{ scale: 1.01 }}
      whileTap={{ scale: 0.97 }}
      className={cn(
        'relative overflow-hidden rounded-xl px-4 py-2 font-semibold text-white transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 disabled:cursor-not-allowed disabled:opacity-60',
        variantClassNames[variant],
        className,
      )}
      disabled={isDisabled}
      onClick={handleClick}
      {...rest}
    >
      <span className="relative z-10 inline-flex items-center justify-center gap-2">{buttonContent}</span>
      {variant === 'gradient' && <span className="pointer-events-none absolute inset-0">{rippleElements}</span>}
    </motion.button>
  )
}

