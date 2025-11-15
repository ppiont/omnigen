import { AnimatePresence, motion } from 'framer-motion'
import { useEffect, useMemo, useRef, useState } from 'react'
import { ChevronDown, Loader2, PlayCircle, Sparkles } from 'lucide-react'
import { DashboardLayout } from '../components/layout/DashboardLayout'
import { SurfaceCard } from '../components/SurfaceCard'
import { PrimaryButton } from '../components/PrimaryButton'
import { Toast } from '../components/ui/Toast'
import { ToggleSwitch } from '../components/ui/ToggleSwitch'
import { VideoCard, VideoCardSkeleton } from '../components/ui/VideoCard'
import { useTheme } from '../context/ThemeContext'
import { cn } from '../lib/cn'

const characterLimit = 2000
const categories = ['Music Video', 'Ad Creative', 'Explainer'] as const
const stylePresets = ['Cinematic', 'Anime', 'Realistic', 'Abstract'] as const
const aspectRatios = ['16:9', '9:16', '1:1'] as const
const durationOptions = [15, 30, 60, 90]

const galleryItems = [
  {
    id: '1',
    prompt: 'Cyberpunk city skyline with neon rain and reflective streets',
    duration: '0:30',
    date: 'Nov 12, 2025',
    thumbnail: 'https://images.unsplash.com/photo-1469474968028-56623f02e42e?auto=format&fit=crop&w=900&q=80',
  },
  {
    id: '2',
    prompt: 'Product showcase for a minimalist smart speaker in a white studio',
    duration: '0:45',
    date: 'Nov 10, 2025',
    thumbnail: 'https://images.unsplash.com/photo-1498050108023-c5249f4df085?auto=format&fit=crop&w=900&q=80',
  },
  {
    id: '3',
    prompt: 'Anime-inspired hero landing sequence with dramatic lighting',
    duration: '0:60',
    date: 'Nov 6, 2025',
    thumbnail: 'https://images.unsplash.com/photo-1469474968028-56623f02e42e?auto=format&fit=crop&w=900&q=80',
  },
] satisfies Array<{ id: string; prompt: string; duration: string; date: string; thumbnail: string }>

