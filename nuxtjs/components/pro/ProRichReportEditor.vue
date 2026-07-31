<template>
  <div
    class="pro-rich-report"
    :class="{ 'pro-rich-report--disabled': disabled || readonly }"
    :data-testid="editorTestId"
  >
    <div
      v-if="!readonly && !disabled"
      class="pro-rich-report__toolbar"
      data-testid="visit-report-rich-toolbar"
    >
      <button
        type="button"
        class="pro-rich-report__btn"
        :class="{ 'is-active': editor?.isActive('bold') }"
        data-testid="visit-report-bold"
        :aria-label="$t('calendar.reportFormatBold')"
        @click="editor?.chain().focus().toggleBold().run()"
      >
        <ProIcon name="format_bold" :size="18" />
      </button>
      <button
        type="button"
        class="pro-rich-report__btn"
        :class="{ 'is-active': editor?.isActive('italic') }"
        data-testid="visit-report-italic"
        :aria-label="$t('calendar.reportFormatItalic')"
        @click="editor?.chain().focus().toggleItalic().run()"
      >
        <ProIcon name="format_italic" :size="18" />
      </button>
      <button
        type="button"
        class="pro-rich-report__btn"
        :class="{ 'is-active': editor?.isActive('bulletList') }"
        data-testid="visit-report-bullet"
        :aria-label="$t('calendar.reportFormatBullet')"
        @click="editor?.chain().focus().toggleBulletList().run()"
      >
        <ProIcon name="format_list_bulleted" :size="18" />
      </button>
      <button
        type="button"
        class="pro-rich-report__btn"
        :class="{ 'is-active': editor?.isActive('orderedList') }"
        data-testid="visit-report-ordered"
        :aria-label="$t('calendar.reportFormatOrdered')"
        @click="editor?.chain().focus().toggleOrderedList().run()"
      >
        <ProIcon name="format_list_numbered" :size="18" />
      </button>
    </div>
    <!-- Tiny markdown mirror for Playwright fill/toHaveValue (not for screen readers). -->
    <textarea
      :id="inputId"
      class="pro-rich-report__mirror"
      :data-testid="textareaTestId"
      :value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      :readonly="readonly"
      aria-hidden="true"
      tabindex="-1"
      rows="1"
      @input="onMirrorInput"
    />
    <EditorContent
      v-if="editor"
      class="pro-input pro-rich-report__editor"
      :data-testid="`${testIdPrefix}-prose`"
      :editor="editor"
    />
  </div>
</template>

<script setup lang="ts">
import { Editor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import {
  canonicalizeReportMarkdown,
  editorHtmlToMarkdown,
  markdownToEditorHtml,
} from '~/utils/reportRichText'

const props = withDefaults(
  defineProps<{
    modelValue: string
    placeholder?: string
    disabled?: boolean
    readonly?: boolean
    inputId?: string
    textareaTestId?: string
    editorTestId?: string
    testIdPrefix?: string
  }>(),
  {
    placeholder: '',
    disabled: false,
    readonly: false,
    inputId: 'visit-report-body',
    textareaTestId: 'visit-report-body',
    editorTestId: 'visit-report-editor',
    testIdPrefix: 'visit-report',
  },
)

const emit = defineEmits<{ 'update:modelValue': [string] }>()

const editor = shallowRef<Editor | null>(null)
let applyingExternal = false

function syncFromMarkdown(md: string) {
  if (!editor.value) return
  const incoming = canonicalizeReportMarkdown(md)
  const current = canonicalizeReportMarkdown(editorHtmlToMarkdown(editor.value.getHTML()))
  if (current === incoming) return
  applyingExternal = true
  editor.value.commands.setContent(markdownToEditorHtml(md), { emitUpdate: false })
  applyingExternal = false
}

function onMirrorInput(ev: Event) {
  const el = ev.target as HTMLTextAreaElement
  emit('update:modelValue', el.value)
}

function emitCanonicalFromEditor(ed: Editor) {
  emit('update:modelValue', canonicalizeReportMarkdown(editorHtmlToMarkdown(ed.getHTML())))
}

onMounted(() => {
  editor.value = new Editor({
    extensions: [
      StarterKit.configure({
        heading: false,
        codeBlock: false,
        code: false,
        blockquote: false,
        horizontalRule: false,
      }),
    ],
    content: markdownToEditorHtml(props.modelValue),
    editable: !props.disabled && !props.readonly,
    editorProps: {
      attributes: {
        class: 'pro-rich-report__prose',
      },
    },
    onUpdate: ({ editor: ed }) => {
      if (applyingExternal) return
      emitCanonicalFromEditor(ed)
    },
  })
})

onBeforeUnmount(() => {
  editor.value?.destroy()
  editor.value = null
})

watch(
  () => props.modelValue,
  (md) => {
    syncFromMarkdown(md)
  },
)

watch(
  () => [props.disabled, props.readonly] as const,
  ([disabled, readonly]) => {
    editor.value?.setEditable(!disabled && !readonly)
  },
)
</script>

<style scoped>
.pro-rich-report {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 0;
  position: relative;
}

.pro-rich-report__toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.pro-rich-report__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  background: var(--pf-vet-surface, #fff);
  color: var(--pf-vet-primary);
  cursor: pointer;
}

.pro-rich-report__btn.is-active,
.pro-rich-report__btn:hover {
  background: var(--pf-vet-bg, #f8fafc);
  border-color: var(--pf-vet-accent);
}

.pro-rich-report__editor {
  min-height: 280px;
  width: 100%;
  padding: 0;
  overflow: auto;
}

.pro-rich-report__editor :deep(.ProseMirror) {
  min-height: 260px;
  padding: 0.75rem 1rem;
  outline: none;
  line-height: 1.45;
}

.pro-rich-report__editor :deep(.ProseMirror p) {
  margin: 0.35rem 0;
}

.pro-rich-report__editor :deep(.ProseMirror ul),
.pro-rich-report__editor :deep(.ProseMirror ol) {
  margin: 0.35rem 0 0.35rem 1.25rem;
}

.pro-rich-report__editor :deep(.ProseMirror strong) {
  font-weight: 600;
  color: var(--pf-vet-primary);
}

.pro-rich-report--disabled .pro-rich-report__editor {
  opacity: 0.85;
}

/* 1×1 opaque: actionable for Playwright fill/toHaveValue, invisible to users. */
.pro-rich-report__mirror {
  position: absolute;
  left: 0;
  top: 0;
  width: 1px;
  height: 1px;
  opacity: 0.01;
  margin: 0;
  padding: 0;
  border: 0;
  overflow: hidden;
  resize: none;
  z-index: 0;
}
</style>
