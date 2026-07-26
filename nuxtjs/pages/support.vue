<template>
  <div class="pro-page" data-testid="support-page">
    <ProPageHeader :title="$t('support.dialogTitle')" :subtitle="$t('support.dialogHint')" />
    <ProCard>
      <form class="pro-form" data-testid="support-form" @submit.prevent="submit">
        <ProInput
          v-model="subject"
          :label="$t('support.subject')"
          required
          maxlength="200"
          test-id="support-subject"
        />
        <div class="pro-field">
          <label class="pro-label" for="support-message">{{ $t('support.message') }}</label>
          <textarea
            id="support-message"
            v-model="message"
            class="pro-textarea"
            rows="5"
            required
            maxlength="8000"
            data-testid="support-message"
          />
        </div>
        <details class="support-diag" data-testid="support-diagnostics-preview">
          <summary>{{ $t('support.diagnosticsPreview', { count: diagCount }) }}</summary>
          <pre class="support-diag__pre">{{ diagnosticsJson }}</pre>
        </details>
        <p v-if="error" class="pro-error" role="alert" data-testid="support-error">{{ error }}</p>
        <p v-if="success" class="pro-hint" data-testid="support-success">{{ $t('support.success') }}</p>
        <div class="pro-form-actions">
          <ProButton type="button" variant="ghost" test-id="support-cancel" @click="navigateTo('/')">
            {{ $t('common.cancel') }}
          </ProButton>
          <!-- Native submit: Cloud Run e2e cannot rely on ProButton Vue emit alone. -->
          <button
            type="submit"
            class="pro-btn pro-btn--primary"
            data-testid="support-submit"
            :disabled="saving"
          >
            {{ saving ? $t('common.loading') : $t('support.submit') }}
          </button>
        </div>
      </form>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
const { t, locale } = useI18n()
const route = useRoute()
const { snapshot } = useSupportDiagnostics()

const subject = ref('')
const message = ref('')
const saving = ref(false)
const error = ref('')
const success = ref(false)
const diagnosticsJson = ref('{}')
const diagCount = ref(0)

const canSubmit = computed(() => subject.value.trim().length > 0 && message.value.trim().length > 0)

onMounted(() => {
  const snap = snapshot()
  diagnosticsJson.value = JSON.stringify(snap, null, 2)
  diagCount.value = (snap.consoleErrors?.length || 0) + (snap.networkEntries?.length || 0)
})

async function submit() {
  if (saving.value) return
  // Read DOM as fallback when v-model missed Playwright fills (SSR hydration races).
  const subjectEl = document.querySelector('[data-testid="support-subject"]') as HTMLInputElement | null
  const messageEl = document.querySelector('[data-testid="support-message"]') as HTMLTextAreaElement | null
  if (subjectEl?.value) subject.value = subjectEl.value
  if (messageEl?.value) message.value = messageEl.value
  if (!canSubmit.value) {
    error.value = t('support.errorGeneric')
    return
  }
  saving.value = true
  error.value = ''
  success.value = false
  try {
    const snap = snapshot()
    await $fetch('/api/support/tickets', {
      method: 'POST',
      body: {
        source: 'nuxt_pro',
        subject: subject.value.trim(),
        message: message.value.trim(),
        diagnostics: snap,
        userAgent: typeof navigator !== 'undefined' ? navigator.userAgent : '',
        appVersion: 'web',
        locale: locale.value,
        route: route.fullPath,
      },
    })
    success.value = true
  } catch (e: any) {
    const code = e?.data?.error?.code || e?.data?.error?.msgKey || ''
    if (code === 'rate_limited' || e?.statusCode === 429) {
      error.value = t('support.errorRateLimit')
    } else if (code === 'diagnostics_too_large' || code === 'payload_too_large' || e?.statusCode === 413) {
      error.value = t('support.errorTooLarge')
    } else {
      error.value = t('support.errorGeneric')
    }
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.support-diag {
  margin: 0.75rem 0;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius);
  padding: 0.5rem 0.75rem;
  background: var(--pf-vet-bg);
}
.support-diag summary {
  cursor: pointer;
  color: var(--pf-vet-text-muted);
  font-size: 0.875rem;
}
.support-diag__pre {
  margin: 0.5rem 0 0;
  max-height: 12rem;
  overflow: auto;
  font-size: 0.7rem;
  white-space: pre-wrap;
  word-break: break-word;
}
.pro-textarea {
  width: 100%;
  min-height: 6rem;
  padding: 0.6rem 0.75rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius);
  background: var(--pf-vet-surface);
  color: var(--pf-vet-text);
  font: inherit;
  resize: vertical;
}
.pro-form-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 1rem;
}
</style>
