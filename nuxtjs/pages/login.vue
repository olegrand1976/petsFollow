<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ $t('auth.login.brandTitle') }}</h2>
      <p>{{ $t('auth.login.brandText') }}</p>
    </aside>
    <div class="pro-login-form-panel">
      <div class="pro-auth-locale">
        <ProLocaleSelect />
      </div>
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
          revealable
          test-id="login-password"
        />
        <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
        <ProButton type="submit" block :loading="loading" test-id="login-submit">
          {{ $t('auth.login.submit') }}
        </ProButton>

        <div v-if="googleEnabled" class="pro-login-divider">
          <span>{{ $t('auth.login.or') }}</span>
        </div>
        <label v-if="googleEnabled" class="pro-consent">
          <input
            v-model="googleConsent"
            type="checkbox"
            name="googleConsent"
            data-testid="login-google-consent"
          >
          <i18n-t keypath="auth.register.consent" tag="span" scope="global">
            <template #terms>
              <NuxtLink to="/legal/terms" target="_blank">{{ $t('legal.terms.link') }}</NuxtLink>
            </template>
            <template #privacy>
              <NuxtLink to="/legal/privacy" target="_blank">{{ $t('legal.privacy.link') }}</NuxtLink>
            </template>
          </i18n-t>
        </label>
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
        method="post"
        action="#"
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
import {
  isAuthSuccess,
  isMFAChallenge,
  isProRole,
  unwrapAuthData,
  finishClientLoginSession,
  homePathForRole,
  clearAuthTokens,
  AUTH_LOGIN_REASON_PRO_ONLY,
  AUTH_POST_LOGIN_RELOAD_PATH,
} from '~/composables/useAuth'
import { mountGoogleSignInButton } from '~/composables/useGoogleAuth'
import { googleLoginBody } from '~/utils/googleLoginBody'

definePageMeta({ layout: false })

const { t } = useI18n()
const { mapError } = useApiError()
const route = useRoute()
const config = useRuntimeConfig()
const googleEnabled = computed(() => !!config.public.googleClientId)

const email = ref(import.meta.dev ? 'vet.demo@petsfollow.test' : '')
const password = ref(import.meta.dev ? 'VetDemo123!' : '')
const totpCode = ref('')
const mfaToken = ref('')
const step = ref<'credentials' | '2fa'>('credentials')
const error = ref(
  String(route.query.reason || '') === AUTH_LOGIN_REASON_PRO_ONLY
    ? t('auth.login.proOnly')
    : '',
)
const loading = ref(false)
const googleBtnRef = ref<HTMLElement | null>(null)
const googleConsent = ref(false)

async function handleAuthResult(res: unknown) {
  const data = unwrapAuthData(res)
  if (isMFAChallenge(data)) {
    mfaToken.value = data.mfaToken
    step.value = '2fa'
    totpCode.value = ''
    return
  }
  if (!isAuthSuccess(data)) {
    error.value = t('auth.login.invalidResponse')
    return
  }
  const role = !isMFAChallenge(data) && typeof data.role === 'string' ? data.role : null
  if (role && !isProRole(role)) {
    await clearAuthTokens()
    error.value = t('auth.login.proOnly')
    return
  }
  // Navigation document vers la home du rôle (évite bounce `/` + XHR /me WebKit).
  // Fallback `/` si role absent (ancien BFF) — auth.global SSR redirige alors.
  const path = role ? homePathForRole(role) : AUTH_POST_LOGIN_RELOAD_PATH
  finishClientLoginSession(path)
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
  // Compte existant : consent ignoré côté API. Create-if-absent → consent_required
  // si la case CGU n'est pas cochée (aligné Flutter login_screen).
  error.value = ''
  loading.value = true
  try {
    const res = await $fetch('/api/auth/google', {
      method: 'POST',
      body: googleLoginBody(idToken, googleConsent.value),
    })
    await handleAuthResult(res)
  } catch (e: any) {
    const code = e?.data?.error?.code
    if (code === 'not_configured') error.value = t('auth.google.notConfigured')
    else if (code === 'forbidden') error.value = t('auth.google.forbidden')
    else if (code === 'consent_required') error.value = t('auth.register.consentRequired')
    else error.value = t('auth.google.failed')
  } finally {
    loading.value = false
  }
}

async function mountGoogleButton() {
  if (!googleEnabled.value || !googleBtnRef.value) return
  try {
    googleBtnRef.value.innerHTML = ''
    await mountGoogleSignInButton(
      googleBtnRef.value,
      config.public.googleClientId,
      handleGoogleCredential,
    )
  } catch {
    /* Google indisponible */
  }
}

onMounted(() => { void mountGoogleButton() })

watch(step, async (s) => {
  if (s === 'credentials') {
    await nextTick()
    await mountGoogleButton()
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
  display: block;
  width: 100%;
  min-height: 44px;
}

.pro-consent {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  font-size: 0.85rem;
  color: var(--pf-vet-text-muted);
  margin-bottom: 0.75rem;
}

.pro-consent input {
  margin-top: 0.2rem;
}

.pro-consent a {
  color: var(--pf-vet-accent);
  font-weight: 600;
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
