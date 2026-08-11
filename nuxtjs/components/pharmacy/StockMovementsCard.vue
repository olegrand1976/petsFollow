<template>
  <ProCard class="pro-mb-lg" data-testid="stock-movements">
    <div class="stock-movements__head">
      <h3 class="pro-mb-sm">{{ $t('pharmacy.stock.movementsTitle') }}</h3>
      <ProButton variant="secondary" test-id="stock-movements-refresh" :disabled="busy" @click="$emit('refresh')">
        {{ $t('pharmacy.stock.movementsRefresh') }}
      </ProButton>
    </div>
    <div v-if="!movements.length" class="pro-empty" data-testid="stock-movements-empty">
      {{ $t('pharmacy.stock.movementsEmpty') }}
    </div>
    <table v-else class="pro-table" data-testid="stock-movements-table">
      <thead>
        <tr>
          <th>{{ $t('pharmacy.stock.movementsColDate') }}</th>
          <th>{{ $t('pharmacy.stock.colMed') }}</th>
          <th>{{ $t('pharmacy.stock.movementsColLot') }}</th>
          <th>{{ $t('pharmacy.stock.movementsColReason') }}</th>
          <th>{{ $t('pharmacy.stock.movementsColDelta') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="m in movements" :key="m.id" :data-testid="`stock-movement-${m.id}`">
          <td>{{ formatDate(m.createdAt) }}</td>
          <td>
            <strong>{{ m.medicationName || '—' }}</strong>
            <div class="pro-hint">{{ m.medicationCnk }}</div>
          </td>
          <td>{{ m.lotNumber || '—' }}</td>
          <td>
            {{ reasonLabel(m.reason) }}
            <span v-if="m.reasonDetail" class="pro-hint"> · {{ detailLabel(m.reasonDetail) }}</span>
          </td>
          <td :class="m.delta < 0 ? 'stock-movements__neg' : 'stock-movements__pos'">
            {{ m.delta > 0 ? `+${m.delta}` : m.delta }}
          </td>
        </tr>
      </tbody>
    </table>
  </ProCard>
</template>

<script setup lang="ts">
import type { PharmacyMovementRow } from '~/composables/usePharmacyStockPage'
import { pharmacyMovementDetailKey, pharmacyMovementReasonKey } from '~/utils/pharmacy-stock'

defineProps<{
  movements: PharmacyMovementRow[]
  busy: boolean
}>()
defineEmits<{ refresh: [] }>()

const { t } = useI18n()

function reasonLabel(reason: string) {
  return t(pharmacyMovementReasonKey(reason))
}

function detailLabel(detail?: string) {
  if (!detail) return ''
  const key = pharmacyMovementDetailKey(detail)
  return key ? t(key) : detail
}

function formatDate(raw: string) {
  if (!raw) return '—'
  const d = new Date(raw)
  if (Number.isNaN(d.getTime())) return raw
  return d.toLocaleString()
}
</script>

<style scoped>
.stock-movements__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}
.stock-movements__neg { color: var(--pf-vet-alert); font-weight: 600; }
.stock-movements__pos { color: var(--pf-vet-accent); font-weight: 600; }
</style>
