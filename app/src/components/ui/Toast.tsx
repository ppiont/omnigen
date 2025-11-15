import type { PropsWithChildren } from 'react'
import { AnimatePresence, motion } from 'framer-motion'

export type ToastProps = PropsWithChildren<{
  isOpen: boolean
}>

export const Toast = ({ isOpen, children }: ToastProps) => {
  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          className="fixed left-1/2 top-6 z-50 -translate-x-1/2 rounded-full border border-white/10 bg-black/80 px-6 py-3 text-sm text-white shadow-xl backdrop-blur"
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -20 }}
        >
          {children}
        </motion.div>
      )}
    </AnimatePresence>
  )
}

