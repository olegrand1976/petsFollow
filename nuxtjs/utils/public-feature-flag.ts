/**
 * Opt-in NUXT_PUBLIC_* feature flag (build-time / process.env).
 * Explicit true/1 or false/0 wins; unset → on in `nuxt dev` (parity `make nuxtjs-dev`),
 * off in production builds.
 */
export function publicFeatureFlag(envKey: string): boolean {
  const v = process.env[envKey]
  if (v === 'true' || v === '1') return true
  if (v === 'false' || v === '0') return false
  return process.env.NODE_ENV !== 'production'
}

/**
 * Runtime truthiness for `runtimeConfig.public.*` flags.
 * Nuxt overrides NUXT_PUBLIC_* as strings — never use Boolean(value)
 * (Boolean("false") === true).
 */
export function isPublicFlagOn(value: unknown): boolean {
  return value === true || value === 'true' || value === '1' || value === 1
}
