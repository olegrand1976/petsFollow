<template>
  <div data-testid="filiation-table">
    <ProListToolbar>
      <template #filters>
        <ProInput
          v-model="search"
          test-id="filiation-search"
          :label="t('filiation.search')"
          :placeholder="t('filiation.searchPlaceholder')"
        />
        <div v-if="showBranchFilter" class="pro-field">
          <label class="pro-label" for="filiation-branch">{{ t('filiation.branch') }}</label>
          <select
            id="filiation-branch"
            v-model="branchId"
            class="pro-select"
            data-testid="filiation-branch"
          >
            <option value="">{{ t('filiation.branchAll') }}</option>
            <option v-for="b in branches" :key="b.id" :value="b.id">
              {{ b.name }} ({{ b.code }})
            </option>
          </select>
        </div>
        <div v-if="showCommercialFilter" class="pro-field">
          <label class="pro-label" for="filiation-commercial">{{ t('filiation.commercial') }}</label>
          <select
            id="filiation-commercial"
            v-model="commercialId"
            class="pro-select"
            data-testid="filiation-commercial"
          >
            <option value="">{{ t('filiation.commercialAll') }}</option>
            <option v-for="c in commercials" :key="c.userId" :value="c.userId">
              {{ c.fullName }} ({{ c.email }})
            </option>
          </select>
        </div>
      </template>
      <template #actions>
        <button
          type="button"
          class="pro-btn pro-btn--ghost"
          data-testid="filiation-export-csv"
          :disabled="!rows.length || loading"
          @click="exportCsv"
        >
          {{ t('filiation.exportCsv') }}
        </button>
        <button
          type="button"
          class="pro-btn pro-btn--ghost"
          data-testid="filiation-export-events-csv"
          :disabled="!events.length || eventsLoading"
          @click="exportEventsCsv"
        >
          {{ t('filiation.exportEventsCsv') }}
        </button>
      </template>
    </ProListToolbar>

    <p v-if="loading" class="pro-hint" data-testid="filiation-loading">{{ t('common.loading') }}</p>
    <p v-else-if="loadError" class="pro-error" data-testid="filiation-error">{{ loadError }}</p>
    <p v-else-if="truncated" class="pro-hint" data-testid="filiation-truncated">
      {{ t('filiation.truncatedPage', { limit: pageLimit }) }}
    </p>
    <p v-if="!loading && !loadError && rows.length" class="pro-hint" data-testid="filiation-csv-hint">
      {{ t('filiation.exportCsvPageHint') }}
    </p>

    <div v-if="!loading && !loadError" class="filiation-pager" data-testid="filiation-pager">
      <button
        type="button"
        class="pro-btn pro-btn--ghost"
        data-testid="filiation-prev"
        :disabled="pageOffset <= 0"
        @click="prevPage"
      >
        {{ t('filiation.prev') }}
      </button>
      <span class="pro-hint">{{ t('filiation.pageHint', { offset: pageOffset, limit: pageLimit }) }}</span>
      <button
        type="button"
        class="pro-btn pro-btn--ghost"
        data-testid="filiation-next"
        :disabled="!truncated && rows.length < pageLimit"
        @click="nextPage"
      >
        {{ t('filiation.next') }}
      </button>
    </div>

    <ProTable
      v-if="!loading && !loadError"
      :empty="!rows.length"
      :empty-title="t('filiation.empty')"
    >
      <thead>
        <tr>
          <th v-if="showOrgCols">{{ t('filiation.branch') }}</th>
          <th v-if="showOrgCols">{{ t('filiation.commercial') }}</th>
          <th>{{ t('filiation.vet') }}</th>
          <th>{{ t('filiation.practice') }}</th>
          <th>{{ t('filiation.client') }}</th>
          <th>{{ t('filiation.inviteCode') }}</th>
          <th>{{ t('filiation.sponsor') }}</th>
          <th>{{ t('filiation.effective') }}</th>
          <th>{{ t('filiation.linkedAt') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(r, idx) in rows"
          :key="rowKey(r, idx)"
          :data-testid="`filiation-row-${idx}`"
        >
          <td v-if="showOrgCols">{{ r.branchName || '—' }}</td>
          <td v-if="showOrgCols">
            <div>{{ r.commercialName }}</div>
            <div class="pro-hint">{{ r.commercialEmail }}</div>
          </td>
          <td>
            <template v-if="r.vetName">
              <div>{{ r.vetName }}</div>
              <div class="pro-hint">{{ r.vetEmail }}</div>
            </template>
            <template v-else>—</template>
          </td>
          <td>{{ r.practiceName || '—' }}</td>
          <td>
            <template v-if="r.clientName">
              <div>{{ r.clientName }}</div>
              <div class="pro-hint">{{ r.clientEmail }}</div>
            </template>
            <template v-else>—</template>
          </td>
          <td>{{ r.inviteCode || '—' }}</td>
          <td>{{ r.sponsorClientName || '—' }}</td>
          <td>
            <ProBadge :variant="sourceVariant(r.effectiveSource)" data-testid="filiation-effective-badge">
              {{ sourceLabel(r.effectiveSource) }}
            </ProBadge>
            <div v-if="r.effectiveCommercialName" class="pro-hint">{{ r.effectiveCommercialName }}</div>
          </td>
          <td>{{ formatDt(r.linkedAt) }}</td>
        </tr>
      </tbody>
    </ProTable>

    <section class="filiation-history" data-testid="filiation-history">
      <h3 class="pro-section-title">{{ t('filiation.historyTitle') }}</h3>
      <div class="filiation-history-filters">
        <div class="pro-field">
          <label class="pro-label" for="filiation-event-type">{{ t('filiation.historyType') }}</label>
          <select
            id="filiation-event-type"
            v-model="eventType"
            class="pro-select"
            data-testid="filiation-event-type"
          >
            <option value="">{{ t('filiation.eventTypeAll') }}</option>
            <option value="vet_assigned">{{ t('filiation.event.vet_assigned') }}</option>
            <option value="vet_unassigned">{{ t('filiation.event.vet_unassigned') }}</option>
            <option value="client_referral">{{ t('filiation.event.client_referral') }}</option>
            <option value="practice_client_linked">{{ t('filiation.event.practice_client_linked') }}</option>
          </select>
        </div>
      </div>
      <p v-if="eventsLoading" class="pro-hint">{{ t('common.loading') }}</p>
      <p v-else-if="eventsError" class="pro-error">{{ eventsError }}</p>
      <p v-else-if="eventsTruncated" class="pro-hint" data-testid="filiation-events-truncated">
        {{ t('filiation.eventsTruncatedPage', { limit: eventsLimit }) }}
      </p>
      <div v-if="!eventsLoading && !eventsError" class="filiation-pager" data-testid="filiation-events-pager">
        <button
          type="button"
          class="pro-btn pro-btn--ghost"
          data-testid="filiation-events-prev"
          :disabled="eventsOffset <= 0"
          @click="prevEventsPage"
        >
          {{ t('filiation.prev') }}
        </button>
        <span class="pro-hint">{{ t('filiation.pageHint', { offset: eventsOffset, limit: eventsLimit }) }}</span>
        <button
          type="button"
          class="pro-btn pro-btn--ghost"
          data-testid="filiation-events-next"
          :disabled="!eventsTruncated && events.length < eventsLimit"
          @click="nextEventsPage"
        >
          {{ t('filiation.next') }}
        </button>
      </div>
      <ProTable
        v-if="!eventsLoading && !eventsError"
        :empty="!events.length"
        :empty-title="t('filiation.historyEmpty')"
      >
        <thead>
          <tr>
            <th>{{ t('filiation.historyWhen') }}</th>
            <th>{{ t('filiation.historyType') }}</th>
            <th v-if="showOrgCols">{{ t('filiation.commercial') }}</th>
            <th>{{ t('filiation.vet') }}</th>
            <th>{{ t('filiation.client') }}</th>
            <th>{{ t('filiation.practice') }}</th>
            <th>{{ t('filiation.inviteCode') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(e, idx) in events"
            :key="e.id"
            :data-testid="`filiation-event-${idx}`"
          >
            <td>{{ formatDt(e.createdAt) }}</td>
            <td>{{ eventTypeLabel(e.eventType) }}</td>
            <td v-if="showOrgCols">{{ e.commercialName || '—' }}</td>
            <td>{{ e.vetName || '—' }}</td>
            <td>{{ e.clientName || '—' }}</td>
            <td>{{ e.practiceName || '—' }}</td>
            <td>{{ e.inviteCode || '—' }}</td>
          </tr>
        </tbody>
      </ProTable>
    </section>
  </div>
</template>

<script setup lang="ts">
export type FiliationRow = {
  branchId?: string
  branchName?: string
  commercialUserId: string
  commercialName: string
  commercialEmail: string
  vetUserId?: string
  vetName?: string
  vetEmail?: string
  practiceId?: string
  practiceName?: string
  clientUserId?: string
  clientName?: string
  clientEmail?: string
  inviteCode?: string
  sponsorClientName?: string
  effectiveCommercialId?: string
  effectiveCommercialName?: string
  effectiveSource: string
  linkedAt?: string
}

type BranchOpt = { id: string, name: string, code: string }
type CommercialOpt = { userId: string, fullName: string, email: string }

type FiliationEvent = {
  id: string
  eventType: string
  commercialUserId?: string
  commercialName?: string
  vetUserId?: string
  vetName?: string
  clientUserId?: string
  clientName?: string
  practiceId?: string
  practiceName?: string
  inviteCode?: string
  createdAt?: string
}

const props = withDefaults(defineProps<{
  apiPath: string
  commercialsApiPath?: string
  showOrgCols?: boolean
  showBranchFilter?: boolean
  showCommercialFilter?: boolean
}>(), {
  commercialsApiPath: '',
  showOrgCols: false,
  showBranchFilter: false,
  showCommercialFilter: false,
})

const { t } = useI18n()
const rows = ref<FiliationRow[]>([])
const events = ref<FiliationEvent[]>([])
const branches = ref<BranchOpt[]>([])
const commercials = ref<CommercialOpt[]>([])
const search = ref('')
const branchId = ref('')
const commercialId = ref('')
const eventType = ref('')
const loading = ref(false)
const eventsLoading = ref(false)
const loadError = ref('')
const eventsError = ref('')
const truncated = ref(false)
const eventsTruncated = ref(false)
const pageLimit = ref(100)
const pageOffset = ref(0)
const eventsLimit = ref(100)
const eventsOffset = ref(0)
let searchTimer: ReturnType<typeof setTimeout> | null = null

const eventsApiPath = computed(() => `${props.apiPath.replace(/\/$/, '')}/events`)

function rowKey(r: FiliationRow, idx: number) {
  return `${r.commercialUserId}-${r.vetUserId || ''}-${r.clientUserId || ''}-${idx}`
}

function sourceLabel(source: string) {
  if (source === 'vet_assignment') return t('filiation.source.vet_assignment')
  if (source === 'client_referral') return t('filiation.source.client_referral')
  if (source === 'none') return t('filiation.source.none')
  return source
}

function eventTypeLabel(type: string) {
  const key = `filiation.event.${type}`
  const label = t(key)
  return label === key ? type : label
}

function sourceVariant(source: string): 'success' | 'warning' | 'neutral' {
  if (source === 'vet_assignment') return 'success'
  if (source === 'client_referral') return 'warning'
  return 'neutral'
}

function formatDt(raw?: string) {
  if (!raw) return '—'
  const d = new Date(raw)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleString()
}

function csvEscape(v: string) {
  if (/[",\n]/.test(v)) return `"${v.replace(/"/g, '""')}"`
  return v
}

function downloadCsv(filename: string, lines: string[]) {
  const blob = new Blob([lines.join('\n')], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

function exportCsv() {
  const headers = [
    ...(props.showOrgCols ? ['branchId', 'branchName', 'commercialUserId', 'commercialName', 'commercialEmail'] : ['commercialUserId']),
    'vetUserId', 'vetName', 'vetEmail',
    'practiceId', 'practiceName',
    'clientUserId', 'clientName', 'clientEmail',
    'inviteCode', 'sponsorClientName',
    'effectiveSource', 'effectiveCommercialId', 'effectiveCommercialName',
    'linkedAt',
  ]
  const lines = [headers.join(',')]
  for (const r of rows.value) {
    const cols = [
      ...(props.showOrgCols
        ? [r.branchId || '', r.branchName || '', r.commercialUserId, r.commercialName, r.commercialEmail]
        : [r.commercialUserId]),
      r.vetUserId || '', r.vetName || '', r.vetEmail || '',
      r.practiceId || '', r.practiceName || '',
      r.clientUserId || '', r.clientName || '', r.clientEmail || '',
      r.inviteCode || '', r.sponsorClientName || '',
      r.effectiveSource, r.effectiveCommercialId || '', r.effectiveCommercialName || '',
      r.linkedAt || '',
    ]
    lines.push(cols.map(c => csvEscape(String(c))).join(','))
  }
  downloadCsv(`filiation-${new Date().toISOString().slice(0, 10)}.csv`, lines)
}

function exportEventsCsv() {
  const headers = [
    'id', 'eventType', 'createdAt',
    'commercialUserId', 'commercialName',
    'vetUserId', 'vetName',
    'clientUserId', 'clientName',
    'practiceId', 'practiceName',
    'inviteCode',
  ]
  const lines = [headers.join(',')]
  for (const e of events.value) {
    lines.push([
      e.id, e.eventType, e.createdAt || '',
      e.commercialUserId || '', e.commercialName || '',
      e.vetUserId || '', e.vetName || '',
      e.clientUserId || '', e.clientName || '',
      e.practiceId || '', e.practiceName || '',
      e.inviteCode || '',
    ].map(c => csvEscape(String(c))).join(','))
  }
  downloadCsv(`filiation-events-${new Date().toISOString().slice(0, 10)}.csv`, lines)
}

function parsePage(res: any): FiliationRow[] {
  const data = res?.data ?? res
  if (Array.isArray(data)) {
    truncated.value = false
    return data
  }
  if (data && Array.isArray(data.items)) {
    truncated.value = Boolean(data.truncated)
    if (typeof data.limit === 'number') pageLimit.value = data.limit
    if (typeof data.offset === 'number') pageOffset.value = data.offset
    return data.items
  }
  return []
}

function parseEventsPage(res: any): FiliationEvent[] {
  const data = res?.data ?? res
  if (Array.isArray(data)) {
    eventsTruncated.value = false
    return data
  }
  if (data && Array.isArray(data.items)) {
    eventsTruncated.value = Boolean(data.truncated)
    if (typeof data.limit === 'number') eventsLimit.value = data.limit
    if (typeof data.offset === 'number') eventsOffset.value = data.offset
    return data.items
  }
  return []
}

async function loadBranches() {
  if (!props.showBranchFilter) return
  try {
    const res: any = await $fetch('/api/admin/sales-branches')
    const list = res.data ?? res ?? []
    branches.value = Array.isArray(list) ? list : (list.branches ?? [])
  } catch {
    branches.value = []
  }
}

async function loadCommercials() {
  if (!props.showCommercialFilter) return
  const path = props.commercialsApiPath || '/api/admin/commercials'
  try {
    const res: any = await $fetch(path)
    const list = res.data ?? res ?? []
    if (!Array.isArray(list)) {
      commercials.value = []
      return
    }
    commercials.value = list.map((c: any) => ({
      userId: c.userId || c.id || '',
      fullName: c.fullName || c.name || '',
      email: c.email || '',
    })).filter((c: CommercialOpt) => c.userId)
  } catch {
    commercials.value = []
  }
}

async function load() {
  loading.value = true
  loadError.value = ''
  truncated.value = false
  try {
    const query: Record<string, string> = {
      limit: String(pageLimit.value),
      offset: String(pageOffset.value),
    }
    const q = search.value.trim()
    if (q) query.q = q
    if (branchId.value.trim()) query.branchId = branchId.value.trim()
    if (commercialId.value.trim()) query.commercialId = commercialId.value.trim()
    const res: any = await $fetch(props.apiPath, { query })
    rows.value = parsePage(res)
  } catch (e: any) {
    rows.value = []
    loadError.value = e?.data?.error?.message || e?.message || t('filiation.loadError')
  } finally {
    loading.value = false
  }
}

async function loadEvents() {
  eventsLoading.value = true
  eventsError.value = ''
  eventsTruncated.value = false
  try {
    const query: Record<string, string> = {
      limit: String(eventsLimit.value),
      offset: String(eventsOffset.value),
    }
    if (commercialId.value.trim()) query.commercialId = commercialId.value.trim()
    if (eventType.value.trim()) query.eventType = eventType.value.trim()
    const res: any = await $fetch(eventsApiPath.value, { query })
    events.value = parseEventsPage(res)
  } catch (e: any) {
    events.value = []
    eventsError.value = e?.data?.error?.message || e?.message || t('filiation.loadError')
  } finally {
    eventsLoading.value = false
  }
}

function resetAndLoad() {
  pageOffset.value = 0
  eventsOffset.value = 0
  void load()
  void loadEvents()
}

function prevPage() {
  pageOffset.value = Math.max(0, pageOffset.value - pageLimit.value)
  void load()
}

function nextPage() {
  pageOffset.value += pageLimit.value
  void load()
}

function prevEventsPage() {
  eventsOffset.value = Math.max(0, eventsOffset.value - eventsLimit.value)
  void loadEvents()
}

function nextEventsPage() {
  eventsOffset.value += eventsLimit.value
  void loadEvents()
}

function scheduleSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { resetAndLoad() }, 300)
}

watch(search, () => { scheduleSearch() })
watch(branchId, () => { resetAndLoad() })
watch(commercialId, () => { resetAndLoad() })
watch(eventType, () => {
  eventsOffset.value = 0
  void loadEvents()
})

onMounted(async () => {
  await Promise.all([loadBranches(), loadCommercials()])
  await Promise.all([load(), loadEvents()])
})

onBeforeUnmount(() => {
  if (searchTimer) clearTimeout(searchTimer)
})
</script>

<style scoped>
.filiation-history {
  margin-top: 1.5rem;
}
.filiation-history-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}
.filiation-pager {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin: 0.5rem 0 0.75rem;
}
.pro-section-title {
  margin: 0 0 0.75rem;
  font-size: 1.05rem;
  font-weight: 600;
}
</style>
