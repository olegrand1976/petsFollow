<template>
  <div data-testid="commercial-prospects-page">
    <ProPageHeader
      :title="$t('commercial.prospects.title')"
      :subtitle="$t('commercial.prospects.subtitle')"
    />

    <ProCard class="pro-mb-lg" data-testid="commercial-prospect-form">
      <h3 class="pro-mb-md">{{ $t('commercial.prospects.create') }}</h3>
      <form class="pro-form" @submit.prevent="createProspect">
        <ProInput v-model="form.practiceName" test-id="prospect-practice" :label="$t('commercial.prospects.practiceName')" required />
        <ProInput v-model="form.contactName" test-id="prospect-contact" :label="$t('commercial.prospects.contactName')" />
        <ProInput v-model="form.contactEmail" test-id="prospect-email" type="email" :label="$t('commercial.prospects.contactEmail')" />
        <ProInput v-model="form.contactPhone" test-id="prospect-phone" :label="$t('commercial.prospects.contactPhone')" />
        <ProInput v-model="form.city" test-id="prospect-city" :label="$t('commercial.prospects.city')" />
        <ProInput v-model="form.notes" test-id="prospect-notes" :label="$t('commercial.prospects.notes')" />
        <label class="pro-label">{{ $t('commercial.prospects.appointmentAt') }}</label>
        <input v-model="form.appointmentAt" class="pro-input" type="datetime-local" data-testid="prospect-appointment">
        <ProButton type="submit" test-id="prospect-submit">{{ $t('commercial.prospects.save') }}</ProButton>
      </form>
    </ProCard>

    <ProCard>
      <ProListToolbar>
        <template #filters>
          <select v-model="statusFilter" class="pro-select" data-testid="prospect-status-filter">
            <option value="">{{ $t('commercial.prospects.statusAll') }}</option>
            <option v-for="s in statuses" :key="s" :value="s">{{ $t(`commercial.prospects.status.${s}`) }}</option>
          </select>
        </template>
      </ProListToolbar>
      <ProTable :empty="!filtered.length" :empty-title="$t('commercial.prospects.empty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.prospects.practiceName') }}</th>
            <th>{{ $t('commercial.prospects.contactName') }}</th>
            <th>{{ $t('commercial.prospects.sourceLabel') }}</th>
            <th>{{ $t('commercial.prospects.statusLabel') }}</th>
            <th>{{ $t('commercial.prospects.appointmentAt') }}</th>
            <th>{{ $t('commercial.prospects.appointmentOutcome') }}</th>
            <th>{{ $t('commercial.prospects.daysInStatus') }}</th>
            <th>{{ $t('commercial.prospects.city') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in filtered" :key="p.id" :data-testid="`prospect-row-${p.id}`">
            <td>{{ p.practiceName }}</td>
            <td>{{ p.contactName }}</td>
            <td>{{ $t(`commercial.prospects.source.${p.source || 'commercial'}`) }}</td>
            <td>
              <select
                class="pro-select"
                :value="p.status"
                :data-testid="`prospect-status-${p.id}`"
                @change="(e) => updateStatus(p.id, (e.target as HTMLSelectElement).value)"
              >
                <option v-for="s in statuses" :key="s" :value="s">{{ $t(`commercial.prospects.status.${s}`) }}</option>
              </select>
              <input
                v-if="p.status === 'lost'"
                class="pro-input pro-mt-sm"
                :value="p.lostReason || ''"
                :placeholder="$t('commercial.prospects.lostReason')"
                @change="(e) => patch(p.id, { lostReason: (e.target as HTMLInputElement).value, status: 'lost' })"
              >
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
            <td>{{ p.daysInStatus }}</td>
            <td>{{ p.city }}</td>
            <td>
              <ProButton
                v-if="p.source !== 'directory'"
                variant="ghost"
                :test-id="`prospect-delete-${p.id}`"
                @click="remove(p.id)"
              >
                {{ $t('common.delete') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const statuses = ['new', 'contacted', 'qualified', 'converted', 'lost'] as const
const outcomes = ['scheduled', 'done', 'no_show', 'cancelled'] as const
const prospects = ref<any[]>([])
const statusFilter = ref('')
const form = reactive({
  practiceName: '',
  contactName: '',
  contactEmail: '',
  contactPhone: '',
  city: '',
  notes: '',
  appointmentAt: '',
})

const filtered = computed(() =>
  statusFilter.value ? prospects.value.filter((p) => p.status === statusFilter.value) : prospects.value,
)

function toLocalInput(iso?: string) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function load() {
  const res: any = await $fetch('/api/commercial/prospects')
  prospects.value = res.data ?? res ?? []
}

async function createProspect() {
  const body: Record<string, unknown> = { ...form }
  delete body.appointmentAt
  if (form.appointmentAt) {
    body.appointmentAt = new Date(form.appointmentAt).toISOString()
    body.appointmentOutcome = 'scheduled'
  }
  await $fetch('/api/commercial/prospects', { method: 'POST', body })
  Object.assign(form, { practiceName: '', contactName: '', contactEmail: '', contactPhone: '', city: '', notes: '', appointmentAt: '' })
  await load()
}

async function patch(id: string, body: Record<string, unknown>) {
  await $fetch(`/api/commercial/prospects/${id}`, { method: 'PATCH', body })
  await load()
}

async function updateStatus(id: string, status: string) {
  await patch(id, { status })
}

async function onAppt(id: string, value: string) {
  if (!value) {
    await patch(id, { clearAppointment: true })
    return
  }
  await patch(id, { appointmentAt: new Date(value).toISOString(), appointmentOutcome: 'scheduled' })
}

async function remove(id: string) {
  await $fetch(`/api/commercial/prospects/${id}`, { method: 'DELETE' })
  await load()
}

onMounted(load)
</script>