export const DashboardPage = () => {
  const { theme } = useTheme()
  const isLight = theme === 'light'
  const [prompt, setPrompt] = useState('')
  const [isAdvancedOpen, setIsAdvancedOpen] = useState(false)
  const [selectedCategory, setSelectedCategory] = useState<(typeof categories)[number]>('Music Video')
  const [selectedStyle, setSelectedStyle] = useState<(typeof stylePresets)[number]>('Cinematic')
  const [durationIndex, setDurationIndex] = useState(1)
  const [selectedAspect, setSelectedAspect] = useState<(typeof aspectRatios)[number]>('16:9')
  const [isGenerating, setIsGenerating] = useState(false)
  const [progress, setProgress] = useState(0)
  const [autoEnhance, setAutoEnhance] = useState(true)
  const [loopVideo, setLoopVideo] = useState(false)
  const [galleryLoading, setGalleryLoading] = useState(true)
  const [showToast, setShowToast] = useState(false)
  const textareaRef = useRef<HTMLTextAreaElement | null>(null)

  const selectedDuration = durationOptions[durationIndex]
  const trimmedPrompt = prompt.trim()

  const estimatedTime = useMemo(() => {
    if (selectedDuration <= 15) return '~30s'
    if (selectedDuration <= 30) return '~45s'
    if (selectedDuration <= 60) return '1 min'
    return '1-2 min'
  }, [selectedDuration])

  const estimatedCost = useMemo(() => {
    const cost = (selectedDuration / 30) * 1.5
    return `$${cost.toFixed(2)}`
  }, [selectedDuration])

  const handlePromptChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = event.target.value.slice(0, characterLimit)
    setPrompt(value)
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`
    }
  }

  const handleGenerate = () => {
    if (!trimmedPrompt) return
    setProgress(5)
    setIsGenerating(true)
  }

  useEffect(() => {
    const timeout = setTimeout(() => setGalleryLoading(false), 900)
    return () => clearTimeout(timeout)
  }, [])

  useEffect(() => {
    if (!isGenerating) {
      return
    }
    const interval = setInterval(() => {
      setProgress((prev) => (prev >= 95 ? 95 : prev + 5))
    }, 250)

    const timeout = setTimeout(() => {
      setProgress(100)
      setTimeout(() => {
        setIsGenerating(false)
        setProgress(0)
        setShowToast(true)
      }, 500)
    }, 6000)

    return () => {
      clearInterval(interval)
      clearTimeout(timeout)
    }
  }, [isGenerating])

  useEffect(() => {
    if (!showToast) return
    const timeout = setTimeout(() => setShowToast(false), 3000)
    return () => clearTimeout(timeout)
  }, [showToast])

  const galleryContainer = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.08,
      },
    },
  }

  const galleryItem = {
    hidden: { opacity: 0, y: 12 },
    visible: { opacity: 1, y: 0 },
  }

  return (
    <DashboardLayout title="Generate" subtitle="Describe your next AI-native video">
      <Toast isOpen={showToast}>Video queued. You&apos;ll be notified soon.</Toast>
      <div className="mx-auto w-full max-w-6xl space-y-8 sm:space-y-10">
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
          className="space-y-6"
        >
          <div className="grid gap-5 sm:gap-6 md:grid-cols-2 xl:grid-cols-[1.15fr,0.85fr]">
            <SurfaceCard
              className={cn(
                'flex h-full flex-col space-y-4 transition',
                isLight
                  ? 'border-light-border focus-within:border-primary/60 focus-within:shadow-primary/20'
                  : 'border-bg-highlight focus-within:border-primary/50 focus-within:shadow-[0_0_35px_rgba(124,255,0,0.25)]',
              )}
            >
              <div className="flex items-center justify-between">
                <div>
                  <h2 className={cn('text-lg font-semibold sm:text-xl', isLight ? 'text-light-text' : 'text-foreground')}>
                    Generate a video
                  </h2>
                  <p className={cn('text-xs sm:text-sm', isLight ? 'text-light-text-secondary' : 'text-foreground-secondary')}>
                    Describe the scene, mood, and style.
                  </p>
                </div>
                <span
                  className={cn('text-[0.7rem] sm:text-xs', isLight ? 'text-light-text-secondary' : 'text-foreground-muted')}
                >
                  {prompt.length}/{characterLimit}
                </span>
              </div>
              <textarea
                ref={textareaRef}
                value={prompt}
                onChange={handlePromptChange}
                placeholder="Describe your video... e.g., 'Create a music video with cyberpunk aesthetics and neon lights'"
                maxLength={characterLimit}
                className={cn(
                  'min-h-[160px] w-full resize-none rounded-2xl border px-4 py-3 text-base outline-none transition focus:ring-2',
                  isLight
                    ? 'border-light-border bg-white text-light-text placeholder:text-light-text-secondary/60 focus:border-primary focus:ring-primary/30'
                    : 'border-bg-highlight bg-bg-highlight text-foreground placeholder:text-foreground-muted focus:border-primary focus:ring-primary/40',
                )}
              />
            </SurfaceCard>

            <SurfaceCard className="flex h-full flex-col space-y-4">
              <div className={cn('flex items-center justify-between', isLight ? 'text-light-text' : 'text-foreground')}>
                <h3 className="text-base font-semibold sm:text-lg">Preview</h3>
                <p className={cn('text-xs', isLight ? 'text-light-text-secondary' : 'text-foreground-secondary')}>
                  {selectedAspect} • {selectedStyle}
                </p>
              </div>

              <div
                className={cn(
                  'rounded-3xl border border-dashed p-4',
                  isLight ? 'border-light-border bg-light-accent/30' : 'border-bg-highlight bg-bg-highlight',
                )}
              >
                <div
                  className={cn(
                    'w-full rounded-2xl',
                    isLight ? 'bg-gradient-to-br from-secondary/10 to-primary/5' : 'bg-gradient-to-br from-white/10 to-white/5',
                  )}
                >
                  <div
                    className={cn(
                      'flex min-h-[200px] items-center justify-center text-center',
                      isLight ? 'text-light-text-secondary' : 'text-foreground-secondary',
                    )}
                  >
                    <AnimatePresence mode="wait">
                      {isGenerating ? (
                        <motion.div
                          key="generating"
                          className="w-full space-y-4"
                          initial={{ opacity: 0 }}
                          animate={{ opacity: 1 }}
                          exit={{ opacity: 0 }}
                        >
                          <div
                            className={cn('h-40 rounded-2xl', isLight ? 'bg-secondary/10' : 'bg-white/10')}
                          />
                          <div className="space-y-2">
                            <div
                              className={cn(
                                'flex items-center justify-between text-xs',
                                isLight ? 'text-light-text-secondary' : 'text-foreground-secondary',
                              )}
                            >
                              <span>Rendering frames</span>
                              <span>{progress.toFixed(0)}%</span>
                            </div>
                            <div
                              className={cn(
                                'h-2 w-full rounded-full',
                                isLight ? 'bg-light-border' : 'bg-bg-highlight',
                              )}
                            >
                              <motion.div
                                className="h-full rounded-full bg-gradient-to-r from-primary to-aurora-teal"
                                animate={{ width: `${progress}%` }}
                                transition={{ ease: 'easeOut', duration: 0.3 }}
                              />
                            </div>
                          </div>
                        </motion.div>
                      ) : (
                        <motion.div
                          key="idle"
                          className={cn(
                            'flex flex-col items-center justify-center gap-3 px-6 text-sm',
                            isLight ? 'text-light-text-secondary' : 'text-foreground-secondary',
                          )}
                          initial={{ opacity: 0 }}
                          animate={{ opacity: 1 }}
                          exit={{ opacity: 0 }}
                        >
                          <PlayCircle
                            className={cn('h-10 w-10', isLight ? 'text-light-text-secondary/60' : 'text-foreground-muted')}
                          />
                          Your video will appear here
                        </motion.div>
                      )}
                    </AnimatePresence>
                  </div>
                </div>
              </div>

              <div
                className={cn(
                  'grid gap-4 rounded-2xl border p-4 text-sm sm:grid-cols-2',
                  isLight
                    ? 'border-light-border bg-light-accent/30 text-light-text'
                    : 'border-bg-highlight bg-bg-highlight text-foreground-secondary',
                )}
              >
                <div>
                  <p
                    className={cn(
                      'text-xs uppercase tracking-wide',
                      isLight ? 'text-light-text-secondary' : 'text-foreground-muted',
                    )}
                  >
                    Estimated time
                  </p>
                  <p className={cn('mt-1 text-lg font-semibold', isLight ? 'text-light-text' : 'text-foreground')}>
                    {estimatedTime}
                  </p>
                </div>
                <div>
                  <p
                    className={cn(
                      'text-xs uppercase tracking-wide',
                      isLight ? 'text-light-text-secondary' : 'text-foreground-muted',
                    )}
                  >
                    Estimated cost
                  </p>
                  <p className={cn('mt-1 text-lg font-semibold', isLight ? 'text-light-text' : 'text-foreground')}>
                    {estimatedCost}
                  </p>
                </div>
              </div>
            </SurfaceCard>
          </div>

          <SurfaceCard className="space-y-4">
            <button
              type="button"
              onClick={() => setIsAdvancedOpen((prev) => !prev)}
              className={cn(
                'flex w-full items-center gap-3 text-left text-sm font-semibold',
                isLight ? 'text-light-text' : 'text-foreground',
              )}
            >
              <span>Advanced options</span>
              <motion.span
                className="ml-auto pr-1"
                animate={{ rotate: isAdvancedOpen ? 0 : -90 }}
                transition={{ duration: 0.2 }}
              >
                <ChevronDown className={cn('h-4 w-4', isLight ? 'text-light-text-secondary' : 'text-foreground-secondary')} />
              </motion.span>
            </button>

            <AnimatePresence initial={false}>
              {isAdvancedOpen && (
                <motion.div
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: 'auto' }}
                  exit={{ opacity: 0, height: 0 }}
                  transition={{ duration: 0.25 }}
                  className="space-y-5 overflow-hidden"
                >
                  <div>
                    <p
                      className={cn(
                        'text-xs font-semibold uppercase tracking-wide',
                        isLight ? 'text-light-text-secondary' : 'text-foreground-muted',
                      )}
                    >
                      Category
                    </p>
                    <select
                      value={selectedCategory}
                      onChange={(event) =>
                        setSelectedCategory(event.target.value as (typeof categories)[number])
                      }
                      className={cn(
                        'mt-2 w-full rounded-2xl border px-4 py-2 text-sm outline-none transition focus:ring-2',
                        isLight
                          ? 'border-light-border bg-white text-light-text focus:border-secondary focus:ring-secondary/20'
                          : 'border-bg-highlight bg-bg-highlight text-foreground focus:border-secondary focus:ring-secondary/30',
                      )}
                    >
                      {categories.map((category) => (
                        <option
                          key={category}
                          value={category}
                          className={isLight ? 'bg-white text-light-text' : 'bg-bg-elevated text-foreground'}
                        >
                          {category}
                        </option>
                      ))}
                    </select>
                  </div>

                  <div>
                    <p
                      className={cn(
                        'text-xs font-semibold uppercase tracking-wide',
                        isLight ? 'text-light-text-secondary' : 'text-foreground-muted',
                      )}
                    >
                      Style preset
                    </p>
                    <div className="mt-3 flex flex-wrap gap-2">
                      {stylePresets.map((preset) => (
                        <button
                          key={preset}
                          type="button"
                          onClick={() => setSelectedStyle(preset)}
                          className={cn(
                            'rounded-full border px-4 py-1.5 text-sm transition',
                            selectedStyle === preset
                              ? isLight
                                ? 'border-primary bg-primary/10 text-primary'
                                : 'border-primary bg-primary/20 text-foreground'
                              : isLight
                                ? 'border-light-border text-light-text-secondary hover:text-light-text hover:bg-light-accent'
                                : 'border-bg-highlight text-foreground-secondary hover:text-foreground',
                          )}
                        >
                          {preset}
                        </button>
                      ))}
                    </div>
                  </div>

                  <div>
                    <p
                      className={cn(
                        'text-xs font-semibold uppercase tracking-wide',
                        isLight ? 'text-light-text-secondary' : 'text-foreground-muted',
                      )}
                    >
                      Duration
                    </p>
                    <div className="mt-4 px-1">
                      <input
                        type="range"
                        min={0}
                        max={durationOptions.length - 1}
                        step={1}
                        value={durationIndex}
                        onChange={(event) => setDurationIndex(Number(event.target.value))}
                        className="w-full accent-primary"
                      />
                      <div
                        className={cn(
                          'mt-2 flex justify-between text-xs',
                          isLight ? 'text-light-text-secondary' : 'text-foreground-muted',
                        )}
                      >
                        {durationOptions.map((duration, index) => (
                          <span
                            key={duration}
                            className={cn(
                              'w-10 text-center',
                              durationIndex === index &&
                                (isLight ? 'text-light-text font-semibold' : 'text-foreground font-semibold'),
                            )}
                          >
                            {duration}s
                          </span>
                        ))}
                      </div>
                    </div>
                  </div>

                  <div>
                    <p
                      className={cn(
                        'text-xs font-semibold uppercase tracking-wide',
                        isLight ? 'text-light-text-secondary' : 'text-foreground-muted',
                      )}
                    >
                      Aspect ratio
                    </p>
                    <div className="mt-3 flex flex-wrap gap-2">
                      {aspectRatios.map((ratio) => (
                        <button
                          key={ratio}
                          type="button"
                          onClick={() => setSelectedAspect(ratio)}
                          className={cn(
                            'rounded-2xl border px-4 py-2 text-sm transition',
                            selectedAspect === ratio
                              ? isLight
                                ? 'border-secondary bg-secondary/15 text-secondary'
                                : 'border-secondary bg-secondary/15 text-foreground'
                              : isLight
                                ? 'border-light-border text-light-text-secondary hover:text-light-text hover:bg-light-accent'
                                : 'border-bg-highlight text-foreground-secondary hover:text-foreground',
                          )}
                        >
                          {ratio}
                        </button>
                      ))}
                    </div>
                  </div>

                  <div
                    className={cn(
                      'flex flex-col gap-4 rounded-2xl border p-4 text-sm',
                      isLight ? 'border-light-border bg-light-accent/40' : 'border-bg-highlight bg-bg-highlight',
                    )}
                  >
                    <div className="flex items-center justify-between gap-4">
                      <div>
                        <p className={cn('font-semibold', isLight ? 'text-light-text' : 'text-foreground')}>Auto-enhance</p>
                        <p className={cn('text-xs', isLight ? 'text-light-text-secondary' : 'text-foreground-secondary')}>
                          Improve details and color balance.
                        </p>
                      </div>
                      <ToggleSwitch checked={autoEnhance} onClick={() => setAutoEnhance((prev) => !prev)} />
                    </div>
                    <div className="flex items-center justify-between gap-4">
                      <div>
                        <p className={cn('font-semibold', isLight ? 'text-light-text' : 'text-foreground')}>Loop video</p>
                        <p className={cn('text-xs', isLight ? 'text-light-text-secondary' : 'text-foreground-secondary')}>
                          Perfect for hero sections.
                        </p>
                      </div>
                      <ToggleSwitch checked={loopVideo} onClick={() => setLoopVideo((prev) => !prev)} />
                    </div>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </SurfaceCard>

          <PrimaryButton
            className="w-full py-4 text-base"
            variant="gradient"
            loading={isGenerating}
            loadingContent={
              <span className="flex items-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" />
                Generating...
              </span>
            }
            disabled={!trimmedPrompt || isGenerating}
            onClick={handleGenerate}
          >
            <span className="flex items-center gap-2">
              <Sparkles className="h-5 w-5" />
              Generate Video
            </span>
          </PrimaryButton>
        </motion.div>
        <motion.section
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="space-y-4"
        >
          <div className="flex items-center justify-between">
            <div>
              <h3 className={cn('text-lg font-semibold', isLight ? 'text-light-text' : 'text-foreground')}>
                Recent videos
              </h3>
              <p className={cn('text-sm', isLight ? 'text-light-text-secondary' : 'text-foreground-secondary')}>
                Your latest generations appear here.
              </p>
            </div>
          </div>

          {galleryLoading ? (
            <div className="grid gap-4 sm:gap-5 md:grid-cols-2 lg:grid-cols-3">
              {Array.from({ length: 3 }).map((_, index) => (
                <VideoCardSkeleton key={index} />
              ))}
            </div>
          ) : galleryItems.length === 0 ? (
            <SurfaceCard
              className={cn(
                'flex flex-col items-center gap-3 py-16 text-center',
                isLight ? 'text-light-text-secondary' : 'text-foreground-secondary',
              )}
            >
              <PlayCircle
                className={cn('h-10 w-10', isLight ? 'text-light-text-secondary/50' : 'text-foreground-muted')}
              />
              <p>No videos yet. Start by generating your first video!</p>
            </SurfaceCard>
          ) : (
            <motion.div
              className="grid gap-4 sm:gap-5 md:grid-cols-2 lg:grid-cols-3"
              variants={galleryContainer}
              initial="hidden"
              animate="visible"
            >
              {galleryItems.map((item) => (
                <motion.div key={item.id} variants={galleryItem}>
                  <VideoCard
                    thumbnail={item.thumbnail}
                    title={item.prompt}
                    duration={item.duration}
                    createdAt={item.date}
                    onClick={() => console.info('Open video', item.id)}
                  />
                </motion.div>
              ))}
            </motion.div>
          )}
        </motion.section>
      </div>
    </DashboardLayout>
  )
}

