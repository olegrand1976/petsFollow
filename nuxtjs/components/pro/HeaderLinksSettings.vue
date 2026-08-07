<template>
  <div class="header-links-settings" data-testid="settings-header-links">
    <p class="pro-profile-form__hint">{{ $t('headerLinks.settingsHint') }}</p>

    <ul v-if="rows.length" class="header-links-settings__list" data-testid="settings-header-links-list">
      <li
        v-for="(row, idx) in rows"
        :key="row.id"
        class="header-links-settings__row"
        :data-testid="`header-link-row-${row.id}`"
      >
        <template v-if="row.kind === 'catalog'">
          <label class="header-links-settings__toggle">
            <input
              v-model="row.enabled"
              type="checkbox"
              :data-testid="`header-link-toggle-${row.id}`"
              @change="emitPrefs"
            >
            <span>{{ catalogLabel(row.id, row.label) }}</span>
          </label>
        </template>
        <template v-else>
          <div class="header-links-settings__custom-fields">
            <ProInput
              v-model="row.label"
              :label="$t('headerLinks.customLabel')"
              :name="`custom-label-${row.id}`"
              maxlength="40"
              required
              @update:model-value="emitPrefs"
            />
            <ProInput
              v-model="row.url"
              :label="$t('headerLinks.customUrl')"
              :name="`custom-url-${row.id}`"
              type="url"
              required
              @update:model-value="emitPrefs"
            />
          </div>
        </template>
        <div class="header-links-settings__order">
          <button
            type="button"
            class="pro-topbar__icon-btn"
            :disabled="idx === 0"
            :aria-label="$t('headerLinks.moveUp')"
            :data-testid="`header-link-up-${row.id}`"
            @click="move(idx, -1)"
          >
            <ProIcon name="keyboard_arrow_up" :size="18" />
          </button>
          <button
            type="button"
            class="pro-topbar__icon-btn"
            :disabled="idx === rows.length - 1"
            :aria-label="$t('headerLinks.moveDown')"
            :data-testid="`header-link-down-${row.id}`"
            @click="move(idx, 1)"
          >
            <ProIcon name="keyboard_arrow_down" :size="18" />
          </button>
          <button
            v-if="row.kind === 'custom'"
            type="button"
            class="pro-topbar__icon-btn"
            :aria-label="$t('headerLinks.remove')"
            :data-testid="`header-link-remove-${row.id}`"
            @click="removeCustom(idx)"
          >
            <ProIcon name="delete" :size="18" />
          </button>
        </div>
      </li>
    </ul>

    <div class="header-links-settings__custom-head">
      <span class="text-muted">{{ customCount }}/{{ maxCustom }}</span>
      <ProButton
        type="button"
        variant="secondary"
        :disabled="customCount >= maxCustom"
        data-testid="header-links-add-custom"
        @click="addCustom"
      >
        {{ $t('headerLinks.addCustom') }}
      </ProButton>
    </div>
    <p v-if="customCount >= maxCustom" class="pro-profile-form__hint">{{ $t('headerLinks.maxReached') }}</p>
  </div>
</template>

<script setup lang="ts">
import type { HeaderLinkCatalogRow, HeaderLinkUnifiedRow, HeaderLinksPrefs } from '~/utils/headerLinks'
import {
  buildUnifiedRows,
  moveRow,
  prefsFromUnifiedRows,
} from '~/utils/headerLinks'

const props = withDefaults(
  defineProps<{
    catalog?: HeaderLinkCatalogRow[]
    maxCustom?: number
    modelValue: HeaderLinksPrefs
  }>(),
  {
    catalog: () => [],
    maxCustom: 10,
  },
)

const emit = defineEmits<{
  'update:modelValue': [HeaderLinksPrefs]
}>()

const { t } = useI18n()

const rows = ref<HeaderLinkUnifiedRow[]>([])
let applyingLocal = false

const customCount = computed(() => rows.value.filter((r) => r.kind === 'custom').length)

watch(
  () => [props.catalog, props.modelValue] as const,
  () => {
    if (applyingLocal) return
    rows.value = buildUnifiedRows(props.catalog || [], props.modelValue)
  },
  { immediate: true, deep: true },
)

function catalogLabel(id: string, fallback: string) {
  const key = `headerLinks.catalog.${id}`
  const translated = t(key)
  return translated === key ? fallback : translated
}

function emitPrefs() {
  applyingLocal = true
  emit('update:modelValue', prefsFromUnifiedRows(rows.value))
  nextTick(() => {
    applyingLocal = false
  })
}

/** Flush local draft fields into v-model (call before parent save). */
function flush() {
  emitPrefs()
}

function move(idx: number, delta: number) {
  rows.value = moveRow(rows.value, idx, delta)
  emitPrefs()
}

function addCustom() {
  if (customCount.value >= props.maxCustom) return
  const id = `custom_${Math.random().toString(36).slice(2, 10)}`
  rows.value = [...rows.value, { kind: 'custom', id, label: '', url: 'https://' }]
  emitPrefs()
}

function removeCustom(idx: number) {
  rows.value = rows.value.filter((_, i) => i !== idx)
  emitPrefs()
}

defineExpose({ flush })
</script>

<style scoped>
.header-links-settings__custom-head {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 0.5rem;
}
.header-links-settings__list {
  list-style: none;
  margin: 0.75rem 0 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.header-links-settings__row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.5rem 0.65rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius-md, 0.5rem);
  background: var(--pf-vet-surface);
}
.header-links-settings__toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  padding-top: 0.35rem;
}
.header-links-settings__order {
  display: inline-flex;
  align-items: center;
  gap: 0.15rem;
  flex-shrink: 0;
}
.header-links-settings__custom-fields {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}
</style>
