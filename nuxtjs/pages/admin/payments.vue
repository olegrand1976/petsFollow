<template>
  <div>
    <ProPageHeader title="Paiements reçus" subtitle="Historique des transactions Stripe." />
    <ProCard>
      <div class="pro-list-toolbar__filters pro-field-inline-wrap">
        <div class="pro-field pro-field-inline">
          <label class="pro-label" for="status-filter">Statut</label>
          <select id="status-filter" v-model="statusFilter" class="pro-select">
            <option value="">Tous</option>
            <option value="active">Actif</option>
            <option value="pending">En attente</option>
            <option value="past_due">Impayé</option>
          </select>
        </div>
      </div>
      <ProTable :empty="!payments.length" empty-title="Aucun paiement">
        <thead>
          <tr>
            <th>Date</th>
            <th>Client</th>
            <th>Animal</th>
            <th>Plan</th>
            <th>Mode</th>
            <th>Montant</th>
            <th>Statut</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in payments" :key="p.id">
            <td>{{ p.createdAt?.substring(0, 10) }}</td>
            <td>{{ p.clientEmail }}</td>
            <td>{{ p.petName }}</td>
            <td>{{ p.planCode }}</td>
            <td>{{ p.billingMode }}</td>
            <td>{{ formatEur(p.amountCents) }}</td>
            <td><ProBadge :variant="statusVariant(p.status)">{{ p.status }}</ProBadge></td>
          </tr>
        </tbody>
      </ProTable>
      <div class="pro-pagination">
        <ProButton variant="secondary" :disabled="page <= 1" @click="page--">Précédent</ProButton>
        <span class="text-muted">Page {{ page }}</span>
        <ProButton variant="secondary" :disabled="!hasMore" @click="page++">Suivant</ProButton>
      </div>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const payments = ref<any[]>([])
const statusFilter = ref('')
const page = ref(1)
const hasMore = ref(false)

function formatEur(cents: number) {
  return `${(cents / 100).toFixed(2)} €`
}

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
