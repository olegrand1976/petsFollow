<template>
  <div data-testid="research-settings-page">
    <ProPageHeader
      :title="$t('research.accountSettings.title')"
      :subtitle="$t('research.accountSettings.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning">{{ $t('nav.tagDev') }}</ProBadge>
      </template>
    </ProPageHeader>

    <ProCard :title="$t('settings.language.title')" class="pro-settings-card pro-mb-md">
      <p class="pro-hint">{{ $t('settings.language.subtitle') }}</p>
      <label class="pro-field">
        <span>{{ $t('settings.language.title') }}</span>
        <select v-model="locale" class="pro-select" data-testid="research-locale-select">
          <option v-for="loc in supportedLocales" :key="loc" :value="loc">
            {{ $t(`settings.language.${loc}`) }}
          </option>
        </select>
      </label>
      <p v-if="localeSaved" class="text-muted" role="status">{{ $t('settings.language.saved') }}</p>
      <p v-if="localeError" class="pro-field-error" role="alert">{{ localeError }}</p>
      <ProButton :loading="localeSaving" test-id="research-locale-save" @click="saveLocale">
        {{ $t('common.save') }}
      </ProButton>
    </ProCard>

    <ProCard :title="$t('settings.password.title')" class="pro-settings-card">
      <p class="pro-hint">{{ $t('settings.password.subtitle') }}</p>
      <form class="pro-form" @submit.prevent="changePassword">
        <ProInput
          v-model="currentPassword"
          type="password"
          :label="$t('settings.password.current')"
          test-id="research-password-current"
        />
        <ProInput
          v-model="newPassword"
          type="password"
          :label="$t('settings.password.new')"
          test-id="research-password-new"
        />
        <p v-if="passwordSaved" class="text-muted" role="status">{{ $t('settings.password.saved') }}</p>
        <p v-if="passwordError" class="pro-field-error" role="alert">{{ passwordError }}</p>
        <ProButton type="submit" :loading="passwordBusy" data-testid="research-password-submit">
          {{ $t('settings.password.save') }}
        </ProButton>
      </form>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
import type { AppLocale } from '~/composables/useLocaleSync'

definePageMeta({
  layout: 'research',
  middleware: ['research-only'],
})

const { t } = useI18n()
const { mapError } = useApiError()
const { saveLocale: persistLocale, supportedLocales } = useLocaleSync()
const { user, fetchUser } = useProUser()

const locale = ref<AppLocale>((user.value?.preferredLocale as AppLocale) || 'fr')
const localeSaving = ref(false)
const localeSaved = ref(false)
const localeError = ref('')
const currentPassword = ref('')
const newPassword = ref('')
const passwordBusy = ref(false)
const passwordSaved = ref(false)
const passwordError = ref('')

onMounted(async () => {
  try {
    const me = await fetchUser(true)
    if (me?.preferredLocale) locale.value = me.preferredLocale as AppLocale
  } catch { /* ignore */ }
})

async function saveLocale() {
  localeSaving.value = true
  localeSaved.value = false
  localeError.value = ''
  try {
    await persistLocale(locale.value)
    localeSaved.value = true
  } catch (e: any) {
    localeError.value = mapError(e) || t('common.save')
  } finally {
    localeSaving.value = false
  }
}

async function changePassword() {
  passwordBusy.value = true
  passwordSaved.value = false
  passwordError.value = ''
  try {
    await $fetch('/api/me/password', {
      method: 'PATCH',
      body: { currentPassword: currentPassword.value, newPassword: newPassword.value },
    })
    passwordSaved.value = true
    currentPassword.value = ''
    newPassword.value = ''
  } catch (e: any) {
    passwordError.value = mapError(e) || t('settings.password.save')
  } finally {
    passwordBusy.value = false
  }
}
</script>
