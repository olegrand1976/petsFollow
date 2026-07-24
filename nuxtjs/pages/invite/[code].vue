<template>
  <div class="pro-login-page" data-testid="app-invite-landing">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ $t('invite.brandTitle') }}</h2>
      <p>{{ $t('invite.brandText') }}</p>
    </aside>
    <div class="pro-login-form-panel">
      <div class="pro-auth-locale">
        <ProLocaleSelect />
      </div>
      <div class="pro-login-form">
        <PetsFollowLogo variant="default" />
        <div v-if="loading" class="text-muted">{{ $t('invite.loading') }}</div>
        <template v-else-if="invite">
          <h1 data-testid="app-invite-ok">{{ $t('invite.title') }}</h1>
          <p class="pro-page-header__subtitle">
            {{ $t('invite.subtitle', { practice: invite.practiceName, vet: invite.vetFullName }) }}
          </p>
          <p class="pro-hint">{{ $t('invite.autoLinkHint') }}</p>
          <div class="invite-actions">
            <a
              v-if="invite.downloadUrl"
              class="pro-btn pro-btn--primary pro-btn--block"
              data-testid="app-invite-download"
              :href="invite.downloadUrl"
              target="_blank"
              rel="noopener noreferrer"
            >
              {{ $t('invite.download') }}
            </a>
            <ProButton
              block
              variant="secondary"
              test-id="app-invite-open-app"
              @click="openApp"
            >
              {{ $t('invite.openApp') }}
            </ProButton>
          </div>
        </template>
        <template v-else>
          <h1 data-testid="app-invite-failed">{{ $t('invite.failedTitle') }}</h1>
          <p class="pro-field-error" role="alert">{{ error || $t('invite.invalid') }}</p>
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
const error = ref('')
const invite = ref<{
  code: string
  practiceName: string
  vetFullName: string
  downloadUrl: string
  deepLink: string
} | null>(null)

function openApp() {
  const link = invite.value?.deepLink
  if (!link) return
  window.location.href = link
}

onMounted(async () => {
  const code = String(route.params.code || '')
  if (!code) {
    error.value = t('invite.invalid')
    loading.value = false
    return
  }
  try {
    const res: any = await $fetch(`/api/public/app-invite/${encodeURIComponent(code)}`)
    const data = res.data ?? res
    invite.value = {
      code: data.code,
      practiceName: data.practiceName || '',
      vetFullName: data.vetFullName || '',
      downloadUrl: data.downloadUrl || '',
      deepLink: data.deepLink || `petsfollow://invite?code=${data.code}`,
    }
    try {
      localStorage.setItem('pf_invite_code', data.code)
    } catch { /* ignore */ }
  } catch (e) {
    error.value = mapError(e)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.invite-actions {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-top: 1.25rem;
}
</style>
