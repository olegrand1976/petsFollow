<template>
  <div data-testid="calendar-page">
    <ProPageHeader :title="$t('calendar.title')" :subtitle="$t('calendar.subtitle')">
      <template #actions>
        <NuxtLink to="/settings#calendar" class="pro-btn pro-btn--secondary">
          {{ $t('calendar.openSettings') }}
        </NuxtLink>
      </template>
    </ProPageHeader>

    <p v-if="!clientBookingEnabled" class="pro-inline-feedback" role="status">
      {{ $t('calendar.bookingDisabledHint') }}
    </p>
    <p v-if="actionSuccess" class="pro-inline-feedback" role="status">{{ actionSuccess }}</p>
    <p v-if="actionError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ actionError }}</p>

    <ProCard :title="$t('calendar.pendingTitle')" class="pro-mb-lg">
      <ProEmptyState
        v-if="!pending.length"
        :title="$t('calendar.pendingEmptyTitle')"
        :description="$t('calendar.pendingEmptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('calendar.columnClient') }}</th>
            <th>{{ $t('calendar.columnPet') }}</th>
            <th>{{ $t('calendar.columnWhen') }}</th>
            <th>{{ $t('calendar.columnStatus') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="v in pending"
            :id="`visit-${v.id}`"
            :key="v.id"
            :data-testid="`visit-request-${v.id}`"
            :class="{ 'calendar-row--focus': focusVisitId === v.id }"
          >
            <td>
              <NuxtLink v-if="v.clientId" :to="`/clients/${v.clientId}`">{{ v.clientName }}</NuxtLink>
              <span v-else>{{ v.clientName }}</span>
            </td>
            <td>{{ v.petName }}</td>
            <td>{{ formatWhen(v) }}</td>
            <td>
              <ProBadge :variant="statusVariant(v.status)">{{ statusLabel(v.status) }}</ProBadge>
            </td>
            <td>
              <div class="pro-flex-gap">
                <ProButton
                  v-if="v.status === 'requested'"
                  :disabled="busyId === v.id"
                  @click="act(v.id, 'confirm')"
                >
                  {{ $t('calendar.confirm') }}
                </ProButton>
                <ProButton
                  v-if="v.status === 'reschedule_pending'"
                  :disabled="busyId === v.id"
                  @click="act(v.id, 'accept_reschedule')"
                >
                  {{ $t('calendar.acceptReschedule') }}
                </ProButton>
                <ProButton
                  v-if="v.status === 'reschedule_pending'"
                  variant="ghost"
                  :disabled="busyId === v.id"
                  @click="act(v.id, 'reject_reschedule')"
                >
                  {{ $t('calendar.rejectReschedule') }}
                </ProButton>
                <ProButton variant="ghost" :disabled="busyId === v.id" @click="act(v.id, 'cancel')">
                  {{ $t('calendar.cancel') }}
                </ProButton>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard :title="agendaTitle">
      <div class="calendar-toolbar pro-mb-md">
        <div class="pro-view-toggle" role="group" :aria-label="$t('calendar.viewToggleAria')">
          <button
            type="button"
            class="pro-view-toggle__btn"
            :class="{ 'pro-view-toggle__btn--active': viewMode === 'week' }"
            data-testid="calendar-view-week"
            @click="setView('week')"
          >
            {{ $t('calendar.viewWeek') }}
          </button>
          <button
            type="button"
            class="pro-view-toggle__btn"
            :class="{ 'pro-view-toggle__btn--active': viewMode === 'month' }"
            data-testid="calendar-view-month"
            @click="setView('month')"
          >
            {{ $t('calendar.viewMonth') }}
          </button>
        </div>
        <div class="calendar-nav">
          <ProButton variant="ghost" data-testid="calendar-prev" @click="shiftPeriod(-1)">
            {{ viewMode === 'week' ? $t('calendar.prevWeek') : $t('calendar.prevMonth') }}
          </ProButton>
          <ProButton variant="secondary" data-testid="calendar-today" @click="goToday">
            {{ $t('calendar.today') }}
          </ProButton>
          <strong class="calendar-nav__label">{{ periodLabel }}</strong>
          <ProButton variant="ghost" data-testid="calendar-next" @click="shiftPeriod(1)">
            {{ viewMode === 'week' ? $t('calendar.nextWeek') : $t('calendar.nextMonth') }}
          </ProButton>
        </div>
      </div>

      <ProCalendarWeekGrid
        v-if="viewMode === 'week'"
        :week-start="weekStart"
        :visits="visits"
        :vacations="vacations"
        :focus-visit-id="focusVisitId"
        @select-visit="openVisitDetail"
      />
      <ProCalendarMonthGrid
        v-else
        :month-start="monthStart"
        :visits="visits"
        :vacations="vacations"
        :focus-visit-id="focusVisitId"
        @select-visit="openVisitDetail"
        @select-day="zoomToDay"
      />
    </ProCard>

    <ProModal v-model:open="detailOpen" :title="$t('calendar.visitDetail')">
      <div v-if="selectedVisit" class="visit-detail">
        <p>
          <strong>{{ $t('calendar.columnClient') }} :</strong>
          <NuxtLink v-if="selectedVisit.clientId" :to="`/clients/${selectedVisit.clientId}`">
            {{ selectedVisit.clientName }}
          </NuxtLink>
          <span v-else>{{ selectedVisit.clientName }}</span>
        </p>
        <p>
          <strong>{{ $t('calendar.columnPet') }} :</strong> {{ selectedVisit.petName }}
        </p>
        <p>
          <strong>{{ $t('calendar.columnWhen') }} :</strong> {{ formatWhen(selectedVisit) }}
        </p>
        <p>
          <strong>{{ $t('calendar.columnStatus') }} :</strong>
          <ProBadge :variant="statusVariant(selectedVisit.status)">
            {{ statusLabel(selectedVisit.status) }}
          </ProBadge>
        </p>
        <div class="pro-field pro-mb-md">
          <ProInput
            v-model="visitAddress"
            :label="$t('calendar.address')"
            test-id="visit-address"
          />
          <div class="pro-flex-gap" style="margin-top: 0.5rem">
            <ProButton
              variant="secondary"
              :disabled="reportBusy || !selectedVisit.id"
              test-id="visit-save-address"
              @click="saveVisitAddress"
            >
              {{ $t('calendar.saveAddress') }}
            </ProButton>
            <a
              v-if="mapsUrl"
              :href="mapsUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="pro-link-btn"
            >
              {{ $t('calendar.openMaps') }}
            </a>
          </div>
        </div>
        <div class="pro-field pro-mb-md">
          <label class="pro-label" for="visit-report-body">{{ $t('calendar.reportTitle') }}</label>
          <textarea
            id="visit-report-body"
            v-model="reportBody"
            class="pro-input"
            rows="6"
            data-testid="visit-report-body"
            :placeholder="$t('calendar.reportHint')"
            :disabled="reportBusy || reportStatus === 'final'"
          />
          <div class="pro-flex-gap" style="margin-top: 0.5rem">
            <ProButton
              variant="secondary"
              :disabled="reportBusy || reportStatus === 'final'"
              test-id="visit-report-save"
              @click="saveVisitReport"
            >
              {{ $t('calendar.saveReport') }}
            </ProButton>
            <ProButton
              :disabled="reportBusy || reportStatus === 'final'"
              test-id="visit-report-improve"
              @click="improveVisitReport"
            >
              {{ $t('calendar.improveReport') }}
            </ProButton>
            <ProButton
              variant="secondary"
              :disabled="reportBusy || reportStatus === 'final' || !reportBody.trim()"
              test-id="visit-report-finalize"
              @click="finalizeVisitReport"
            >
              {{ $t('calendar.finalizeReport') }}
            </ProButton>
            <label
              v-if="reportStatus !== 'final'"
              class="pro-link-btn"
              style="cursor: pointer"
            >
              <input
                type="file"
                accept="audio/*,.mp3,.m4a,.wav,.ogg,.webm"
                style="display: none"
                data-testid="visit-report-audio"
                :disabled="reportBusy"
                @change="onReportAudioSelected"
              >
              {{ $t('calendar.transcribeAudio') }}
            </label>
          </div>
          <p v-if="reportStatus === 'final'" class="pro-hint">{{ $t('calendar.reportFinal') }}</p>
          <p v-if="reportMsg" class="pro-hint">{{ reportMsg }}</p>
        </div>
        <div class="pro-flex-gap create-client-actions">
          <ProButton
            v-if="selectedVisit.status === 'requested'"
            :disabled="busyId === selectedVisit.id"
            @click="actFromDetail('confirm')"
          >
            {{ $t('calendar.confirm') }}
          </ProButton>
          <ProButton
            v-if="selectedVisit.status === 'reschedule_pending'"
            :disabled="busyId === selectedVisit.id"
            @click="actFromDetail('accept_reschedule')"
          >
            {{ $t('calendar.acceptReschedule') }}
          </ProButton>
          <ProButton
            v-if="selectedVisit.status === 'reschedule_pending'"
            variant="ghost"
            :disabled="busyId === selectedVisit.id"
            @click="actFromDetail('reject_reschedule')"
          >
            {{ $t('calendar.rejectReschedule') }}
          </ProButton>
          <ProButton
            v-if="selectedVisit.status === 'confirmed'"
            variant="secondary"
            :disabled="busyId === selectedVisit.id"
            @click="openReschedule(selectedVisit)"
          >
            {{ $t('calendar.proposeMove') }}
          </ProButton>
          <ProButton
            variant="ghost"
            :disabled="busyId === selectedVisit.id"
            @click="actFromDetail('cancel')"
          >
            {{ $t('calendar.cancel') }}
          </ProButton>
        </div>
      </div>
    </ProModal>

    <ProModal v-model:open="rescheduleOpen" :title="$t('calendar.proposeMove')">
      <form class="pro-form" @submit.prevent="submitReschedule">
        <ProInput
          v-model="rescheduleAt"
          type="datetime-local"
          :label="$t('calendar.newSlot')"
          required
        />
        <p v-if="rescheduleError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">
          {{ rescheduleError }}
        </p>
        <div class="create-client-actions">
          <ProButton variant="secondary" type="button" @click="closeReschedule">
            {{ $t('common.cancel') }}
          </ProButton>
          <ProButton
            type="submit"
            :loading="busyId === rescheduleVisitId"
            :disabled="busyId === rescheduleVisitId"
          >
            {{ $t('calendar.sendPropose') }}
          </ProButton>
        </div>
      </form>
    </ProModal>
    <ProModal v-model:open="audioConsentOpen" :title="$t('calendar.audioConsentTitle')">
      <p class="pro-hint">{{ $t('calendar.audioConsent') }}</p>
      <template #footer>
        <ProButton variant="ghost" test-id="audio-consent-cancel" @click="cancelAudioConsent">
          {{ $t('calendar.cancel') }}
        </ProButton>
        <ProButton test-id="audio-consent-accept" @click="acceptAudioConsent">
          {{ $t('calendar.audioConsentAccept') }}
        </ProButton>
      </template>
    </ProModal>
  </div>
</template>

<script setup lang="ts">
import type { CalendarVacation, CalendarVisit } from '~/composables/useCalendarGrid'

definePageMeta({ middleware: 'vet-only' })

type CalendarViewMode = 'week' | 'month'

const route = useRoute()
const { t } = useI18n()
const { formatDate, dateLocale } = useFormatters()
const { mapError } = useApiError()
const {
  startOfDay,
  startOfWeek,
  startOfMonth,
  monthGridRange,
  visitDisplayAt,
  statusVariant,
} = useCalendarGrid()

const pending = ref<CalendarVisit[]>([])
const visits = ref<CalendarVisit[]>([])
const vacations = ref<CalendarVacation[]>([])
const clientBookingEnabled = ref(false)
const actionError = ref('')
const actionSuccess = ref('')
const busyId = ref('')
const focusVisitId = ref('')
const anchorDate = ref(startOfDay(new Date()))

const viewMode = ref<CalendarViewMode>('week')
if (import.meta.client) {
  const saved = localStorage.getItem('pf-calendar-view')
  if (saved === 'week' || saved === 'month') viewMode.value = saved
}

const detailOpen = ref(false)
const selectedVisit = ref<CalendarVisit | null>(null)
const visitAddress = ref('')
const reportBody = ref('')
const reportStatus = ref('')
const reportBusy = ref(false)
const reportMsg = ref('')
const audioConsentOpen = ref(false)
const pendingAudioFile = ref<File | null>(null)
const rescheduleOpen = ref(false)
const rescheduleVisitId = ref('')
const rescheduleAt = ref('')
const rescheduleError = ref('')

const mapsUrl = computed(() => {
  const v = selectedVisit.value
  if (!v) return ''
  if (v.lat != null && v.lng != null) {
    return `https://www.google.com/maps/dir/?api=1&destination=${v.lat},${v.lng}`
  }
  const addr = (visitAddress.value || v.addressText || '').trim()
  if (!addr) return ''
  return `https://www.google.com/maps/dir/?api=1&destination=${encodeURIComponent(addr)}`
})

const weekStart = computed(() => startOfWeek(anchorDate.value))
const monthStart = computed(() => startOfMonth(anchorDate.value))

const agendaTitle = computed(() =>
  viewMode.value === 'week' ? t('calendar.weekTitle') : t('calendar.monthTitle'),
)

const periodLabel = computed(() => {
  if (viewMode.value === 'week') {
    const end = new Date(weekStart.value)
    end.setDate(end.getDate() + 6)
    const opts: Intl.DateTimeFormatOptions = { day: 'numeric', month: 'short', year: 'numeric' }
    return `${weekStart.value.toLocaleDateString(dateLocale(), opts)} – ${end.toLocaleDateString(dateLocale(), opts)}`
  }
  return monthStart.value.toLocaleDateString(dateLocale(), { month: 'long', year: 'numeric' })
})

function setView(mode: CalendarViewMode) {
  viewMode.value = mode
  if (import.meta.client) localStorage.setItem('pf-calendar-view', mode)
  load()
}

function shiftPeriod(delta: number) {
  if (viewMode.value === 'week') {
    const n = new Date(anchorDate.value)
    n.setDate(n.getDate() + delta * 7)
    anchorDate.value = startOfDay(n)
  } else {
    // Évite le débordement JS (31 jan + 1 mois → mars)
    const m = monthStart.value
    anchorDate.value = startOfDay(new Date(m.getFullYear(), m.getMonth() + delta, 1))
  }
  load()
}

function goToday() {
  anchorDate.value = startOfDay(new Date())
  load()
}

function zoomToDay(day: Date) {
  anchorDate.value = startOfDay(day)
  setView('week')
}

function formatWhen(v: CalendarVisit) {
  if (v.proposedScheduledAt) {
    return `${formatDate(v.proposedScheduledAt)} (${t('calendar.proposed')})`
  }
  if (v.scheduledAt) return formatDate(v.scheduledAt)
  if (v.createdAt) return `${formatDate(v.createdAt)} (${t('calendar.unscheduled')})`
  return t('calendar.unscheduled')
}

function statusLabel(status: string) {
  return t(`calendar.status.${status}` as any) || status
}

function toLocalRFC3339(d: Date) {
  const pad = (n: number) => String(n).padStart(2, '0')
  const offMin = -d.getTimezoneOffset()
  const sign = offMin >= 0 ? '+' : '-'
  const abs = Math.abs(offMin)
  const oh = pad(Math.floor(abs / 60))
  const om = pad(abs % 60)
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
    + `T${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}${sign}${oh}:${om}`
}

function rangeForView() {
  if (viewMode.value === 'week') {
    const from = weekStart.value
    const to = new Date(from)
    to.setDate(to.getDate() + 7)
    return { from, to }
  }
  const { gridStart, gridEnd } = monthGridRange(monthStart.value)
  return { from: gridStart, to: gridEnd }
}

async function load() {
  actionError.value = ''
  const { from, to } = rangeForView()
  try {
    const [calRes, schedRes]: any[] = await Promise.all([
      $fetch(`/api/vet/calendar?from=${encodeURIComponent(toLocalRFC3339(from))}&to=${encodeURIComponent(toLocalRFC3339(to))}`),
      $fetch('/api/vet/schedule'),
    ])
    const cal = calRes.data ?? calRes
    pending.value = cal.pending ?? []
    visits.value = cal.visits ?? []
    vacations.value = cal.vacations ?? []
    const sched = schedRes.data ?? schedRes
    clientBookingEnabled.value = !!sched.clientBookingEnabled
  } catch (e: any) {
    actionSuccess.value = ''
    actionError.value = mapError(e)
  }
}

async function act(id: string, action: string) {
  busyId.value = id
  actionError.value = ''
  try {
    await $fetch(`/api/visits/${id}`, { method: 'PATCH', body: { action } })
    detailOpen.value = false
    await load()
  } catch (e: any) {
    actionError.value = mapError(e)
  } finally {
    busyId.value = ''
  }
}

async function actFromDetail(action: string) {
  if (!selectedVisit.value) return
  await act(selectedVisit.value.id, action)
}

function openVisitDetail(v: CalendarVisit) {
  selectedVisit.value = v
  visitAddress.value = v.addressText || ''
  reportBody.value = ''
  reportStatus.value = ''
  reportMsg.value = ''
  focusVisitId.value = v.id
  detailOpen.value = true
  void loadVisitReport(v.id)
}

async function loadVisitReport(visitId: string) {
  try {
    const res: any = await $fetch(`/api/visits/${visitId}/report`)
    const data = res.data ?? res
    reportBody.value = data.bodyText || data.transcriptText || ''
    reportStatus.value = data.status || ''
  } catch {
    reportBody.value = ''
    reportStatus.value = ''
  }
}

async function saveVisitAddress() {
  if (!selectedVisit.value) return
  reportBusy.value = true
  reportMsg.value = ''
  try {
    const res: any = await $fetch(`/api/visits/${selectedVisit.value.id}/location`, {
      method: 'PATCH',
      body: { addressText: visitAddress.value.trim() },
    })
    const data = res.data ?? res
    selectedVisit.value = {
      ...selectedVisit.value,
      addressText: data.addressText || visitAddress.value,
      lat: data.lat ?? selectedVisit.value.lat,
      lng: data.lng ?? selectedVisit.value.lng,
    }
    reportMsg.value = t('calendar.addressSaved')
    await load()
  } catch (e: any) {
    reportMsg.value = mapError(e)
  } finally {
    reportBusy.value = false
  }
}

async function saveVisitReport() {
  if (!selectedVisit.value || reportStatus.value === 'final') return
  reportBusy.value = true
  reportMsg.value = ''
  try {
    await $fetch(`/api/visits/${selectedVisit.value.id}/report`, {
      method: 'PUT',
      body: { bodyText: reportBody.value },
    })
    reportMsg.value = t('calendar.reportSaved')
  } catch (e: any) {
    reportMsg.value = mapError(e)
  } finally {
    reportBusy.value = false
  }
}

async function improveVisitReport() {
  if (!selectedVisit.value || reportStatus.value === 'final') return
  reportBusy.value = true
  reportMsg.value = ''
  try {
    await $fetch(`/api/visits/${selectedVisit.value.id}/report`, {
      method: 'PUT',
      body: { bodyText: reportBody.value },
    })
    const res: any = await $fetch(`/api/visits/${selectedVisit.value.id}/report/improve`, {
      method: 'POST',
    })
    const data = res.data ?? res
    reportBody.value = data.bodyText || data.improvedText || reportBody.value
    reportMsg.value = t('calendar.reportImproved')
  } catch (e: any) {
    reportMsg.value = mapError(e)
  } finally {
    reportBusy.value = false
  }
}

async function finalizeVisitReport() {
  if (!selectedVisit.value || reportStatus.value === 'final') return
  reportBusy.value = true
  reportMsg.value = ''
  try {
    await $fetch(`/api/visits/${selectedVisit.value.id}/report`, {
      method: 'PUT',
      body: { bodyText: reportBody.value },
    })
    const res: any = await $fetch(`/api/visits/${selectedVisit.value.id}/report/finalize`, {
      method: 'POST',
    })
    const data = res.data ?? res
    reportStatus.value = data.status || 'final'
    reportBody.value = data.bodyText || reportBody.value
    reportMsg.value = t('calendar.reportFinalized')
  } catch (e: any) {
    reportMsg.value = mapError(e)
  } finally {
    reportBusy.value = false
  }
}

async function onReportAudioSelected(ev: Event) {
  if (!selectedVisit.value || reportStatus.value === 'final') return
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  pendingAudioFile.value = file
  audioConsentOpen.value = true
}

function cancelAudioConsent() {
  pendingAudioFile.value = null
  audioConsentOpen.value = false
}

async function acceptAudioConsent() {
  const file = pendingAudioFile.value
  pendingAudioFile.value = null
  audioConsentOpen.value = false
  if (!file || !selectedVisit.value || reportStatus.value === 'final') return
  reportBusy.value = true
  reportMsg.value = ''
  try {
    const form = new FormData()
    form.append('audio', file, file.name)
    const res: any = await $fetch(`/api/visits/${selectedVisit.value.id}/report/transcribe`, {
      method: 'POST',
      body: form,
    })
    const data = res.data ?? res
    reportBody.value = data.bodyText || data.transcriptText || reportBody.value
    reportMsg.value = t('calendar.reportTranscribed')
  } catch (e: any) {
    reportMsg.value = mapError(e)
  } finally {
    reportBusy.value = false
  }
}

function openReschedule(v: CalendarVisit) {
  rescheduleVisitId.value = v.id
  rescheduleAt.value = ''
  rescheduleError.value = ''
  actionSuccess.value = ''
  rescheduleOpen.value = true
  detailOpen.value = false
}

function closeReschedule() {
  rescheduleOpen.value = false
  rescheduleError.value = ''
}

async function submitReschedule() {
  if (!rescheduleAt.value) return
  busyId.value = rescheduleVisitId.value
  rescheduleError.value = ''
  actionSuccess.value = ''
  try {
    const at = new Date(rescheduleAt.value)
    if (Number.isNaN(at.getTime())) {
      rescheduleError.value = t('errors.invalid_proposed')
      return
    }
    const iso = at.toISOString()
    await $fetch(`/api/visits/${rescheduleVisitId.value}`, {
      method: 'PATCH',
      body: { action: 'propose_reschedule', proposedScheduledAt: iso },
    })
    rescheduleOpen.value = false
    rescheduleError.value = ''
    await load()
    actionSuccess.value = t('calendar.proposeSent')
  } catch (e: any) {
    rescheduleError.value = mapError(e)
  } finally {
    busyId.value = ''
  }
}

async function findVisitById(id: string): Promise<CalendarVisit | null> {
  const local = pending.value.find((v) => v.id === id) || visits.value.find((v) => v.id === id)
  if (local) return local
  const statuses = ['confirmed', 'requested', 'reschedule_pending'] as const
  try {
    const results = await Promise.all(
      statuses.map((status) =>
        $fetch(`/api/vet/visits?status=${encodeURIComponent(status)}`).catch(() => null),
      ),
    )
    for (const res of results) {
      const list = ((res as any)?.data ?? res ?? []) as CalendarVisit[]
      const found = list.find((v) => v.id === id)
      if (found) return found
    }
  } catch {
    /* ignore lookup errors */
  }
  return null
}

function revealVisit(v: CalendarVisit) {
  const at = visitDisplayAt(v)
  if (at) {
    const day = startOfDay(at)
    const { from, to } = rangeForView()
    if (day < from || day >= to) {
      anchorDate.value = day
      load().then(() => {
        const fresh = visits.value.find((x) => x.id === v.id) || v
        openVisitDetail(fresh)
        nextTick(() => {
          document.getElementById(`calendar-chip-${v.id}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
        })
      })
      return
    }
  }
  openVisitDetail(v)
  nextTick(() => {
    document.getElementById(`calendar-chip-${v.id}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  })
}

async function focusVisit(id: string) {
  focusVisitId.value = id
  const inPending = pending.value.find((v) => v.id === id)
  if (inPending) {
    nextTick(() => {
      document.getElementById(`visit-${id}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
    })
    return
  }
  const v = await findVisitById(id)
  if (v) revealVisit(v)
}

onMounted(async () => {
  await load()
  if (typeof route.query.visit === 'string' && route.query.visit) {
    await focusVisit(route.query.visit)
  }
})

watch(
  () => route.query.visit,
  (visit) => {
    if (typeof visit === 'string' && visit) void focusVisit(visit)
  },
)
</script>

<style scoped>
.calendar-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}
.calendar-nav {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}
.calendar-nav__label {
  min-width: 10rem;
  text-align: center;
}
.calendar-row--focus {
  background: color-mix(in srgb, var(--pf-vet-accent) 12%, transparent);
}
.create-client-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 1rem;
}
.visit-detail p {
  margin: 0.4rem 0;
}
.pro-inline-feedback--error {
  background: color-mix(in srgb, var(--pf-vet-alert) 10%, var(--pf-vet-surface));
  border-color: color-mix(in srgb, var(--pf-vet-alert) 35%, transparent);
  color: var(--pf-vet-alert);
}
</style>
