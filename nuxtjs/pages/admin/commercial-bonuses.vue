<template>
  <div data-testid="admin-commercial-bonuses-page">
    <ProPageHeader
      :title="$t('admin.commercialBonuses.title')"
      :subtitle="$t('admin.commercialBonuses.subtitle')"
    />

    <ProCard>
      <div class="pro-list-toolbar__filters pro-field-inline-wrap">
        <div class="pro-field pro-field-inline">
          <label class="pro-label" for="bonus-status">{{ $t('admin.commercialBonuses.filterStatus') }}</label>
          <select id="bonus-status" v-model="statusFilter" class="pro-input" data-testid="bonus-filter-status" @change="load">
            <option value="">{{ $t('admin.commercialBonuses.filterAll') }}</option>
            <option value="in_progress">{{ $t('commissionSheet.status.in_progress') }}</option>
            <option value="earned">{{ $t('commissionSheet.status.earned') }}</option>
            <option value="paid">{{ $t('commissionSheet.status.paid') }}</option>
          </select>
        </div>
        <div class="pro-field pro-field-inline">
          <label class="pro-label" for="bonus-commercial">{{ $t('admin.commercialBonuses.filterCommercial') }}</label>
          <select id="bonus-commercial" v-model="commercialFilter" class="pro-input" data-testid="bonus-filter-commercial" @change="load">
            <option value="">{{ $t('admin.commercialBonuses.filterAll') }}</option>
            <option v-for="c in commercials" :key="c.userId" :value="c.userId">
              {{ c.fullName }} ({{ c.email }})
            </option>
          </select>
        </div>
      </div>

      <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>

      <ProTable :empty="!items.length" :empty-title="$t('admin.commercialBonuses.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.commercialBonuses.colCommercial') }}</th>
            <th>{{ $t('admin.commercialBonuses.colBonus') }}</th>
            <th>{{ $t('admin.commercialBonuses.colProgress') }}</th>
            <th>{{ $t('admin.commercialBonuses.colDetail') }}</th>
            <th>{{ $t('admin.commercialBonuses.colAmount') }}</th>
            <th>{{ $t('admin.commercialBonuses.colStatus') }}</th>
            <th>{{ $t('admin.commercialBonuses.colActions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, idx) in items" :key="row.awardId || `${row.commercialUserId}-${row.bonusCode}-${row.vetUserId || row.periodYm}-${idx}`">
            <td>
              <div>{{ row.commercialFullName }}</div>
              <div class="text-muted">{{ row.commercialEmail }}</div>
            </td>
            <td>{{ $t(`commissionSheet.bonusTitles.${row.bonusCode}`) }}</td>
            <td>{{ row.progress }}/{{ row.target }}{{ row.bonusCode === 'commercial_mix' ? ' %' : '' }}</td>
            <td>
              <template v-if="row.bonusCode === 'commercial_ramp'">
                {{ row.vetFullName || row.vetEmail || '—' }}
              </template>
              <template v-else>
                {{ row.periodYm || '—' }}
              </template>
            </td>
            <td>{{ formatCurrency(row.amountCents) }}</td>
            <td>
              <ProBadge :variant="statusVariant(row.status)">
                {{ $t(`commissionSheet.status.${row.status}`) }}
              </ProBadge>
            </td>
            <td>
              <ProButton
                v-if="row.status === 'earned' && row.awardId"
                variant="secondary"
                :loading="payingId === row.awardId"
                :test-id="`bonus-mark-paid-${row.awardId}`"
                @click="markPaid(row)"
              >
                {{ $t('admin.commercialBonuses.markPaid') }}
              </ProButton>
              <span v-else class="text-muted">—</span>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t } = useI18n()
const { formatCurrency } = useFormatters()
const { mapError } = useApiError()

const items = ref<any[]>([])
const commercials = ref<any[]>([])
const statusFilter = ref('')
const commercialFilter = ref('')
const error = ref('')
const payingId = ref('')

function statusVariant(status: string): 'success' | 'warning' | 'neutral' {
  if (status === 'paid' || status === 'earned') return 'success'
  if (status === 'in_progress') return 'warning'
  return 'neutral'
}

async function load() {
  error.value = ''
  try {
    const query: Record<string, string> = {}
    if (statusFilter.value) query.status = statusFilter.value
    if (commercialFilter.value) query.commercialId = commercialFilter.value
    const res: any = await $fetch('/api/admin/commercial-bonuses', { query })
    const data = res.data ?? res
    items.value = data.items ?? []
  } catch (e: any) {
    error.value = mapError(e)
  }
}

async function markPaid(row: any) {
  if (!row.awardId) return
  if (!confirm(t('admin.commercialBonuses.confirmMarkPaid'))) return
  payingId.value = row.awardId
  error.value = ''
  try {
    await $fetch(`/api/admin/commercial-bonuses/${row.awardId}/mark-paid`, { method: 'POST' })
    await load()
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    payingId.value = ''
  }
}

onMounted(async () => {
  const commercialsRes: any = await $fetch('/api/admin/commercials').catch(() => null)
  commercials.value = commercialsRes?.data ?? commercialsRes ?? []
  await load()
})
</script>

<style scoped>
.pro-field-inline-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin-bottom: 1rem;
}
.text-muted {
  color: var(--pf-vet-muted, #6b7280);
  font-size: 0.875rem;
}
</style>
