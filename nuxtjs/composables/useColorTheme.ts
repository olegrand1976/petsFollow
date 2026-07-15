export type ColorTheme = 'light' | 'dark'

const THEME_COOKIE = 'pf_theme'

export function useColorTheme() {
  const theme = useCookie<ColorTheme>(THEME_COOKIE, {
    default: () => 'light',
    maxAge: 60 * 60 * 24 * 365,
  })

  const isDark = computed(() => theme.value === 'dark')

  useHead({
    htmlAttrs: {
      class: computed(() => (isDark.value ? 'dark' : '')),
    },
  })

  function toggleTheme() {
    theme.value = isDark.value ? 'light' : 'dark'
  }

  function setTheme(value: ColorTheme) {
    theme.value = value
  }

  return { theme, isDark, toggleTheme, setTheme }
}
