<template>
  <ProModal :open="open" size="md" :title="title" @update:open="emit('update:open', $event)">
    <div v-if="loading" class="text-muted">{{ $t('common.loading') }}</div>
    <template v-else-if="invite">
      <p class="pro-hint pro-mb-md">{{ hintText }}</p>
      <div class="app-invite-qr" data-testid="app-invite-qr">
        <img :src="invite.qrCodeDataUrl" :alt="$t('clients.appInvite.qrAlt')" width="200" height="200">
      </div>
      <div v-if="invite.qrAndroid || invite.qrIos" class="app-invite-store-qrs" data-testid="app-invite-store-qrs">
        <figure v-if="invite.qrAndroid?.publicUrl">
          <img :src="invite.qrAndroid.publicUrl" :alt="$t('clients.appInvite.qrAndroidAlt')" width="96" height="96">
          <figcaption>{{ $t('clients.appInvite.android') }}</figcaption>
        </figure>
        <figure v-if="invite.qrIos?.publicUrl">
          <img :src="invite.qrIos.publicUrl" :alt="$t('clients.appInvite.qrIosAlt')" width="96" height="96">
          <figcaption>{{ $t('clients.appInvite.ios') }}</figcaption>
        </figure>
      </div>
      <p class="app-invite-meta">
        <strong>{{ metaPrimary }}</strong>
        <span v-if="metaSecondary" class="text-muted"> — {{ metaSecondary }}</span>
      </p>
      <p class="app-invite-code text-muted">{{ $t('clients.appInvite.codeLabel') }} {{ invite.code }}</p>
      <div class="app-invite-actions">
        <ProInput
          :model-value="invite.inviteUrl"
          disabled
          :label="hasVetRegister ? $t('clients.appInvite.linkLabelClient') : $t('clients.appInvite.linkLabel')"
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
        <template v-if="hasVetRegister">
          <ProInput
            :model-value="invite.vetRegisterUrl"
            disabled
            :label="$t('clients.appInvite.linkLabelCabinet')"
            test-id="app-invite-vet-register-url"
          />
          <ProButton
            type="button"
            variant="secondary"
            test-id="app-invite-copy-vet-register"
            @click="copyVetRegister"
          >
            <ProIcon name="content_copy" />
            {{ vetCopied ? $t('clients.appInvite.copied') : $t('clients.appInvite.copyCabinet') }}
          </ProButton>
        </template>
        <ProButton
          type="button"
          variant="ghost"
          test-id="app-invite-refresh-pack"
          :disabled="loading"
          @click="loadInvite"
        >
          {{ $t('clients.appInvite.refreshPack') }}
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
const vetCopied = ref(false)
const copyError = ref('')
const invite = ref<{
  code: string
  role: string
  inviteUrl: string
  vetRegisterUrl?: string
  qrCodeDataUrl: string
  practiceName: string
  displayName: string
  qrAndroid?: { publicUrl?: string, storeUrl?: string } | null
  qrIos?: { publicUrl?: string, storeUrl?: string } | null
} | null>(null)

const hasVetRegister = computed(() => Boolean(invite.value?.vetRegisterUrl?.trim()))

const hintText = computed(() => {
  switch (invite.value?.role) {
    case 'care_pro':
      return t('clients.appInvite.hintCarePro')
    case 'commercial':
    case 'commercial_manager':
      return t('clients.appInvite.hintCommercial')
    case 'vet':
    default:
      return t('clients.appInvite.hint')
  }
})

const metaPrimary = computed(() => {
  const inv = invite.value
  if (!inv) return ''
  if (inv.practiceName.trim()) return inv.practiceName
  return inv.displayName || t('clients.appInvite.fallbackName')
})

const metaSecondary = computed(() => {
  const inv = invite.value
  if (!inv) return ''
  if (inv.practiceName.trim() && inv.displayName.trim()) return inv.displayName
  return ''
})

async function loadInvite() {
  loading.value = true
  error.value = ''
  invite.value = null
  try {
    const res: any = await $fetch('/api/me/app-invite')
    const data = res.data ?? res
    invite.value = {
      code: data.code,
      role: data.role || 'vet',
      inviteUrl: data.inviteUrl,
      vetRegisterUrl: data.vetRegisterUrl || '',
      qrCodeDataUrl: data.qrCodeDataUrl,
      practiceName: data.practiceName || '',
      displayName: data.displayName || data.vetFullName || '',
      qrAndroid: data.qrAndroid || null,
      qrIos: data.qrIos || null,
    }
  } catch (e) {
    error.value = mapError(e)
  } finally {
    loading.value = false
  }
}

async function copyText(url: string, mark: 'client' | 'vet') {
  copyError.value = ''
  copied.value = false
  vetCopied.value = false
  if (!url) return
  try {
    await navigator.clipboard.writeText(url)
    if (mark === 'vet') vetCopied.value = true
    else copied.value = true
  } catch {
    copyError.value = t('clients.appInvite.copyError')
  }
}

async function copyLink() {
  await copyText(invite.value?.inviteUrl || '', 'client')
}

async function copyVetRegister() {
  await copyText(invite.value?.vetRegisterUrl || '', 'vet')
}

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    copied.value = false
    vetCopied.value = false
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
.app-invite-store-qrs {
  display: flex;
  justify-content: center;
  gap: 1rem;
  margin-bottom: 1rem;
}
.app-invite-store-qrs figure {
  margin: 0;
  text-align: center;
  font-size: 0.8rem;
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
