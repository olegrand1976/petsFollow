<template>
  <div
    v-if="show"
    class="pro-eid-reader"
    data-testid="eid-reader"
  >
    <div class="pro-eid-reader__heading">
      <h4 class="pro-eid-reader__title">{{ $t('clients.eid.title') }}</h4>
      <ProBadge variant="warning" data-testid="eid-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
    </div>
    <p class="pro-hint">{{ $t('clients.eid.hint') }}</p>

    <div v-if="localEidHint" class="pro-eid-reader__info" data-testid="eid-local-hint">
      {{ $t('clients.eid.localSoftwareHint') }}
    </div>

    <div class="pro-eid-reader__actions">
      <label class="pro-eid-reader__file" data-testid="eid-import-label">
        <ProButton
          type="button"
          variant="secondary"
          :disabled="busy"
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
      <ProButton
        type="button"
        variant="secondary"
        test-id="eid-read-card"
        :disabled="busy"
        @click="onReadCard"
      >
        {{ $t('clients.eid.readCard') }}
      </ProButton>
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

    <p v-if="msg" class="pro-hint" data-testid="eid-reader-msg">{{ msg }}</p>
    <p v-if="error" class="pro-error" data-testid="eid-reader-error">{{ error }}</p>
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
  detectInstalledEidExtensions,
  checkWebEidStatus,
  uploadViewerFile,
  readCard,
} = useEidPrefill()

const busy = ref(false)
const msg = ref('')
const error = ref('')
const webEidReady = ref(false)
const localEidHint = ref(false)

const show = computed(() => enabled.value && props.practiceIsBe === true)

const { mapError } = useApiError()

let probeRetryTimer: ReturnType<typeof setTimeout> | null = null

async function probeLocalTools() {
  try {
    const ext = await detectInstalledEidExtensions()
    localEidHint.value = Boolean(ext.beidconnect || ext.eid_chrome)
  } catch { /* ignore */ }
  try {
    const st = await checkWebEidStatus()
    webEidReady.value = st.ok && st.hasExtension && st.hasNativeApp
  } catch {
    webEidReady.value = false
  }
}

function scheduleProbeRetry() {
  if (probeRetryTimer) {
    clearTimeout(probeRetryTimer)
    probeRetryTimer = null
  }
  // Native app / extension may appear a moment after page open.
  probeRetryTimer = setTimeout(() => {
    probeRetryTimer = null
    if (show.value && !webEidReady.value) void probeLocalTools()
  }, 2000)
}

// Country arrives async (overview) — (re)probe whenever the block becomes visible.
watch(show, (visible) => {
  if (!visible) {
    if (probeRetryTimer) {
      clearTimeout(probeRetryTimer)
      probeRetryTimer = null
    }
    return
  }
  void probeLocalTools().then(() => {
    if (show.value && !webEidReady.value) scheduleProbeRetry()
  })
}, { immediate: true })

onBeforeUnmount(() => {
  if (probeRetryTimer) clearTimeout(probeRetryTimer)
})

async function applyIdentity(identity: EidIdentity) {
  const keys = applyEidIdentityToForm(props.form, identity)
  msg.value = t('clients.eid.prefillSuccess')
  error.value = ''
  emit('filled', keys.map(String), identity)
}

async function onFile(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  busy.value = true
  msg.value = ''
  error.value = ''
  try {
    const identity = await uploadViewerFile(file)
    await applyIdentity(identity)
  } catch (e: any) {
    error.value = mapError(e) || t('clients.eid.importError')
  } finally {
    busy.value = false
  }
}

async function onReadCard() {
  busy.value = true
  msg.value = ''
  error.value = ''
  try {
    if (!webEidReady.value) {
      await probeLocalTools()
    }
    if (!webEidReady.value) {
      error.value = t('clients.eid.webEidNotReady')
      return
    }
    const identity = await readCard()
    await applyIdentity(identity)
  } catch (e: any) {
    error.value = mapError(e) || t('clients.eid.readError')
  } finally {
    busy.value = false
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
.pro-eid-reader__info {
  margin-top: 0.5rem;
  font-size: 0.875rem;
  color: var(--pf-vet-primary, #1e3a5f);
}
</style>
