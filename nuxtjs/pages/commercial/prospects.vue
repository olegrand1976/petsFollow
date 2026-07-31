<template>
  <div data-testid="commercial-prospects-page">
    <ProPageHeader
      :title="$t('commercial.prospects.title')"
      :subtitle="$t('commercial.prospects.subtitle')"
    />

    <p v-if="actionError" class="pro-field-error pro-mb-md" role="alert">{{ actionError }}</p>

    <ProCard class="pro-mb-lg" data-testid="commercial-prospect-lookup">
      <strong>{{ $t('commercial.prospects.lookupTitle') }}</strong>
      <p class="pro-hint">{{ $t('commercial.prospects.lookupHint') }}</p>
      <div class="pf-lookup-row pro-mt-md">
        <ProInput
          v-model="lookupQ"
          test-id="prospect-lookup-q"
          :label="$t('commercial.prospects.lookupLabel')"
          :placeholder="$t('commercial.prospects.searchPlaceholder')"
        />
        <ProButton test-id="prospect-lookup-btn" :loading="lookupLoading" @click="runLookup">
          {{ $t('commercial.prospects.lookupAction') }}
        </ProButton>
      </div>
      <p
        v-if="lookupResult?.status === 'owned'"
        class="pro-field-error pro-mt-md"
        data-testid="prospect-lookup-owned"
        role="status"
      >
        {{ $t('commercial.prospects.ownedBy', { name: lookupResult.ownerName || '—' }) }}
      </p>
      <div v-else-if="lookupResult?.status === 'free'" class="pro-mt-md" data-testid="prospect-lookup-free">
        <p class="pro-hint">{{ $t('commercial.prospects.lookupFree') }}</p>
        <ProButton
          v-if="lookupResult.prospect?.id"
          test-id="prospect-claim-btn"
          @click="claimProspect(lookupResult.prospect.id)"
        >
          {{ $t('commercial.prospects.claim') }}
        </ProButton>
      </div>
      <p v-else-if="lookupResult?.status === 'not_found'" class="pro-hint pro-mt-md" data-testid="prospect-lookup-missing">
        {{ $t('commercial.prospects.lookupMissing') }}
      </p>
    </ProCard>

    <ProCard class="pro-mb-lg" data-testid="commercial-prospect-form">
      <button
        type="button"
        class="pf-collapse-toggle"
        data-testid="prospect-create-toggle"
        @click="toggleCreate"
      >
        <strong>{{ $t('commercial.prospects.create') }}</strong>
        <span>{{ showCreate ? '−' : '+' }}</span>
      </button>
      <form v-if="showCreate" class="pro-form pro-mt-md" @submit.prevent="createProspect">
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
      <ProListToolbar v-model:view-mode="viewMode">
        <template #filters>
          <ProInput
            v-model="q"
            test-id="prospect-search"
            :label="$t('commercial.prospects.search')"
            :placeholder="$t('commercial.prospects.searchPlaceholder')"
          />
          <select v-model="sourceFilter" class="pro-select" data-testid="prospect-source-filter">
            <option value="">{{ $t('commercial.prospects.sourceAll') }}</option>
            <option value="directory">{{ $t('commercial.prospects.source.directory') }}</option>
            <option value="commercial">{{ $t('commercial.prospects.source.commercial') }}</option>
            <option value="vet_referral">{{ $t('commercial.prospects.source.vet_referral') }}</option>
          </select>
          <select v-model="statusFilter" class="pro-select" data-testid="prospect-status-filter">
            <option value="">{{ $t('commercial.prospects.statusAll') }}</option>
            <option v-for="s in statuses" :key="s" :value="s">{{ $t(`commercial.prospects.status.${s}`) }}</option>
          </select>
        </template>
      </ProListToolbar>

      <p class="pro-hint pro-mb-md" data-testid="prospect-total">
        {{ $t('commercial.prospects.totalCount', { total, from: rangeFrom, to: rangeTo }) }}
      </p>

      <ProTable
        v-if="viewMode === 'table'"
        :empty="!prospects.length"
        :empty-title="$t('commercial.prospects.empty')"
      >
        <thead>
          <tr>
            <th>{{ $t('commercial.prospects.practiceName') }}</th>
            <th>{{ $t('commercial.prospects.city') }}</th>
            <th>{{ $t('commercial.prospects.contactName') }}</th>
            <th>{{ $t('commercial.prospects.contactEmail') }}</th>
            <th>{{ $t('commercial.prospects.contactPhone') }}</th>
            <th>{{ $t('commercial.prospects.sourceLabel') }}</th>
            <th>{{ $t('commercial.prospects.statusLabel') }}</th>
            <th>{{ $t('commercial.prospects.appointmentAt') }}</th>
            <th>{{ $t('commercial.prospects.appointmentOutcome') }}</th>
            <th>{{ $t('commercial.prospects.daysInStatus') }}</th>
            <th>{{ $t('commercial.prospects.notes') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in prospects" :key="p.id" :data-testid="`prospect-row-${p.id}`">
            <td>{{ p.practiceName }}</td>
            <td>{{ p.city }}</td>
            <td>{{ p.contactName || '—' }}</td>
            <td>{{ p.contactEmail }}</td>
            <td>{{ p.contactPhone }}</td>
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
            <td>
              <span
                v-if="p.inactive"
                class="pf-inactive-dot"
                :title="$t('commercial.prospects.inactiveSince', { n: p.inactiveDays })"
                data-testid="prospect-inactive"
              />
              {{ p.daysInStatus }}
            </td>
            <td>
              <span class="pf-notes" :title="p.notes || ''">{{ truncate(p.notes) }}</span>
            </td>
            <td class="pf-actions">
              <NuxtLink
                v-if="p.status !== 'converted'"
                :to="`/commercial/vets?prospectId=${p.id}`"
                class="pro-link"
                :data-testid="`prospect-encode-${p.id}`"
              >
                {{ $t('commercial.prospects.encode') }}
              </NuxtLink>
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

      <ProKanban v-else>
        <ProKanbanColumn
          v-for="col in kanbanColumns"
          :key="col.key"
          :title="col.title"
          :count="col.items.length"
          :empty="!col.items.length"
          :empty-title="$t('common.empty')"
        >
          <div
            v-for="p in col.items"
            :key="p.id"
            class="pro-kanban-card"
            :data-testid="`prospect-kanban-${p.id}`"
          >
            <strong>{{ p.practiceName }}</strong>
            <p class="pro-kanban-card__meta">{{ p.city || '—' }} · {{ p.contactName || p.contactEmail || '—' }}</p>
            <ProBadge variant="neutral">{{ $t(`commercial.prospects.source.${p.source || 'commercial'}`) }}</ProBadge>
            <p v-if="p.daysInStatus != null" class="pro-kanban-card__meta">
              <span
                v-if="p.inactive"
                class="pf-inactive-dot"
                :title="$t('commercial.prospects.inactiveSince', { n: p.inactiveDays })"
              />
              {{ $t('commercial.prospects.daysShort', { n: p.daysInStatus }) }}
            </p>
            <NuxtLink
              v-if="p.status !== 'converted'"
              :to="`/commercial/vets?prospectId=${p.id}`"
              class="pro-link"
            >
              {{ $t('commercial.prospects.encode') }}
            </NuxtLink>
          </div>
        </ProKanbanColumn>
      </ProKanban>

      <div v-if="viewMode === 'table'" class="pf-pager" data-testid="prospect-pager">
        <ProButton
          variant="secondary"
          test-id="prospect-prev"
          :disabled="offset <= 0 || loading"
          @click="prevPage"
        >
          {{ $t('commercial.prospects.prev') }}
        </ProButton>
        <ProButton
          variant="secondary"
          test-id="prospect-next"
          :disabled="!hasNext || loading"
          @click="nextPage"
        >
          {{ $t('commercial.prospects.next') }}
        </ProButton>
      </div>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const { t } = useI18n()
const { mapError } = useApiError()
const statuses = ['new', 'contacted', 'qualified', 'converted', 'lost'] as const
const outcomes = ['scheduled', 'done', 'no_show', 'cancelled'] as const
const pageSize = 50
const prospects = ref<any[]>([])
const total = ref(0)
const offset = ref(0)
const q = ref('')
const sourceFilter = ref('')
const statusFilter = ref('')
const showCreate = ref(false)
const loading = ref(false)
const actionError = ref('')
const lookupQ = ref('')
const lookupLoading = ref(false)
const lookupResult = ref<{ status: string, ownerName?: string, prospect?: any } | null>(null)
const { viewMode } = useListView('commercial-prospects', 'table')
let searchTimer: ReturnType<typeof setTimeout> | null = null
let loadSeq = 0

const kanbanColumns = computed(() =>
  statuses.map((key) => ({
    key,
    title: t(`commercial.prospects.status.${key}`),
    items: prospects.value.filter((p) => p.status === key),
  })),
)

function toggleCreate() {
  showCreate.value = !showCreate.value
}
const form = reactive({
  practiceName: '',
  contactName: '',
  contactEmail: '',
  contactPhone: '',
  city: '',
  notes: '',
  appointmentAt: '',
})

const rangeFrom = computed(() => (total.value === 0 ? 0 : offset.value + 1))
const rangeTo = computed(() => Math.min(offset.value + prospects.value.length, total.value))
const hasNext = computed(() => offset.value + pageSize < total.value)

function truncate(notes?: string) {
  if (!notes) return '—'
  return notes.length > 48 ? `${notes.slice(0, 48)}…` : notes
}

function toLocalInput(iso?: string) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function load() {
  const seq = ++loadSeq
  loading.value = true
  try {
    const isKanban = viewMode.value === 'kanban'
    const res: any = await $fetch('/api/commercial/prospects', {
      query: {
        q: q.value || undefined,
        source: sourceFilter.value || undefined,
        status: statusFilter.value || undefined,
        limit: isKanban ? 500 : pageSize,
        offset: isKanban ? 0 : offset.value,
      },
    })
    if (seq !== loadSeq) return
    const data = res.data ?? res
    prospects.value = data.items ?? (Array.isArray(data) ? data : [])
    total.value = Number(data.total ?? prospects.value.length)
  } finally {
    if (seq === loadSeq) loading.value = false
  }
}

function scheduleReload() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    offset.value = 0
    load()
  }, 300)
}

