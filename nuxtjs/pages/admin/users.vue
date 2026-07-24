<template>
  <div data-testid="admin-users-page">
    <ProPageHeader :title="$t('admin.users.title')" :subtitle="$t('admin.users.subtitle')" />
    <ProCard class="pro-mb-lg" data-testid="admin-create-commercial">
      <h3 class="pro-mb-md">{{ $t('admin.users.createCommercial') }}</h3>
      <form class="pro-form" @submit.prevent="createCommercial">
        <ProInput v-model="cForm.fullName" test-id="admin-commercial-name" :label="$t('admin.users.commercialName')" required />
        <ProInput v-model="cForm.email" test-id="admin-commercial-email" type="email" :label="$t('admin.users.commercialEmail')" required />
        <ProInput v-model="cForm.password" test-id="admin-commercial-password" type="password" :label="$t('admin.users.commercialPassword')" required />
        <div class="pro-field">
          <label class="pro-label" for="admin-commercial-role">{{ $t('admin.users.commercialRole') }}</label>
          <select id="admin-commercial-role" v-model="cForm.role" class="pro-select" data-testid="admin-commercial-role">
            <option value="commercial">{{ $t('admin.users.roleCommercial') }}</option>
            <option value="commercial_manager">{{ $t('admin.users.roleCommercialManager') }}</option>
          </select>
        </div>
        <div v-if="cForm.role === 'commercial'" class="pro-field">
          <label class="pro-label" for="admin-commercial-manager">{{ $t('admin.users.commercialManager') }}</label>
          <select id="admin-commercial-manager" v-model="cForm.managerUserId" class="pro-select" data-testid="admin-commercial-manager">
            <option value="">{{ $t('admin.users.commercialManagerNone') }}</option>
            <option v-for="m in managers" :key="m.userId" :value="m.userId">
              {{ m.fullName }} ({{ m.email }})
            </option>
          </select>
        </div>
        <p v-if="cMsg" class="pro-hint" data-testid="admin-commercial-msg">{{ cMsg }}</p>
        <ProButton type="submit" test-id="admin-commercial-submit" :disabled="cSaving">{{ $t('admin.users.createCommercial') }}</ProButton>
      </form>
    </ProCard>
    <ProCard class="pro-mb-lg" data-testid="admin-create-vet">
      <h3 class="pro-mb-md">{{ $t('admin.users.createVet') }}</h3>
      <form class="pro-form" @submit.prevent="createVet">
        <ProInput v-model="vForm.fullName" test-id="admin-vet-name" :label="$t('admin.users.vetName')" required />
        <ProInput v-model="vForm.practiceName" test-id="admin-vet-practice" :label="$t('admin.users.vetPractice')" required />
        <ProInput v-model="vForm.email" test-id="admin-vet-email" type="email" :label="$t('admin.users.vetEmail')" required />
        <ProInput v-model="vForm.password" test-id="admin-vet-password" type="password" :label="$t('admin.users.vetPassword')" required />
        <p v-if="vMsg" class="pro-hint" data-testid="admin-vet-msg">{{ vMsg }}</p>
        <ProButton type="submit" test-id="admin-vet-submit" :disabled="vSaving">{{ $t('admin.users.createVet') }}</ProButton>
      </form>
    </ProCard>
    <ProCard class="pro-mb-lg" data-testid="admin-create-client">
      <h3 class="pro-mb-md">{{ $t('admin.users.createClient') }}</h3>
      <form class="pro-form" @submit.prevent="createClient">
        <div class="pro-field">
          <label class="pro-label" for="admin-client-vet">{{ $t('admin.users.clientVet') }}</label>
          <select id="admin-client-vet" v-model="clForm.vetUserId" class="pro-select" required data-testid="admin-client-vet">
            <option value="">{{ $t('admin.users.clientVetPlaceholder') }}</option>
            <option v-for="v in vetOptions" :key="v.userId" :value="v.userId">
              {{ v.fullName }} — {{ v.practiceName }}
            </option>
          </select>
        </div>
        <ProInput v-model="clForm.fullName" test-id="admin-client-name" :label="$t('admin.users.clientName')" required />
        <ProInput v-model="clForm.email" test-id="admin-client-email" type="email" :label="$t('admin.users.clientEmail')" required />
        <ProInput v-model="clForm.password" test-id="admin-client-password" type="password" :label="$t('admin.users.clientPassword')" required />
        <p v-if="clMsg" class="pro-hint" data-testid="admin-client-msg">{{ clMsg }}</p>
        <ProButton type="submit" test-id="admin-client-submit" :disabled="clSaving">{{ $t('admin.users.createClient') }}</ProButton>
      </form>
    </ProCard>
    <ProCard class="pro-mb-lg" data-testid="admin-create-care-pro">
      <h3 class="pro-mb-md">{{ $t('admin.users.createCarePro') }}</h3>
      <p class="pro-hint pro-mb-md">{{ $t('admin.users.createCareProHint') }}</p>
      <form class="pro-form" @submit.prevent="createCarePro">
        <ProInput v-model="cpForm.fullName" test-id="admin-care-pro-name" :label="$t('admin.users.careProName')" required />
        <ProInput v-model="cpForm.email" test-id="admin-care-pro-email" type="email" :label="$t('admin.users.careProEmail')" required />
        <ProInput v-model="cpForm.password" test-id="admin-care-pro-password" type="password" :label="$t('admin.users.careProPassword')" required />
        <div class="pro-field">
          <label class="pro-label" for="admin-care-pro-specialty">{{ $t('admin.users.careProSpecialty') }}</label>
          <select id="admin-care-pro-specialty" v-model="cpForm.specialty" class="pro-select" required data-testid="admin-care-pro-specialty">
            <option v-for="s in careProSpecialties" :key="s" :value="s">
              {{ $t(`admin.users.specialty.${s}`) }}
            </option>
          </select>
        </div>
        <p v-if="cpMsg" class="pro-hint" data-testid="admin-care-pro-msg">{{ cpMsg }}</p>
        <ProButton type="submit" test-id="admin-care-pro-submit" :disabled="cpSaving">
          {{ $t('admin.users.createCarePro') }}
        </ProButton>
      </form>
    </ProCard>
    <ProCard>
      <ProListToolbar v-model:view-mode="viewMode">
        <template #filters>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="role-filter">{{ $t('admin.users.role') }}</label>
            <select id="role-filter" v-model="roleFilter" class="pro-select" data-testid="admin-role-filter">
              <option value="">{{ $t('admin.users.roleAll') }}</option>
              <option value="client">{{ $t('admin.users.roleClient') }}</option>
              <option value="vet">{{ $t('admin.users.roleVet') }}</option>
              <option value="care_pro">{{ $t('admin.users.roleCarePro') }}</option>
              <option value="commercial">{{ $t('admin.users.roleCommercial') }}</option>
              <option value="commercial_manager">{{ $t('admin.users.roleCommercialManager') }}</option>
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
            <td><ProBadge variant="neutral">{{ roleLabel(u.role) }}</ProBadge></td>
            <td>{{ u.createdAt?.substring(0, 10) }}</td>
            <td>{{ u.petCount }}</td>
            <td><ProBadge :variant="paymentVariant(u.paymentLabel)">{{ paymentLabel(u.paymentLabel) }}</ProBadge></td>
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
              <ProBadge variant="neutral">{{ roleLabel(u.role) }}</ProBadge>
              <ProBadge :variant="paymentVariant(u.paymentLabel)">{{ paymentLabel(u.paymentLabel) }}</ProBadge>
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
const { roleLabel, paymentLabel } = useCodeLabels()

