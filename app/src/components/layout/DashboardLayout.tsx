import { AnimatePresence, motion } from 'framer-motion'
import {
  ChevronLeft,
  ChevronRight,
  Film,
  Menu,
  Settings,
  Sparkles,
  Sun,
  Moon,
  User,
  CreditCard,
  UserCircle2,
} from 'lucide-react'
import { PropsWithChildren, useEffect, useMemo, useRef, useState } from 'react'
import { NavLink } from 'react-router-dom'
import { useTheme } from '../../context/ThemeContext'
import { cn } from '../../lib/cn'

type PlanTier = 'free' | 'pro' | 'max'

type DashboardLayoutProps = PropsWithChildren<{
  title: string
  subtitle?: string
  plan?: PlanTier
}>

const navItems = [
  { label: 'Generate', icon: Sparkles, to: '/dashboard', active: true },
  { label: 'My Videos', icon: Film, badge: 'Soon', disabled: true },
]

export const DashboardLayout = ({ title, subtitle, plan = 'max', children }: DashboardLayoutProps) => {
  const [isCollapsed, setIsCollapsed] = useState(false)
  const [isMobileOpen, setIsMobileOpen] = useState(false)
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false)
  const { theme, toggleTheme } = useTheme()

  const sidebarWidth = useMemo(() => (isCollapsed ? 72 : 256), [isCollapsed])
  const isLight = theme === 'light'
  const planLabel: Record<PlanTier, string> = {
    free: 'Free plan',
    pro: 'Pro plan',
    max: 'Max plan',
  }

  const userMenuRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (!isUserMenuOpen) return
    const handler = (event: MouseEvent) => {
      if (!userMenuRef.current?.contains(event.target as Node)) {
        setIsUserMenuOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [isUserMenuOpen])

  const SidebarContent = (
    <div className="relative flex h-full flex-col">
      <div className="flex items-center justify-between px-5 py-4">
        <div className="flex items-center gap-3">
          {isCollapsed ? (
            <button
              type="button"
              onClick={() => setIsCollapsed(false)}
              className={cn(
                'flex h-10 w-10 items-center justify-center rounded-full border text-foreground-secondary transition hover:scale-105',
                isLight
                  ? 'border-light-border bg-light-surface text-light-text hover:bg-light-accent'
                  : 'border-bg-highlight bg-bg-elevated/80',
              )}
              aria-label="Expand sidebar"
            >
              <ChevronRight className="h-5 w-5" />
            </button>
          ) : (
            <div
              className={cn(
                'flex h-10 w-10 items-center justify-center rounded-2xl text-primary',
                isLight ? 'bg-primary/15' : 'bg-primary/25',
              )}
            >
              <Sparkles className="h-5 w-5" />
            </div>
          )}
          {!isCollapsed && (
            <p className={cn('text-sm font-semibold', isLight ? 'text-light-text' : 'text-foreground')}>OmniGen Studio</p>
          )}
        </div>
        {!isCollapsed && (
          <button
            type="button"
            onClick={() => setIsCollapsed(true)}
            className={cn(
              'hidden rounded-full border p-2 transition hover:scale-105 md:inline-flex',
              isLight
                ? 'border-light-border bg-light-surface text-light-text hover:bg-light-accent'
                : 'border-bg-highlight bg-bg-elevated/60 text-foreground-secondary hover:text-foreground',
            )}
            aria-label="Collapse sidebar"
          >
            <ChevronLeft className="h-4 w-4" />
          </button>
        )}
      </div>

      <nav className="mt-2 flex flex-col gap-1 px-3">
        {navItems.map((item) =>
          item.to && !item.disabled ? (
            <NavLink
              key={item.label}
              to={item.to}
              className={({ isActive }) =>
                cn(
                  'group flex items-center rounded-2xl px-3 py-3 text-sm font-medium transition',
                  isLight
                    ? isActive
                      ? 'bg-gradient-to-r from-primary/20 to-secondary/15 text-primary'
                      : 'text-light-text-secondary hover:text-light-text hover:bg-light-accent'
                    : isActive
                      ? 'bg-gradient-to-r from-primary/30 to-secondary/20 text-foreground'
                      : 'text-foreground-secondary hover:text-foreground',
                )
              }
              onClick={() => setIsMobileOpen(false)}
            >
              <motion.span whileHover={{ rotate: 4, scale: 1.05 }}>
                <item.icon className={cn('h-5 w-5', !isCollapsed && 'mr-3')} />
              </motion.span>
              <AnimatePresence>
                {!isCollapsed && (
                  <motion.span
                    initial={{ opacity: 0, width: 0 }}
                    animate={{ opacity: 1, width: 'auto' }}
                    exit={{ opacity: 0, width: 0 }}
                    className="overflow-hidden"
                  >
                    {item.label}
                  </motion.span>
                )}
              </AnimatePresence>
            </NavLink>
          ) : (
            <button
              key={item.label}
              type="button"
              disabled
              className={cn(
                'group flex items-center rounded-2xl px-3 py-3 text-sm font-medium transition',
                isLight ? 'text-light-text-secondary/50' : 'text-foreground-muted',
              )}
            >
              <motion.span whileHover={{ rotate: 4, scale: 1.05 }}>
                <item.icon className={cn('h-5 w-5', !isCollapsed && 'mr-3')} />
              </motion.span>
              <AnimatePresence>
                {!isCollapsed && (
                  <motion.span
                    initial={{ opacity: 0, width: 0 }}
                    animate={{ opacity: 1, width: 'auto' }}
                    exit={{ opacity: 0, width: 0 }}
                    className="flex items-center gap-2 overflow-hidden"
                  >
                    {item.label}
                    {item.badge && (
                      <span
                        className={cn(
                          'rounded-full px-2 py-0.5 text-xs font-semibold',
                          isLight ? 'bg-light-accent text-light-text-secondary' : 'bg-bg-highlight text-foreground-secondary',
                        )}
                      >
                        {item.badge}
                      </span>
                    )}
                  </motion.span>
                )}
              </AnimatePresence>
            </button>
          ),
        )}
      </nav>

    </div>
  )

  return (
    <div
      className={cn(
        'flex min-h-screen flex-col transition-colors md:flex-row',
        isLight ? 'bg-light-bg text-light-text' : 'bg-background text-foreground',
      )}
    >
      <aside className="hidden md:block">
        <motion.div
          animate={{ width: sidebarWidth }}
          className={cn(
            'relative h-full min-h-screen shadow-lg transition-all',
            isLight ? 'bg-light-surface border-r border-light-border shadow-secondary/5' : 'bg-surface shadow-black/20',
          )}
          transition={{ duration: 0.25, ease: 'easeInOut' }}
        >
          {SidebarContent}
        </motion.div>
      </aside>

      <AnimatePresence>
        {isMobileOpen && (
          <>
            <motion.div
              className="fixed inset-0 z-40 bg-background/80 backdrop-blur"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => setIsMobileOpen(false)}
            />
            <motion.aside
              className="fixed inset-y-0 left-0 z-50 w-64 bg-surface shadow-xl shadow-black/40"
              initial={{ x: '-100%' }}
              animate={{ x: 0 }}
              exit={{ x: '-100%' }}
            >
              <button
                type="button"
                onClick={() => setIsMobileOpen(false)}
                className="absolute right-4 top-4 z-10 rounded-full border border-bg-highlight bg-bg-elevated/60 p-2 text-foreground-secondary transition hover:text-foreground"
                aria-label="Close sidebar"
              >
                <ChevronLeft className="h-4 w-4" />
              </button>
              <div className="h-full overflow-y-auto px-3 pt-3">{SidebarContent}</div>
            </motion.aside>
          </>
        )}
      </AnimatePresence>

      <div className="flex flex-1 flex-col">
        <header
          className={cn(
            'flex min-h-16 items-center justify-between border-b px-4 py-3 text-sm transition-all sm:px-6 sm:text-base lg:px-8',
            isLight ? 'border-light-border bg-light-surface/80 text-light-text' : 'border-bg-highlight bg-bg-elevated/80 text-foreground',
          )}
        >
          <div className="flex items-center gap-3">
            <button
              type="button"
              className={cn(
                'rounded-full border p-2 transition md:hidden',
                isLight
                  ? 'border-light-border bg-light-accent text-light-text hover:bg-secondary/10'
                  : 'border-bg-highlight bg-bg-highlight text-foreground-secondary hover:text-foreground',
              )}
              onClick={() => setIsMobileOpen(true)}
            >
              <Menu className="h-4 w-4" />
            </button>
            <div>
              <p
                className={cn(
                  'text-xs uppercase tracking-wide',
                  isLight ? 'text-light-text-secondary' : 'text-foreground-muted',
                )}
              >
                Dashboard
              </p>
              <h1 className={cn('text-lg font-semibold', isLight ? 'text-light-text' : 'text-foreground')}>{title}</h1>
              {subtitle && (
                <p className={cn('text-xs', isLight ? 'text-light-text-secondary' : 'text-foreground-muted')}>{subtitle}</p>
              )}
            </div>
          </div>

          <div className="flex items-center gap-3 sm:gap-4">
            <span
              className={cn(
                'rounded-full border px-3 py-1 text-xs font-semibold uppercase tracking-wide sm:px-4 sm:text-sm',
                isLight
                  ? 'border-light-border bg-light-surface text-primary shadow-inner shadow-secondary/10'
                  : 'border-primary/30 bg-primary/10 text-primary',
              )}
            >
              {planLabel[plan]}
            </span>
            <div className="relative" ref={userMenuRef}>
              <button
                type="button"
                onClick={() => setIsUserMenuOpen((prev) => !prev)}
                className={cn(
                  'flex items-center gap-2 rounded-full border px-3 py-1.5 text-sm transition',
                  isLight
                    ? 'border-light-border bg-light-surface text-light-text hover:bg-light-accent'
                    : 'border-bg-highlight bg-bg-highlight text-foreground-secondary',
                )}
              >
                <UserCircle2 className="h-6 w-6" />
                <span className="hidden md:inline">Akhil Patel</span>
              </button>
              <AnimatePresence>
                {isUserMenuOpen && (
                  <motion.div
                    initial={{ opacity: 0, y: -8, scale: 0.98 }}
                    animate={{ opacity: 1, y: 0, scale: 1 }}
                    exit={{ opacity: 0, y: -8, scale: 0.98 }}
                    className={cn(
                      'absolute right-0 mt-3 w-60 rounded-3xl border p-3 shadow-[0px_24px_60px_rgba(15,15,30,0.25)]',
                      isLight
                        ? 'border-light-border bg-white/95 text-light-text'
                        : 'border-bg-highlight bg-bg-elevated text-foreground',
                    )}
                  >
                    <button
                      type="button"
                      className={cn(
                        'flex w-full items-center gap-3 rounded-2xl px-3 py-2 text-left text-sm transition',
                        isLight ? 'hover:bg-light-accent' : 'hover:bg-bg-highlight',
                      )}
                      onClick={() => setIsUserMenuOpen(false)}
                    >
                      <User className="h-4 w-4" />
                      View profile
                    </button>
                    <button
                      type="button"
                      className={cn(
                        'flex w-full items-center gap-3 rounded-2xl px-3 py-2 text-left text-sm transition',
                        isLight ? 'hover:bg-light-accent' : 'hover:bg-bg-highlight',
                      )}
                      onClick={() => {
                        toggleTheme()
                        setIsUserMenuOpen(false)
                      }}
                    >
                      {isLight ? <Moon className="h-4 w-4" /> : <Sun className="h-4 w-4" />}
                      {isLight ? 'Switch to dark mode' : 'Switch to light mode'}
                    </button>
                    <button
                      type="button"
                      className={cn(
                        'flex w-full items-center gap-3 rounded-2xl px-3 py-2 text-left text-sm transition',
                        isLight ? 'hover:bg-light-accent' : 'hover:bg-bg-highlight',
                      )}
                      onClick={() => setIsUserMenuOpen(false)}
                    >
                      <CreditCard className="h-4 w-4" />
                      Manage subscription
                    </button>
                    <button
                      type="button"
                      className={cn(
                        'flex w-full items-center gap-3 rounded-2xl px-3 py-2 text-left text-sm transition',
                        isLight ? 'hover:bg-light-accent' : 'hover:bg-bg-highlight',
                      )}
                      onClick={() => {
                        setIsUserMenuOpen(false)
                      }}
                    >
                      <Settings className="h-4 w-4" />
                      Settings
                    </button>
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          </div>
        </header>

        <main className="flex-1 overflow-y-auto px-4 pb-28 pt-6 transition-[padding] sm:px-6 md:pb-10 lg:px-8">
          {children}
        </main>
      </div>

      <nav
        className={cn(
          'fixed bottom-0 left-0 right-0 z-40 flex items-center justify-around border-t px-4 py-3 text-xs backdrop-blur md:hidden',
          isLight ? 'border-light-border bg-light-surface/95 text-light-text' : 'border-bg-highlight bg-background/95 text-foreground',
        )}
      >
        {navItems.map((item) =>
          item.to && !item.disabled ? (
            <NavLink
              key={item.label}
              to={item.to}
              className={({ isActive }) =>
                cn(
                  'flex flex-col items-center gap-1 rounded-full px-3 py-1 transition',
                  isLight
                    ? isActive
                      ? 'text-primary'
                      : 'text-light-text-secondary'
                    : isActive
                      ? 'text-foreground'
                      : 'text-foreground-secondary',
                )
              }
              onClick={() => setIsMobileOpen(false)}
            >
              <item.icon className="h-5 w-5" />
              <span className="text-[0.65rem]">{item.label}</span>
            </NavLink>
          ) : (
            <div
              key={item.label}
              className={cn(
                'flex flex-col items-center gap-1',
                isLight ? 'text-light-text-secondary/50' : 'text-foreground-muted',
              )}
            >
              <item.icon className="h-5 w-5" />
              <span className="text-[0.65rem]">{item.label}</span>
            </div>
          ),
        )}
      </nav>
    </div>
  )
}

