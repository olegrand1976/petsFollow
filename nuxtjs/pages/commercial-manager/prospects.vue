<template>
  <div data-testid="manager-prospects-page">
    <ProPageHeader
      :title="$t('manager.prospects.title')"
      :subtitle="$t('manager.prospects.subtitle')"
    />

    <ProCard>
      <ProListToolbar>
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
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in prospects" :key="p.id">
            <td>{{ p.practiceName }}</td>
            <td>{{ p.commercialName || p.commercialEmail || (p.source === 'directory' ? $t('commercial.prospects.source.directory') : '—') }}</td>
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
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial-manager', middleware: 'commercial-manager-only' })

const statuses = ['new', 'contacted', 'qualified', 'converted', 'lost'] as const
const outcomes = ['scheduled', 'done', 'no_show', 'cancelled'] as const
const prospects = ref<any[]>([])
const team = ref<any[]>([])
const statusFilter = ref('')
const commercialFilter = ref('')

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

watch([statusFilter, commercialFilter], load)

onMounted(async () => {
  const teamRes: any = await $fetch('/api/commercial-manager/team')
  team.value = teamRes.data ?? teamRes ?? []
  await load()
})
</script>
