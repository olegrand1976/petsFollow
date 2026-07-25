<template>
  <div data-testid="admin-commercials-page">
    <ProPageHeader :title="$t('admin.commercials.title')" :subtitle="$t('admin.commercials.subtitle')" />

    <ProCard class="pro-mb-lg" data-testid="admin-assign-manager">
      <h3 class="pro-mb-md">{{ $t('admin.commercials.assignManagerTitle') }}</h3>
      <form class="pro-form" @submit.prevent="assignManager">
        <div class="pro-field">
          <label class="pro-label" for="mgr-commercial">{{ $t('admin.commercials.assignCommercial') }}</label>
          <select id="mgr-commercial" v-model="managerForm.commercialId" class="pro-select" required data-testid="admin-manager-commercial">
            <option value="" disabled>{{ $t('admin.commercials.assignCommercialPlaceholder') }}</option>
            <option v-for="c in rows" :key="c.userId" :value="c.userId">
              {{ c.fullName }} ({{ c.email }})
            </option>
          </select>
        </div>
        <div class="pro-field">
          <label class="pro-label" for="mgr-manager">{{ $t('admin.commercials.assignManager') }}</label>
          <select id="mgr-manager" v-model="managerForm.managerUserId" class="pro-select" data-testid="admin-manager-select">
            <option value="">{{ $t('admin.commercials.assignManagerNone') }}</option>
            <option v-for="m in managers" :key="m.userId" :value="m.userId">
              {{ m.fullName }} ({{ m.email }})
            </option>
          </select>
        </div>
        <p v-if="managerMsg" class="pro-hint">{{ managerMsg }}</p>
        <p v-if="managerError" class="pro-error">{{ managerError }}</p>
        <ProButton type="submit" test-id="admin-manager-submit" :disabled="managerSaving || !managerForm.commercialId">
          {{ $t('admin.commercials.assignManagerSubmit') }}
        </ProButton>
      </form>
    </ProCard>

    <ProCard class="pro-mb-lg" data-testid="admin-assign-vet">
      <h3 class="pro-mb-md">{{ $t('admin.commercials.assignTitle') }}</h3>
      <form class="pro-form" @submit.prevent="assignVet">
        <div class="pro-field">
          <label class="pro-label" for="assign-commercial">{{ $t('admin.commercials.assignCommercial') }}</label>
          <select
            id="assign-commercial"
            v-model="assignForm.commercialId"
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
            v-model="assignForm.vetUserId"
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
        <p v-if="assignMsg" class="pro-hint" data-testid="admin-assign-msg">{{ assignMsg }}</p>
        <p v-if="assignError" class="pro-error">{{ assignError }}</p>
        <ProButton type="submit" test-id="admin-assign-submit" :disabled="assignSaving || !assignForm.commercialId || !assignForm.vetUserId">
          {{ $t('admin.commercials.assignSubmit') }}
        </ProButton>
      </form>
    </ProCard>

    <ProCard class="pro-mb-lg" data-testid="admin-commercial-base-location">
      <h3 class="pro-mb-md">{{ $t('admin.commercials.baseLocationTitle') }}</h3>
      <form class="pro-form" @submit.prevent="saveBaseLocation">
        <div class="pro-field">
          <label class="pro-label" for="base-commercial">{{ $t('admin.commercials.assignCommercial') }}</label>
          <select id="base-commercial" v-model="baseForm.commercialId" class="pro-select" required data-testid="admin-base-commercial">
            <option value="" disabled>{{ $t('admin.commercials.assignCommercialPlaceholder') }}</option>
            <option v-for="c in rows" :key="c.userId" :value="c.userId">
              {{ c.fullName }} ({{ c.email }})
            </option>
          </select>
        </div>
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
        <p v-if="baseMsg" class="pro-hint">{{ baseMsg }}</p>
        <p v-if="baseError" class="pro-error">{{ baseError }}</p>
        <ProButton type="submit" test-id="admin-base-submit" :disabled="baseSaving || !baseForm.commercialId">
          {{ $t('admin.commercials.baseLocationSubmit') }}
        </ProButton>
      </form>
    </ProCard>

    <ProCard>
      <ProTable :empty="!rows.length" :empty-title="$t('admin.commercials.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.commercials.columnName') }}</th>
            <th>{{ $t('admin.commercials.columnEmail') }}</th>
            <th>{{ $t('admin.commercials.columnVets') }}</th>
            <th>{{ $t('admin.commercials.columnDue') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in rows" :key="r.userId">
            <td>{{ r.fullName }}</td>
            <td>{{ r.email }}</td>
            <td>{{ r.clientCount }}</td>
            <td>{{ formatDue(r.userId) }}</td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t, locale } = useI18n()

const rows = ref<any[]>([])
const managers = ref<any[]>([])
const vets = ref<any[]>([])
const dueByCommercial = ref<Record<string, number>>({})
const assignForm = reactive({ commercialId: '', vetUserId: '' })
const assignSaving = ref(false)
const assignMsg = ref('')
const assignError = ref('')
const managerForm = reactive({ commercialId: '', managerUserId: '' })
const managerSaving = ref(false)
const managerMsg = ref('')
const managerError = ref('')
const baseForm = reactive({ commercialId: '', city: '', postalCode: '', lat: '', lng: '' })
const baseSaving = ref(false)
const baseMsg = ref('')
const baseError = ref('')

function formatDue(commercialId: string) {
  const cents = dueByCommercial.value[commercialId]
  if (cents == null) return '—'
  return new Intl.NumberFormat(locale.value, { style: 'currency', currency: 'EUR' }).format(cents / 100)
}

async function loadCommissions(commercialId: string) {
  try {
    const res: any = await $fetch(`/api/admin/commercials/${commercialId}/commissions`)
    const data = res.data ?? res
    dueByCommercial.value[commercialId] = Number(data.lifetimeEarnedCents ?? 0)
  } catch {
    dueByCommercial.value[commercialId] = 0
  }
}

async function load() {
  const [commercialsRes, managersRes, usersRes]: any[] = await Promise.all([
    $fetch('/api/admin/commercials'),
    $fetch('/api/admin/commercial-managers').catch(() => null),
    $fetch('/api/admin/users', { query: { role: 'vet' } }),
  ])
  rows.value = commercialsRes.data ?? commercialsRes ?? []
  managers.value = managersRes?.data ?? managersRes ?? []
  const users = usersRes.data ?? usersRes ?? []
  vets.value = Array.isArray(users) ? users : []
  await Promise.all(rows.value.map((r: any) => loadCommissions(r.userId)))
}

async function assignManager() {
  managerSaving.value = true
  managerMsg.value = ''
  managerError.value = ''
  try {
    await $fetch(`/api/admin/commercials/${managerForm.commercialId}/manager`, {
      method: 'PATCH',
      body: { managerUserId: managerForm.managerUserId },
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
    await $fetch(`/api/admin/commercials/${assignForm.commercialId}/assign`, {
      method: 'PATCH',
      body: { vetUserId: assignForm.vetUserId },
    })
    assignMsg.value = t('admin.commercials.assignSuccess')
    assignForm.vetUserId = ''
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
    await $fetch(`/api/admin/commercials/${baseForm.commercialId}/base-location`, {
      method: 'PATCH',
      body: {
        baseCity: baseForm.city,
        basePostalCode: baseForm.postalCode,
        baseLat,
        baseLng,
      },
    })
    baseMsg.value = t('admin.commercials.baseLocationSuccess')
  } catch {
    baseError.value = t('admin.commercials.baseLocationFailed')
  } finally {
    baseSaving.value = false
  }
}

onMounted(load)
</script>
