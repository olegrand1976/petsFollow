<template>
  <div data-testid="consultations-page">
    <ProPageHeader :title="$t('consultations.title')" :subtitle="$t('consultations.subtitle')" />
    <p v-if="loadError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ loadError }}</p>

    <ProCard>
      <ProListToolbar :show-view-toggle="false">
        <template #filters>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="consult-search">{{ $t('consultations.search') }}</label>
            <input
              id="consult-search"
              v-model="query"
              type="search"
              class="pro-input"
              :placeholder="$t('consultations.searchPlaceholder')"
              data-testid="consultations-search"
            >
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="consult-status">{{ $t('consultations.statusFilter') }}</label>
            <select
              id="consult-status"
              v-model="statusFilter"
              class="pro-select"
              data-testid="consultations-status-filter"
            >
              <option value="">{{ $t('consultations.statusAll') }}</option>
              <option value="confirmed">{{ $t('consultations.status.confirmed') }}</option>
              <option value="done">{{ $t('consultations.status.done') }}</option>
              <option value="cancelled">{{ $t('consultations.status.cancelled') }}</option>
            </select>
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="consult-from">{{ $t('consultations.from') }}</label>
            <input
              id="consult-from"
              v-model="fromDate"
              type="date"
              class="pro-input"
              data-testid="consultations-from"
            >
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="consult-to">{{ $t('consultations.to') }}</label>
            <input
              id="consult-to"
              v-model="toDate"
              type="date"
              class="pro-input"
              data-testid="consultations-to"
            >
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-checkbox-label" for="consult-audio">
              <input
                id="consult-audio"
                v-model="audioOnly"
                type="checkbox"
                class="pro-checkbox"
                data-testid="consultations-audio-only"
              >
              {{ $t('consultations.audioOnly') }}
            </label>
          </div>
        </template>
      </ProListToolbar>

      <ProTable
        :empty="!rows.length"
        :empty-title="$t('consultations.emptyTitle')"
        :empty-description="$t('consultations.emptyDescription')"
      >
        <thead>
          <tr>
            <th>{{ $t('consultations.columnDate') }}</th>
            <th>{{ $t('consultations.columnClient') }}</th>
            <th>{{ $t('consultations.columnPet') }}</th>
            <th>{{ $t('consultations.columnStatus') }}</th>
            <th>{{ $t('consultations.columnReport') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in rows"
            :key="row.id"
            :data-testid="`consultation-row-${row.id}`"
          >
            <td>{{ formatDate(row.scheduledAt || row.createdAt) }}</td>
            <td>
              <NuxtLink
                v-if="row.clientId"
                :to="`/clients/${row.clientId}`"
              >
                {{ row.clientName || $t('common.dash') }}
              </NuxtLink>
              <template v-else>{{ row.clientName || $t('common.dash') }}</template>
            </td>
            <td>
              <NuxtLink
                v-if="row.clientId && row.petId"
                :to="`/clients/${row.clientId}/pets/${row.petId}`"
              >
                {{ row.petName || $t('common.dash') }}
              </NuxtLink>
              <template v-else>{{ row.petName || $t('common.dash') }}</template>
            </td>
            <td>
              <ProBadge :variant="statusVariant(row.status)">
                {{ statusLabel(row.status) }}
              </ProBadge>
            </td>
            <td>
              <span v-if="row.hasReport">{{ reportStatusLabel(row.reportStatus) }}</span>
              <span v-else class="text-muted">{{ $t('common.dash') }}</span>
            </td>
            <td>
              <div class="pro-flex-gap">
                <ProIconAction
                  v-if="canWriteClinical && row.hasAudio"
                  icon="play_arrow"
                  :label="$t('consultations.listen')"
                  :test-id="`consultation-audio-${row.id}`"
                  @click="playAudio(row)"
                />
                <ProIconAction
                  v-if="canReadPets"
                  icon="description"
                  :label="$t('consultations.openReport')"
                  :test-id="`consultation-open-cr-${row.id}`"
                  @click="openReport(row)"
                />
                <NuxtLink
                  v-else-if="row.clientId && row.petId"
                  :to="`/clients/${row.clientId}/pets/${row.petId}`"
                  class="pro-link"
                  :data-testid="`consultation-pet-link-${row.id}`"
                >
                  {{ $t('common.profile') }}
                </NuxtLink>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
      <p
        v-if="rows.length >= 200"
        class="pro-hint"
        data-testid="consultations-limit-hint"
      >
        {{ $t('consultations.limitHint') }}
      </p>
    </ProCard>

    <ProModal
      v-model:open="reportOpen"
      size="lg"
      :title="$t('consultations.reportModalTitle')"
      test-id="consultation-history-report-modal"
    >
      <ProVisitReportPanel
        v-if="selectedVisitId"
        :visit-id="selectedVisitId"
        :visit-scheduled-at="selectedVisitScheduledAt || undefined"
        :readonly="!canWriteClinical"
      />
    </ProModal>

    <ProModal
      v-model:open="audioOpen"
      size="sm"
      :title="$t('consultations.audioModalTitle')"
      test-id="consultation-audio-modal"
      @update:open="onAudioModalOpen"
    >
      <p v-if="audioError" class="pro-error" role="alert">{{ audioError }}</p>
      <p v-else-if="audioLoading" class="pro-hint">{{ $t('consultations.audioLoading') }}</p>
      <audio
        v-if="audioUrl"
        :src="audioUrl"
        controls
        autoplay
        class="consultations-audio"
        data-testid="consultation-audio-player"
      />
    </ProModal>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'calendar.manage' })