const roleFilter = ref('')
const paymentFilter = ref<'all' | 'active' | 'pending' | 'past'>('all')
const users = ref<AdminUser[]>([])
const page = ref(1)
const hasMore = ref(false)
const { viewMode } = useListView('pf-admin-users-view', 'table')
const cSaving = ref(false)
const cMsg = ref('')
const cForm = reactive({ fullName: '', email: '', password: '', role: 'commercial', managerUserId: '' })
const managers = ref<{ userId: string; fullName: string; email: string }[]>([])
const vSaving = ref(false)
const vMsg = ref('')
const vForm = reactive({ fullName: '', practiceName: '', email: '', password: '' })
const clSaving = ref(false)
const clMsg = ref('')
const clForm = reactive({ vetUserId: '', fullName: '', email: '', password: '' })
const vetOptions = ref<{ userId: string; fullName: string; practiceName: string }[]>([])
const cpSaving = ref(false)
const cpMsg = ref('')
const cpForm = reactive({ fullName: '', email: '', password: '', specialty: 'vet_light' })
const careProSpecialties = [
  'vet_light',
  'farrier',
  'physio',
  'behaviorist',
  'groomer',
  'breeder',
] as const

function paymentVariant(label: string): 'success' | 'warning' | 'danger' | 'neutral' {
  const l = (label || '').toLowerCase()
  if (l === 'active' || l === 'paid' || l === 'succeeded') return 'success'
  if (l === 'pending' || l === 'processing') return 'warning'
  if (l === 'past_due' || l === 'failed') return 'danger'
  return 'neutral'
}

