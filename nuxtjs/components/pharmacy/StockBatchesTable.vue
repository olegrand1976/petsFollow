<template>
  <ProCard data-testid="stock-batches">
    <div class="stock-toolbar">
      <input
        :value="query"
        class="pro-input"
        :placeholder="$t('pharmacy.stock.searchPlaceholder')"
        data-testid="stock-search"
        @input="$emit('update:query', ($event.target as HTMLInputElement).value)"
        @keyup.enter="$emit('search')"
      >
      <ProButton variant="secondary" test-id="stock-refresh" :disabled="busy" @click="$emit('refresh')">
        {{ $t('pharmacy.stock.refresh') }}
      </ProButton>
    </div>
    <div v-if="!batches.length" class="pro-empty" data-testid="stock-empty">
      {{ $t('pharmacy.stock.empty') }}
    </div>
    <table v-else class="pro-table" data-testid="stock-table">
      <thead>
        <tr>
          <th>{{ $t('pharmacy.stock.colMed') }}</th>
          <th>{{ $t('pharmacy.stock.lot') }}</th>
          <th>{{ $t('pharmacy.stock.expiresOn') }}</th>
          <th>{{ $t('pharmacy.stock.qty') }}</th>
          <th>{{ $t('pharmacy.stock.band') }}</th>
          <th />
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in batches" :key="row.id" :data-testid="`stock-row-${row.id}`">
          <td>
            <strong>{{ row.medicationName }}</strong>
            <div class="pro-hint">{{ row.medicationCnk }} · {{ row.depositCode }}</div>
          </td>
          <td>{{ row.lotNumber }}</td>
          <td>{{ row.expiresOn }}</td>
          <td>{{ row.qtyOnHand }} {{ row.unit }}</td>
          <td>
            <ProBadge :variant="bandVariant(row.expiryBand)">{{ bandLabel(row.expiryBand) }}</ProBadge>
          </td>
          <td class="stock-row-actions">
            <ProButton
              v-if="canWrite && row.status === 'active'"
              variant="secondary"
              :test-id="`stock-quarantine-${row.id}`"
              :disabled="busy"
              @click="$emit('quarantine', row.id)"
            >
              {{ $t('pharmacy.stock.quarantine') }}
            </ProButton>
            <ProButton
              v-if="canWrite"
              variant="ghost"
              :test-id="`stock-waste-${row.id}`"
              :disabled="busy"
              @click="$emit('waste', row.id)"
            >
              {{ $t('pharmacy.stock.waste') }}
            </ProButton>
          </td>
        </tr>
      </tbody>
    </table>
  </ProCard>
</template>

<script setup lang="ts">
import type { PharmacyBatchRow } from '~/composables/usePharmacyStockPage'

defineProps<{
  batches: PharmacyBatchRow[]
  query: string
  busy: boolean
  canWrite: boolean
  bandVariant: (band: string) => 'success' | 'warning' | 'danger' | 'neutral'
  bandLabel: (band: string) => string
}>()
defineEmits<{
  'update:query': [value: string]
  search: []
  refresh: []
  quarantine: [id: string]
  waste: [id: string]
}>()
</script>
