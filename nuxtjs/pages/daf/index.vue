<template>
  <div data-testid="daf-page">
    <ProPageHeader
      :title="$t('pharmacy.daf.title')"
      :subtitle="$t('pharmacy.daf.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="daf-page-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
        <ProButton
          v-if="canWritePharmacy"
          variant="primary"
          test-id="daf-new"
          @click="navigateTo('/daf/nouveau')"
        >
          {{ $t('pharmacy.daf.new') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <details class="pharmacy-legal">
      <summary>{{ $t('pharmacy.legal.title') }}</summary>
      <div class="pharmacy-legal__body">
        <p>{{ $t('pharmacy.legal.daf') }}</p>
        <p>{{ $t('pharmacy.legal.dafFinalize') }}</p>
        <p>{{ $t('pharmacy.legal.fefo') }}</p>
      </div>
    </details>

    <p v-if="error" class="pro-alert">{{ error }}</p>

    <ProCard>
      <div v-if="!items.length" class="pro-empty" data-testid="daf-empty">{{ $t('pharmacy.daf.empty') }}</div>
      <table v-else class="pro-table" data-testid="daf-table">
        <thead>
          <tr>
            <th>{{ $t('pharmacy.daf.colNumber') }}</th>
            <th>{{ $t('pharmacy.daf.colStatus') }}</th>
            <th>{{ $t('pharmacy.daf.colDate') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in items"
            :key="row.id"
            class="daf-row"
            :data-testid="`daf-row-${row.id}`"
            @click="navigateTo(`/daf/${row.id}`)"
          >
            <td>{{ row.displayNumber || '—' }}</td>
            <td><ProBadge :variant="statusVariant(row.status)">{{ statusLabel(row.status) }}</ProBadge></td>
            <td>{{ row.createdAt?.slice?.(0, 16) || row.createdAt }}</td>
          </tr>
        </tbody>
      </table>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: ['auth', 'vet-only', 'practice-perm'], practicePerm: 'pharmacy.read' })
const { t } = useI18n()
const { canPractice } = usePracticePerms()
const canWritePharmacy = computed(() => canPractice('pharmacy.write'))
const error = ref('')
const items = ref<any[]>([])

function unwrap(res: any) {
  return res?.data ?? res
}

function statusVariant(s: string) {
  if (s === 'finalized') return 'success' as const
  if (s === 'cancelled') return 'neutral' as const
  return 'warning' as const
}

function statusLabel(s: string) {
  if (s === 'draft') return t('pharmacy.daf.statusDraft')
  if (s === 'finalized') return t('pharmacy.daf.statusFinalized')
  if (s === 'cancelled') return t('pharmacy.daf.statusCancelled')
  return s
}

onMounted(async () => {
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/daf')
    items.value = unwrap(res)?.items ?? []
  }
  catch (e: any) {
    error.value = e?.data?.error?.message || t('pharmacy.daf.error')
  }
})
</script>

<style scoped>
.pharmacy-legal {
  margin: 0 0 1rem;
  padding: 0.65rem 0.85rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  background: var(--pf-vet-surface);
}
.pharmacy-legal__body { margin-top: 0.5rem; display: grid; gap: 0.35rem; }
.pharmacy-legal__body p { margin: 0; font-size: 0.9rem; color: var(--pf-vet-muted); }
.daf-row { cursor: pointer; }
</style>
