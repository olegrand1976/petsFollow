<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ $t('auth.forgot.brandTitle') }}</h2>
      <p>{{ $t('auth.forgot.brandText') }}</p>
    </aside>
    <div class="pro-login-form-panel">
      <div class="pro-auth-locale">
        <ProLocaleSelect />
      </div>
      <form
        v-if="!sent"
        class="pro-login-form"
        data-testid="forgot-form"
        method="post"
        action="#"
        @submit.prevent="submit"
      >
        <PetsFollowLogo variant="default" />
        <h1>{{ $t('auth.forgot.title') }}</h1>
        <p class="pro-page-header__subtitle">{{ $t('auth.forgot.subtitle') }}</p>
        <ProInput
          v-model="email"
          :label="$t('auth.fields.email')"
          type="email"
          name="email"
          autocomplete="email"
          required
          test-id="forgot-email"
        />
        <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
        <ProButton type="submit" block :loading="loading" test-id="forgot-submit">
          {{ $t('auth.forgot.submit') }}
        </ProButton>
        <p class="pro-login-form__footer">
          <NuxtLink to="/login">{{ $t('auth.forgot.backToLogin') }}</NuxtLink>
        </p>
        <ProLegalFooter />
      </form>

      <div v-else class="pro-login-form" data-testid="forgot-sent">
        <PetsFollowLogo variant="default" />
        <h1>{{ $t('auth.forgot.sentTitle') }}</h1>
        <p class="pro-page-header__subtitle">
          {{ $t('auth.forgot.sentSubtitle', { email }) }}
        </p>
        <p class="text-muted">{{ $t('auth.forgot.sentInstructions') }}</p>

        <div v-if="devLink" class="pro-dev-link" data-testid="forgot-dev-link">
          <p class="pro-dev-link__label">{{ $t('auth.forgot.devLinkLabel') }}</p>
          <NuxtLink :to="devLink" class="pro-dev-link__url">{{ devLink }}</NuxtLink>
        </div>

        <ProButton block class="pro-mt-lg" test-id="forgot-go-login" @click="navigateTo('/login')">
          {{ $t('auth.forgot.backToLogin') }}
        </ProButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const { t } = useI18n()
const { mapError } = useApiError()

const email = ref('')
const error = ref('')
const loading = ref(false)
const sent = ref(false)
const devLink = ref('')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const res: any = await $fetch('/api/auth/forgot-password', {
      method: 'POST',
      body: { email: email.value },
    })
    const data = res.data ?? res
    sent.value = true
    if (import.meta.dev && data.resetPath) {
      devLink.value = String(data.resetPath)
    }
  } catch (e: any) {
    error.value = mapError(e) || t('auth.forgot.failed')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.pro-dev-link {
  margin-top: 1.5rem;
  padding: 1rem;
  background: var(--pf-vet-bg);
  border-radius: var(--pf-vet-radius);
  border: 1px dashed var(--pf-vet-border);
}

.pro-dev-link__label {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--pf-vet-text-muted);
  margin-bottom: 0.5rem;
}

.pro-dev-link__url {
  font-family: var(--pf-font-mono, monospace);
  font-size: 0.8rem;
  word-break: break-all;
  color: var(--pf-vet-accent);
}

.pro-mt-lg {
  margin-top: 1.25rem;
}
</style>