watch([sourceFilter, statusFilter], () => {
  offset.value = 0
  load()
})
watch(viewMode, () => {
  offset.value = 0
  load()
})
watch(q, scheduleReload)

async function runLookup() {
  actionError.value = ''
  lookupResult.value = null
  const q = lookupQ.value.trim()
  if (!q) return
  lookupLoading.value = true
  try {
    const res: any = await $fetch('/api/commercial/prospects/lookup', { query: { q } })
    lookupResult.value = res.data ?? res
  } catch (e: any) {
    actionError.value = mapError(e)
  } finally {
    lookupLoading.value = false
  }
}

async function claimProspect(id: string) {
  actionError.value = ''
  try {
    await $fetch(`/api/commercial/prospects/${id}/claim`, { method: 'POST' })
    lookupResult.value = null
    offset.value = 0
    await load()
  } catch (e: any) {
    const details = e?.data?.error?.details || e?.data?.details
    if (details?.ownerName) {
      actionError.value = t('commercial.prospects.ownedBy', { name: details.ownerName })
    } else {
      actionError.value = mapError(e)
    }
  }
}

async function createProspect() {
  actionError.value = ''
  const body: Record<string, unknown> = { ...form }
  delete body.appointmentAt
  if (form.appointmentAt) {
    body.appointmentAt = new Date(form.appointmentAt).toISOString()
    body.appointmentOutcome = 'scheduled'
  }
  try {
    await $fetch('/api/commercial/prospects', { method: 'POST', body })
    Object.assign(form, { practiceName: '', contactName: '', contactEmail: '', contactPhone: '', city: '', notes: '', appointmentAt: '' })
    showCreate.value = false
    lookupResult.value = null
    offset.value = 0
    if (sourceFilter.value === 'commercial') {
      await load()
    } else {
      // watch(sourceFilter) will reload once
      sourceFilter.value = 'commercial'
    }
  } catch (e: any) {
    const details = e?.data?.error?.details || e?.data?.details
    if (details?.ownerName) {
      actionError.value = t('commercial.prospects.ownedBy', { name: details.ownerName })
    } else {
      actionError.value = mapError(e)
    }
  }
}

