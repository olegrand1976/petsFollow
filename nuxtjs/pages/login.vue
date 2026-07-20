<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ $t('auth.login.brandTitle') }}</h2>
      <p>{{ $t('auth.login.brandText') }}</p>
    </aside>
    <div class="pro-login-form-panel">
      <form
        v-if="step === 'credentials'"
        class="pro-login-form"
        data-testid="login-form"
        method="post"
        action="#"
        @submit.prevent="submit"
      >
        <PetsFollowLogo variant="default" />
        <h1>{{ $t('auth.login.title') }}</h1>
        <p class="pro-page-header__subtitle">{{ $t('auth.login.subtitle') }}</p>
        <ProInput
          v-model="email"
          :label="$t('auth.fields.email')"
          type="email"
          name="email"
          autocomplete="email"
          required
          test-id="login-email"
        />
        <ProInput
          v-model="password"
          :label="$t('auth.fields.password')"
          type="password"
          name="password"
          autocomplete="current-password"
          required
          test-id="login-password"
        />
        <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
        <ProButton type="submit" block :loading="loading" test-id="login-submit">
          {{ $t('auth.login.submit') }}
        </ProButton>

        <div v-if="googleEnabled" class="pro-login-divider">
          <span>{{ $t('auth.login.or') }}</span>
        </div>
        <div
          v-if="googleEnabled"
          ref="googleBtnRef"
          class="pro-login-google"
          data-testid="login-google"
        />

        <p class="pro-login-form__footer">
          <NuxtLink to="/forgot-password" data-testid="login-forgot-link">
            {{ $t('auth.login.forgotLink') }}
          </NuxtLink>
        </p>
        <p class="pro-login-form__footer">
          {{ $t('auth.login.noAccount') }}
          <NuxtLink to="/register">{{ $t('auth.login.registerLink') }}</NuxtLink>
        </p>
        <ProLegalFooter />
      </form>

      <form
        v-else
        class="pro-login-form"
        data-testid="login-2fa-form"
        @submit.prevent="submit2FA"
      >
        <PetsFollowLogo variant="default" />
        <h1>{{ $t('auth.twoFa.title') }}</h1>
        <p class="pro-page-header__subtitle">{{ $t('auth.twoFa.subtitle') }}</p>
        <ProInput
          v-model="totpCode"
          :label="$t('auth.twoFa.codeLabel')"
          type="text"
          name="totp"
          inputmode="numeric"
          autocomplete="one-time-code"
          maxlength="6"
          required
          test-id="login-2fa-code"
        />
        <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
        <ProButton type="submit" block :loading="loading" test-id="login-2fa-submit">
          {{ $t('auth.twoFa.submit') }}
        </ProButton>
        <button type="button" class="pro-login-back" @click="reset2FA">
          {{ $t('auth.twoFa.back') }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { extractAccessToken, isMFAChallenge, unwrapAuthData, persistAuthTokens, clearAuthTokens } from '~/composables/useAuth'
import { mountGoogleSignInButton } from '~/composables/useGoogleAuth'

definePageMeta({ layout: false })

const { t } = useI18n()
const { mapError } = useApiError()
const { syncFromUser } = useLocaleSync()
const config = useRuntimeConfig()
const googleEnabled = computed(() => !!config.public.googleClientId)

const email = ref(import.meta.dev ? 'vet.demo@petsfollow.test' : '')
const password = ref(import.meta.dev ? 'VetDemo123!' : '')
const totpCode = ref('')
const mfaToken = ref('')
const step = ref<'credentials' | '2fa'>('credentials')
const error = ref('')
const loading = ref(false)
const googleBtnRef = ref<HTMLElement | null>(null)

async function redirectAfterLogin() {
  await syncFromUser()
  const me: any = await $fetch('/api/me')
  const role = me.data?.role || me.role || parseJwtRole(useCookie('pf_token').value)
  const profileComplete = me.data?.profileComplete ?? me.profileComplete
  const mustChangePassword = me.data?.mustChangePassword ?? me.mustChangePassword
  if (mustChangePassword === true) {
    await navigateTo('/change-password')
    return
  }
  if (!isProRole(role)) {
    clearAuthTokens()
    error.value = t('auth.login.proOnly')
    return
  }
  await navigateTo(homePathForRole(role, { profileComplete }))
}

async function handleAuthResult(res: unknown) {
  const data = unwrapAuthData(res)
  if (isMFAChallenge(data)) {
    mfaToken.value = data.mfaToken
    step.value = '2fa'
    totpCode.value = ''
    return
  }
  if (!extractAccessToken(data)) {
    error.value = t('auth.login.invalidResponse')
    return
  }
  persistAuthTokens(data)
  await redirectAfterLogin()
}

function mapAuthError(e: any) {
  return mapError(e)
}

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const res = await $fetch('/api/auth/login', {
      method: 'POST',
      body: { email: email.value, password: password.value },
    })
    await handleAuthResult(res)
  } catch (e: any) {
    error.value = mapAuthError(e)
  } finally {
    loading.value = false
  }
}

async function submit2FA() {
  error.value = ''
  loading.value = true
  try {
    const res = await $fetch('/api/auth/2fa/verify', {
      method: 'POST',
      body: { mfaToken: mfaToken.value, code: totpCode.value },
    })
    await handleAuthResult(res)
  } catch {
    error.value = t('auth.twoFa.invalidCode')
  } finally {
    loading.value = false
  }
}

function reset2FA() {
  step.value = 'credentials'
  mfaToken.value = ''
  totpCode.value = ''
  error.value = ''
}

async function handleGoogleCredential(idToken: string) {
  error.value = ''
  loading.value = true
  try {
    const res = await $fetch('/api/auth/google', { method: 'POST', body: { idToken } })
    await handleAuthResult(res)
  } catch (e: any) {
    const code = e?.data?.error?.code
    if (code === 'not_configured') error.value = t('auth.google.notConfigured')
    else if (code === 'forbidden') error.value = t('auth.google.forbidden')
    else error.value = t('auth.google.failed')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  if (!googleEnabled.value || !googleBtnRef.value) return
  try {
    await mountGoogleSignInButton(
      googleBtnRef.value,
      config.public.googleClientId,
      handleGoogleCredential,
    )
  } catch {
    /* Google indisponible */
  }
})
</script>

<style scoped>
.pro-login-form__footer {
  margin-top: 1.25rem;
  text-align: center;
  font-size: 0.9rem;
  color: var(--pf-vet-text-muted);
}

.pro-login-form__footer a {
  color: var(--pf-vet-accent);
  font-weight: 600;
}

.pro-login-divider {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin: 1.25rem 0 1rem;
  color: var(--pf-vet-text-muted);
  font-size: 0.85rem;
}

.pro-login-divider::before,
.pro-login-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--pf-vet-border);
}

.pro-login-google {
  display: flex;
  justify-content: center;
  min-height: 44px;
}

.pro-login-back {
  margin-top: 1rem;
  width: 100%;
  background: none;
  border: none;
  color: var(--pf-vet-accent);
  font-weight: 600;
  cursor: pointer;
  padding: 0.5rem;
}
</style>
