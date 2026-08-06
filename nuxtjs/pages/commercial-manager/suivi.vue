<template>
  <div data-testid="manager-followups-page">
    <ProPageHeader
      :title="$t('manager.followups.title')"
      :subtitle="$t('manager.followups.subtitle')"
    >
      <template #actions>
        <NuxtLink to="/commercial-manager/agenda" class="pro-link" data-testid="mgr-followups-agenda-link">
          {{ $t('manager.crm.agendaTitle') }}
        </NuxtLink>
      </template>
    </ProPageHeader>

    <ProCard class="pro-mb-lg" data-testid="manager-overdue-tasks">
      <div class="pf-card-head">
        <h3>{{ $t('manager.crm.overdueTitle') }} ({{ overdueCount }})</h3>
        <ProButton test-id="mgr-assign-open" @click="openAssign()">
          {{ $t('manager.crm.assignAction') }}
        </ProButton>
      </div>
      <ProTable :empty="!overdue.length" :empty-title="$t('manager.crm.overdueEmpty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.crm.activityTitle') }}</th>
            <th>{{ $t('commercial.prospects.practiceName') }}</th>
            <th>{{ $t('manager.followups.colCommercial') }}</th>
            <th>{{ $t('commercial.crm.dueAt') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="a in overdue" :key="a.id" :data-testid="`mgr-overdue-${a.id}`">
            <td>{{ a.title }}</td>
            <td>
              <NuxtLink v-if="a.prospectId" :to="`/commercial/prospects/${a.prospectId}`" class="pro-link">
                {{ a.practiceName || '—' }}
              </NuxtLink>
            </td>
            <td>{{ a.assigneeName || '—' }}</td>
            <td>{{ formatDt(a.dueAt) }}</td>
            <td class="pf-row-actions">
              <ProButton
                variant="ghost"
                :test-id="`mgr-overdue-assign-${a.id}`"
                @click="openAssign({ id: a.prospectId, practiceName: a.practiceName, commercialUserId: a.assigneeUserId })"
              >
                {{ $t('manager.crm.assignAction') }}
              </ProButton>
              <ProButton variant="ghost" :test-id="`mgr-overdue-done-${a.id}`" @click="markDone(a.id)">
                {{ $t('commercial.crm.markDone') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard class="pro-mb-lg">
      <h3 class="pro-mb-md">{{ $t('manager.followups.upcomingTitle') }}</h3>
      <ProTable :empty="!upcoming.length" :empty-title="$t('manager.followups.upcomingEmpty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.prospects.practiceName') }}</th>
            <th>{{ $t('manager.followups.colCommercial') }}</th>
            <th>{{ $t('commercial.prospects.appointmentAt') }}</th>
            <th>{{ $t('commercial.prospects.statusLabel') }}</th>
            <th>{{ $t('commercial.prospects.city') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in upcoming" :key="p.id">
            <td>
              <NuxtLink :to="`/commercial/prospects/${p.id}`" class="pro-link">{{ p.practiceName }}</NuxtLink>
            </td>
            <td>{{ p.commercialName || p.commercialEmail || '—' }}</td>
            <td>{{ formatDt(p.appointmentAt) }}</td>
            <td>{{ $t(`commercial.prospects.status.${p.status}`) }}</td>
            <td>{{ p.city }}</td>
            <td>
              <ProButton variant="ghost" :test-id="`mgr-upcoming-assign-${p.id}`" @click="openAssign(p)">
                {{ $t('manager.crm.assignAction') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard>
      <h3 class="pro-mb-md">{{ $t('manager.followups.staleTitle') }}</h3>
      <ProTable :empty="!stale.length" :empty-title="$t('manager.followups.staleEmpty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.prospects.practiceName') }}</th>
            <th>{{ $t('manager.followups.colCommercial') }}</th>
            <th>{{ $t('commercial.prospects.statusLabel') }}</th>
            <th>{{ $t('commercial.prospects.daysInStatus') }}</th>
            <th>{{ $t('commercial.prospects.city') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in stale" :key="p.id">
            <td>
              <NuxtLink :to="`/commercial/prospects/${p.id}`" class="pro-link">{{ p.practiceName }}</NuxtLink>
            </td>
            <td>{{ p.commercialName || p.commercialEmail || '—' }}</td>
            <td>{{ $t(`commercial.prospects.status.${p.status}`) }}</td>
            <td>{{ p.daysInStatus }}</td>
            <td>{{ p.city }}</td>
            <td>
              <ProButton variant="ghost" :test-id="`mgr-stale-assign-${p.id}`" @click="openAssign(p)">
                {{ $t('manager.crm.assignAction') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProModal
      :open="assignOpen"
      :title="$t('manager.crm.assignTitle')"
      test-id="mgr-assign-modal"
      @update:open="(v) => { assignOpen = v }"
    >
      <p v-if="assignError" class="pro-field-error" role="alert">{{ assignError }}</p>
      <p v-if="assignOk" class="pro-hint" role="status">{{ $t('manager.crm.assignOk') }}</p>
      <form class="pro-form" @submit.prevent="assignTask">
        <label class="pro-label">{{ $t('commercial.prospects.practiceName') }}</label>
        <select v-model="assignForm.prospectId" class="pro-select" data-testid="mgr-assign-prospect" required>
          <option value="" disabled>—</option>
          <option v-for="p in assignProspectOptions" :key="p.id" :value="p.id">
            {{ p.practiceName }}{{ p.commercialName ? ` · ${p.commercialName}` : '' }}
          </option>
        </select>
        <ProInput v-model="assignForm.title" test-id="mgr-assign-title" :label="$t('commercial.crm.activityTitle')" required />
        <label class="pro-label">{{ $t('manager.followups.colCommercial') }}</label>
        <select v-model="assignForm.assigneeUserId" class="pro-select" data-testid="mgr-assign-commercial">
          <option value="">{{ $t('manager.crm.assigneeDefault') }}</option>
          <option v-for="m in team" :key="m.userId" :value="m.userId">{{ m.fullName || m.email }}</option>
        </select>
        <label class="pro-label">{{ $t('commercial.crm.dueAt') }}</label>
        <input v-model="assignForm.dueAt" class="pro-input" type="datetime-local" data-testid="mgr-assign-due">
        <ProButton type="submit" test-id="mgr-assign-submit" :loading="assignSaving">
          {{ $t('manager.crm.assignAction') }}
        </ProButton>
      </form>
    </ProModal>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial-manager', middleware: 'commercial-manager-only' })

const { t } = useI18n()
const { formatDate } = useFormatters()
const upcoming = ref<any[]>([])
const stale = ref<any[]>([])
const overdue = ref<any[]>([])
const overdueCount = ref(0)
const team = ref<any[]>([])
const teamProspects = ref<any[]>([])

const assignOpen = ref(false)
const assignSaving = ref(false)
const assignError = ref('')
const assignOk = ref(false)
const assignForm = reactive({
  prospectId: '',
  title: '',
  assigneeUserId: '',
  dueAt: '',
})

const assignProspectOptions = computed(() => {
  const map = new Map<string, { id: string, practiceName: string, commercialName?: string, commercialUserId?: string }>()
  for (const p of [...stale.value, ...upcoming.value, ...teamProspects.value]) {
    if (p?.id) map.set(p.id, {
      id: p.id,
      practiceName: p.practiceName || p.id,
      commercialName: p.commercialName,
      commercialUserId: p.commercialUserId,
    })
  }
  for (const a of overdue.value) {
    if (a?.prospectId && !map.has(a.prospectId)) {
      map.set(a.prospectId, {
        id: a.prospectId,
        practiceName: a.practiceName || a.prospectId,
        commercialName: a.assigneeName,
        commercialUserId: a.assigneeUserId,
      })
    }
  }
  return [...map.values()].sort((a, b) => a.practiceName.localeCompare(b.practiceName))
})

function formatDt(v?: string) {
  if (!v) return '—'
  try {
    return formatDate(v)
  } catch {
    return v
  }
}

function defaultDueLocal() {
  const d = new Date(Date.now() + 86400000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T10:00`
}

function openAssign(p?: { id?: string, practiceName?: string, commercialUserId?: string, commercialName?: string }) {
  assignError.value = ''
  assignOk.value = false
  assignForm.prospectId = p?.id || ''
  assignForm.title = p?.practiceName ? t('manager.crm.assignDefaultTitle', { practice: p.practiceName }) : ''
  assignForm.assigneeUserId = p?.commercialUserId || ''
  assignForm.dueAt = defaultDueLocal()
  assignOpen.value = true
}

async function load() {
  const res: any = await $fetch('/api/commercial-manager/followups')
  const data = res.data ?? res ?? {}
  upcoming.value = data.upcomingAppointments ?? []
  stale.value = data.staleProspects ?? []
  overdue.value = data.overdueActivities ?? []
  overdueCount.value = Number(data.overdueActivitiesCount ?? overdue.value.length)
}

async function loadTeam() {
  try {
    const [teamRes, prospRes]: any[] = await Promise.all([
      $fetch('/api/commercial-manager/team'),
      $fetch('/api/commercial-manager/prospects', { query: { limit: 200 } }),
    ])
    team.value = teamRes.data ?? teamRes ?? []
    const raw = prospRes.data ?? prospRes
    teamProspects.value = raw.items ?? (Array.isArray(raw) ? raw : [])
  } catch {
    team.value = []
    teamProspects.value = []
  }
}

async function markDone(id: string) {
  try {
    await $fetch(`/api/commercial/activities/${id}`, { method: 'PATCH', body: { status: 'done' } })
    await load()
  } catch (e: any) {
    assignError.value = e?.data?.error?.message || e?.message || t('commercial.crm.loadError')
  }
}

async function assignTask() {
  assignSaving.value = true
  assignError.value = ''
  assignOk.value = false
  try {
    const body: Record<string, unknown> = {
      prospectId: assignForm.prospectId.trim(),
      title: assignForm.title.trim(),
      kind: 'follow_up',
      assigneeUserId: assignForm.assigneeUserId || undefined,
    }
    if (assignForm.dueAt) body.dueAt = new Date(assignForm.dueAt).toISOString()
    await $fetch('/api/commercial-manager/activities', { method: 'POST', body })
    assignOk.value = true
    assignForm.prospectId = ''
    assignForm.title = ''
    assignForm.dueAt = ''
    await load()
  } catch (e: any) {
    assignError.value = e?.data?.error?.message || e?.message || t('commercial.crm.loadError')
  } finally {
    assignSaving.value = false
  }
}

onMounted(async () => {
  await Promise.all([load(), loadTeam()])
})
</script>

<style scoped>
.pf-card-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}
.pf-row-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}
</style>
