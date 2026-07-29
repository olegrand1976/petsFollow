<template>
  <div class="pro-md-report" data-testid="visit-report-editor">
    <div class="pro-flex-gap pro-md-report__tabs">
      <ProButton
        :variant="mode === 'preview' ? 'primary' : 'secondary'"
        test-id="visit-report-preview-tab"
        :disabled="disabled"
        @click="mode = 'preview'"
      >
        {{ $t('calendar.reportPreview') }}
      </ProButton>
      <ProButton
        :variant="mode === 'edit' ? 'primary' : 'secondary'"
        test-id="visit-report-edit-tab"
        :disabled="disabled"
        @click="mode = 'edit'"
      >
        {{ $t('calendar.reportEditSource') }}
      </ProButton>
    </div>
    <div
      v-if="mode === 'preview'"
      class="pro-md-report__preview pro-input"
      data-testid="visit-report-preview"
      v-html="safeHtml"
    />
    <textarea
      v-else
      :id="inputId"
      class="pro-input pro-md-report__textarea"
      rows="14"
      data-testid="visit-report-body"
      :value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      :readonly="readonly"
      @input="onInput"
    />
  </div>
</template>

<script setup lang="ts">
import { renderSafeMarkdown } from '~/utils/safeMarkdown'

const props = withDefaults(
  defineProps<{
    modelValue: string
    placeholder?: string
    disabled?: boolean
    readonly?: boolean
    inputId?: string
    /**
     * Incrementing tick forces preview mode (boolean preferPreview alone would not
     * re-fire the watcher when already true after a second improve/transcribe).
     */
    preferPreviewTick?: number
  }>(),
  {
    placeholder: '',
    disabled: false,
    readonly: false,
    inputId: 'visit-report-body',
    preferPreviewTick: 0,
  },
)

const emit = defineEmits<{ 'update:modelValue': [string] }>()

const mode = ref<'preview' | 'edit'>('edit')

watch(
  () => props.preferPreviewTick,
  (tick) => {
    if (tick > 0) mode.value = 'preview'
  },
)

const safeHtml = computed(() => renderSafeMarkdown(props.modelValue))

function onInput(ev: Event) {
  const el = ev.target as HTMLTextAreaElement
  emit('update:modelValue', el.value)
}
</script>

<style scoped>
.pro-md-report__tabs {
  margin-bottom: 0.5rem;
}

.pro-md-report__textarea,
.pro-md-report__preview {
  min-height: 42vh;
  width: 100%;
  resize: vertical;
}

.pro-md-report__preview {
  overflow: auto;
  padding: 0.75rem 1rem;
  line-height: 1.45;
  white-space: normal;
}

.pro-md-report__preview :deep(h1),
.pro-md-report__preview :deep(h2),
.pro-md-report__preview :deep(h3),
.pro-md-report__preview :deep(h4) {
  margin: 0.75rem 0 0.35rem;
  color: var(--pf-vet-primary);
  font-size: 1rem;
}

.pro-md-report__preview :deep(p) {
  margin: 0.35rem 0;
}

.pro-md-report__preview :deep(ul),
.pro-md-report__preview :deep(ol) {
  margin: 0.35rem 0 0.35rem 1.25rem;
}

.pro-md-report__preview :deep(strong) {
  font-weight: 600;
}
</style>