function matchesPayment(label: string) {
  const l = (label || '').toLowerCase()
  if (paymentFilter.value === 'all') return true
  if (paymentFilter.value === 'active') return l === 'active'
  if (paymentFilter.value === 'pending') return l === 'pending'
  if (paymentFilter.value === 'past') return l === 'none' || l === 'past_due'
  return true
}

const filtered = computed(() =>
  users.value.filter((u) => matchesPayment(u.paymentLabel)),
)

const kanbanColumns = computed(() => {
  const roles = [
    { role: 'client', title: t('admin.users.roleClient') },
    { role: 'vet', title: t('admin.users.roleVet') },
    { role: 'care_pro', title: t('admin.users.roleCarePro') },
    { role: 'commercial', title: t('admin.users.roleCommercial') },
    { role: 'commercial_manager', title: t('admin.users.roleCommercialManager') },
    { role: 'admin', title: t('admin.users.roleAdmin') },
  ]
  return roles.map((r) => ({
    ...r,
    items: filtered.value.filter((u) => u.role === r.role),
  }))
})

async function createCommercial() {
  cSaving.value = true
  cMsg.value = ''
  try {
    const body: Record<string, string> = {
      fullName: cForm.fullName,
      email: cForm.email,
      password: cForm.password,
      role: cForm.role,
    }
    if (cForm.role === 'commercial' && cForm.managerUserId) {
      body.managerUserId = cForm.managerUserId
    }
    await $fetch('/api/admin/commercials', { method: 'POST', body })
    cMsg.value = t('admin.users.commercialCreated')
    Object.assign(cForm, { fullName: '', email: '', password: '', role: 'commercial', managerUserId: '' })
    await Promise.all([load(), loadManagers()])
  } catch {
    cMsg.value = t('admin.users.commercialFailed')
  } finally {
    cSaving.value = false
  }
}

async function createVet() {
  vSaving.value = true
  vMsg.value = ''
  try {
    await $fetch('/api/admin/vets', { method: 'POST', body: { ...vForm } })
    vMsg.value = t('admin.users.vetCreated')
    Object.assign(vForm, { fullName: '', practiceName: '', email: '', password: '' })
    await Promise.all([load(), loadVets()])
  } catch {
    vMsg.value = t('admin.users.vetFailed')
  } finally {
    vSaving.value = false
  }
}

async function createClient() {
  clSaving.value = true
  clMsg.value = ''
  try {
    await $fetch('/api/admin/clients', { method: 'POST', body: { ...clForm } })
    clMsg.value = t('admin.users.clientCreated')
    Object.assign(clForm, { vetUserId: '', fullName: '', email: '', password: '' })
    await load()
  } catch {
    clMsg.value = t('admin.users.clientFailed')
  } finally {
    clSaving.value = false
  }
}

async function createCarePro() {
  cpSaving.value = true
  cpMsg.value = ''
  try {
    await $fetch('/api/admin/care-pros', { method: 'POST', body: { ...cpForm } })
    cpMsg.value = t('admin.users.careProCreated')
    Object.assign(cpForm, { fullName: '', email: '', password: '', specialty: 'vet_light' })
    await load()
  } catch {
    cpMsg.value = t('admin.users.careProFailed')
  } finally {
    cpSaving.value = false
  }
}

async function loadVets() {
  const res: any = await $fetch('/api/admin/vets')
  vetOptions.value = res.data ?? res ?? []
}

async function loadManagers() {
  const res: any = await $fetch('/api/admin/commercial-managers').catch(() => null)
  managers.value = res?.data ?? res ?? []
}

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
onMounted(async () => {
  await Promise.all([load(), loadVets(), loadManagers()])
})
</script>

<style scoped>
.pro-kanban-card--static {
  cursor: default;
}
</style>
