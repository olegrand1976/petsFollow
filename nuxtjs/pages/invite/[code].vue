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
          <div class="invite-code-row" data-testid="app-invite-code">
            <p class="invite-code">
              <span class="text-muted">{{ $t('invite.codeLabel') }}</span>
              <strong>{{ invite.code }}</strong>
            </p>
            <ProButton
              type="button"
              variant="secondary"
              test-id="app-invite-copy-code"
              @click="copyCode"
            >
              <ProIcon name="content_copy" />
              {{ codeCopied ? $t('invite.codeCopied') : $t('invite.copyCode') }}
            </ProButton>
          </div>
          <p v-if="copyError" class="pro-field-error" role="alert">{{ copyError }}</p>
          <p class="pro-hint invite-code-hint">{{ codeHint }}</p>
          <div class="invite-actions">
            <ProButton
              block
              test-id="app-invite-open-app"
              @click="openApp"
            >
              {{ $t('invite.openApp') }}
            </ProButton>
            <NuxtLink
              v-if="isCommercialInvite"
              class="pro-btn pro-btn--secondary pro-btn--block"
              data-testid="app-invite-cabinet-cta"
              :to="cabinetRegisterPath"
            >
              {{ $t('invite.cabinetCta') }}
            </NuxtLink>
            <a
              v-if="invite.downloadUrl"
              class="pro-btn pro-btn--secondary pro-btn--block"
              data-testid="app-invite-download"
              :href="invite.downloadUrl"
              target="_blank"
              rel="noopener noreferrer"
            >
              {{ $t('invite.download') }}
            </a>
            <div v-if="invite.qrAndroid || invite.qrIos" class="invite-store-qrs" data-testid="invite-store-qrs">
              <a
                v-if="invite.qrAndroid?.publicUrl"
                :href="invite.qrAndroid.storeUrl || invite.downloadUrl || '#'"
                target="_blank"
                rel="noopener noreferrer"
              >
                <img :src="invite.qrAndroid.publicUrl" alt="Android" width="120" height="120">
              </a>
              <a
                v-if="invite.qrIos?.publicUrl"
                :href="invite.qrIos.storeUrl || invite.downloadUrl || '#'"
                target="_blank"
                rel="noopener noreferrer"
              >
                <img :src="invite.qrIos.publicUrl" alt="iOS" width="120" height="120">
              </a>
            </div>
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

type InviteRole = 'vet' | 'care_pro' | 'commercial' | 'commercial_manager' | 'client'

const { t } = useI18n()
const { mapError } = useApiError()
const route = useRoute()

const loading = ref(true)
const error = ref('')
const codeCopied = ref(false)
const copyError = ref('')
const invite = ref<{
  code: string
  role: InviteRole
  practiceName: string
  displayName: string
  downloadUrl: string
  deepLink: string
  vetRegisterUrl?: string
  qrAndroid?: { publicUrl?: string, storeUrl?: string } | null
  qrIos?: { publicUrl?: string, storeUrl?: string } | null
} | null>(null)

function normalizeRole(raw: unknown): InviteRole {
  switch (raw) {
    case 'care_pro':
      return 'care_pro'
    case 'commercial':
      return 'commercial'
    case 'commercial_manager':
      return 'commercial_manager'
    case 'client':
      return 'client'
    case 'vet':
      return 'vet'
    default:
      return 'vet'
  }
}

const isCommercialInvite = computed(() => {
  const role = invite.value?.role
  return role === 'commercial' || role === 'commercial_manager'
})

const cabinetRegisterPath = computed(() => {
  const code = invite.value?.code || ''
  return `/register?invite=${encodeURIComponent(code)}`
})

const brandTitle = computed(() => {
  const role = invite.value?.role ?? 'vet'
  switch (role) {
    case 'care_pro':
      return t('invite.brandTitleCarePro')
    case 'commercial':
    case 'commercial_manager':
      return t('invite.brandTitleCommercial')
    case 'client':
      return t('invite.brandTitleClient')
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
    case 'client':
      return t('invite.subtitleClient', { name: name || 'petsFollow' })
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
    case 'client':
      return t('invite.autoLinkHintClient')
    case 'vet':
      return t('invite.autoLinkHint')
    default: {
      const _exhaustive: never = role
      return _exhaustive
    }
  }
})

const codeHint = computed(() => {
  const role = invite.value?.role ?? 'vet'
  switch (role) {
    case 'commercial':
    case 'commercial_manager':
      return t('invite.codeHintCommercial')
    case 'client':
      return t('invite.codeHintClient')
    case 'care_pro':
    case 'vet':
      return t('invite.codeHint')
    default: {
      const _exhaustive: never = role
      return _exhaustive
    }
  }
})

async function copyCode() {
  copyError.value = ''
  codeCopied.value = false
  const code = invite.value?.code
  if (!code) return
  try {
    await navigator.clipboard.writeText(code)
    codeCopied.value = true
  } catch {
    copyError.value = t('invite.copyError')
  }
}

function openApp() {
  const inv = invite.value
  if (!inv) return
  const preconsult = String(route.query.preconsult || '').trim()
  if (preconsult) {
    const code = encodeURIComponent(inv.code || '')
    const visit = encodeURIComponent(preconsult)
    const inviteQs = code ? `&inviteCode=${code}` : ''
    window.location.href = `petsfollow://preconsult?visitId=${visit}${inviteQs}`
    return
  }
  const link = inv.deepLink
  if (!link) return
  window.location.href = link
}

function isLikelyMobile(): boolean {
  if (!import.meta.client) return false
  return /Android|iPhone|iPad|iPod|Mobile/i.test(navigator.userAgent || '')
}

/** Attempt deep link on mobile so an already-installed app captures the invite code. */
function tryAutoOpenApp() {
  if (!invite.value?.deepLink) return
  // Skip auto-open when the user explicitly landed for download-only flows.
  if (String(route.query.download || '') === '1') return
  // Opt-in via ?open=1, or auto on mobile UA only (avoid desktop custom-scheme dialogs).
  const forceOpen = String(route.query.open || '') === '1'
  if (!forceOpen && !isLikelyMobile()) return
  openApp()
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
    const preconsult = String(route.query.preconsult || '').trim()
    const inviteCode = data.code || code
    invite.value = {
      code: inviteCode,
      role: normalizeRole(data.role),
      practiceName: data.practiceName || '',
      displayName: data.displayName || data.vetFullName || '',
      downloadUrl: data.downloadUrl || '',
      deepLink: preconsult
        ? `petsfollow://preconsult?visitId=${encodeURIComponent(preconsult)}&inviteCode=${encodeURIComponent(inviteCode)}`
        : (data.deepLink || `petsfollow://invite?code=${inviteCode}`),
      vetRegisterUrl: data.vetRegisterUrl || '',
      qrAndroid: data.qrAndroid || null,
      qrIos: data.qrIos || null,
    }
    // Defer slightly so the landing paints before the custom-scheme navigation.
    setTimeout(tryAutoOpenApp, 400)
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
.invite-code-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
  margin: 1rem 0 0.25rem;
}
.invite-code {
  margin: 0;
  font-size: 1.05rem;
  letter-spacing: 0.06em;
}
.invite-code strong {
  margin-left: 0.35rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
.invite-code-hint {
  margin-bottom: 0;
}
.invite-store-qrs {
  display: flex;
  justify-content: center;
  gap: 1rem;
  flex-wrap: wrap;
}
.invite-store-qrs img {
  border-radius: 0.5rem;
  background: #fff;
}
</style>
