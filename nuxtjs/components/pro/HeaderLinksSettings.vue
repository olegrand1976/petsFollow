<template>
  <div class="header-links-settings" data-testid="settings-header-links">
    <p class="pro-profile-form__hint">{{ $t('headerLinks.settingsHint') }}</p>

    <section v-if="catalog.length" class="header-links-settings__block" data-testid="settings-header-links-catalog">
      <h3 class="header-links-settings__subtitle">{{ $t('headerLinks.catalogTitle') }}</h3>
      <ul class="header-links-settings__list">
        <li v-for="(row, idx) in catalogRows" :key="row.id" class="header-links-settings__row">
          <label class="header-links-settings__toggle">
            <input
              v-model="row.enabled"
              type="checkbox"
              :data-testid="`header-link-toggle-${row.id}`"
              @change="emitPrefs"
            >
            <span>{{ catalogLabel(row.id, row.label) }}</span>
          </label>
          <div class="header-links-settings__order">
            <button
              type="button"
              class="pro-topbar__icon-btn"
              :disabled="idx === 0"
              :aria-label="$t('headerLinks.moveUp')"
              @click="moveCatalog(idx, -1)"
            >
              <ProIcon name="keyboard_arrow_up" :size="18" />
            </button>
            <button
              type="button"
              class="pro-topbar__icon-btn"
              :disabled="idx === catalogRows.length - 1"
              :aria-label="$t('headerLinks.moveDown')"
              @click="moveCatalog(idx, 1)"
            >
              <ProIcon name="keyboard_arrow_down" :size="18" />
            </button>
          </div>
        </li>
      </ul>
    </section>

    <section class="header-links-settings__block" data-testid="settings-header-links-custom">
      <div class="header-links-settings__custom-head">
        <h3 class="header-links-settings__subtitle">{{ $t('headerLinks.customTitle') }}</h3>
        <span class="text-muted">{{ custom.length }}/{{ maxCustom }}</span>
      </div>

      <ul v-if="custom.length" class="header-links-settings__list">
        <li v-for="(c, idx) in custom" :key="c.id" class="header-links-settings__custom-row">
          <div class="header-links-settings__custom-fields">
            <ProInput
              v-model="c.label"
              :label="$t('headerLinks.customLabel')"
              :name="`custom-label-${c.id}`"
              maxlength="40"
              required
              @update:model-value="emitPrefs"
            />
            <ProInput
              v-model="c.url"
              :label="$t('headerLinks.customUrl')"
              :name="`custom-url-${c.id}`"
              type="url"
              required
              @update:model-value="emitPrefs"
            />
          </div>
          <div class="header-links-settings__order">
            <button
              type="button"
              class="pro-topbar__icon-btn"
              :disabled="idx === 0"
              :aria-label="$t('headerLinks.moveUp')"
              @click="moveCustom(idx, -1)"
            >
              <ProIcon name="keyboard_arrow_up" :size="18" />
            </button>
            <button
              type="button"
              class="pro-topbar__icon-btn"
              :disabled="idx === custom.length - 1"
              :aria-label="$t('headerLinks.moveDown')"
              @click="moveCustom(idx, 1)"
            >
              <ProIcon name="keyboard_arrow_down" :size="18" />
            </button>
            <button
              type="button"
              class="pro-topbar__icon-btn"
              :aria-label="$t('headerLinks.remove')"
              :data-testid="`header-link-remove-${c.id}`"
              @click="removeCustom(idx)"
            >
              <ProIcon name="delete" :size="18" />
            </button>
          </div>
        </li>
      </ul>

      <ProButton
        type="button"
        variant="secondary"
        :disabled="custom.length >= maxCustom"
        data-testid="header-links-add-custom"
        @click="addCustom"
      >
        {{ $t('headerLinks.addCustom') }}
      </ProButton>
      <p v-if="custom.length >= maxCustom" class="pro-profile-form__hint">{{ $t('headerLinks.maxReached') }}</p>
    </section>
  </div>
</template>

<script setup lang="ts">
export type HeaderLinkCustom = { id: string, label: string, url: string }
export type HeaderLinkCatalogRow = { id: string, label: string, url: string, enabled: boolean }
export type HeaderLinksPrefs = {
  configured: boolean
  enabled: string[]
  order: string[]
  custom: HeaderLinkCustom[]
}

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

const catalogRows = ref<HeaderLinkCatalogRow[]>([])
const custom = ref<HeaderLinkCustom[]>([])

watch(
  () => [props.catalog, props.modelValue] as const,
  () => {
    syncFromProps()
  },
  { immediate: true, deep: true },
)

function syncFromProps() {
  const prefs = props.modelValue
  const cat = (props.catalog || []).map((c) => ({ ...c }))
  // Order catalog rows by prefs.order when present.
  if (prefs.order?.length) {
    cat.sort((a, b) => {
      const ia = prefs.order.indexOf(a.id)
      const ib = prefs.order.indexOf(b.id)
      if (ia === -1 && ib === -1) return 0
      if (ia === -1) return 1
      if (ib === -1) return -1
      return ia - ib
    })
  }
  catalogRows.value = cat
  custom.value = (prefs.custom || []).map((c) => ({ ...c }))
}

function catalogLabel(id: string, fallback: string) {
  const key = `headerLinks.catalog.${id}`
  const translated = t(key)
  return translated === key ? fallback : translated
}

function buildPrefs(): HeaderLinksPrefs {
  const enabled = catalogRows.value.filter((r) => r.enabled).map((r) => r.id)
  for (const c of custom.value) {
    if (c.id && !enabled.includes(c.id)) enabled.push(c.id)
  }
  const order = [
    ...catalogRows.value.filter((r) => r.enabled).map((r) => r.id),
    ...custom.value.map((c) => c.id),
  ]
  return {
    configured: true,
    enabled,
    order,
    custom: custom.value.map((c) => ({
      id: c.id,
      label: c.label.trim(),
      url: c.url.trim(),
    })),
  }
}

function emitPrefs() {
  emit('update:modelValue', buildPrefs())
}

function moveCatalog(idx: number, delta: number) {
  const next = idx + delta
  if (next < 0 || next >= catalogRows.value.length) return
  const arr = [...catalogRows.value]
  const [row] = arr.splice(idx, 1)
  arr.splice(next, 0, row)
  catalogRows.value = arr
  emitPrefs()
}

function moveCustom(idx: number, delta: number) {
  const next = idx + delta
  if (next < 0 || next >= custom.value.length) return
  const arr = [...custom.value]
  const [row] = arr.splice(idx, 1)
  arr.splice(next, 0, row)
  custom.value = arr
  emitPrefs()
}

function addCustom() {
  if (custom.value.length >= props.maxCustom) return
  const id = `custom_${Math.random().toString(36).slice(2, 10)}`
  custom.value = [...custom.value, { id, label: '', url: 'https://' }]
  emitPrefs()
}

function removeCustom(idx: number) {
  custom.value = custom.value.filter((_, i) => i !== idx)
  emitPrefs()
}
</script>

<style scoped>
.header-links-settings__block {
  margin-top: 1rem;
}
.header-links-settings__subtitle {
  margin: 0 0 0.5rem;
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--pf-vet-primary);
}
.header-links-settings__custom-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 1rem;
}
.header-links-settings__list {
  list-style: none;
  margin: 0 0 0.75rem;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.header-links-settings__row,
.header-links-settings__custom-row {
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
