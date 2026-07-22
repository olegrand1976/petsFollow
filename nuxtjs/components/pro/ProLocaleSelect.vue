<template>
  <select
    :value="locale"
    class="pro-select pro-locale-select"
    data-testid="locale-select"
    :aria-label="$t('settings.language.title')"
    :disabled="saving"
    @change="onChange"
  >
    <option v-for="loc in supportedLocales" :key="loc" :value="loc">
      {{ $t(`settings.language.${loc}`) }}
    </option>
  </select>
</template>

<script setup lang="ts">
import type { AppLocale } from '~/composables/useLocaleSync'

const props = withDefaults(defineProps<{ persist?: boolean }>(), { persist: false })

const { locale, switchLocale, saveLocale, supportedLocales } = useLocaleSync()
const saving = ref(false)

async function onChange(e: Event) {
  const value = (e.target as HTMLSelectElement).value as AppLocale
  saving.value = true
  try {
    if (props.persist) {
      await saveLocale(value)
    } else {
      await switchLocale(value)
    }
  } finally {
    saving.value = false
  }
}
</script>
