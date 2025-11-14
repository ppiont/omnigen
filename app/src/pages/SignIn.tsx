import { motion } from 'framer-motion'
import { Eye, EyeOff, Lock, Mail, Sparkles } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Link } from 'react-router-dom'
import { PrimaryButton } from '../components/PrimaryButton'
import { Checkbox } from '../components/ui/Checkbox'
import type { SignInFormValues } from '../types/forms'

export const SignInPage = () => {
  const [showPassword, setShowPassword] = useState(false)
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<SignInFormValues>({
    defaultValues: {
      email: '',
      password: '',
      rememberMe: true,
    },
    mode: 'onBlur',
  })

  const onSubmit = async (values: SignInFormValues) => {
    await new Promise((resolve) => setTimeout(resolve, 800))
    console.info('Sign in', values)
  }

  return (
    <div className="relative min-h-screen overflow-hidden bg-[#0a0a0a] px-4 pb-24 pt-16 sm:px-8">
      <motion.div
        aria-hidden
        className="pointer-events-none absolute inset-0 opacity-70"
        initial={{ backgroundPosition: '0% 50%' }}
        animate={{ backgroundPosition: ['0% 50%', '100% 50%', '0% 50%'] }}
        transition={{ duration: 18, repeat: Infinity }}
        style={{
          backgroundImage:
            'linear-gradient(120deg, rgba(139,92,246,0.35), rgba(59,130,246,0.15), rgba(139,92,246,0.35))',
          backgroundSize: '200% 200%',
          filter: 'blur(70px)',
        }}
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 opacity-20"
        style={{
          backgroundImage: 'radial-gradient(rgba(255,255,255,0.08) 1px, transparent 1px)',
          backgroundSize: '26px 26px',
        }}
      />

      <div className="relative z-10 mx-auto flex min-h-[70vh] max-w-6xl items-center justify-center">
        <motion.div
          initial={{ opacity: 0, y: 24, scale: 0.98 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          transition={{ duration: 0.6, ease: 'easeOut' }}
          className="w-full max-w-md rounded-3xl border border-white/10 bg-white/5 p-6 text-sm shadow-2xl shadow-primary/20 backdrop-blur-2xl sm:p-8 sm:text-base"
        >
          <div className="mb-8 space-y-4 text-center">
            <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-primary to-secondary text-white shadow-lg shadow-primary/30">
              <Sparkles className="h-7 w-7" />
            </div>
            <div>
              <h1 className="text-xl font-semibold text-white sm:text-2xl">Welcome to OmniGen</h1>
              <p className="text-xs text-white/70 sm:text-sm">Generate AI videos from text</p>
            </div>
          </div>

          <form className="space-y-5" onSubmit={handleSubmit(onSubmit)} noValidate>
            <div className="space-y-2">
              <label className="text-[0.7rem] font-semibold uppercase tracking-wide text-white/60 sm:text-xs">
                Email
              </label>
              <div className="relative">
                <Mail className="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-white/40" />
                <input
                  type="email"
                  placeholder="you@example.com"
                  className="w-full rounded-2xl border border-white/15 bg-black/30 px-12 py-3 text-white outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/40"
                  {...register('email', {
                    required: 'Email is required',
                    pattern: { value: /\S+@\S+\.\S+/, message: 'Enter a valid email address' },
                  })}
                />
              </div>
              {errors.email && <p className="text-xs text-secondary">{errors.email.message}</p>}
            </div>

            <div className="space-y-2">
              <label className="text-[0.7rem] font-semibold uppercase tracking-wide text-white/60 sm:text-xs">
                Password
              </label>
              <div className="relative">
                <Lock className="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-white/40" />
                <input
                  type={showPassword ? 'text' : 'password'}
                  placeholder="••••••••"
                  className="w-full rounded-2xl border border-white/15 bg-black/30 px-12 py-3 pr-12 text-white outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/40"
                  {...register('password', {
                    required: 'Password is required',
                    minLength: { value: 8, message: 'Use at least 8 characters' },
                  })}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((prev) => !prev)}
                  className="absolute right-4 top-1/2 -translate-y-1/2 text-white/50 transition hover:text-white/80"
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
              {errors.password && <p className="text-xs text-secondary">{errors.password.message}</p>}
            </div>

            <Checkbox label="Remember me" {...register('rememberMe')} />

            <PrimaryButton type="submit" loading={isSubmitting} className="w-full py-3" variant="gradient">
              {isSubmitting ? 'Signing in...' : 'Sign In'}
            </PrimaryButton>

            <div className="flex items-center gap-4 text-xs uppercase tracking-wide text-white/40">
              <span className="h-px flex-1 bg-white/10" />
              or
              <span className="h-px flex-1 bg-white/10" />
            </div>

            <button
              type="button"
              className="flex w-full items-center justify-center gap-3 rounded-2xl border border-white/15 bg-white/5 px-4 py-3 text-sm font-semibold text-white/80 transition hover:border-white/30 hover:bg-white/10"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                aria-hidden
                viewBox="0 0 24 24"
                className="h-5 w-5"
              >
                <path
                  fill="#4285F4"
                  d="M23.04 12.261c0-.815-.073-1.596-.209-2.348H12v4.44h6.211c-.268 1.44-1.079 2.662-2.3 3.478v2.89h3.713c2.173-2 3.416-4.946 3.416-8.46"
                />
                <path
                  fill="#34A853"
                  d="M12 24c3.24 0 5.951-1.073 7.934-2.879l-3.713-2.89c-1.035.696-2.356 1.108-4.221 1.108-3.247 0-5.993-2.192-6.976-5.146H1.194v3.043C3.166 21.316 7.245 24 12 24"
                />
                <path
                  fill="#FBBC05"
                  d="M5.024 14.193A7.213 7.213 0 0 1 4.642 12c0-.763.132-1.507.372-2.193V6.764H1.194A11.997 11.997 0 0 0 0 12c0 1.947.465 3.788 1.194 5.236z"
                />
                <path
                  fill="#EA4335"
                  d="M12 4.749c1.763 0 3.343.607 4.587 1.797L19 4.133C15.951 1.31 12 0 12 0 7.245 0 3.166 2.684 1.194 6.764l3.82 3.043C6.007 6.941 8.753 4.749 12 4.749"
                />
              </svg>
              Continue with Google
            </button>
          </form>

          <p className="mt-8 text-center text-sm text-white/60">
            Don&apos;t have an account?{' '}
            <Link to="/dashboard" className="text-secondary transition hover:text-secondary/80">
              Sign up
            </Link>
          </p>
        </motion.div>
      </div>
    </div>
  )
}

