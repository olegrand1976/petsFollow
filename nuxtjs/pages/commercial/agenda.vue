<template>
  <div data-testid="commercial-agenda-page">
    <ProPageHeader
      :title="$t('commercial.crm.agendaTitle')"
      :subtitle="$t('commercial.crm.agendaSubtitle')"
    >
      <template #actions>
        <div class="pf-week-nav">
          <ProButton variant="secondary" test-id="agenda-prev" @click="shiftWeek(-1)">←</ProButton>
          <ProButton variant="secondary" test-id="agenda-today" @click="goToday">{{ $t('commercial.crm.today') }}</ProButton>
          <ProButton variant="secondary" test-id="agenda-next" @click="shiftWeek(1)">→</ProButton>
        </div>
      </template>
    </ProPageHeader>

    <p class="pro-hint pro-mb-md" data-testid="agenda-range">{{ rangeLabel }}</p>
    <p v-if="loadError" class="pro-field-error" role="alert">{{ loadError }}</p>

    <ProCard>
      <ProTable :empty="!items.length" :empty-title="$t('commercial.crm.agendaEmpty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.crm.when') }}</th>
            <th>{{ $t('commercial.crm.itemType') }}</th>
            <th>{{ $t('commercial.prospects.practiceName') }}</th>
            <th>{{ $t('commercial.crm.activityTitle') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="it in items" :key="`${it.source}-${it.id}`" :data-testid="`agenda-item-${it.id}`">
            <td>{{ formatDt(it.startsAt) }}</td>
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

    <ProCard class="pro-mt-lg" data-testid="commercial-tasks-block">
      <h3 class="pro-mb-md">{{ $t('commercial.crm.myTasks') }}</h3>
      <ProTable :empty="!tasks.length" :empty-title="$t('commercial.crm.tasksEmpty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.crm.activityTitle') }}</th>
            <th>{{ $t('commercial.prospects.practiceName') }}</th>
            <th>{{ $t('commercial.crm.dueAt') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="a in tasks" :key="a.id" :data-testid="`task-${a.id}`">
            <td>{{ a.title }}</td>
            <td>
              <NuxtLink v-if="a.prospectId" :to="`/commercial/prospects/${a.prospectId}`" class="pro-link">
                {{ a.practiceName || '—' }}
              </NuxtLink>
            </td>
            <td>{{ formatDt(a.dueAt) }}</td>
            <td>
              <ProButton variant="ghost" :test-id="`task-done-${a.id}`" @click="markDone(a.id)">
                {{ $t('commercial.crm.markDone') }}
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

const { t } = useI18n()
const { formatDate } = useFormatters()

const weekStart = ref(startOfWeek(new Date()))
const items = ref<any[]>([])
const tasks = ref<any[]>([])
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

/** Local midnight as RFC3339 so API range matches the browser week, not UTC date-only. */
function localDayISO(d: Date) {
  const x = new Date(d)
  x.setHours(0, 0, 0, 0)
  return x.toISOString()
}

const rangeLabel = computed(() => {
  const from = weekStart.value
  const to = addDays(from, 6)
  return `${toISODate(from)} → ${toISODate(to)}`
})

function formatDt(v?: string) {
  if (!v) return '—'
  try { return formatDate(v) } catch { return v }
}

function shiftWeek(n: number) {
  weekStart.value = addDays(weekStart.value, n * 7)
}

function goToday() {
  weekStart.value = startOfWeek(new Date())
}

async function load() {
  loadError.value = ''
  const from = localDayISO(weekStart.value)
  const to = localDayISO(addDays(weekStart.value, 7))
  try {
    const [agendaRes, tasksRes]: any[] = await Promise.all([
      $fetch('/api/commercial/agenda', { query: { from, to } }),
      $fetch('/api/commercial/activities', { query: { status: 'open', limit: 50 } }),
    ])
    items.value = agendaRes.data ?? agendaRes ?? []
    tasks.value = tasksRes.data ?? tasksRes ?? []
  } catch (e: any) {
    loadError.value = e?.data?.error?.message || e?.message || t('commercial.crm.loadError')
  }
}

async function markDone(id: string) {
  try {
    await $fetch(`/api/commercial/activities/${id}`, { method: 'PATCH', body: { status: 'done' } })
    await load()
  } catch (e: any) {
    loadError.value = e?.data?.error?.message || e?.message || t('commercial.crm.loadError')
  }
}

watch(weekStart, () => { load() }, { immediate: true })
</script>

<style scoped>
.pf-week-nav {
  display: flex;
  gap: 0.5rem;
}
</style>
