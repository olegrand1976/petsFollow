<template>
  <div>
    <ProPageHeader :title="$t('admin.payments.title')" :subtitle="$t('admin.payments.subtitle')" />
    <ProCard>
      <div class="pro-list-toolbar__filters pro-field-inline-wrap">
        <div class="pro-field pro-field-inline">
          <label class="pro-label" for="status-filter">{{ $t('admin.payments.status') }}</label>
          <select id="status-filter" v-model="statusFilter" class="pro-select">
            <option value="">{{ $t('admin.payments.statusAll') }}</option>
            <option value="active">{{ $t('admin.payments.statusActive') }}</option>
            <option value="pending">{{ $t('admin.payments.statusPending') }}</option>
            <option value="past_due">{{ $t('admin.payments.statusPastDue') }}</option>
          </select>
        </div>
      </div>
      <ProTable :empty="!payments.length" :empty-title="$t('admin.payments.emptyTitle')">
        <thead>
          <tr>
            <th>{{ $t('admin.payments.columnDate') }}</th>
            <th>{{ $t('admin.payments.columnClient') }}</th>
            <th>{{ $t('admin.payments.columnPet') }}</th>
            <th>{{ $t('admin.payments.columnPlan') }}</th>
            <th>{{ $t('admin.payments.columnMode') }}</th>
            <th>{{ $t('admin.payments.columnAmount') }}</th>
            <th>{{ $t('admin.payments.columnStatus') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in payments" :key="p.id">
            <td>{{ p.createdAt?.substring(0, 10) }}</td>
            <td>{{ p.clientEmail }}</td>
            <td>{{ p.petName }}</td>
            <td>{{ planLabel(p.planCode) }}</td>
            <td>{{ billingModeLabel(p.billingMode) }}</td>
            <td>{{ formatCurrency(p.amountCents) }}</td>
            <td><ProBadge :variant="statusVariant(p.status)">{{ paymentLabel(p.status) }}</ProBadge></td>
          </tr>
        </tbody>
      </ProTable>
      <div class="pro-pagination">
        <ProButton variant="secondary" :disabled="page <= 1" @click="page--">{{ $t('common.previous') }}</ProButton>
        <span class="text-muted">{{ $t('common.page', { page }) }}</span>
        <ProButton variant="secondary" :disabled="!hasMore" @click="page++">{{ $t('common.next') }}</ProButton>
      </div>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { formatCurrency } = useFormatters()
const { planLabel, billingModeLabel, paymentLabel } = useCodeLabels()

const payments = ref<any[]>([])
const statusFilter = ref('')
const page = ref(1)
const hasMore = ref(false)

function statusVariant(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
  const s = (status || '').toLowerCase()
  if (s === 'paid' || s === 'succeeded' || s === 'active') return 'success'
  if (s === 'pending' || s === 'processing') return 'warning'
  if (s === 'failed' || s === 'past_due') return 'danger'
  return 'neutral'
}

async function load() {
  const res: any = await $fetch('/api/admin/payments', {
    query: {
      status: statusFilter.value || undefined,
      page: page.value,
    },
  })
  const rows = res.data ?? res ?? []
  payments.value = rows
  hasMore.value = rows.length >= 50
}

watch([statusFilter, page], load)
onMounted(load)
</script>
