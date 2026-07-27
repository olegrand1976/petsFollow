/** Environnement front Nuxt : local | staging | production */
export type AppEnv = 'local' | 'staging' | 'production'

export function parseAppEnv(raw: unknown): AppEnv {
  const v = String(raw || 'local').toLowerCase()
  return (['local', 'staging', 'production'].includes(v) ? v : 'local') as AppEnv
}

export function useAppEnv() {
  const appEnv = parseAppEnv(useRuntimeConfig().public.appEnv)
  /** Pages use cases : staging (et local pour développer). Jamais production. */
  const isStagingLike = computed(() => appEnv === 'staging' || appEnv === 'local')
  const isStaging = computed(() => appEnv === 'staging')
  return { appEnv, isStaging, isStagingLike }
}
