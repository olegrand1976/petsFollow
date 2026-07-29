<template>
  <div data-testid="admin-commercials-page">
    <ProPageHeader :title="$t('admin.commercials.title')" :subtitle="$t('admin.commercials.subtitle')" />

    <ProSectionTabs
      v-model="activeTab"
      :tabs="commercialTabs"
      :aria-label="$t('admin.commercials.tabsLabel')"
    />

    <ProCard
      v-show="activeTab === 'manager'"
      class="pro-mb-lg"
      role="tabpanel"
      aria-labelledby="tab-manager"
      data-testid="admin-assign-manager"
    >
      <h3 class="pro-mb-md">{{ $t('admin.commercials.assignManagerTitle') }}</h3>
      <form class="pro-form" @submit.prevent="assignManager">
        <div class="pro-grid-2">
          <div class="pro-field">
            <label class="pro-label" for="mgr-commercial">{{ $t('admin.commercials.assignCommercial') }}</label>
            <select
              id="mgr-commercial"
              v-model="selectedCommercialId"
              class="pro-select"
              required
              data-testid="admin-manager-commercial"
            >
              <option value="" disabled>{{ $t('admin.commercials.assignCommercialPlaceholder') }}</option>
              <option v-for="c in rows" :key="c.userId" :value="c.userId">
                {{ c.fullName }} ({{ c.email }})
              </option>
            </select>
          </div>
          <div class="pro-field">
            <label class="pro-label" for="mgr-manager">{{ $t('admin.commercials.assignManager') }}</label>
            <select id="mgr-manager" v-model="managerUserId" class="pro-select" data-testid="admin-manager-select">
              <option value="">{{ $t('admin.commercials.assignManagerNone') }}</option>
              <option v-for="m in managers" :key="m.userId" :value="m.userId">
                {{ m.fullName }} ({{ m.email }})
              </option>
            </select>
          </div>
        </div>
        <p v-if="managerMsg" class="pro-hint">{{ managerMsg }}</p>
        <p v-if="managerError" class="pro-error">{{ managerError }}</p>
        <ProButton type="submit" test-id="admin-manager-submit" :disabled="managerSaving || !selectedCommercialId">
          {{ $t('admin.commercials.assignManagerSubmit') }}
        </ProButton>
      </form>
    </ProCard>

    <ProCard
      v-show="activeTab === 'assign'"
      class="pro-mb-lg"
      role="tabpanel"
      aria-labelledby="tab-assign"
      data-testid="admin-assign-vet"
    >
      <h3 class="pro-mb-md">{{ $t('admin.commercials.assignTitle') }}</h3>
      <form class="pro-form" @submit.prevent="assignVet">
        <div class="pro-grid-2">
          <div class="pro-field">
            <label class="pro-label" for="assign-commercial">{{ $t('admin.commercials.assignCommercial') }}</label>
            <select
              id="assign-commercial"
              v-model="selectedCommercialId"
              class="pro-select"
              data-testid="admin-assign-commercial"
              required
            >
              <option value="" disabled>{{ $t('admin.commercials.assignCommercialPlaceholder') }}</option>
              <option v-for="c in rows" :key="c.userId" :value="c.userId">
                {{ c.fullName }} ({{ c.email }})
              </option>
            </select>
          </div>
          <div class="pro-field">
            <label class="pro-label" for="assign-vet">{{ $t('admin.commercials.assignVet') }}</label>
            <select
              id="assign-vet"
              v-model="vetUserId"
              class="pro-select"
              data-testid="admin-assign-vet-select"
              required
            >
              <option value="" disabled>{{ $t('admin.commercials.assignVetPlaceholder') }}</option>
              <option v-for="v in vets" :key="v.id" :value="v.id">
                {{ v.fullName }} ({{ v.email }})
              </option>
            </select>
          </div>
        </div>
        <p v-if="assignMsg" class="pro-hint" data-testid="admin-assign-msg">{{ assignMsg }}</p>
        <p v-if="assignError" class="pro-error">{{ assignError }}</p>
        <ProButton type="submit" test-id="admin-assign-submit" :disabled="assignSaving || !selectedCommercialId || !vetUserId">
          {{ $t('admin.commercials.assignSubmit') }}
        </ProButton>
      </form>
    </ProCard>

    <ProCard
      v-show="activeTab === 'base'"
      class="pro-mb-lg"
      role="tabpanel"
      aria-labelledby="tab-base"
      data-testid="admin-commercial-base-location"
    >
      <h3 class="pro-mb-md">{{ $t('admin.commercials.baseLocationTitle') }}</h3>
      <form class="pro-form" @submit.prevent="saveBaseLocation">
        <div class="pro-field">
          <label class="pro-label" for="base-commercial">{{ $t('admin.commercials.assignCommercial') }}</label>
          <select
            id="base-commercial"
            v-model="selectedCommercialId"
            class="pro-select"
            required
            data-testid="admin-base-commercial"
          >
            <option value="" disabled>{{ $t('admin.commercials.assignCommercialPlaceholder') }}</option>
            <option v-for="c in rows" :key="c.userId" :value="c.userId">
              {{ c.fullName }} ({{ c.email }})
            </option>
          </select>
        </div>
        <div class="pro-grid-2">
          <div class="pro-field">
            <label class="pro-label" for="base-city">{{ $t('admin.commercials.baseCity') }}</label>
            <input id="base-city" v-model="baseForm.city" class="pro-input" data-testid="admin-base-city">
          </div>
          <div class="pro-field">
            <label class="pro-label" for="base-postal">{{ $t('admin.commercials.basePostalCode') }}</label>
            <input id="base-postal" v-model="baseForm.postalCode" class="pro-input" data-testid="admin-base-postal">
          </div>
          <div class="pro-field">
            <label class="pro-label" for="base-lat">{{ $t('admin.commercials.baseLat') }}</label>
            <input id="base-lat" v-model="baseForm.lat" class="pro-input" inputmode="decimal" data-testid="admin-base-lat">
          </div>
          <div class="pro-field">
            <label class="pro-label" for="base-lng">{{ $t('admin.commercials.baseLng') }}</label>
            <input id="base-lng" v-model="baseForm.lng" class="pro-input" inputmode="decimal" data-testid="admin-base-lng">
          </div>
        </div>
        <p v-if="baseMsg" class="pro-hint">{{ baseMsg }}</p>
        <p v-if="baseError" class="pro-error">{{ baseError }}</p>
        <ProButton type="submit" test-id="admin-base-submit" :disabled="baseSaving || !selectedCommercialId">
          {{ $t('admin.commercials.baseLocationSubmit') }}
        </ProButton>
      </form>
    </ProCard>

    <ProCard data-testid="admin-commercials-list">
      <h3 class="pro-mb-md">{{ $t('admin.commercials.listTitle') }}</h3>
      <p v-if="loadError" class="pro-error">{{ loadError }}</p>
      <p v-else-if="loading" class="pro-hint">{{ $t('common.loading') }}</p>
      <ProTable v-else :empty="!rows.length" :empty-title="$t('admin.commercials.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.commercials.columnName') }}</th>
            <th>{{ $t('admin.commercials.columnEmail') }}</th>
            <th>{{ $t('admin.commercials.colManager') }}</th>
            <th>{{ $t('admin.commercials.colBranch') }}</th>
            <th>{{ $t('admin.commercials.columnVets') }}</th>
            <th>{{ $t('admin.commercials.columnDue') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in rows" :key="r.userId">
            <td>{{ r.fullName }}</td>
            <td>{{ r.email }}</td>
            <td>{{ r.managerName || $t('common.dash') }}</td>
            <td>{{ r.branchName || $t('common.dash') }}</td>
            <td>{{ r.clientCount }}</td>
            <td>{{ formatDue(r.userId) }}</td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-or-dev' })

type CommercialTab = 'manager' | 'assign' | 'base'

type CommercialRow = {
  userId: string
  fullName: string
  email: string
  clientCount: number
  managerUserId?: string
  managerName?: string
  branchName?: string
  baseCity?: string
  basePostalCode?: string
}

type ManagerOpt = { userId: string; fullName: string; email: string }
type VetOpt = { id: string; fullName: string; email: string }

const { t, locale } = useI18n()

const activeTab = ref<CommercialTab>('assign')
const selectedCommercialId = ref('')
const managerUserId = ref('')
const vetUserId = ref('')
const rows = ref<CommercialRow[]>([])
const managers = ref<ManagerOpt[]>([])
const vets = ref<VetOpt[]>([])
const dueByCommercial = ref<Record<string, number>>({})
const loading = ref(false)
const loadError = ref('')

const assignSaving = ref(false)
const assignMsg = ref('')
const assignError = ref('')
const managerSaving = ref(false)
const managerMsg = ref('')
const managerError = ref('')
const baseForm = reactive({ city: '', postalCode: '', lat: '', lng: '' })
const baseSaving = ref(false)
const baseMsg = ref('')
const baseError = ref('')

const commercialTabs = computed(() => [
  { id: 'manager', label: t('admin.commercials.tabManager') },
  { id: 'assign', label: t('admin.commercials.tabAssign') },
  { id: 'base', label: t('admin.commercials.tabBase') },
])

function formatDue(commercialId: string) {
  const cents = dueByCommercial.value[commercialId]
  if (cents == null) return t('common.dash')
  return new Intl.NumberFormat(locale.value, { style: 'currency', currency: 'EUR' }).format(cents / 100)
}

function applyCommercialPrefill(id: string, resetCoords = true) {
  const row = rows.value.find(r => r.userId === id)
  if (!row) {
    managerUserId.value = ''
    baseForm.city = ''
    baseForm.postalCode = ''
    if (resetCoords) {
      baseForm.lat = ''
      baseForm.lng = ''
    }
    return
  }
  managerUserId.value = row.managerUserId || ''
  baseForm.city = row.baseCity || ''
  baseForm.postalCode = row.basePostalCode || ''
  // lat/lng absents de la liste admin — ne pas les effacer au reload du même commercial
  if (resetCoords) {
    baseForm.lat = ''
    baseForm.lng = ''
  }
}

watch(selectedCommercialId, (id) => {
  applyCommercialPrefill(id, true)
  assignMsg.value = ''
  assignError.value = ''
  managerMsg.value = ''
  managerError.value = ''
  baseMsg.value = ''
  baseError.value = ''
})

async function loadCommissions(commercialId: string) {
  try {
    const res: any = await $fetch(`/api/admin/commercials/${commercialId}/commissions`)
    const data = res.data ?? res
    dueByCommercial.value[commercialId] = Number(data.lifetimeEarnedCents ?? 0)
  } catch {
    dueByCommercial.value[commercialId] = 0
  }
}

async function loadAllVets(): Promise<VetOpt[]> {
  const all: VetOpt[] = []
  let page = 1
  for (;;) {
    const res: any = await $fetch('/api/admin/users', { query: { role: 'vet', page } })
    const batch = res.data ?? res ?? []
    if (!Array.isArray(batch) || batch.length === 0) break
    for (const u of batch) {
      if (u?.id) all.push({ id: u.id, fullName: u.fullName ?? '', email: u.email ?? '' })
    }
    if (batch.length < 50) break
    page += 1
    if (page > 40) break
  }
  return all
}

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    const [commercialsRes, managersRes, vetList]: any[] = await Promise.all([
      $fetch('/api/admin/commercials'),
      $fetch('/api/admin/commercial-managers').catch(() => null),
      loadAllVets(),
    ])
    rows.value = commercialsRes.data ?? commercialsRes ?? []
    managers.value = managersRes?.data ?? managersRes ?? []
    vets.value = vetList
    if (selectedCommercialId.value) {
      applyCommercialPrefill(selectedCommercialId.value, false)
    }
    await Promise.all(rows.value.map(r => loadCommissions(r.userId)))
  } catch {
    loadError.value = t('admin.commercials.loadFailed')
  } finally {
    loading.value = false
  }
}

