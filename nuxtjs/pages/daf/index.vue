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

    <ProListToolbar :show-view-toggle="false">
      <template #filters>
        <div class="pro-field pro-field-inline">
          <label class="pro-label" for="daf-status-filter">{{ $t('pharmacy.daf.colStatus') }}</label>
          <select
            id="daf-status-filter"
            v-model="statusFilter"
            class="pro-select"
            data-testid="daf-status-filter"
          >
            <option value="">{{ $t('consultations.statusAll') }}</option>
            <option value="draft">{{ $t('pharmacy.daf.statusDraft') }}</option>
            <option value="finalized">{{ $t('pharmacy.daf.statusFinalized') }}</option>
            <option value="cancelled">{{ $t('pharmacy.daf.statusCancelled') }}</option>
          </select>
        </div>
      </template>
    </ProListToolbar>

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
            <td>
              {{ row.displayNumber || '—' }}
              <ProBadge
                v-if="row.status === 'draft' && isDraftStale(row.createdAt)"
                variant="warning"
                class="daf-age-badge"
                data-testid="daf-draft-age-hint"
              >
                {{ draftAgeHint(row.createdAt) }}
              </ProBadge>
            </td>
            <td><ProBadge :variant="statusVariant(row.status)">{{ statusLabel(row.status) }}</ProBadge></td>
            <td>{{ row.createdAt?.slice?.(0, 16) || row.createdAt }}</td>
          </tr>
        </tbody>
      </table>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'pharmacy.read' })
const { t } = useI18n()
function pharmacyErr(e: any, fallbackKey: string): string {
  return pharmacyErrorMessage(t, e, fallbackKey)
}
const { canPractice } = usePracticePerms()
const canWritePharmacy = computed(() => canPractice('pharmacy.write'))
const error = ref('')
const items = ref<any[]>([])
const statusFilter = ref('')

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

function isDraftStale(createdAt?: string): boolean {
  if (!createdAt) return false
  const ageMs = Date.now() - new Date(createdAt).getTime()
  return ageMs >= 60 * 60 * 1000
}

function draftAgeHint(createdAt?: string): string {
  if (!createdAt) return ''
  const hours = Math.floor((Date.now() - new Date(createdAt).getTime()) / (60 * 60 * 1000))
  if (hours < 1) return t('clients.consultation.draftsOrphanBadge')
  return `${t('clients.consultation.draftsOrphanBadge')} (${hours}h)`
}

async function load() {
  error.value = ''
  try {
    const query: Record<string, string> = {}
    if (statusFilter.value) query.status = statusFilter.value
    const res = await $fetch<any>('/api/vet/pharmacy/daf', { query })
    items.value = unwrap(res)?.items ?? []
  }
  catch (e: any) {
    error.value = pharmacyErr(e, 'pharmacy.daf.error')
    items.value = []
  }
}

watch(statusFilter, () => { void load() })

onMounted(() => { void load() })
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
.daf-age-badge {
  margin-left: 0.35rem;
  vertical-align: middle;
}
</style>