type ConsultationRow = {
  id: string
  petId: string
  clientId?: string
  petName?: string
  clientName?: string
  status: string
  scheduledAt?: string
  createdAt: string
  hasReport?: boolean
  hasAudio?: boolean
  reportStatus?: string
}

const { t, locale } = useI18n()
const { mapError } = useApiError()
const { canPractice } = usePracticePerms()
const canWriteClinical = computed(() => canPractice('pets.write_clinical'))
const canReadPets = computed(() => canPractice('pets.read'))

const rows = ref<ConsultationRow[]>([])
const loadError = ref('')
const query = ref('')
const statusFilter = ref('')
const fromDate = ref('')
const toDate = ref('')
const audioOnly = ref(false)
const loading = ref(false)

const reportOpen = ref(false)
const selectedVisitId = ref('')
const selectedVisitScheduledAt = ref('')
const audioOpen = ref(false)
const audioUrl = ref('')
const audioLoading = ref(false)
const audioError = ref('')

function formatDate(iso?: string) {
  if (!iso) return t('common.dash')
  try {
    return new Intl.DateTimeFormat(locale.value || 'fr', {
      dateStyle: 'short',
      timeStyle: 'short',
    }).format(new Date(iso))
  }
  catch {
    return iso
  }
}

function statusLabel(status: string) {
  const key = `consultations.status.${status}`
  const translated = t(key)
  return translated === key ? status : translated
}

function statusVariant(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
  switch (status) {
    case 'done': return 'success'
    case 'confirmed': return 'warning'
    case 'cancelled': return 'danger'
    default: return 'neutral'
  }
}

function reportStatusLabel(status?: string) {
  if (status === 'final') return t('consultations.reportFinal')
  if (status === 'draft') return t('consultations.reportDraft')
  return t('common.dash')
}

