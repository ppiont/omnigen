import { motion } from 'framer-motion'
import { PlayCircle } from 'lucide-react'
import { cn } from '../../lib/cn'

export type VideoCardProps = {
  thumbnail: string
  title: string
  duration: string
  createdAt: string | Date
  onClick?: () => void
  className?: string
}

const formatDate = (value: string | Date) => {
  const date = typeof value === 'string' ? new Date(value) : value
  return date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

export const VideoCard = ({ thumbnail, title, duration, createdAt, onClick, className }: VideoCardProps) => {
  return (
    <motion.button
      type="button"
      whileHover={{ y: -6, scale: 1.02 }}
      className={cn(
        'group w-full rounded-3xl bg-gradient-to-br from-white/10 via-white/5 to-transparent p-[1px] text-left focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary',
        className,
      )}
      onClick={onClick}
      aria-label={`Play video: ${title}`}
    >
      <div className="h-full rounded-[calc(1.5rem-1px)] bg-bg-elevated/80 p-4 transition group-hover:bg-bg-elevated/90">
        <div className="relative overflow-hidden rounded-2xl">
          <img src={thumbnail} alt={title} className="aspect-video w-full object-cover" />
          <span className="absolute left-3 top-3 rounded-full bg-black/70 px-3 py-1 text-xs font-semibold text-white">
            {duration}
          </span>
          <motion.div
            className="absolute inset-0 flex items-center justify-center bg-black/60 text-white opacity-0"
            initial={false}
            whileHover={{ opacity: 1 }}
            transition={{ duration: 0.2 }}
          >
            <PlayCircle className="h-10 w-10" />
          </motion.div>
        </div>
        <div className="mt-4 space-y-1 text-sm text-foreground-secondary">
          <p className="line-clamp-2 font-semibold text-foreground">{title}</p>
          <p className="text-xs text-foreground-muted">{formatDate(createdAt)}</p>
        </div>
      </div>
    </motion.button>
  )
}

export const VideoCardSkeleton = () => {
  return (
    <div className="rounded-3xl border border-bg-highlight bg-bg-elevated/60 p-4">
      <div className="relative overflow-hidden rounded-2xl">
        <div className="shimmer aspect-video w-full rounded-2xl bg-white/5" />
        <span className="absolute left-3 top-3 h-5 w-12 rounded-full bg-black/30" />
      </div>
      <div className="mt-4 space-y-2">
        <div className="shimmer h-4 w-3/4 rounded-full bg-white/10" />
        <div className="shimmer h-3 w-1/2 rounded-full bg-white/5" />
      </div>
    </div>
  )
}

