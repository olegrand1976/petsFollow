<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>Suivi cardiaque et dossier client en un seul outil</h2>
      <p>
        Bienvenue sur petsFollow Pro — l'espace pensé pour les pros du soin animal,
        avec une touche de bonne humeur.
      </p>
    </aside>
    <div class="pro-login-form-panel">
      <form
        v-if="step === 'credentials'"
        class="pro-login-form"
        data-testid="login-form"
        @submit.prevent="submit"
      >
        <PetsFollowLogo variant="default" />
        <h1>Connexion Pro</h1>
        <p class="pro-page-header__subtitle">Accès réservé aux vétérinaires et administrateurs.</p>
        <ProInput
          v-model="email"
          label="Email"
          type="email"
          name="email"
          autocomplete="email"
          required
          test-id="login-email"
        />
        <ProInput
          v-model="password"
          label="Mot de passe"
          type="password"
          name="password"
          autocomplete="current-password"
          required
          test-id="login-password"
        />
        <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
        <ProButton type="submit" block :loading="loading" test-id="login-submit">
          Se connecter
        </ProButton>

        <div v-if="googleEnabled" class="pro-login-divider">
          <span>ou</span>
        </div>
        <div
          v-if="googleEnabled"
          ref="googleBtnRef"
          class="pro-login-google"
          data-testid="login-google"
        />

        <p class="pro-login-form__footer">
          Pas encore de compte ?
          <NuxtLink to="/register">S'inscrire gratuitement</NuxtLink>
        </p>
      </form>

      <form
        v-else
        class="pro-login-form"
        data-testid="login-2fa-form"
        @submit.prevent="submit2FA"
      >
        <PetsFollowLogo variant="default" />
        <h1>Vérification 2FA</h1>
        <p class="pro-page-header__subtitle">
          Saisissez le code à 6 chiffres de votre application d'authentification.
        </p>
        <ProInput
          v-model="totpCode"
          label="Code authenticator"
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
          Valider
        </ProButton>
        <button type="button" class="pro-login-back" @click="reset2FA">
          ← Retour à la connexion
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { extractAccessToken, isMFAChallenge, unwrapAuthData } from '~/composables/useAuth'
import { mountGoogleSignInButton } from '~/composables/useGoogleAuth'

definePageMeta({ layout: false })

const config = useRuntimeConfig()
const googleEnabled = computed(() => !!config.public.googleClientId)

const email = ref(import.meta.dev ? 'vet.demo@petsfollow.test' : '')
const password = ref(import.meta.dev ? 'VetDemo123!' : '')
const totpCode = ref('')
const mfaToken = ref('')
const step = ref<'credentials' | '2fa'>('credentials')
const error = ref('')
const loading = ref(false)
const token = useCookie('pf_token')
const googleBtnRef = ref<HTMLElement | null>(null)

async function redirectAfterLogin() {
  const me: any = await $fetch('/api/me')
  const role = me.data?.role || me.role
  const profileComplete = me.data?.profileComplete ?? me.profileComplete
  if (role === 'admin') await navigateTo('/admin')
  else if (role === 'vet') {
    if (profileComplete === false) await navigateTo('/onboarding')
    else await navigateTo('/dashboard')
  } else {
    token.value = null
    error.value = 'Accès réservé aux profils Pro (véto / admin)'
  }
}

async function handleAuthResult(res: unknown) {
  const data = unwrapAuthData(res)
  if (isMFAChallenge(data)) {
    mfaToken.value = data.mfaToken
    step.value = '2fa'
    totpCode.value = ''
    return
  }
  const accessToken = extractAccessToken(data)
  if (!accessToken) {
    error.value = 'Réponse d\'authentification invalide'
    return
  }
  token.value = accessToken
  await redirectAfterLogin()
}

function mapAuthError(e: any) {
  const code = e?.data?.error?.code
  if (code === 'email_not_verified') return 'Confirmez votre email avant de vous connecter.'
  if (code === 'use_google_sign_in') return 'Ce compte utilise Google — connectez-vous avec Google.'
  return 'Identifiants invalides'
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
    error.value = 'Code 2FA invalide ou expiré'
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
    if (code === 'not_configured') error.value = 'Connexion Google non configurée.'
    else if (code === 'forbidden') error.value = 'Accès Google réservé aux profils Pro.'
    else error.value = 'Connexion Google impossible'
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
    /* Google indisponible — bouton masqué implicitement */
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
