<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ $t('auth.forceChange.brandTitle') }}</h2>
      <p>{{ $t('auth.forceChange.brandText') }}</p>
    </aside>
    <div class="pro-login-form-panel">
      <form
        class="pro-login-form"
        data-testid="force-change-password-form"
        method="post"
        action="#"
        @submit.prevent="submit"
      >
        <PetsFollowLogo variant="default" />
        <h1>{{ $t('auth.forceChange.title') }}</h1>
        <p class="pro-page-header__subtitle">{{ $t('auth.forceChange.subtitle') }}</p>
        <ProInput
          v-model="password"
          :label="$t('auth.forceChange.password')"
          type="password"
          name="password"
          autocomplete="new-password"
          required
          test-id="force-change-password"
        />
        <ProInput
          v-model="confirmPassword"
          :label="$t('auth.forceChange.passwordConfirm')"
          type="password"
          name="password-confirm"
          autocomplete="new-password"
          required
          test-id="force-change-password-confirm"
        />
        <p class="pro-field-hint">{{ $t('auth.forceChange.passwordHint') }}</p>
        <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
        <ProButton type="submit" block :loading="loading" test-id="force-change-submit">
          {{ $t('auth.forceChange.submit') }}
        </ProButton>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const { t } = useI18n()
const { mapError } = useApiError()
const { fetchUser } = useProUser()

const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)

async function redirectAfterChange() {
  const me = await fetchUser(true)
  const role = me?.role || parseJwtRole(useCookie('pf_token').value)
  await navigateTo(homePathForRole(role, { profileComplete: me?.profileComplete }))
}

async function submit() {
  error.value = ''
  if (password.value.length < 8) {
    error.value = t('auth.forceChange.passwordHint')
    return
  }
  if (password.value !== confirmPassword.value) {
    error.value = t('auth.forceChange.passwordMismatch')
    return
  }
  loading.value = true
  try {
    await $fetch('/api/me/password', {
      method: 'PATCH',
      body: { newPassword: password.value },
    })
    await redirectAfterChange()
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    loading.value = false
  }
}
</script>
