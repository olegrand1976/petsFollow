<template>
  <div
    v-if="show"
    class="pro-eid-reader"
    data-testid="eid-reader"
    :aria-busy="busy"
  >
    <div class="pro-eid-reader__heading">
      <h4 class="pro-eid-reader__title">{{ $t('clients.eid.title') }}</h4>
      <ProBadge variant="warning" data-testid="eid-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
    </div>
    <p class="pro-hint">{{ $t('clients.eid.hint') }}</p>
    <p class="pro-hint">{{ $t('clients.eid.pinRequiredHint') }}</p>

    <div class="pro-eid-reader__actions">
      <ProButton
        type="button"
        test-id="eid-read-card"
        :disabled="busy"
        :loading="busy && busyKind === 'card'"
        @click="onReadCard"
      >
        {{ $t('clients.eid.readCard') }}
      </ProButton>
      <label class="pro-eid-reader__file" data-testid="eid-import-label">
        <ProButton
          type="button"
          variant="secondary"
          :disabled="busy"
          :loading="busy && busyKind === 'file'"
          tabindex="-1"
        >
          {{ $t('clients.eid.importViewer') }}
        </ProButton>
        <input
          data-testid="eid-import-input"
          type="file"
          accept=".eid,.xml,application/xml,text/xml"
          :disabled="busy"
          @change="onFile"
        >
      </label>
      <a
        class="pro-eid-reader__help"
        data-testid="eid-help-link"
        href="https://eid.belgium.be/"
        target="_blank"
        rel="noopener noreferrer"
      >
        {{ $t('clients.eid.helpBosa') }}
      </a>
    </div>

    <p
      v-if="busy && busyHint"
      class="pro-hint pro-eid-reader__busy"
      data-testid="eid-reader-busy"
      role="status"
    >
      {{ busyHint }}
    </p>
    <p v-if="msg" class="pro-hint" data-testid="eid-reader-msg">{{ msg }}</p>
    <p v-if="error" class="pro-error" data-testid="eid-reader-error" role="alert">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import type { EidFormTarget, EidIdentity } from '~/utils/eid-prefill'
import { applyEidIdentityToForm } from '~/utils/eid-prefill'

const props = defineProps<{
  form: EidFormTarget
  /** Cabinet BE — si false/undefined, le bloc est masqué (fail-closed). */
  practiceIsBe?: boolean
}>()

const emit = defineEmits<{
  filled: [keys: string[], identity: EidIdentity]
}>()

const { t } = useI18n()
const {
  enabled,
  checkWebEidStatus,
  uploadViewerFile,
  readCard,
} = useEidPrefill()

const busy = ref(false)
const busyKind = ref<'card' | 'file' | null>(null)
const busyHint = ref('')
const msg = ref('')
const error = ref('')
const webEidReady = ref(false)

const show = computed(() => enabled.value && props.practiceIsBe === true)

const { mapError } = useApiError()

let probeRetryTimer: ReturnType<typeof setTimeout> | null = null
let probeGen = 0

async function probeWebEid() {
  const gen = ++probeGen
  try {
    const st = await checkWebEidStatus()
    if (gen !== probeGen) return
    webEidReady.value = st.ok && st.hasExtension && st.hasNativeApp
  } catch {
    if (gen !== probeGen) return
    webEidReady.value = false
  }
}

function scheduleProbeRetry() {
  if (probeRetryTimer) {
    clearTimeout(probeRetryTimer)
    probeRetryTimer = null
  }
  probeRetryTimer = setTimeout(() => {
    probeRetryTimer = null
    if (show.value && !webEidReady.value) void probeWebEid()
  }, 2000)
}

watch(show, (visible) => {
  if (!import.meta.client || !visible) {
    probeGen++
    if (probeRetryTimer) {
      clearTimeout(probeRetryTimer)
      probeRetryTimer = null
    }
    return
  }
  void probeWebEid().then(() => {
    if (show.value && !webEidReady.value) scheduleProbeRetry()
  })
}, { immediate: true })

