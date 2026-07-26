<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ $t('auth.registerSent.brandTitle') }}</h2>
      <p>{{ $t('auth.registerSent.brandText') }}</p>
    </aside>
    <div class="pro-login-form-panel">
      <div class="pro-auth-locale">
        <ProLocaleSelect />
      </div>
      <div class="pro-login-form">
        <PetsFollowLogo variant="default" />
        <h1>{{ $t('auth.registerSent.title') }}</h1>
        <p class="pro-page-header__subtitle">
          {{ $t('auth.registerSent.subtitle', { email }) }}
        </p>
        <p class="text-muted">{{ $t('auth.registerSent.instructions') }}</p>
        <p v-if="resendInfo" class="text-muted" data-testid="register-resend-info">{{ resendInfo }}</p>
        <p v-if="resendError" class="alert" data-testid="register-resend-error">{{ resendError }}</p>

        <div v-if="devLink" class="pro-dev-link">
          <p class="pro-dev-link__label">{{ $t('auth.registerSent.devLinkLabel') }}</p>
          <NuxtLink :to="devLink" class="pro-dev-link__url">{{ devLink }}</NuxtLink>
        </div>

        <ProButton
          v-if="email"
          variant="secondary"
          block
          class="pro-mt-lg"
          :disabled="resendBusy"
          test-id="register-resend"
          @click="resend"
        >
          {{ $t('auth.registerSent.resend') }}
        </ProButton>
        <ProButton block :class="email ? '' : 'pro-mt-lg'" @click="navigateTo('/login')">
          {{ $t('auth.registerSent.goToLogin') }}
        </ProButton>
        <ProButton variant="ghost" block @click="navigateTo('/')">
          {{ $t('auth.registerSent.backHome') }}
        </ProButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const { t } = useI18n()
const route = useRoute()
const email = computed(() => String(route.query.email || ''))
const devLink = computed(() => route.query.devLink ? String(route.query.devLink) : '')
const resendBusy = ref(false)
const resendInfo = ref('')
const resendError = ref('')

async function resend() {
  const addr = email.value.trim().toLowerCase()
  if (!addr || resendBusy.value) return
  resendBusy.value = true
  resendInfo.value = ''
  resendError.value = ''
  try {
    await $fetch('/api/auth/resend-confirmation', {
      method: 'POST',
      body: { email: addr },
    })
    resendInfo.value = t('auth.registerSent.resendOk')
  } catch {
    resendError.value = t('auth.registerSent.resendFailed')
  } finally {
    resendBusy.value = false
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
