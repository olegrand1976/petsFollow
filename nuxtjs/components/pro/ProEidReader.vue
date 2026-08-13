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
        <span class="pro-btn pro-btn--secondary">{{ $t('clients.eid.importViewer') }}</span>
        <input
          data-testid="eid-import-input"
          type="file"
          accept=".eid,.xml,application/xml,text/xml"
          :disabled="busy"
          @change="onFile"
        >
      </label>
      <ProButton
        v-if="webEidReady"
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
  /** Cabinet BE — si false, le bloc est masqué. */
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

const show = computed(() => enabled.value && props.practiceIsBe !== false)

const { mapError } = useApiError()

onMounted(async () => {
  if (!show.value) return
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
.pro-btn {
  display: inline-flex;
  align-items: center;
  padding: 0.5rem 0.9rem;
  border-radius: 6px;
  border: 1px solid var(--pf-vet-border, #d0d7de);
  background: var(--pf-vet-surface, #fff);
  font: inherit;
}
</style>
