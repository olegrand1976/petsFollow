<template>
  <div class="pro-avatar-upload" data-testid="avatar-upload">
    <ProAvatar :src="previewUrl || modelValue" :name="name" size="lg" :alt="name" />
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

async function onFile(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  error.value = ''
  loading.value = true
  previewUrl.value = URL.createObjectURL(file)
  try {
    const fd = new FormData()
    fd.append('file', file)
    const res: any = await $fetch(props.uploadUrl, { method: 'POST', body: fd })
    const data = res.data ?? res
    const url = data.avatarUrl || data.photoUrl || ''
    if (url) emit('update:modelValue', url)
    emit('uploaded', data)
  } catch (e: any) {
    error.value = mapError(e)
    previewUrl.value = null
  } finally {
    loading.value = false
    if (inputEl.value) inputEl.value.value = ''
  }
}
</script>

<style scoped>
.pro-avatar-upload {
  display: flex;
  align-items: center;
  gap: 1rem;
}
.pro-avatar-upload__actions {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}
.pro-avatar-upload__label {
  display: inline-block;
  cursor: pointer;
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