/** Civil day bounds in Europe/Brussels → RFC3339 UTC. */
function dateBounds(dateStr: string, endOfDay: boolean): string | undefined {
  if (!dateStr || !/^\d{4}-\d{2}-\d{2}$/.test(dateStr)) return undefined
  const y = Number(dateStr.slice(0, 4))
  const m = Number(dateStr.slice(5, 7))
  const d = Number(dateStr.slice(8, 10))
  if (!y || !m || !d) return undefined
  const h = endOfDay ? 23 : 0
  const mi = endOfDay ? 59 : 0
  const s = endOfDay ? 59 : 0
  const ms = endOfDay ? 999 : 0
  // Treat Y-M-D H:M:S as Brussels wall time.
  const utcGuess = new Date(Date.UTC(y, m - 1, d, h, mi, s, ms))
  const dtf = new Intl.DateTimeFormat('en-US', {
    timeZone: 'Europe/Brussels',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hourCycle: 'h23',
  })
  const parts = Object.fromEntries(
    dtf.formatToParts(utcGuess)
      .filter(p => p.type !== 'literal')
      .map(p => [p.type, p.value]),
  ) as Record<string, string>
  const asBrusselsMs = Date.UTC(
    Number(parts.year),
    Number(parts.month) - 1,
    Number(parts.day),
    Number(parts.hour),
    Number(parts.minute),
    Number(parts.second),
    ms,
  )
  const offset = asBrusselsMs - utcGuess.getTime()
  return new Date(utcGuess.getTime() - offset).toISOString()
}

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    const params = new URLSearchParams()
    if (query.value.trim()) params.set('q', query.value.trim())
    if (statusFilter.value) params.set('status', statusFilter.value)
    const from = dateBounds(fromDate.value, false)
    const to = dateBounds(toDate.value, true)
    if (from) params.set('from', from)
    if (to) params.set('to', to)
    if (audioOnly.value) params.set('hasAudio', 'true')
    const qs = params.toString()
    const res: any = await $fetch(`/api/vet/consultations${qs ? `?${qs}` : ''}`)
    const list = (res?.data ?? res) as ConsultationRow[]
    rows.value = Array.isArray(list) ? list : []
  }
  catch (e: any) {
    loadError.value = mapError(e) || t('consultations.loadError')
    rows.value = []
  }
  finally {
    loading.value = false
  }
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null
function scheduleLoad() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => { void load() }, 250)
}

watch([query, statusFilter, fromDate, toDate, audioOnly], scheduleLoad)

function openReport(row: ConsultationRow) {
  if (!canReadPets.value) return
  selectedVisitId.value = row.id
  selectedVisitScheduledAt.value = row.scheduledAt || row.createdAt || ''
  reportOpen.value = true
}

function revokeAudioUrl() {
  if (audioUrl.value) {
    URL.revokeObjectURL(audioUrl.value)
    audioUrl.value = ''
  }
}

function onAudioModalOpen(v: boolean) {
  if (!v) {
    revokeAudioUrl()
    audioError.value = ''
  }
}

async function playAudio(row: ConsultationRow) {
  if (!canWriteClinical.value) return
  audioError.value = ''
  audioLoading.value = true
  revokeAudioUrl()
  audioOpen.value = true
  try {
    const blob = await $fetch<Blob>(`/api/visits/${row.id}/report/audio`, {
      responseType: 'blob',
    })
    if (blob.type && blob.type.includes('application/json')) {
      const text = await blob.text()
      try {
        const parsed = JSON.parse(text) as { error?: { msgKey?: string } }
        audioError.value = parsed?.error?.msgKey || t('consultations.audioError')
      }
      catch {
        audioError.value = t('consultations.audioError')
      }
      return
    }
    audioUrl.value = URL.createObjectURL(blob)
  }
  catch (e: any) {
    audioError.value = mapError(e) || t('consultations.audioError')
  }
  finally {
    audioLoading.value = false
  }
}

onMounted(() => { void load() })
onBeforeUnmount(() => {
  revokeAudioUrl()
  if (debounceTimer) clearTimeout(debounceTimer)
})
</script>

<style scoped>
.consultations-audio {
  width: 100%;
  margin-top: 0.5rem;
}
</style>
