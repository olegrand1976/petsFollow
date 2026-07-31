<script setup lang="ts">
/**
 * Autocomplete combobox for Pro (pharmacy CNK search, etc.).
 * Debounced search via `searchFn`; emits the selected item.
 */
export type ProComboboxItem = {
  id: string
  label: string
  hint?: string
  badge?: string
  raw?: unknown
}

const props = withDefaults(
  defineProps<{
    modelValue?: ProComboboxItem | null
    /** Forwards to the input — use with an external <label for="…">. */
    inputId?: string
    placeholder?: string
    disabled?: boolean
    debounceMs?: number
    minChars?: number
    searchFn: (q: string) => Promise<ProComboboxItem[]>
  }>(),
  {
    modelValue: null,
    inputId: undefined,
    placeholder: '',
    disabled: false,
    debounceMs: 250,
    minChars: 2,
  },
)

const emit = defineEmits<{
  'update:modelValue': [ProComboboxItem | null]
  select: [ProComboboxItem]
}>()

const query = ref(props.modelValue?.label ?? '')
const items = ref<ProComboboxItem[]>([])
const open = ref(false)
const loading = ref(false)
const activeIdx = ref(-1)
let timer: ReturnType<typeof setTimeout> | null = null
let blurTimer: ReturnType<typeof setTimeout> | null = null
let reqSeq = 0

onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
  if (blurTimer) clearTimeout(blurTimer)
})

watch(
  () => props.modelValue,
  (v) => {
    if (v) query.value = v.label
    else if (!open.value) query.value = ''
  },
)

function onInput() {
  emit('update:modelValue', null)
  open.value = true
  activeIdx.value = -1
  if (timer) clearTimeout(timer)
  timer = setTimeout(runSearch, props.debounceMs)
}

async function runSearch() {
  const q = query.value.trim()
  if (q.length < props.minChars) {
    items.value = []
    return
  }
  const seq = ++reqSeq
  loading.value = true
  try {
    const res = await props.searchFn(q)
    if (seq !== reqSeq) return
    items.value = res
    open.value = true
  } catch {
    if (seq !== reqSeq) return
    items.value = []
  } finally {
    if (seq === reqSeq) loading.value = false
  }
}

function select(item: ProComboboxItem) {
  query.value = item.label
  items.value = []
  open.value = false
  emit('update:modelValue', item)
  emit('select', item)
}

function onKeydown(e: KeyboardEvent) {
  if (!open.value || items.value.length === 0) return
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    activeIdx.value = Math.min(activeIdx.value + 1, items.value.length - 1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    activeIdx.value = Math.max(activeIdx.value - 1, 0)
  } else if (e.key === 'Enter' && activeIdx.value >= 0) {
    e.preventDefault()
    select(items.value[activeIdx.value]!)
  } else if (e.key === 'Escape') {
    open.value = false
  }
}

function onBlur() {
  if (blurTimer) clearTimeout(blurTimer)
  blurTimer = setTimeout(() => {
    open.value = false
  }, 150)
}
</script>

<template>
  <div class="pro-combobox" data-testid="pro-combobox">
    <input
      :id="inputId"
      v-model="query"
      type="search"
      class="pro-input__field"
      :placeholder="placeholder"
      :disabled="disabled"
      autocomplete="off"
      role="combobox"
      :aria-expanded="open"
      aria-autocomplete="list"
      data-testid="pro-combobox-input"
      @input="onInput"
      @keydown="onKeydown"
      @focus="open = items.length > 0"
      @blur="onBlur"
    >
    <ul
      v-if="open && (items.length > 0 || loading)"
      class="pro-combobox__list"
      role="listbox"
      data-testid="pro-combobox-list"
    >
      <li v-if="loading" class="pro-combobox__empty">…</li>
      <li
        v-for="(item, idx) in items"
        :key="item.id"
        class="pro-combobox__option"
        :class="{ 'pro-combobox__option--active': idx === activeIdx }"
        role="option"
        :aria-selected="idx === activeIdx"
        @mousedown.prevent="select(item)"
      >
        <span class="pro-combobox__label">{{ item.label }}</span>
        <span v-if="item.hint" class="pro-combobox__hint">{{ item.hint }}</span>
        <ProBadge v-if="item.badge" variant="warning" class="pro-combobox__badge">
          {{ item.badge }}
        </ProBadge>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.pro-combobox {
  position: relative;
  width: 100%;
}
.pro-combobox .pro-input__field {
  width: 100%;
  box-sizing: border-box;
  padding: 0.55rem 0.75rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  background: var(--pf-vet-surface);
  color: var(--pf-vet-text, inherit);
  font: inherit;
}
.pro-combobox__list {
  position: absolute;
  z-index: 20;
  left: 0;
  right: 0;
  margin: 0.25rem 0 0;
  padding: 0.25rem 0;
  list-style: none;
  max-height: 16rem;
  overflow: auto;
  background: var(--pf-vet-surface);
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  box-shadow: var(--pf-vet-shadow-md);
}
.pro-combobox__option {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem 0.5rem;
  padding: 0.5rem 0.75rem;
  cursor: pointer;
}
.pro-combobox__option--active,
.pro-combobox__option:hover {
  background: color-mix(in srgb, var(--pf-vet-accent) 12%, transparent);
}
.pro-combobox__label {
  font-weight: 600;
  flex: 1 1 auto;
}
.pro-combobox__hint {
  font-size: 0.8rem;
  color: var(--pf-vet-muted, #64748b);
}
.pro-combobox__empty {
  padding: 0.5rem 0.75rem;
  color: var(--pf-vet-muted, #64748b);
}
</style>