async function patch(id: string, body: Record<string, unknown>) {
  actionError.value = ''
  try {
    await $fetch(`/api/commercial/prospects/${id}`, { method: 'PATCH', body })
    await load()
  } catch (e: any) {
    actionError.value = mapError(e)
  }
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
  actionError.value = ''
  try {
    await $fetch(`/api/commercial/prospects/${id}`, { method: 'DELETE' })
    await load()
  } catch (e: any) {
    actionError.value = mapError(e)
  }
}

function prevPage() {
  offset.value = Math.max(0, offset.value - pageSize)
  load()
}

function nextPage() {
  if (!hasNext.value) return
  offset.value += pageSize
  load()
}

onMounted(load)
</script>

<style scoped>
.pf-collapse-toggle {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  background: transparent;
  border: 0;
  padding: 0;
  cursor: pointer;
  font: inherit;
  color: inherit;
  text-align: left;
}
.pro-mt-md { margin-top: 1rem; }
.pro-mb-md { margin-bottom: 0.75rem; }
.pf-notes {
  display: inline-block;
  max-width: 12rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: bottom;
}
.pf-pager {
  display: flex;
  gap: 0.75rem;
  margin-top: 1rem;
  justify-content: flex-end;
}
.pf-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.35rem;
}
.pro-link {
  color: var(--pf-vet-accent);
  text-decoration: none;
  font-size: 0.875rem;
}
.pro-link:hover {
  text-decoration: underline;
}
.pf-lookup-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: flex-end;
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
</style>
