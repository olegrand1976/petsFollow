<template>
  <div>
    <ProPageHeader :title="$t('admin.users.title')" :subtitle="$t('admin.users.subtitle')" />
    <ProCard>
      <ProListToolbar v-model:view-mode="viewMode">
        <template #filters>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="role-filter">{{ $t('admin.users.role') }}</label>
            <select id="role-filter" v-model="roleFilter" class="pro-select">
              <option value="">{{ $t('admin.users.roleAll') }}</option>
              <option value="client">{{ $t('admin.users.roleClient') }}</option>
              <option value="vet">{{ $t('admin.users.roleVet') }}</option>
              <option value="admin">{{ $t('admin.users.roleAdmin') }}</option>
            </select>
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="payment-filter">{{ $t('admin.users.payment') }}</label>
            <select id="payment-filter" v-model="paymentFilter" class="pro-select">
              <option value="all">{{ $t('admin.users.paymentAll') }}</option>
              <option value="active">{{ $t('admin.users.paymentActive') }}</option>
              <option value="pending">{{ $t('admin.users.paymentPending') }}</option>
              <option value="past">{{ $t('admin.users.paymentPast') }}</option>
            </select>
          </div>
        </template>
      </ProListToolbar>

      <ProTable v-if="viewMode === 'table'" :empty="!filtered.length" :empty-title="$t('admin.users.emptyTitle')">
        <thead>
          <tr>
            <th>{{ $t('admin.users.columnEmail') }}</th>
            <th>{{ $t('admin.users.columnName') }}</th>
            <th>{{ $t('admin.users.columnRole') }}</th>
            <th>{{ $t('admin.users.columnRegistered') }}</th>
            <th>{{ $t('admin.users.columnPets') }}</th>
            <th>{{ $t('admin.users.columnPayment') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in filtered" :key="u.id">
            <td>{{ u.email }}</td>
            <td>{{ u.fullName }}</td>
            <td><ProBadge variant="neutral">{{ u.role }}</ProBadge></td>
            <td>{{ u.createdAt?.substring(0, 10) }}</td>
            <td>{{ u.petCount }}</td>
            <td><ProBadge :variant="paymentVariant(u.paymentLabel)">{{ u.paymentLabel }}</ProBadge></td>
          </tr>
        </tbody>
      </ProTable>

      <ProKanban v-else>
        <ProKanbanColumn
          v-for="col in kanbanColumns"
          :key="col.role"
          :title="col.title"
          :count="col.items.length"
          :empty="!col.items.length"
          :empty-title="$t('common.none')"
        >
          <article v-for="u in col.items" :key="u.id" class="pro-kanban-card pro-kanban-card--static">
            <strong>{{ u.fullName }}</strong>
            <p class="pro-kanban-card__meta">{{ u.email }}</p>
            <div class="pro-flex-gap">
              <ProBadge variant="neutral">{{ u.role }}</ProBadge>
              <ProBadge :variant="paymentVariant(u.paymentLabel)">{{ u.paymentLabel }}</ProBadge>
            </div>
          </article>
        </ProKanbanColumn>
      </ProKanban>

      <div v-if="viewMode === 'table'" class="pro-pagination">
        <ProButton variant="secondary" :disabled="page <= 1" @click="page--">{{ $t('common.previous') }}</ProButton>
        <span class="text-muted">{{ $t('common.page', { page }) }}</span>
        <ProButton variant="secondary" :disabled="!hasMore" @click="page++">{{ $t('common.next') }}</ProButton>
      </div>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

type AdminUser = {
  id: string
  email: string
  fullName: string
  role: string
  createdAt?: string
  petCount: number
  paymentLabel: string
}

const { t } = useI18n()

const roleFilter = ref('')
const paymentFilter = ref<'all' | 'active' | 'pending' | 'past'>('all')
const users = ref<AdminUser[]>([])
const page = ref(1)
const hasMore = ref(false)
const { viewMode } = useListView('pf-admin-users-view', 'table')

function paymentVariant(label: string): 'success' | 'warning' | 'danger' | 'neutral' {
  const l = (label || '').toLowerCase()
  if (l.includes('actif') || l.includes('payé') || l.includes('active') || l.includes('paid')) return 'success'
  if (l.includes('attente') || l.includes('pending')) return 'warning'
  if (l.includes('impayé') || l.includes('past')) return 'danger'
  return 'neutral'
}

function matchesPayment(label: string) {
  const l = (label || '').toLowerCase()
  if (paymentFilter.value === 'all') return true
  if (paymentFilter.value === 'active') return l.includes('actif') || l.includes('payé') || l.includes('active') || l.includes('paid')
  if (paymentFilter.value === 'pending') return l.includes('attente') || l.includes('pending')
  if (paymentFilter.value === 'past') return l.includes('impayé') || l.includes('past')
  return true
}

const filtered = computed(() =>
  users.value.filter((u) => matchesPayment(u.paymentLabel)),
)

const kanbanColumns = computed(() => {
  const roles = [
    { role: 'client', title: t('admin.users.roleClient') },
    { role: 'vet', title: t('admin.users.roleVet') },
    { role: 'admin', title: t('admin.users.roleAdmin') },
  ]
  return roles.map((r) => ({
    ...r,
    items: filtered.value.filter((u) => u.role === r.role),
  }))
})

async function load() {
  const res: any = await $fetch('/api/admin/users', {
    query: { role: roleFilter.value || undefined, page: page.value },
  })
  const rows = res.data ?? res ?? []
  users.value = rows
  hasMore.value = rows.length >= 50
}

watch([roleFilter, page], load)
watch(paymentFilter, () => { /* client-side filter */ })
onMounted(load)
</script>

<style scoped>
.pro-kanban-card--static {
  cursor: default;
}
</style>
