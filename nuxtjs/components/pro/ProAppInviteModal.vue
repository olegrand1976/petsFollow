<template>
  <ProModal :open="open" size="md" :title="title" @update:open="emit('update:open', $event)">
    <div v-if="loading" class="text-muted">{{ $t('common.loading') }}</div>
    <template v-else-if="invite">
      <p class="pro-hint pro-mb-md">{{ $t('clients.appInvite.hint') }}</p>
      <div class="app-invite-qr" data-testid="app-invite-qr">
        <img :src="invite.qrCodeDataUrl" :alt="$t('clients.appInvite.qrAlt')" width="200" height="200">
      </div>
      <p class="app-invite-meta">
        <strong>{{ invite.practiceName }}</strong>
        <span v-if="invite.vetFullName" class="text-muted"> — {{ invite.vetFullName }}</span>
      </p>
      <p class="app-invite-code text-muted">{{ $t('clients.appInvite.codeLabel') }} {{ invite.code }}</p>
      <div class="app-invite-actions">
        <ProInput
          :model-value="invite.inviteUrl"
          disabled
          :label="$t('clients.appInvite.linkLabel')"
          test-id="app-invite-url"
        />
        <ProButton
          type="button"
          variant="secondary"
          test-id="app-invite-copy"
          @click="copyLink"
        >
          <ProIcon name="content_copy" />
          {{ copied ? $t('clients.appInvite.copied') : $t('clients.appInvite.copy') }}
        </ProButton>
      </div>
      <p v-if="copyError" class="pro-error">{{ copyError }}</p>
    </template>
    <p v-else-if="error" class="pro-error" role="alert">{{ error }}</p>
  </ProModal>
</template>

<script setup lang="ts">
const props = defineProps<{
  open: boolean
  title?: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const { t } = useI18n()
const { mapError } = useApiError()

const title = computed(() => props.title || t('clients.appInvite.title'))
const loading = ref(false)
const error = ref('')
const copied = ref(false)
const copyError = ref('')
const invite = ref<{
  code: string
  inviteUrl: string
  qrCodeDataUrl: string
  practiceName: string
  vetFullName: string
} | null>(null)

async function loadInvite() {
  loading.value = true
  error.value = ''
  invite.value = null
  try {
    const res: any = await $fetch('/api/vet/app-invite')
    const data = res.data ?? res
    invite.value = {
      code: data.code,
      inviteUrl: data.inviteUrl,
      qrCodeDataUrl: data.qrCodeDataUrl,
      practiceName: data.practiceName || '',
      vetFullName: data.vetFullName || '',
    }
  } catch (e) {
    error.value = mapError(e)
  } finally {
    loading.value = false
  }
}

async function copyLink() {
  copyError.value = ''
  copied.value = false
  const url = invite.value?.inviteUrl
  if (!url) return
  try {
    await navigator.clipboard.writeText(url)
    copied.value = true
  } catch {
    copyError.value = t('clients.appInvite.copyError')
  }
}

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    copied.value = false
    copyError.value = ''
    loadInvite()
  }
})
</script>

<style scoped>
.app-invite-qr {
  display: flex;
  justify-content: center;
  margin-bottom: 1rem;
}
.app-invite-qr img {
  border-radius: var(--pf-vet-radius-md, 8px);
  background: #fff;
  padding: 0.5rem;
  border: 1px solid var(--pf-vet-border);
}
.app-invite-meta {
  text-align: center;
  margin: 0 0 0.25rem;
}
.app-invite-code {
  text-align: center;
  font-family: ui-monospace, monospace;
  letter-spacing: 0.06em;
  margin: 0 0 1rem;
}
.app-invite-actions {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
</style>
