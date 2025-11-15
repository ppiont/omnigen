import type { PropsWithChildren } from 'react'
import { useTheme } from '../context/ThemeContext'
import { cn } from '../lib/cn'

type SurfaceCardProps = {
  className?: string
}

export const SurfaceCard = ({ children, className }: PropsWithChildren<SurfaceCardProps>) => {
  const { theme } = useTheme()
  const isLight = theme === 'light'
  return (
    <div
      className={cn(
        'rounded-2xl p-5 sm:p-6 lg:p-7 backdrop-blur transition-all',
        isLight
          ? 'border border-bg-highlight bg-bg-elevated text-foreground shadow-lg shadow-primary/10'
          : 'border border-bg-highlight bg-bg-elevated/90 text-foreground shadow-2xl shadow-primary/5',
        className,
      )}
    >
      {children}
    </div>
  )
}

