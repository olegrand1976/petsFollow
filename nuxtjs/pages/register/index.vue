<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ $t('auth.register.brandTitle') }}</h2>
      <p>{{ $t('auth.register.brandText') }}</p>
    </aside>
    <div class="pro-login-form-panel">
      <div class="pro-auth-locale">
        <ProLocaleSelect />
      </div>
      <form
        class="pro-login-form"
        data-testid="register-form"
        method="post"
        action="#"
        @submit.prevent="submit"
      >
        <PetsFollowLogo variant="default" />
        <h1>{{ $t('auth.register.title') }}</h1>
        <p class="pro-page-header__subtitle">{{ $t('auth.register.subtitle') }}</p>

        <ProInput
          v-model="fullName"
          :label="$t('auth.register.fullName')"
          name="fullName"
          autocomplete="name"
          required
          test-id="register-fullname"
        />
        <ProInput
          v-model="practiceName"
          :label="$t('auth.register.practiceName')"
          name="practiceName"
          required
          test-id="register-practice"
        />
        <ProInput
          v-model="email"
          :label="$t('auth.register.email')"
          type="email"
          name="email"
          autocomplete="email"
          required
          test-id="register-email"
        />
        <ProInput
          v-model="password"
          :label="$t('auth.register.password')"
          type="password"
          name="password"
          autocomplete="new-password"
          required
          test-id="register-password"
        />
        <ProInput
          v-model="confirmPassword"
          :label="$t('auth.register.passwordConfirm')"
          type="password"
          name="passwordConfirm"
          autocomplete="new-password"
          required
          test-id="register-password-confirm"
        />
        <p class="pro-field-hint">{{ $t('auth.register.passwordHint') }}</p>

        <label class="pro-consent">
          <input
            v-model="consent"
            type="checkbox"
            name="consent"
            required
            data-testid="register-consent"
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

        <section class="nearby-commercial" data-testid="register-nearby-commercial">
          <h2 class="nearby-commercial__title">{{ $t('auth.register.nearbyCommercial.title') }}</h2>
          <p class="nearby-commercial__hint">{{ $t('auth.register.nearbyCommercial.hint') }}</p>
          <div class="nearby-commercial__actions">
            <ProButton
              type="button"
              variant="secondary"
              :loading="nearbyLoading"
              test-id="register-nearby-geo"
              @click="findByGeo"
            >
              {{ $t('auth.register.nearbyCommercial.useLocation') }}
            </ProButton>
          </div>
          <div class="nearby-commercial__postal">
            <ProInput
              v-model="postalCode"
              :label="$t('auth.register.nearbyCommercial.postalCode')"
              name="postalCode"
              test-id="register-nearby-postal"
            />
            <ProButton
              type="button"
              variant="secondary"
              :loading="nearbyLoading"
              :disabled="!postalCode.trim()"
              test-id="register-nearby-postal-btn"
              @click="findByPostal"
            >
              {{ $t('auth.register.nearbyCommercial.searchPostal') }}
            </ProButton>
          </div>
          <p v-if="nearbyError" class="pro-field-error" role="alert">{{ nearbyError }}</p>
          <p v-else-if="nearbyEmpty" class="pro-field-hint">{{ $t('auth.register.nearbyCommercial.empty') }}</p>
          <fieldset v-if="nearbyRows.length" class="nearby-commercial__list">
            <legend class="sr-only">{{ $t('auth.register.nearbyCommercial.title') }}</legend>
            <label
              v-for="row in nearbyRows"
              :key="row.userId"
              class="nearby-commercial__option"
              :data-testid="`register-nearby-option-${row.userId}`"
            >
              <input
                v-model="selectedCommercialId"
                type="radio"
                name="assignedCommercialId"
                :value="row.userId"
              >
              <span>
                <strong>{{ row.fullName }}</strong>
                <span v-if="row.city"> — {{ row.city }}</span>
                <span v-if="row.distanceKm" class="nearby-commercial__dist">
                  ({{ $t('auth.register.nearbyCommercial.distance', { km: row.distanceKm }) }})
                </span>
              </span>
            </label>
            <button
              type="button"
              class="nearby-commercial__skip"
              data-testid="register-nearby-skip"
              @click="selectedCommercialId = ''"
            >
              {{ $t('auth.register.nearbyCommercial.skip') }}
            </button>
          </fieldset>
        </section>

        <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>

        <ProButton type="submit" block :loading="loading" test-id="register-submit">
          {{ $t('auth.register.submit') }}
        </ProButton>

        <p class="pro-login-form__footer">
          {{ $t('auth.register.alreadyRegistered') }}
          <NuxtLink to="/login">{{ $t('auth.register.loginLink') }}</NuxtLink>
        </p>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const { t } = useI18n()
