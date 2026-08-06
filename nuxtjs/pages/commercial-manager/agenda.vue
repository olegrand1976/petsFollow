<template>
  <div data-testid="manager-agenda-page">
    <ProPageHeader
      :title="$t('manager.crm.agendaTitle')"
      :subtitle="$t('manager.crm.agendaSubtitle')"
    >
      <template #actions>
        <div class="pf-week-nav">
          <ProButton variant="secondary" test-id="mgr-agenda-prev" @click="shiftWeek(-1)">←</ProButton>
          <ProButton variant="secondary" test-id="mgr-agenda-today" @click="goToday">{{ $t('commercial.crm.today') }}</ProButton>
          <ProButton variant="secondary" test-id="mgr-agenda-next" @click="shiftWeek(1)">→</ProButton>
        </div>
      </template>
    </ProPageHeader>

    <ProCard class="pro-mb-lg">
      <label class="pro-label">{{ $t('manager.followups.colCommercial') }}</label>
      <select v-model="commercialUserId" class="pro-select" data-testid="mgr-agenda-commercial">
        <option value="">{{ $t('manager.prospects.allCommercials') }}</option>
        <option v-for="m in team" :key="m.userId" :value="m.userId">{{ m.fullName || m.email }}</option>
      </select>
    </ProCard>

    <p class="pro-hint pro-mb-md">{{ rangeLabel }}</p>
    <p v-if="loadError" class="pro-field-error" role="alert">{{ loadError }}</p>

    <ProCard>
      <ProTable :empty="!items.length" :empty-title="$t('commercial.crm.agendaEmpty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.crm.when') }}</th>
            <th>{{ $t('manager.followups.colCommercial') }}</th>
            <th>{{ $t('commercial.crm.itemType') }}</th>
            <th>{{ $t('commercial.prospects.practiceName') }}</th>
            <th>{{ $t('commercial.crm.activityTitle') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="it in items" :key="`${it.source}-${it.id}`" :data-testid="`mgr-agenda-item-${it.id}`">
            <td>{{ formatDt(it.startsAt) }}</td>
            <td>{{ it.commercialName || '—' }}</td>
            <td>{{ $t(`commercial.crm.agendaType.${it.source}`, it.source) }}</td>
            <td>
              <NuxtLink v-if="it.prospectId" :to="`/commercial/prospects/${it.prospectId}`" class="pro-link">
                {{ it.practiceName || '—' }}
              </NuxtLink>
              <span v-else>{{ it.practiceName || '—' }}</span>
            </td>
            <td>{{ it.title || '—' }}</td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial-manager', middleware: 'commercial-manager-only' })

const { t } = useI18n()
const { formatDate } = useFormatters()

const weekStart = ref(startOfWeek(new Date()))
const commercialUserId = ref('')
const team = ref<any[]>([])
const items = ref<any[]>([])
const loadError = ref('')

function startOfWeek(d: Date) {
  const x = new Date(d)
  x.setHours(0, 0, 0, 0)
  const day = x.getDay() || 7
  x.setDate(x.getDate() - (day - 1))
  return x
}
function addDays(d: Date, n: number) {
  const x = new Date(d)
  x.setDate(x.getDate() + n)
  return x
}
function toISODate(d: Date) {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}
function localDayISO(d: Date) {
  const x = new Date(d)
  x.setHours(0, 0, 0, 0)
  return x.toISOString()
}
const rangeLabel = computed(() => {
  const from = weekStart.value
  return `${toISODate(from)} → ${toISODate(addDays(from, 6))}`
})
function formatDt(v?: string) {
  if (!v) return '—'
  try { return formatDate(v) } catch { return v }
}
function shiftWeek(n: number) { weekStart.value = addDays(weekStart.value, n * 7) }
function goToday() { weekStart.value = startOfWeek(new Date()) }

async function loadTeam() {
  try {
    const res: any = await $fetch('/api/commercial-manager/team')
    team.value = res.data ?? res ?? []
  } catch {
    team.value = []
  }
}

async function load() {
  loadError.value = ''
  const from = localDayISO(weekStart.value)
  const to = localDayISO(addDays(weekStart.value, 7))
  try {
    const res: any = await $fetch('/api/commercial-manager/agenda', {
      query: {
        from,
        to,
        commercialUserId: commercialUserId.value || undefined,
      },
    })
    items.value = res.data ?? res ?? []
  } catch (e: any) {
    loadError.value = e?.data?.error?.message || e?.message || t('commercial.crm.loadError')
  }
}

onMounted(loadTeam)
watch([weekStart, commercialUserId], () => { load() }, { immediate: true })
</script>

<style scoped>
.pf-week-nav { display: flex; gap: 0.5rem; }
</style>
