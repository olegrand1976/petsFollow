<template>
  <div data-testid="medicaments-page">
    <ProPageHeader
      :title="$t('pharmacy.medicaments.title')"
      :subtitle="$t('pharmacy.medicaments.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="medicaments-page-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
      </template>
    </ProPageHeader>

    <details class="pharmacy-legal" data-testid="pharmacy-legal">
      <summary>{{ $t('pharmacy.legal.title') }}</summary>
      <div class="pharmacy-legal__body">
        <p>{{ $t('pharmacy.legal.fefo') }}</p>
        <p>{{ $t('pharmacy.legal.daf') }}</p>
        <p>{{ $t('pharmacy.legal.vamreg') }}</p>
      </div>
    </details>

    <ProCard>
      <label class="pro-label" for="med-search">{{ $t('pharmacy.medicaments.searchLabel') }}</label>
      <ProCombobox
        input-id="med-search"
        v-model="selected"
        :placeholder="$t('pharmacy.medicaments.searchPlaceholder')"
        :search-fn="searchMedications"
        data-testid="medicaments-search"
      />
      <p v-if="selected" class="pro-hint" data-testid="medicaments-selected">
        <strong>{{ selected.label }}</strong>
        <span v-if="selected.hint"> — {{ selected.hint }}</span>
        <ProBadge v-if="selected.badge" variant="warning">{{ selected.badge }}</ProBadge>
      </p>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'

definePageMeta({ middleware: ['auth', 'vet-only', 'practice-perm'], practicePerm: 'pharmacy.read' })

const { t } = useI18n()
const selected = ref<ProComboboxItem | null>(null)

type MedRow = {
  id: string
  cnk: string
  name: string
  atcCode?: string
  pharmaceuticalForm?: string
  packSize?: string
  isAntibiotic?: boolean
}

async function searchMedications(q: string): Promise<ProComboboxItem[]> {
  const res = await $fetch<{ data?: { items?: MedRow[] } } | { items?: MedRow[] }>(
    '/api/vet/pharmacy/medications/search',
    { query: { q, limit: '20' } },
  )
  const payload = (res as { data?: { items?: MedRow[] } })?.data ?? res
  const items = (payload as { items?: MedRow[] })?.items ?? []
  return items.map((m) => ({
    id: m.id,
    label: m.name,
    hint: m.cnk + (m.pharmaceuticalForm ? ` · ${m.pharmaceuticalForm}` : ''),
    badge: m.isAntibiotic ? t('pharmacy.antibioticWarning') : undefined,
    raw: m,
  }))
}
</script>

<style scoped>
.pharmacy-legal {
  margin: 0 0 1rem;
  padding: 0.65rem 0.85rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  background: var(--pf-vet-surface);
}
.pharmacy-legal summary {
  cursor: pointer;
  font-weight: 600;
  font-size: 0.9rem;
}
.pharmacy-legal__body {
  margin-top: 0.65rem;
  display: grid;
  gap: 0.5rem;
  font-size: 0.85rem;
  color: var(--pf-vet-muted, #64748b);
}
.pharmacy-legal__body p {
  margin: 0;
}
.pro-label {
  display: block;
  margin-bottom: 0.35rem;
  font-size: 0.85rem;
  font-weight: 600;
}
.pro-hint {
  margin-top: 0.75rem;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.4rem;
}
</style>
