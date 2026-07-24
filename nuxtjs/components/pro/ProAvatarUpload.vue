<template>
  <div class="pro-avatar-upload" data-testid="avatar-upload">
    <ProAvatar
      :src="displaySrc"
      :name="name"
      size="xl"
      fit="contain"
      :alt="name"
    />
    <div class="pro-avatar-upload__actions">
      <label class="pro-avatar-upload__label">
        <input
          ref="inputEl"
          type="file"
          accept="image/jpeg,image/png,image/webp"
          class="pro-avatar-upload__input"
          :disabled="loading"
          data-testid="avatar-upload-input"
          @change="onFile"
        >
        <span class="pro-btn pro-btn--secondary">{{ loading ? '…' : label }}</span>
      </label>
      <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
      <p v-if="hint" class="pro-settings-hint">{{ hint }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    modelValue?: string | null
    name?: string
    uploadUrl: string
    label?: string
    hint?: string
  }>(),
  {
    modelValue: null,
    name: '',
    label: undefined,
    hint: undefined,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
  uploaded: [payload: any]
}>()

const { t } = useI18n()
const { mapError } = useApiError()

const loading = ref(false)
const error = ref('')
const previewUrl = ref<string | null>(null)
const inputEl = ref<HTMLInputElement | null>(null)

const label = computed(() => props.label || t('settings.avatar.change'))
const displaySrc = computed(() => previewUrl.value || props.modelValue || null)

function clearPreview() {
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = null
  }
}

async function onFile(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  error.value = ''
  loading.value = true
  clearPreview()
  previewUrl.value = URL.createObjectURL(file)
  try {
    const fd = new FormData()
    fd.append('file', file)
    const res: any = await $fetch(props.uploadUrl, { method: 'POST', body: fd })
    const data = res.data ?? res
    const url = data.avatarUrl || data.photoUrl || ''
    if (url) {
      clearPreview()
      emit('update:modelValue', url)
    }
    emit('uploaded', data)
  } catch (e: any) {
    error.value = mapError(e)
    clearPreview()
  } finally {
    loading.value = false
    if (inputEl.value) inputEl.value.value = ''
  }
}

onBeforeUnmount(() => clearPreview())
</script>

<style scoped>
.pro-avatar-upload {
  display: flex;
  align-items: center;
  gap: 1.25rem;
}

.pro-avatar-upload :deep(.pro-avatar--xl) {
  width: 6.5rem;
  height: 6.5rem;
  border: 1px solid var(--pf-vet-border);
  background: var(--pf-vet-bg);
  box-shadow: var(--pf-vet-shadow-sm);
}

.pro-avatar-upload :deep(.pro-avatar__img) {
  padding: 0.4rem;
  box-sizing: border-box;
}

.pro-avatar-upload__actions {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  min-width: 0;
}

.pro-avatar-upload__label {
  display: inline-block;
  cursor: pointer;
  width: fit-content;
}

.pro-avatar-upload__input {
  position: absolute;
  width: 1px;
  height: 1px;
  opacity: 0;
  overflow: hidden;
}

.pro-avatar-upload__label .pro-btn {
  pointer-events: none;
}
</style>
