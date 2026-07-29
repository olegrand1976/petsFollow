<template>
  <ProCard class="pro-mb-lg" data-testid="stock-inventory">
    <h3 class="pro-mb-sm">{{ $t('pharmacy.stock.invTitle') }}</h3>
    <p class="pro-hint pro-mb-sm">{{ $t('pharmacy.stock.invHint') }}</p>
    <div v-if="!session || session.status !== 'open'" class="stock-form__actions pro-mb-sm">
      <ProButton variant="secondary" test-id="stock-inv-start" :disabled="busy" @click="$emit('start')">
        {{ $t('pharmacy.stock.invStart') }}
      </ProButton>
    </div>
    <template v-if="session?.status === 'open'">
      <p class="pro-hint pro-mb-sm" data-testid="stock-inv-meta">
        {{ $t('pharmacy.stock.invProgress', { counted: session.countedCount ?? 0, total: session.lineCount ?? 0 }) }}
      </p>
      <div class="stock-table-wrap">
        <table class="pro-table" data-testid="stock-inv-table">
          <thead>
            <tr>
              <th>{{ $t('pharmacy.stock.colMed') }}</th>
              <th>{{ $t('pharmacy.stock.lot') }}</th>
              <th>{{ $t('pharmacy.stock.invSystem') }}</th>
              <th>{{ $t('pharmacy.stock.invCounted') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="ln in (session.lines || [])" :key="ln.id">
              <td>
                <strong>{{ ln.medicationName }}</strong>
                <div class="pro-hint">{{ ln.medicationCnk }}</div>
              </td>
              <td>{{ ln.lotNumber }}</td>
              <td>{{ ln.systemQty }}</td>
              <td>
                <input
                  :value="counts[ln.id] ?? ln.countedQty ?? ln.systemQty"
                  type="number"
                  min="0"
                  step="any"
                  class="pro-input"
                  :data-testid="`stock-inv-count-${ln.id}`"
                  @change="$emit('count', ln.id, $event)"
                >
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="stock-form__actions">
        <ProButton variant="primary" test-id="stock-inv-close" :disabled="busy" @click="$emit('close')">
          {{ $t('pharmacy.stock.invClose') }}
        </ProButton>
        <ProButton variant="ghost" test-id="stock-inv-cancel" :disabled="busy" @click="$emit('cancel')">
          {{ $t('pharmacy.stock.invCancel') }}
        </ProButton>
        <ProButton variant="secondary" test-id="stock-inv-export" :disabled="busy" @click="$emit('export')">
          {{ $t('pharmacy.stock.invExport') }}
        </ProButton>
      </div>
    </template>
    <p v-if="msg" class="pro-hint" data-testid="stock-inv-msg">{{ msg }}</p>
  </ProCard>
</template>

<script setup lang="ts">
import type { PharmacyInvSession } from '~/composables/usePharmacyStockPage'

defineProps<{
  session: PharmacyInvSession | null
  counts: Record<string, number>
  msg: string
  busy: boolean
}>()
defineEmits<{
  start: []
  close: []
  cancel: []
  export: []
  count: [lineId: string, event: Event]
}>()
</script>
