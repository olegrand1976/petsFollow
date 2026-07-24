<template>
  <div class="pro-login-page" data-testid="app-invite-landing">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ brandTitle }}</h2>
      <p>{{ brandText }}</p>
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
          <p class="pro-page-header__subtitle">{{ subtitle }}</p>
          <p class="pro-hint">{{ autoLinkHint }}</p>
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

type InviteRole = 'vet' | 'care_pro' | 'commercial' | 'commercial_manager'

const { t } = useI18n()
const { mapError } = useApiError()
const route = useRoute()

const loading = ref(true)
const error = ref('')
const invite = ref<{
  code: string
  role: InviteRole
  practiceName: string
  displayName: string
  downloadUrl: string
  deepLink: string
} | null>(null)

function normalizeRole(raw: unknown): InviteRole {
  switch (raw) {
    case 'care_pro':
      return 'care_pro'
    case 'commercial':
      return 'commercial'
    case 'commercial_manager':
      return 'commercial_manager'
    case 'vet':
    default:
      return 'vet'
  }
}

const brandTitle = computed(() => {
  const role = invite.value?.role ?? 'vet'
  switch (role) {
    case 'care_pro':
      return t('invite.brandTitleCarePro')
    case 'commercial':
    case 'commercial_manager':
      return t('invite.brandTitleCommercial')
    case 'vet':
      return t('invite.brandTitle')
    default: {
      const _exhaustive: never = role
      return _exhaustive
    }
  }
})

const brandText = computed(() => t('invite.brandText'))

const subtitle = computed(() => {
  const inv = invite.value
  if (!inv) return ''
  const name = inv.displayName || ''
  const practice = inv.practiceName || ''
  switch (inv.role) {
    case 'care_pro':
      return t('invite.subtitleCarePro', { name: name || 'petsFollow' })
    case 'commercial':
    case 'commercial_manager':
      return t('invite.subtitleCommercial', { name: name || 'petsFollow' })
    case 'vet':
      return t('invite.subtitle', {
        practice: practice || 'petsFollow',
        vet: name,
      })
    default: {
      const _exhaustive: never = inv.role
      return _exhaustive
    }
  }
})

const autoLinkHint = computed(() => {
  const role = invite.value?.role ?? 'vet'
  switch (role) {
    case 'care_pro':
      return t('invite.autoLinkHintCarePro')
    case 'commercial':
    case 'commercial_manager':
      return t('invite.autoLinkHintCommercial')
    case 'vet':
      return t('invite.autoLinkHint')
    default: {
      const _exhaustive: never = role
      return _exhaustive
    }
  }
})

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
      role: normalizeRole(data.role),
      practiceName: data.practiceName || '',
      displayName: data.displayName || data.vetFullName || '',
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