const { mapError } = useApiError()

const fullName = ref('')
const practiceName = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const consent = ref(false)
const error = ref('')
const loading = ref(false)

type NearbyRow = { userId: string, fullName: string, city?: string, distanceKm?: number, postalCode?: string }
const postalCode = ref('')
const nearbyRows = ref<NearbyRow[]>([])
const selectedCommercialId = ref('')
const nearbyLoading = ref(false)
const nearbyError = ref('')
const nearbyEmpty = ref(false)

async function loadNearby(query: Record<string, string | number>) {
  nearbyLoading.value = true
  nearbyError.value = ''
  nearbyEmpty.value = false
  try {
    const res: any = await $fetch('/api/commercials/nearby', { query })
    const rows = (res.data ?? res) as NearbyRow[]
    nearbyRows.value = Array.isArray(rows) ? rows : []
    nearbyEmpty.value = nearbyRows.value.length === 0
    if (!nearbyRows.value.some(r => r.userId === selectedCommercialId.value)) {
      selectedCommercialId.value = ''
    }
  } catch (e: any) {
    nearbyError.value = mapError(e)
    nearbyRows.value = []
  } finally {
    nearbyLoading.value = false
  }
}

async function findByGeo() {
  nearbyError.value = ''
  if (!navigator.geolocation) {
    nearbyError.value = t('auth.register.nearbyCommercial.geoUnsupported')
    return
  }
  nearbyLoading.value = true
  navigator.geolocation.getCurrentPosition(
    async (pos) => {
      await loadNearby({ lat: pos.coords.latitude, lng: pos.coords.longitude, limit: 5 })
    },
    () => {
      nearbyLoading.value = false
      nearbyError.value = t('auth.register.nearbyCommercial.geoDenied')
    },
    { enableHighAccuracy: false, timeout: 10000 },
  )
}

async function findByPostal() {
  const code = postalCode.value.trim()
  if (!code) return
  await loadNearby({ postalCode: code, limit: 5 })
}

async function submit() {
  error.value = ''
  if (!consent.value) {
    error.value = t('auth.register.consentRequired')
    return
  }
  if (password.value !== confirmPassword.value) {
    error.value = t('auth.register.passwordMismatch')
    return
  }
  if (password.value.length < 8) {
    error.value = t('errors.password_too_short')
    return
  }
  loading.value = true
  try {
    const res: any = await $fetch('/api/auth/register', {
      method: 'POST',
      body: {
        fullName: fullName.value,
        practiceName: practiceName.value,
        email: email.value,
        password: password.value,
        consent: consent.value,
        ...(selectedCommercialId.value
          ? { assignedCommercialId: selectedCommercialId.value }
          : {}),
      },
    })
    const data = res.data ?? res
    await navigateTo({
      path: '/register/sent',
      query: { email: email.value, devLink: import.meta.dev ? data.confirmPath : undefined },
    })
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
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

.nearby-commercial {
  margin: 1rem 0 1.25rem;
  padding: 1rem 0 0;
  border-top: 1px solid var(--pf-vet-border);
}

.nearby-commercial__title {
  margin: 0 0 0.35rem;
  font-size: 1rem;
  font-weight: 600;
  color: var(--pf-vet-text);
}

.nearby-commercial__hint {
  margin: 0 0 0.75rem;
  font-size: 0.85rem;
  color: var(--pf-vet-text-muted);
}

.nearby-commercial__actions {
  margin-bottom: 0.75rem;
}

.nearby-commercial__postal {
  display: grid;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.nearby-commercial__list {
  border: 0;
  margin: 0.5rem 0 0;
  padding: 0;
  display: grid;
  gap: 0.5rem;
}

.nearby-commercial__option {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  font-size: 0.9rem;
  cursor: pointer;
}

.nearby-commercial__dist {
  color: var(--pf-vet-text-muted);
  font-weight: 400;
}

.nearby-commercial__skip {
  background: none;
  border: 0;
  padding: 0.25rem 0;
  color: var(--pf-vet-accent);
  font-size: 0.85rem;
  cursor: pointer;
  text-align: left;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.pro-field-hint {
  font-size: 0.8rem;
  color: var(--pf-vet-text-muted);
  margin-top: -0.5rem;
  margin-bottom: 0.75rem;
}

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
</style>