async function assignManager() {
  managerSaving.value = true
  managerMsg.value = ''
  managerError.value = ''
  try {
    await $fetch(`/api/admin/commercials/${selectedCommercialId.value}/manager`, {
      method: 'PATCH',
      body: { managerUserId: managerUserId.value },
    })
    managerMsg.value = t('admin.commercials.assignManagerSuccess')
    await load()
  } catch {
    managerError.value = t('admin.commercials.assignManagerFailed')
  } finally {
    managerSaving.value = false
  }
}

async function assignVet() {
  assignSaving.value = true
  assignMsg.value = ''
  assignError.value = ''
  try {
    await $fetch(`/api/admin/commercials/${selectedCommercialId.value}/assign`, {
      method: 'PATCH',
      body: { vetUserId: vetUserId.value },
    })
    assignMsg.value = t('admin.commercials.assignSuccess')
    vetUserId.value = ''
    await load()
  } catch {
    assignError.value = t('admin.commercials.assignFailed')
  } finally {
    assignSaving.value = false
  }
}

async function saveBaseLocation() {
  baseSaving.value = true
  baseMsg.value = ''
  baseError.value = ''
  const latRaw = baseForm.lat.trim()
  const lngRaw = baseForm.lng.trim()
  let baseLat: number | null = null
  let baseLng: number | null = null
  if (latRaw || lngRaw) {
    baseLat = Number(latRaw)
    baseLng = Number(lngRaw)
    if (!Number.isFinite(baseLat) || !Number.isFinite(baseLng)) {
      baseError.value = t('admin.commercials.baseLocationFailed')
      baseSaving.value = false
      return
    }
  }
  try {
    await $fetch(`/api/admin/commercials/${selectedCommercialId.value}/base-location`, {
      method: 'PATCH',
      body: {
        baseCity: baseForm.city,
        basePostalCode: baseForm.postalCode,
        baseLat,
        baseLng,
      },
    })
    baseMsg.value = t('admin.commercials.baseLocationSuccess')
    await load()
  } catch {
    baseError.value = t('admin.commercials.baseLocationFailed')
  } finally {
    baseSaving.value = false
  }
}

onMounted(load)
</script>
