<template>
  <select
    :value="locale"
    class="pro-select pro-locale-select"
    data-testid="locale-select"
    :aria-label="$t('settings.language.title')"
    @change="onChange"
  >
    <option v-for="loc in supportedLocales" :key="loc" :value="loc">
      {{ $t(`settings.language.${loc}`) }}
    </option>
  </select>
</template>

<script setup lang="ts">
import type { AppLocale } from '~/composables/useLocaleSync'

const { locale, switchLocale, supportedLocales } = useLocaleSync()

async function onChange(e: Event) {
  const value = (e.target as HTMLSelectElement).value as AppLocale
  await switchLocale(value)
}
</script>
