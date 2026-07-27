<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="pro-desk-lock"
      :class="{ 'pro-desk-lock--modal': promptMode === 'switch' }"
      role="dialog"
      aria-modal="true"
      :aria-labelledby="titleId"
      data-testid="pro-desk-lock"
    >
      <div class="pro-desk-lock__panel">
        <PetsFollowLogo variant="compact" />
        <h2 :id="titleId" class="pro-desk-lock__title">
          {{ promptMode === 'lock' ? $t('desk.lockTitle') : $t('desk.switchTitle') }}
        </h2>
        <p class="pro-desk-lock__hint">
          {{ promptMode === 'lock' ? $t('desk.lockHint') : $t('desk.switchHint') }}
        </p>

        <div v-if="roster.length && step === 'credentials'" class="pro-desk-lock__roster" data-testid="pro-desk-lock-roster">
          <button
            v-for="m in roster"
            :key="m.email"
            type="button"
            class="pro-desk-lock__member"
            :class="{ 'pro-desk-lock__member--selected': selectedEmail === m.email }"
            :data-testid="`pro-desk-lock-user-${m.email}`"
            @click="selectedEmail = m.email"
          >
            <ProAvatar :name="m.fullName" size="sm" />
            <span class="pro-desk-lock__member-name">{{ m.fullName }}</span>
          </button>
        </div>

        <form v-if="step === 'credentials'" class="pro-desk-lock__form" data-testid="pro-desk-lock-form" @submit.prevent="submitPassword">
          <p class="pro-desk-lock__email" data-testid="pro-desk-lock-email">{{ selectedEmail }}</p>
          <ProInput
            v-model="password"
            type="password"
            :label="$t('desk.password')"
            autocomplete="current-password"
            required
            revealable
            test-id="pro-desk-lock-password"
          />
          <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
          <div class="pro-desk-lock__actions">
            <ProButton
              v-if="promptMode === 'switch'"
              type="button"
              variant="ghost"
              test-id="pro-desk-lock-cancel"
              @click="onCancel"
            >
              {{ $t('desk.cancel') }}
            </ProButton>
            <ProButton
              v-if="promptMode === 'lock'"
              type="button"
              variant="ghost"
              test-id="pro-desk-lock-login"
              @click="onAbandon"
            >
              {{ $t('desk.goLogin') }}
            </ProButton>
            <ProButton type="submit" :loading="loading" test-id="pro-desk-lock-submit">
              {{ promptMode === 'lock' ? $t('desk.unlock') : $t('desk.switch') }}
            </ProButton>
          </div>
        </form>

        <form v-else class="pro-desk-lock__form" data-testid="pro-desk-lock-2fa" @submit.prevent="submit2fa">
          <h3 class="pro-desk-lock__title">{{ $t('desk.twoFaTitle') }}</h3>
          <p class="pro-desk-lock__hint">{{ $t('desk.twoFaHint') }}</p>
          <ProInput
            v-model="totpCode"
            type="text"
            :label="$t('desk.twoFaTitle')"
            autocomplete="one-time-code"
            required
            test-id="pro-desk-lock-2fa-code"
          />
          <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
          <div class="pro-desk-lock__actions">
            <ProButton type="button" variant="ghost" test-id="pro-desk-lock-2fa-back" @click="backFrom2fa">
              {{ $t('desk.cancel') }}
            </ProButton>
            <ProButton type="submit" :loading="loading" test-id="pro-desk-lock-2fa-submit">
              {{ $t('desk.twoFaSubmit') }}
            </ProButton>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
const desk = useDeskSession()
const { t } = useI18n()

const titleId = 'pro-desk-lock-title'
const password = ref('')
const totpCode = ref('')
const mfaToken = ref('')
const step = ref<'credentials' | '2fa'>('credentials')
const loading = ref(false)
const error = ref('')
const selectedEmail = ref('')

const promptMode = computed(() => desk.promptMode.value)
const roster = computed(() => desk.roster.value)
const visible = computed(() => promptMode.value === 'lock' || promptMode.value === 'switch')

watch(
  () => [promptMode.value, desk.pendingEmail.value] as const,
  ([mode, email]) => {
    if (!mode) {
      password.value = ''
      totpCode.value = ''
      mfaToken.value = ''
      step.value = 'credentials'
      error.value = ''
      return
    }
    selectedEmail.value = email || roster.value[0]?.email || ''
    password.value = ''
    totpCode.value = ''
    step.value = 'credentials'
    error.value = ''
  },
  { immediate: true },
)

function onCancel() {
  desk.cancelPrompt()
}

async function onAbandon() {
  await desk.abandonToLogin()
}

function backFrom2fa() {
  step.value = 'credentials'
  totpCode.value = ''
  mfaToken.value = ''
  error.value = ''
}

function mapAuthFail(reason: 'error' | 'proOnly' | 'mfa') {
  if (reason === 'proOnly') return t('desk.proOnly')
  return t('desk.error')
}

async function submitPassword() {
  error.value = ''
  if (!selectedEmail.value || !password.value) return
  loading.value = true
  try {
    const result = await desk.authenticate(selectedEmail.value, password.value)
    if (result.ok) {
      desk.completeUnlock(selectedEmail.value, result.role)
      return
    }
    if (result.reason === 'mfa') {
      mfaToken.value = result.mfaToken
      step.value = '2fa'
      totpCode.value = ''
      return
    }
    error.value = mapAuthFail(result.reason)
  } finally {
    loading.value = false
  }
}

async function submit2fa() {
  error.value = ''
  loading.value = true
  try {
    const result = await desk.verify2fa(mfaToken.value, totpCode.value)
    if (result.ok) {
      desk.completeUnlock(selectedEmail.value, result.role)
      return
    }
    if (result.reason === 'mfa') {
      mfaToken.value = result.mfaToken
      error.value = t('desk.twoFaInvalid')
      return
    }
    error.value = result.reason === 'proOnly' ? t('desk.proOnly') : t('desk.twoFaInvalid')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.pro-desk-lock {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  background: rgba(15, 23, 42, 0.88);
  backdrop-filter: blur(6px);
}

.pro-desk-lock--modal {
  /* Switch : même opacité que la veille — pas de PHI lisible derrière. */
  background: rgba(15, 23, 42, 0.88);
}

.pro-desk-lock__panel {
  width: min(100%, 26rem);
  padding: 1.75rem 1.5rem;
  border-radius: var(--pf-vet-radius-lg);
  background: var(--pf-vet-surface);
  box-shadow: var(--pf-vet-shadow-md);
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.pro-desk-lock__title {
  margin: 0;
  font-size: 1.25rem;
  color: var(--pf-vet-primary);
}

.pro-desk-lock__hint {
  margin: 0;
  font-size: 0.9rem;
  color: var(--pf-vet-text-muted);
}

.pro-desk-lock__roster {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  max-height: 12rem;
  overflow-y: auto;
}

.pro-desk-lock__member {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.45rem 0.6rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius-lg);
  background: var(--pf-vet-bg);
  cursor: pointer;
  text-align: left;
  font: inherit;
  color: inherit;
}

.pro-desk-lock__member--selected {
  border-color: var(--pf-vet-accent);
  box-shadow: 0 0 0 1px var(--pf-vet-accent);
}

.pro-desk-lock__member-name {
  font-size: 0.9rem;
  font-weight: 600;
}

.pro-desk-lock__email {
  margin: 0;
  font-size: 0.85rem;
  color: var(--pf-vet-text-muted);
  word-break: break-all;
}

.pro-desk-lock__form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.pro-desk-lock__actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  flex-wrap: wrap;
}
</style>
