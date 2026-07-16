<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ confirmed ? $t('auth.confirmEmail.brandConfirmed') : $t('auth.confirmEmail.brandPending') }}</h2>
      <p v-if="confirmed">{{ $t('auth.confirmEmail.brandConfirmedText') }}</p>
      <p v-else-if="error">{{ $t('auth.confirmEmail.brandInvalidLink') }}</p>
    </aside>
    <div class="pro-login-form-panel">
      <div class="pro-login-form">
        <PetsFollowLogo variant="default" />
        <div v-if="loading" class="text-muted">{{ $t('auth.confirmEmail.loading') }}</div>
        <template v-else-if="confirmed">
          <h1 data-testid="confirm-email-success">{{ $t('auth.confirmEmail.title') }}</h1>
          <p class="pro-page-header__subtitle">{{ welcomeMessage }}</p>
          <ProButton block test-id="confirm-email-continue" @click="continueAfterConfirm">
            {{ $t('auth.confirmEmail.discover') }}
          </ProButton>
        </template>
        <template v-else>
          <h1 data-testid="confirm-email-failed">{{ $t('auth.confirmEmail.failedTitle') }}</h1>
          <p class="pro-field-error" role="alert">{{ error }}</p>
          <ProButton block @click="navigateTo('/register')">{{ $t('auth.confirmEmail.retry') }}</ProButton>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const { t } = useI18n()
const { mapError } = useApiError()

const route = useRoute()
const loading = ref(true)
const confirmed = ref(false)
const confirmedEmail = ref('')
const sessionReady = ref(false)
const error = ref('')

const welcomeMessage = computed(() =>
  t('auth.confirmEmail.welcome', {
    emailPart: confirmedEmail.value ? `, ${confirmedEmail.value}` : '',
  }),
)

async function continueAfterConfirm() {
  if (sessionReady.value) {
    await navigateTo('/welcome')
    return
  }
  await navigateTo('/login')
}

onMounted(async () => {
  const token = String(route.query.token || '')
  if (!token) {
    error.value = t('auth.confirmEmail.invalidLink')
    loading.value = false
    return
  }
  try {
    const res: any = await $fetch('/api/auth/confirm-email', {
      method: 'POST',
      body: { token },
    })
    const data = res.data ?? res
    confirmed.value = true
    confirmedEmail.value = data.email || ''
    if (data.accessToken) {
      persistAuthTokens({
        accessToken: data.accessToken,
        refreshToken: data.refreshToken,
        expiresIn: data.expiresIn,
      })
      sessionReady.value = true
    }
  } catch (e: any) {
    error.value = mapError(e) || t('auth.confirmEmail.expiredLink')
  } finally {
    loading.value = false
  }
})
</script>