onBeforeUnmount(() => {
  probeGen++
  if (probeRetryTimer) clearTimeout(probeRetryTimer)
})

async function applyIdentity(identity: EidIdentity) {
  const keys = applyEidIdentityToForm(props.form, identity)
  msg.value = t('clients.eid.prefillSuccess')
  error.value = ''
  emit('filled', keys.map(String), identity)
}

function mapReadError(e: unknown): string {
  const any = e as any
  const raw = String(any?.message || any || '')
  if (raw.startsWith('eid_origin_mismatch')) return t('clients.eid.originMismatch')
  if (raw === 'web_eid_extension_unavailable') {
    webEidReady.value = false
    return t('clients.eid.extensionMissing')
  }
  if (raw === 'web_eid_loopback_unavailable') {
    webEidReady.value = false
    return t('clients.eid.loopbackUnavailable')
  }
  if (raw === 'web_eid_native_unavailable') {
    webEidReady.value = false
    return t('clients.eid.nativeAppMissing')
  }
  if (raw === 'web_eid_version_mismatch') {
    webEidReady.value = false
    return t('clients.eid.versionMismatch')
  }
  if (raw === 'web_eid_context_insecure') {
    return t('clients.eid.contextInsecure')
  }
  if (raw === 'web_eid_unavailable') {
    webEidReady.value = false
    return t('clients.eid.webEidNotReady')
  }
  if (raw === 'web_eid_timeout') return t('clients.eid.timeout')
  if (raw === 'web_eid_cancelled') return t('clients.eid.cancelled')
  if (raw === 'eid_nonce_missing' || raw === 'eid_nonce_expired') return t('clients.eid.nonceExpired')
  if (raw === 'eid_cert_untrusted') return t('clients.eid.certUntrusted')
  if (raw === 'eid_token_signature') return t('clients.eid.tokenSignature')
  if (raw === 'eid_token_invalid' || raw === 'eid_token_parse' || raw === 'eid_identity_empty') {
    return t('clients.eid.verifyFailed')
  }
  return mapError(e) || t('clients.eid.readError')
}

async function onFile(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  busy.value = true
  busyKind.value = 'file'
  busyHint.value = t('clients.eid.importing')
  msg.value = ''
  error.value = ''
  try {
    const identity = await uploadViewerFile(file)
    await applyIdentity(identity)
  } catch (e: any) {
    error.value = mapError(e) || t('clients.eid.importError')
  } finally {
    busy.value = false
    busyKind.value = null
    busyHint.value = ''
  }
}

async function onReadCard() {
  busy.value = true
  busyKind.value = 'card'
  busyHint.value = t('clients.eid.reading')
  msg.value = ''
  error.value = ''
  try {
    // Soft probe only — status() can flake; authenticate is the source of truth.
    const identity = await readCard()
    webEidReady.value = true
    await applyIdentity(identity)
  } catch (e: any) {
    error.value = mapReadError(e)
  } finally {
    busy.value = false
    busyKind.value = null
    busyHint.value = ''
  }
}
</script>

<style scoped>
.pro-eid-reader {
  margin-bottom: var(--pf-space-md, 1rem);
  padding: var(--pf-space-md, 1rem);
  border: 1px dashed var(--pf-vet-border, #d0d7de);
  border-radius: var(--pf-radius-md, 8px);
  background: var(--pf-vet-surface, #fff);
}
.pro-eid-reader__heading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-bottom: 0.25rem;
}
.pro-eid-reader__title {
  margin: 0;
  font-size: 1rem;
}
.pro-eid-reader__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
  margin-top: 0.75rem;
}
.pro-eid-reader__file {
  position: relative;
  display: inline-flex;
  cursor: pointer;
}
.pro-eid-reader__file input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}
.pro-eid-reader__help {
  font-size: 0.875rem;
  color: var(--pf-vet-accent, #0d9488);
  text-decoration: underline;
}
.pro-eid-reader__busy {
  margin-top: 0.5rem;
  font-weight: 500;
}
</style>
