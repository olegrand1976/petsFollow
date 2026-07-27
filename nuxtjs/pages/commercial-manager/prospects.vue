<template>
  <div data-testid="manager-prospects-page">
    <ProPageHeader
      :title="$t('manager.prospects.title')"
      :subtitle="$t('manager.prospects.subtitle')"
    />

    <ProCard class="pro-mb-lg">
      <div class="pf-manager-actions">
        <ProButton
          variant="secondary"
          test-id="manager-release-inactive"
          :loading="releasing"
          @click="releaseInactive"
        >
          {{ $t('manager.prospects.releaseInactive') }}
        </ProButton>
        <p v-if="actionMsg" class="pro-hint" role="status">{{ actionMsg }}</p>
        <p v-if="actionError" class="pro-field-error" role="alert">{{ actionError }}</p>
      </div>
    </ProCard>

    <ProCard>
      <ProListToolbar :show-view-toggle="false">
        <template #filters>
          <select v-model="statusFilter" class="pro-select" data-testid="manager-prospect-status-filter">
            <option value="">{{ $t('commercial.prospects.statusAll') }}</option>
            <option v-for="s in statuses" :key="s" :value="s">{{ $t(`commercial.prospects.status.${s}`) }}</option>
          </select>
          <select v-model="commercialFilter" class="pro-select" data-testid="manager-prospect-commercial-filter">
            <option value="">{{ $t('manager.prospects.allCommercials') }}</option>
            <option v-for="m in team" :key="m.userId" :value="m.userId">{{ m.fullName }}</option>
          </select>
        </template>
      </ProListToolbar>
      <ProTable :empty="!prospects.length" :empty-title="$t('commercial.prospects.empty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.prospects.practiceName') }}</th>
            <th>{{ $t('manager.followups.colCommercial') }}</th>
            <th>{{ $t('commercial.prospects.sourceLabel') }}</th>
            <th>{{ $t('commercial.prospects.statusLabel') }}</th>
            <th>{{ $t('commercial.prospects.lostReason') }}</th>
            <th>{{ $t('commercial.prospects.appointmentAt') }}</th>
            <th>{{ $t('commercial.prospects.appointmentOutcome') }}</th>
            <th>{{ $t('commercial.prospects.city') }}</th>
            <th>{{ $t('commercial.prospects.daysInStatus') }}</th>
            <th>{{ $t('manager.prospects.reassign') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in prospects" :key="p.id">
            <td>{{ p.practiceName }}</td>
            <td>{{ p.commercialName || p.commercialEmail || (!p.commercialUserId ? $t('manager.prospects.unassigned') : '—') }}</td>
            <td>{{ $t(`commercial.prospects.source.${p.source || 'commercial'}`) }}</td>
            <td>
              <select
                class="pro-select"
                :value="p.status"
                @change="(e) => patch(p.id, { status: (e.target as HTMLSelectElement).value })"
              >
                <option v-for="s in statuses" :key="s" :value="s">{{ $t(`commercial.prospects.status.${s}`) }}</option>
              </select>
            </td>
            <td>
              <input
                v-if="p.status === 'lost'"
                class="pro-input"
                :value="p.lostReason || ''"
                :placeholder="$t('commercial.prospects.lostReason')"
                @change="(e) => patch(p.id, { lostReason: (e.target as HTMLInputElement).value, status: 'lost' })"
              >
              <span v-else>—</span>
            </td>
            <td>
              <input
                class="pro-input"
                type="datetime-local"
                :value="toLocalInput(p.appointmentAt)"
                @change="(e) => onAppt(p.id, (e.target as HTMLInputElement).value)"
              >
            </td>
            <td>
              <select
                class="pro-select"
                :value="p.appointmentOutcome || ''"
                @change="(e) => patch(p.id, { appointmentOutcome: (e.target as HTMLSelectElement).value })"
              >
                <option value="">—</option>
                <option v-for="o in outcomes" :key="o" :value="o">{{ $t(`commercial.prospects.outcome.${o}`) }}</option>
              </select>
            </td>
            <td>{{ p.city }}</td>
            <td>
              <span
                v-if="p.inactive"
                class="pf-inactive-dot"
                :title="$t('commercial.prospects.inactiveSince', { n: p.inactiveDays })"
                data-testid="manager-prospect-inactive"
              />
              {{ p.daysInStatus ?? '—' }}
            </td>
            <td>
              <select
                class="pro-select"
                :value="p.commercialUserId || ''"
                data-testid="manager-prospect-reassign"
                @change="(e) => reassign(p.id, (e.target as HTMLSelectElement).value)"
              >
                <option value="">{{ $t('manager.prospects.unassigned') }}</option>
                <option v-for="m in team" :key="m.userId" :value="m.userId">{{ m.fullName }}</option>
              </select>
            </td>
            <td>
              <ProButton
                v-if="p.commercialUserId && p.status !== 'converted'"
                variant="ghost"
                :test-id="`manager-release-${p.id}`"
                @click="releaseOne(p.id)"
              >
                {{ $t('manager.prospects.release') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial-manager', middleware: 'commercial-manager-only' })

const { t } = useI18n()
const { mapError } = useApiError()
const statuses = ['new', 'contacted', 'qualified', 'converted', 'lost'] as const
const outcomes = ['scheduled', 'done', 'no_show', 'cancelled'] as const
const prospects = ref<any[]>([])
const team = ref<any[]>([])
const statusFilter = ref('')
const commercialFilter = ref('')
const releasing = ref(false)
const actionMsg = ref('')
const actionError = ref('')

function toLocalInput(iso?: string) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function load() {
  const q: Record<string, string> = {}
  if (statusFilter.value) q.status = statusFilter.value
  if (commercialFilter.value) q.commercialUserId = commercialFilter.value
  const res: any = await $fetch('/api/commercial-manager/prospects', { query: q })
  prospects.value = res.data ?? res ?? []
}

async function loadTeam() {
  const res: any = await $fetch('/api/commercial-manager/team')
  team.value = res.data ?? res ?? []
}

async function patch(id: string, body: Record<string, unknown>) {
  await $fetch(`/api/commercial-manager/prospects/${id}`, { method: 'PATCH', body })
  await load()
}

async function onAppt(id: string, value: string) {
  if (!value) {
    await patch(id, { clearAppointment: true })
    return
  }
  await patch(id, { appointmentAt: new Date(value).toISOString(), appointmentOutcome: 'scheduled' })
}

async function reassign(id: string, commercialUserId: string) {
  await $fetch(`/api/commercial-manager/prospects/${id}/reassign`, {
    method: 'PATCH',
    body: { commercialUserId },
  })
  await load()
}

async function releaseOne(id: string) {
  actionError.value = ''
  actionMsg.value = ''
  try {
    await $fetch(`/api/commercial-manager/prospects/${id}/release`, { method: 'POST' })
    actionMsg.value = t('manager.prospects.releaseOk')
    await load()
  } catch (e: any) {
    actionError.value = mapError(e)
  }
}

async function releaseInactive() {
  actionError.value = ''
  actionMsg.value = ''
  releasing.value = true
  try {
    const res: any = await $fetch('/api/commercial-manager/prospects/release-inactive', { method: 'POST' })
    const data = res.data ?? res
    actionMsg.value = t('manager.prospects.releaseInactiveOk', { n: data.count ?? 0 })
    await load()
  } catch (e: any) {
    actionError.value = mapError(e)
  } finally {
    releasing.value = false
  }
}

watch([statusFilter, commercialFilter], load)
onMounted(async () => {
  await Promise.all([loadTeam(), load()])
})
</script>

<style scoped>
.pf-manager-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
}
.pf-inactive-dot {
  display: inline-block;
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 50%;
  background: var(--pf-vet-alert);
  margin-right: 0.35rem;
  vertical-align: middle;
}
.pro-mb-lg { margin-bottom: 1.25rem; }
</style>
