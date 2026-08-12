<template>
  <div data-testid="calendar-page" :data-calendar-ready="calendarReady ? '1' : undefined">
    <ProPageHeader :title="$t('calendar.title')" :subtitle="$t('calendar.subtitle')">
      <template #actions>
        <ProButton data-testid="calendar-new-appointment" @click="openNewAppointment()">
          {{ $t('calendar.newAppointment') }}
        </ProButton>
        <NuxtLink to="/settings?tab=calendar" class="pro-btn pro-btn--secondary">
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
                <ProIconAction
                  v-if="v.status === 'requested' && v.pendingActionBy === 'vet'"
                  icon="check"
                  :label="$t('calendar.confirm')"
                  :disabled="busyId === v.id"
                  @click="act(v.id, 'confirm')"
                />
                <ProIconAction
                  v-if="v.status === 'reschedule_pending' && v.pendingActionBy === 'vet'"
                  icon="event_available"
                  :label="$t('calendar.acceptReschedule')"
                  :disabled="busyId === v.id"
                  @click="act(v.id, 'accept_reschedule')"
                />
                <ProIconAction
                  v-if="v.status === 'reschedule_pending' && v.pendingActionBy === 'vet'"
                  icon="close"
                  :label="$t('calendar.rejectReschedule')"
                  :disabled="busyId === v.id"
                  @click="act(v.id, 'reject_reschedule')"
                />
                <ProIconAction
                  v-if="v.status === 'confirmed' && v.consultationSession"
                  icon="task_alt"
                  :label="$t('calendar.markDone')"
                  :disabled="busyId === v.id"
                  test-id="calendar-walkin-done"
                  @click="act(v.id, 'done')"
                />
                <ProIconAction
                  v-if="!v.consultationSession"
                  icon="cancel"
                  variant="danger"
                  :label="$t('calendar.cancel')"
                  :disabled="busyId === v.id"
                  @click="act(v.id, 'cancel')"
                />
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
            :class="{ 'pro-view-toggle__btn--active': viewMode === 'day' }"
            :aria-pressed="viewMode === 'day'"
            data-testid="calendar-view-day"
            @click="setView('day')"
          >
            {{ $t('calendar.viewDay') }}
          </button>
          <button
            type="button"
            class="pro-view-toggle__btn"
            :class="{ 'pro-view-toggle__btn--active': viewMode === 'week' }"
            :aria-pressed="viewMode === 'week'"
            data-testid="calendar-view-week"
            @click="setView('week')"
          >
            {{ $t('calendar.viewWeek') }}
          </button>
          <button
            type="button"
            class="pro-view-toggle__btn"
            :class="{ 'pro-view-toggle__btn--active': viewMode === 'month' }"
            :aria-pressed="viewMode === 'month'"
            data-testid="calendar-view-month"
            @click="setView('month')"
          >
            {{ $t('calendar.viewMonth') }}
          </button>
        </div>
        <div
          v-if="showResourceToggle"
          class="pro-view-toggle"
          role="group"
          :aria-label="$t('calendar.resourceToggleAria')"
          data-testid="calendar-resource-toggle"
        >
          <button
            type="button"
            class="pro-view-toggle__btn"
            :class="{ 'pro-view-toggle__btn--active': resourceMode === 'people' }"
            :aria-pressed="resourceMode === 'people'"
            data-testid="calendar-resource-people"
            @click="resourceMode = 'people'"
          >
            {{ $t('calendar.resourcePeople') }}
          </button>
          <button
            type="button"
            class="pro-view-toggle__btn"
            :class="{ 'pro-view-toggle__btn--active': resourceMode === 'rooms' }"
            :aria-pressed="resourceMode === 'rooms'"
            data-testid="calendar-resource-rooms"
            @click="resourceMode = 'rooms'"
          >
            {{ $t('calendar.resourceRooms') }}
          </button>
        </div>
        <div class="calendar-nav">
          <ProButton variant="ghost" data-testid="calendar-prev" @click="shiftPeriod(-1)">
            {{ navPrevLabel }}
          </ProButton>
          <ProButton variant="secondary" data-testid="calendar-today" @click="goToday">
            {{ $t('calendar.today') }}
            <span v-if="todayVisitCount > 0" class="calendar-today-count">{{ todayVisitCount }}</span>
          </ProButton>
          <strong class="calendar-nav__label">{{ periodLabel }}</strong>
          <ProButton variant="ghost" data-testid="calendar-next" @click="shiftPeriod(1)">
            {{ navNextLabel }}
          </ProButton>
        </div>
      </div>

      <p
        v-if="viewMode === 'day' && sitesUiEnabled && isAggregatedView"
        class="pro-inline-feedback"
        role="status"
        data-testid="calendar-day-needs-site"
      >
        {{ $t('calendar.dayNeedsSite') }}
      </p>
      <ProCalendarDayGrid
        v-else-if="viewMode === 'day'"
        :day="anchorDate"
        :visits="visits"
        :columns="dayColumns"
        :resource-mode="dayResourceMode"
        :focus-visit-id="focusVisitId"
        :busy-id="busyId"
        @select-visit="openVisitDetail"
        @visit-action="onChipVisitAction"
      />
      <ProCalendarWeekGrid
        v-else-if="viewMode === 'week'"
        :week-start="weekStart"
        :visits="visits"
        :vacations="vacations"
        :focus-visit-id="focusVisitId"
        :busy-id="busyId"
        :show-site-label="multiSite && isAggregatedView"
        @select-visit="openVisitDetail"
        @visit-action="onChipVisitAction"
      />
      <ProCalendarMonthGrid
        v-else
        :month-start="monthStart"
        :visits="visits"
        :vacations="vacations"
        :focus-visit-id="focusVisitId"
        :busy-id="busyId"
        :show-site-label="multiSite && isAggregatedView"
        @select-visit="openVisitDetail"
        @visit-action="onChipVisitAction"
        @select-day="zoomToDay"
      />
    </ProCard>

    <ProModal v-model:open="detailOpen" size="xl" :title="$t('calendar.visitDetail')">
      <div v-if="selectedVisit" class="visit-detail">
        <p>
          <strong>{{ $t('calendar.columnClient') }} :</strong>
          <NuxtLink v-if="selectedVisit.clientId" :to="`/clients/${selectedVisit.clientId}`">
            {{ selectedVisit.clientName }}
          </NuxtLink>
          <span v-else>{{ selectedVisit.clientName }}</span>
          <ProBadge
            v-if="selectedVisit.isWalkinPlaceholder"
            variant="warning"
            data-testid="walkin-visit-badge"
          >
            {{ $t('clients.walkin.badge') }}
          </ProBadge>
        </p>
        <p
          v-if="selectedVisit.callbackPhone"
          data-testid="visit-callback-phone"
        >
          <strong>{{ $t('clients.walkin.callbackPhone') }} :</strong>
          <a :href="`tel:${selectedVisit.callbackPhone}`">{{ selectedVisit.callbackPhone }}</a>
        </p>
        <p>
          <strong>{{ $t('calendar.columnPet') }} :</strong> {{ selectedVisit.petName }}
        </p>
        <p v-if="selectedVisit.siteName" data-testid="visit-site-label">
          <strong>{{ $t('calendar.columnSite') }} :</strong> {{ selectedVisit.siteName }}
        </p>
        <div
          v-if="canEditVisitResources"
          class="visit-resources pro-mb-md"
          data-testid="visit-resources"
        >
          <div class="pro-field">
            <label class="pro-label" for="visit-assignee">{{ $t('calendar.assignee') }}</label>
            <select
              id="visit-assignee"
              v-model="editAssigneeId"
              class="pro-select"
              data-testid="visit-assignee"
            >
              <option value="">{{ $t('calendar.unassigned') }}</option>
              <option v-for="m in detailAssigneeOptions" :key="m.id" :value="m.id">{{ m.fullName }}</option>
            </select>
          </div>
          <div v-if="sitesUiEnabled" class="pro-field">
            <label class="pro-label" for="visit-room">{{ $t('calendar.room') }}</label>
            <select
              id="visit-room"
              v-model="editRoomId"
              class="pro-select"
              data-testid="visit-room"
            >
              <option value="">{{ $t('calendar.noRoom') }}</option>
              <option v-for="r in detailRoomOptions" :key="r.id" :value="r.id">{{ r.name }}</option>
            </select>
          </div>
          <ProButton
            type="button"
            variant="secondary"
            :loading="resourcesBusy"
            data-testid="visit-resources-save"
            @click="saveVisitResources"
          >
            {{ $t('calendar.saveResources') }}
          </ProButton>
        </div>
        <template v-else>
          <p v-if="selectedVisit.assigneeName" data-testid="visit-assignee-label">
            <strong>{{ $t('calendar.assignee') }} :</strong> {{ selectedVisit.assigneeName }}
          </p>
          <p v-if="sitesUiEnabled && selectedVisit.roomName" data-testid="visit-room-label">
            <strong>{{ $t('calendar.room') }} :</strong> {{ selectedVisit.roomName }}
          </p>
        </template>
        <p>
          <strong>{{ $t('calendar.columnWhen') }} :</strong> {{ formatWhen(selectedVisit) }}
        </p>
        <p v-if="selectedVisit.visitTypeName" data-testid="visit-type-label">
          <strong>{{ $t('calendar.visitType') }} :</strong>
          <span
            v-if="selectedVisit.visitTypeColor"
            class="visit-type-dot"
            :style="{ background: selectedVisit.visitTypeColor }"
            aria-hidden="true"
          />
          {{ selectedVisit.visitTypeName }}
          <span v-if="selectedVisit.durationMinutes" class="text-muted">
            ({{ selectedVisit.durationMinutes }} min)
          </span>
        </p>
        <p class="pro-flex-gap" style="align-items: center">
          <strong>{{ $t('calendar.columnStatus') }} :</strong>
          <ProBadge :variant="statusVariant(selectedVisit.status)">
            {{ statusLabel(selectedVisit.status) }}
          </ProBadge>
          <ProBadge
            v-if="selectedVisit.consultationSession"
            variant="warning"
            data-testid="visit-walkin-badge"
          >
            {{ $t('calendar.walkInSession') }}
          </ProBadge>
          <ProBadge
            v-else-if="selectedVisit.source === 'care_pro'"
            variant="neutral"
            data-testid="visit-care-pro-badge"
          >
            {{ $t('calendar.sourceCarePro') }}
          </ProBadge>
          <ProBadge
            v-if="selectedVisit.waitingRoomAt"
            variant="warning"
            data-testid="visit-waiting-room-badge"
            :title="$t('calendar.waitingRoomTooltip')"
          >
            {{ $t('calendar.waitingRoomTag') }}
          </ProBadge>
        </p>
        <p v-if="selectedVisit.preconsultStatus || preconsult" data-testid="visit-preconsult-status">
          <strong>{{ $t('calendar.preconsultLabel') }} :</strong>
          <ProBadge :variant="preconsultBadgeVariant">
            {{ preconsultStatusLabel }}
          </ProBadge>
          <ProBadge
            v-if="preconsultIsUrgent"
            variant="danger"
            data-testid="visit-preconsult-urgent"
          >
            {{ $t('calendar.preconsultUrgentShort') }}
          </ProBadge>
        </p>
        <div
          v-if="preconsult?.status === 'submitted'"
          class="preconsult-answers"
          data-testid="visit-preconsult-answers"
        >
          <h3 class="pro-section-title">{{ $t('calendar.preconsultAnswers') }}</h3>
          <dl class="preconsult-dl">
            <div>
              <dt>{{ $t('calendar.preconsultComplaint') }}</dt>
              <dd>{{ preconsult.answers?.chiefComplaint || '—' }}</dd>
            </div>
            <div>
              <dt>{{ $t('calendar.preconsultDuration') }}</dt>
              <dd>{{ preconsultEnumLabel('duration', preconsult.answers?.duration) }}</dd>
            </div>
            <div>
              <dt>{{ $t('calendar.preconsultBehavior') }}</dt>
              <dd>{{ preconsultEnumLabel('behavior', preconsult.answers?.behavior) }}</dd>
            </div>
            <div>
              <dt>{{ $t('calendar.preconsultAppetite') }}</dt>
              <dd>{{ preconsultEnumLabel('scale', preconsult.answers?.appetite) }}</dd>
            </div>
            <div>
              <dt>{{ $t('calendar.preconsultThirst') }}</dt>
              <dd>{{ preconsultEnumLabel('scale', preconsult.answers?.thirst) }}</dd>
            </div>
            <div>
              <dt>{{ $t('calendar.preconsultElimination') }}</dt>
              <dd>{{ preconsultEnumLabel('scale', preconsult.answers?.elimination) }}</dd>
            </div>
            <div>
              <dt>{{ $t('calendar.preconsultUrgency') }}</dt>
              <dd>{{ preconsultEnumLabel('urgency', preconsult.answers?.urgency) }}</dd>
            </div>
            <div v-if="preconsult.answers?.comment">
              <dt>{{ $t('calendar.preconsultComment') }}</dt>
              <dd>{{ preconsult.answers.comment }}</dd>
            </div>
          </dl>
          <div
            v-if="preconsult.aiUrgency || preconsult.aiSummary"
            class="preconsult-ai"
            data-testid="visit-preconsult-ai"
          >
            <h3 class="pro-section-title">{{ $t('calendar.preconsultAiTitle') }}</h3>
            <p class="pro-settings-hint">{{ $t('calendar.preconsultAiDisclaimer') }}</p>
            <dl class="preconsult-dl">
              <div v-if="preconsult.aiUrgency">
                <dt>{{ $t('calendar.preconsultAiLevel') }}</dt>
                <dd>
                  <ProBadge :variant="preconsultAiBadgeVariant">
                    {{ preconsultAiLevelLabel }}
                  </ProBadge>
                </dd>
              </div>
              <div v-if="preconsult.aiSummary">
                <dt>{{ $t('calendar.preconsultAiSummary') }}</dt>
                <dd>{{ preconsult.aiSummary }}</dd>
              </div>
            </dl>
          </div>
        </div>
        <div v-if="canWriteClinical" class="pro-field pro-mb-md">
          <ProInput
            v-model="visitAddress"
            :label="$t('calendar.address')"
            test-id="visit-address"
          />
          <div class="pro-flex-gap" style="margin-top: 0.5rem">
            <ProButton
              variant="secondary"
              :disabled="addressBusy || !selectedVisit.id"
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
          <p v-if="addressMsg" class="pro-hint">{{ addressMsg }}</p>
        </div>
        <div class="pro-field pro-mb-md" data-testid="calendar-desk-note">
          <label class="pro-label" for="calendar-desk-note-input">{{ $t('calendar.deskNote') }}</label>
          <textarea
            id="calendar-desk-note-input"
            v-model="deskNote"
            class="pro-textarea"
            rows="4"
            data-testid="calendar-desk-note-input"
            :placeholder="$t('calendar.deskNotePlaceholder')"
          />
          <div class="pro-flex-gap" style="margin-top: 0.5rem">
            <ProButton
              variant="secondary"
              :disabled="deskNoteBusy || !selectedVisit.id"
              test-id="calendar-desk-note-save"
              @click="saveDeskNote"
            >
              {{ $t('calendar.deskNoteSave') }}
            </ProButton>
          </div>
          <p v-if="deskNoteMsg" class="pro-hint">{{ deskNoteMsg }}</p>
        </div>
        <div class="pro-flex-gap create-client-actions">
          <ProButton
            v-if="consultationCta === 'start'"
            :disabled="busyId === selectedVisit.id"
            test-id="calendar-open-consultation"
            @click="openConsultationFromDetail"
          >
            {{ $t('calendar.openConsultation') }}
          </ProButton>
          <ProButton
            v-else-if="consultationCta === 'view'"
            variant="secondary"
            test-id="calendar-view-consultation"
            @click="viewConsultationFromDetail"
          >
            {{ $t('calendar.viewConsultation') }}
          </ProButton>
          <label
            v-if="selectedVisit.status === 'requested' && selectedVisit.pendingActionBy === 'vet'"
            class="pro-checkbox-label"
            data-testid="calendar-request-preconsult"
          >
            <input v-model="requestPreconsult" type="checkbox" class="pro-checkbox">
            {{ $t('calendar.requestPreconsult') }}
          </label>
          <ProIconAction
            v-if="selectedVisit.status === 'requested' && selectedVisit.pendingActionBy === 'vet'"
            icon="check"
            :label="$t('calendar.confirm')"
            :disabled="busyId === selectedVisit.id"
            @click="actFromDetail('confirm')"
          />
          <ProIconAction
            v-if="selectedVisit.status === 'reschedule_pending' && selectedVisit.pendingActionBy === 'vet'"
            icon="event_available"
            :label="$t('calendar.acceptReschedule')"
            :disabled="busyId === selectedVisit.id"
            @click="actFromDetail('accept_reschedule')"
          />
          <ProIconAction
            v-if="selectedVisit.status === 'reschedule_pending' && selectedVisit.pendingActionBy === 'vet'"
            icon="close"
            :label="$t('calendar.rejectReschedule')"
            :disabled="busyId === selectedVisit.id"
            @click="actFromDetail('reject_reschedule')"
          />
          <ProIconAction
            v-if="canSendPreconsult"
            icon="mail"
            :label="$t('calendar.sendPreconsult')"
            :disabled="busyId === selectedVisit.id"
            test-id="calendar-send-preconsult"
            @click="actFromDetail('send_preconsult')"
          />
          <ProIconAction
            v-if="selectedVisit.status === 'confirmed' && !selectedVisit.consultationSession"
            icon="event"
            :label="$t('calendar.changeTime')"
            :disabled="busyId === selectedVisit.id"
            test-id="calendar-change-time"
            @click="openReschedule(selectedVisit)"
          />
          <ProIconAction
            v-if="selectedVisit.status === 'confirmed' && selectedVisit.consultationSession && canWriteClinical"
            icon="task_alt"
            :label="$t('calendar.markDone')"
            :disabled="busyId === selectedVisit.id"
            test-id="calendar-walkin-done-detail"
            @click="actFromDetail('done')"
          />
          <ProIconAction
            v-if="!selectedVisit.consultationSession && !selectedVisit.waitingRoomAt"
            icon="hourglass_top"
            :label="$t('calendar.waitingRoomOn')"
            :disabled="busyId === selectedVisit.id"
            test-id="calendar-waiting-room-on"
            @click="actFromDetail('mark_waiting_room')"
          />
          <ProIconAction
            v-if="!selectedVisit.consultationSession && selectedVisit.waitingRoomAt"
            icon="hourglass_bottom"
            :label="$t('calendar.waitingRoomOff')"
            :disabled="busyId === selectedVisit.id"
            test-id="calendar-waiting-room-off"
            @click="actFromDetail('clear_waiting_room')"
          />
          <ProIconAction
            v-if="!selectedVisit.consultationSession"
            icon="cancel"
            variant="danger"
            :label="canWriteClinical ? $t('calendar.cancel') : $t('calendar.deleteVisit')"
            :disabled="busyId === selectedVisit.id"
            test-id="calendar-delete-visit"
            @click="actFromDetail('cancel')"
          />
        </div>
      </div>
    </ProModal>

    <ProModal v-model:open="rescheduleOpen" :title="$t('calendar.changeTime')">
      <form class="pro-form" @submit.prevent="submitReschedule">
        <ProInput
          v-model="rescheduleAt"
          type="datetime-local"
          :label="$t('calendar.newSlot')"
          required
        />
        <fieldset class="pro-fieldset" data-testid="calendar-reschedule-mode">
          <legend class="pro-label">{{ $t('calendar.rescheduleMode') }}</legend>
          <label class="pro-checkbox-label">
            <input v-model="rescheduleMode" type="radio" value="propose" class="pro-radio">
            {{ $t('calendar.rescheduleModePropose') }}
          </label>
          <label class="pro-checkbox-label">
            <input v-model="rescheduleMode" type="radio" value="direct" class="pro-radio">
            {{ $t('calendar.rescheduleModeDirect') }}
          </label>
        </fieldset>
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
            test-id="calendar-reschedule-submit"
          >
            {{ rescheduleMode === 'direct' ? $t('calendar.rescheduleConfirmDirect') : $t('calendar.sendPropose') }}
          </ProButton>
        </div>
      </form>
    </ProModal>

    <ProNewAppointmentModal
      v-model:open="newApptOpen"
      :visits="visits"
      :default-day="newApptDay"
      @created="onAppointmentCreated"
    />
  </div>
</template>

<script setup lang="ts">
import type { CalendarDeskAction, CalendarVacation, CalendarVisit } from '~/composables/useCalendarGrid'
import { visitConsultationCta } from '~/composables/useCalendarGrid'
import { useActiveConsultation } from '~/composables/useActiveConsultation'
import {
  extractTeamMembersList,
  mapTeamMembersForCalendar,
  type CalendarTeamMember,
} from '~/utils/calendarTeam'

// Client-only: date grids + view toggle are click-dead during SSR hydration races
// (Playwright / fast clicks hit static HTML before Vue binds @click).
definePageMeta({
  middleware: ['vet-only', 'practice-perm'],
  practicePerm: 'calendar.manage',
  ssr: false,
})

type CalendarViewMode = 'day' | 'week' | 'month'
type ResourceMode = 'people' | 'rooms'
type TeamMemberRow = CalendarTeamMember
type RoomRow = { id: string; name: string; active?: boolean }

const route = useRoute()
const { t } = useI18n()
const { formatDate, dateLocale } = useFormatters()
const { mapError } = useApiError()
const { canPractice } = usePracticePerms()
const {
  withSiteQuery,
  concreteSiteId,
  initFromStorage,
  multiSite,
  isAggregatedView,
  sitesUiEnabled,
} = usePracticeSites()
const activeConsult = useActiveConsultation()
const canWriteClinical = computed(() => canPractice('pets.write_clinical'))
const canViewConsultations = computed(() => canPractice('consultations.history.read'))
const {
  startOfDay,
  startOfWeek,
  startOfMonth,
  monthGridRange,
  visitDisplayAt,
  statusVariant,
  dayKey,
} = useCalendarGrid()

const pending = ref<CalendarVisit[]>([])
const visits = ref<CalendarVisit[]>([])
const vacations = ref<CalendarVacation[]>([])
const clientBookingEnabled = ref(false)
const actionError = ref('')
const actionSuccess = ref('')
const busyId = ref('')
const requestPreconsult = ref(false)
const focusVisitId = ref('')
const anchorDate = ref(startOfDay(new Date()))
const newApptOpen = ref(false)
const newApptDay = ref('')

function openNewAppointment(day?: Date) {
  newApptDay.value = day ? dayKey(startOfDay(day)) : dayKey(startOfDay(new Date()))
  newApptOpen.value = true
}

async function onAppointmentCreated() {
  actionSuccess.value = t('calendar.newAppointmentCreated')
  await load()
}

const viewMode = ref<CalendarViewMode>((() => {
  if (!import.meta.client) return 'week'
  const saved = localStorage.getItem('pf-calendar-view')
  if (saved === 'day' || saved === 'week' || saved === 'month') return saved
  return 'week'
})())
/** True once the page component is mounted (clicks safe — avoids dead SSR/shell DOM). */
const calendarReady = ref(false)

const resourceMode = ref<ResourceMode>('people')
const teamMembers = ref<TeamMemberRow[]>([])
const siteRooms = ref<RoomRow[]>([])
const editAssigneeId = ref('')
const editRoomId = ref('')
const resourcesBusy = ref(false)
const detailRooms = computed(() => siteRooms.value.filter((r) => r.active !== false))

/** Include current assignee even if filtered out by defaultSiteId (orphan column case). */
const detailAssigneeOptions = computed(() => {
  const opts = [...teamMembers.value]
  const cur = selectedVisit.value
  if (cur?.assigneeUserId && !opts.some((m) => m.id === cur.assigneeUserId)) {
    opts.push({
      id: cur.assigneeUserId,
      fullName: cur.assigneeName || cur.assigneeUserId,
      defaultSiteId: '',
    })
  }
  return opts
})

/** Include current room even if deactivated / missing from active list. */
const detailRoomOptions = computed(() => {
  const opts = [...detailRooms.value]
  const cur = selectedVisit.value
  if (cur?.roomId && !opts.some((r) => r.id === cur.roomId)) {
    opts.push({
      id: cur.roomId,
      name: cur.roomName || cur.roomId,
      active: false,
    })
  }
  return opts
})

/** People columns in day view even when sites UI is off; rooms toggle only with sites. */
const showDayPeopleResources = computed(
  () => viewMode.value === 'day' && !isAggregatedView.value,
)
const showResourceToggle = computed(
  () => showDayPeopleResources.value && sitesUiEnabled.value,
)
const dayResourceMode = computed<ResourceMode | null>(() => {
  if (!showDayPeopleResources.value) return null
  if (!sitesUiEnabled.value) return 'people'
  return resourceMode.value
})

watch(
  [sitesUiEnabled, showDayPeopleResources],
  () => {
    if (!sitesUiEnabled.value && resourceMode.value === 'rooms') {
      resourceMode.value = 'people'
    }
  },
  { immediate: true },
)

const detailOpen = ref(false)
const selectedVisit = ref<CalendarVisit | null>(null)
const visitAddress = ref('')
const addressBusy = ref(false)
const addressMsg = ref('')
const deskNote = ref('')
const deskNoteBusy = ref(false)
const deskNoteMsg = ref('')
const preconsult = ref<{
  status?: string
  aiUrgency?: string
  aiSummary?: string
  answers?: {
    chiefComplaint?: string
    duration?: string
    behavior?: string
    appetite?: string
    thirst?: string
    elimination?: string
    urgency?: string
    comment?: string
  }
} | null>(null)

const preconsultIsUrgent = computed(() => {
  return (
    preconsult.value?.answers?.urgency === 'high'
    || preconsult.value?.aiUrgency === 'red'
    || selectedVisit.value?.preconsultAlert === 'urgent'
  )
})

const preconsultAiLevelLabel = computed(() => {
  const level = preconsult.value?.aiUrgency
  if (!level) return ''
  const key = `calendar.preconsultAiLevels.${level}`
  const translated = t(key)
  return translated === key ? level : translated
})

const preconsultAiBadgeVariant = computed(() => {
  switch (preconsult.value?.aiUrgency) {
    case 'red':
      return 'danger'
    case 'orange':
      return 'warning'
    case 'green':
      return 'success'
    default:
      return 'neutral'
  }
})

const preconsultStatusLabel = computed(() => {
  const st = preconsult.value?.status || selectedVisit.value?.preconsultStatus || ''
  switch (st) {
    case 'submitted':
      return t('calendar.preconsultSubmitted')
    case 'pending':
      return t('calendar.preconsultPending')
    case 'skipped':
      return t('calendar.preconsultSkipped')
    default:
      return t('calendar.preconsultNone')
  }
})

const preconsultBadgeVariant = computed(() => {
  if (preconsultIsUrgent.value) return 'danger'
  const st = preconsult.value?.status || selectedVisit.value?.preconsultStatus || ''
  switch (st) {
    case 'submitted':
      return 'success'
    case 'pending':
      return 'warning'
    default:
      return 'neutral'
  }
})

function preconsultEnumLabel(kind: 'duration' | 'behavior' | 'scale' | 'urgency', value?: string) {
  if (!value) return '—'
  const key = `calendar.preconsultEnums.${kind}.${value}`
  const translated = t(key)
  return translated === key ? value : translated
}
const rescheduleOpen = ref(false)
const rescheduleVisitId = ref('')
const rescheduleAt = ref('')
const rescheduleError = ref('')
const rescheduleMode = ref<'propose' | 'direct'>('propose')

const canSendPreconsult = computed(() => {
  const v = selectedVisit.value
  if (!v || v.consultationSession || v.status !== 'confirmed') return false
  const st = preconsult.value?.status || v.preconsultStatus || ''
  return !st
})

/** CTA consultation : écran « Nouvelle consultation » (RDV à venir) ou lecture du CR (RDV passé). */
const consultationCta = computed<'start' | 'view' | null>(() => {
  const v = selectedVisit.value
  if (!v) return null
  const cta = visitConsultationCta(v)
  if (cta === 'start') return canWriteClinical.value ? 'start' : null
  if (cta === 'view') return canViewConsultations.value ? 'view' : null
  return null
})

function openConsultationFromDetail() {
  const v = selectedVisit.value
  if (!v?.clientId) return
  detailOpen.value = false
  activeConsult.openForVisit({ visitId: v.id, clientId: v.clientId, petId: v.petId })
}

async function viewConsultationFromDetail() {
  const v = selectedVisit.value
  if (!v?.clientId) return
  detailOpen.value = false
  activeConsult.openForVisit({ visitId: v.id, clientId: v.clientId, petId: v.petId })
}

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

const agendaTitle = computed(() => {
  switch (viewMode.value) {
    case 'day':
      return t('calendar.dayTitle')
    case 'week':
      return t('calendar.weekTitle')
    case 'month':
      return t('calendar.monthTitle')
    default: {
      const _exhaustive: never = viewMode.value
      return _exhaustive
    }
  }
})

const navPrevLabel = computed(() => {
  switch (viewMode.value) {
    case 'day':
      return t('calendar.prevDay')
    case 'week':
      return t('calendar.prevWeek')
    case 'month':
      return t('calendar.prevMonth')
    default: {
      const _exhaustive: never = viewMode.value
      return _exhaustive
    }
  }
})

const navNextLabel = computed(() => {
  switch (viewMode.value) {
    case 'day':
      return t('calendar.nextDay')
    case 'week':
      return t('calendar.nextWeek')
    case 'month':
      return t('calendar.nextMonth')
    default: {
      const _exhaustive: never = viewMode.value
      return _exhaustive
    }
  }
})

const periodLabel = computed(() => {
  if (viewMode.value === 'day') {
    return anchorDate.value.toLocaleDateString(dateLocale(), {
      weekday: 'long',
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    })
  }
  if (viewMode.value === 'week') {
    const end = new Date(weekStart.value)
    end.setDate(end.getDate() + 6)
    const opts: Intl.DateTimeFormatOptions = { day: 'numeric', month: 'short', year: 'numeric' }
    return `${weekStart.value.toLocaleDateString(dateLocale(), opts)} – ${end.toLocaleDateString(dateLocale(), opts)}`
  }
  return monthStart.value.toLocaleDateString(dateLocale(), { month: 'long', year: 'numeric' })
})

const canEditVisitResources = computed(() => {
  const v = selectedVisit.value
  if (!v || v.status === 'cancelled' || v.status === 'done') return false
  return true
})

const dayColumns = computed(() => {
  if (!showDayPeopleResources.value) {
    return [{ id: '', label: periodLabel.value }]
  }
  const dayKeyNow = dayKey(startOfDay(anchorDate.value))
  const dayVisits = visits.value.filter((v) => {
    const at = visitDisplayAt(v)
    return !!at && dayKey(at) === dayKeyNow
  })
  if (dayResourceMode.value === 'rooms') {
    const cols: { id: string; label: string }[] = [
      { id: '', label: t('calendar.noRoom') },
      ...detailRooms.value.map((r) => ({ id: r.id, label: r.name })),
    ]
    const known = new Set(cols.map((c) => c.id))
    for (const v of dayVisits) {
      const rid = v.roomId || ''
      if (!rid || known.has(rid)) continue
      known.add(rid)
      cols.push({ id: rid, label: v.roomName || rid })
    }
    return cols
  }
  const cols: { id: string; label: string }[] = [
    { id: '', label: t('calendar.unassigned') },
    ...teamMembers.value.map((m) => ({ id: m.id, label: m.fullName })),
  ]
  const known = new Set(cols.map((c) => c.id))
  for (const v of dayVisits) {
    const aid = v.assigneeUserId || ''
    if (!aid || known.has(aid)) continue
    known.add(aid)
    cols.push({ id: aid, label: v.assigneeName || aid })
  }
  return cols
})

const todayVisitCount = ref(0)

function setView(mode: CalendarViewMode) {
  viewMode.value = mode
  if (import.meta.client) localStorage.setItem('pf-calendar-view', mode)
  void load()
}

function shiftPeriod(delta: number) {
  if (viewMode.value === 'day') {
    const n = new Date(anchorDate.value)
    n.setDate(n.getDate() + delta)
    anchorDate.value = startOfDay(n)
  } else if (viewMode.value === 'week') {
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
  setView('day')
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
  if (viewMode.value === 'day') {
    const from = startOfDay(anchorDate.value)
    const to = new Date(from)
    to.setDate(to.getDate() + 1)
    return { from, to }
  }
  if (viewMode.value === 'week') {
    const from = weekStart.value
    const to = new Date(from)
    to.setDate(to.getDate() + 7)
    return { from, to }
  }
  const { gridStart, gridEnd } = monthGridRange(monthStart.value)
  return { from: gridStart, to: gridEnd }
}

async function loadDayResources() {
  if (isAggregatedView.value) {
    teamMembers.value = []
    siteRooms.value = []
    return
  }
  const siteId = concreteSiteId.value
  try {
    const teamRes: any = await $fetch('/api/vet/team')
    teamMembers.value = mapTeamMembersForCalendar(
      extractTeamMembersList(teamRes),
      sitesUiEnabled.value ? siteId : '',
    )
    if (sitesUiEnabled.value && siteId) {
      const roomsRes: any = await $fetch(`/api/vet/sites/${encodeURIComponent(siteId)}/rooms`)
      const roomsList = roomsRes?.data ?? roomsRes ?? []
      siteRooms.value = (Array.isArray(roomsList) ? roomsList : [])
        .filter((r: any) => r?.id)
        .map((r: any) => ({
          id: String(r.id),
          name: String(r.name || r.id),
          active: r.active !== false,
        }))
    }
    else {
      siteRooms.value = []
    }
  } catch {
    teamMembers.value = []
    siteRooms.value = []
  }
}

async function load() {
  actionError.value = ''
  const { from, to } = rangeForView()
  const day0 = startOfDay(new Date())
  const day1 = new Date(day0)
  day1.setDate(day1.getDate() + 1)
  const todayFrom = toLocalRFC3339(day0)
  const todayTo = toLocalRFC3339(day1)
  initFromStorage()
  try {
    const fetches: Promise<any>[] = [
      ($fetch as any)(withSiteQuery(`/api/vet/calendar?from=${encodeURIComponent(toLocalRFC3339(from))}&to=${encodeURIComponent(toLocalRFC3339(to))}`)),
      ($fetch as any)(withSiteQuery('/api/vet/schedule', concreteSiteId.value)),
      ($fetch as any)(withSiteQuery(`/api/vet/calendar?from=${encodeURIComponent(todayFrom)}&to=${encodeURIComponent(todayTo)}`)),
    ]
    if (viewMode.value === 'day' && !isAggregatedView.value) {
      fetches.push(loadDayResources())
    }
    const [calRes, schedRes, todayRes] = await Promise.all(fetches)
    const cal = calRes.data ?? calRes
    pending.value = cal.pending ?? []
    visits.value = cal.visits ?? []
    vacations.value = cal.vacations ?? []
    const todayCal = todayRes.data ?? todayRes
    todayVisitCount.value = (todayCal.visits ?? []).filter(
      (v: CalendarVisit) => !v.consultationSession,
    ).length
    const sched = schedRes.data ?? schedRes
    clientBookingEnabled.value = !!sched.clientBookingEnabled
  } catch (e: any) {
    actionSuccess.value = ''
    actionError.value = mapError(e)
  }
}

async function act(id: string, action: string) {
  if (action === 'cancel') {
    const msg = canWriteClinical.value ? t('calendar.cancelConfirm') : t('calendar.deleteConfirm')
    if (!window.confirm(msg)) return
  }
  if (action === 'reject_reschedule' && !window.confirm(t('calendar.rejectRescheduleConfirm'))) return
  busyId.value = id
  actionError.value = ''
  try {
    const body: Record<string, unknown> = { action }
    if (action === 'confirm' && requestPreconsult.value) {
      body.requestPreconsult = true
    }
    const res: any = await $fetch(`/api/visits/${id}`, { method: 'PATCH', body })
    const data = res?.data ?? res
    if (action === 'mark_waiting_room' || action === 'clear_waiting_room' || action === 'send_preconsult') {
      if (selectedVisit.value?.id === id) {
        selectedVisit.value = {
          ...selectedVisit.value,
          ...data,
          waitingRoomAt:
            data?.waitingRoomAt
            ?? (action === 'clear_waiting_room' ? null : selectedVisit.value.waitingRoomAt),
          preconsultStatus: data?.preconsultStatus || selectedVisit.value.preconsultStatus,
        }
        if (action === 'send_preconsult') {
          await loadVisitPreconsult(id)
        }
      }
      await load()
      return
    }
    detailOpen.value = false
    requestPreconsult.value = false
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

function onChipVisitAction(payload: { visit: CalendarVisit; action: CalendarDeskAction }) {
  void act(payload.visit.id, payload.action)
}

function openVisitDetail(v: CalendarVisit) {
  selectedVisit.value = v
  visitAddress.value = v.addressText || ''
  addressMsg.value = ''
  deskNote.value = v.notes || ''
  deskNoteMsg.value = ''
  editAssigneeId.value = v.assigneeUserId || ''
  editRoomId.value = v.roomId || ''
  preconsult.value = null
  focusVisitId.value = v.id
  detailOpen.value = true
  void loadVisitPreconsult(v.id)
  void ensureResourcesForDetail(v.siteId || concreteSiteId.value)
}

async function ensureResourcesForDetail(siteId: string) {
  const needTeam = !teamMembers.value.length
  const needRooms = sitesUiEnabled.value && (!siteRooms.value.length || (siteId && siteId !== concreteSiteId.value))
  if (!needTeam && !needRooms) return
  try {
    const teamRes: any = await $fetch('/api/vet/team')
    teamMembers.value = mapTeamMembersForCalendar(
      extractTeamMembersList(teamRes),
      sitesUiEnabled.value ? siteId : '',
    )
    if (sitesUiEnabled.value && siteId) {
      const roomsRes: any = await $fetch(`/api/vet/sites/${encodeURIComponent(siteId)}/rooms`)
      const roomsList = roomsRes?.data ?? roomsRes ?? []
      siteRooms.value = (Array.isArray(roomsList) ? roomsList : [])
        .filter((r: any) => r?.id)
        .map((r: any) => ({
          id: String(r.id),
          name: String(r.name || r.id),
          active: r.active !== false,
        }))
    }
  } catch { /* keep previous */ }
}

async function saveVisitResources() {
  if (!selectedVisit.value) return
  resourcesBusy.value = true
  actionError.value = ''
  try {
    const res: any = await $fetch(`/api/visits/${selectedVisit.value.id}`, {
      method: 'PATCH',
      body: {
        action: 'set_resources',
        assigneeUserId: editAssigneeId.value,
        roomId: editRoomId.value,
      },
    })
    const data = res?.data ?? res
    selectedVisit.value = {
      ...selectedVisit.value,
      ...data,
      assigneeUserId: data?.assigneeUserId ?? editAssigneeId.value,
      assigneeName: data?.assigneeName
        ?? teamMembers.value.find((m) => m.id === editAssigneeId.value)?.fullName
        ?? '',
      roomId: data?.roomId ?? editRoomId.value,
      roomName: data?.roomName
        ?? siteRooms.value.find((r) => r.id === editRoomId.value)?.name
        ?? '',
    }
    actionSuccess.value = t('calendar.resourcesSaved')
    await load()
  } catch (e: any) {
    actionError.value = mapError(e)
  } finally {
    resourcesBusy.value = false
  }
}

async function saveDeskNote() {
  if (!selectedVisit.value) return
  deskNoteBusy.value = true
  deskNoteMsg.value = ''
  try {
    const res: any = await $fetch(`/api/visits/${selectedVisit.value.id}/notes`, {
      method: 'PATCH',
      body: { notes: deskNote.value },
    })
    const data = res.data ?? res
    selectedVisit.value = { ...selectedVisit.value, notes: data.notes ?? deskNote.value }
    deskNoteMsg.value = t('calendar.deskNoteSaved')
    await load()
  } catch (e: any) {
    deskNoteMsg.value = mapError(e)
  } finally {
    deskNoteBusy.value = false
  }
}

async function loadVisitPreconsult(visitId: string) {
  try {
    const res: any = await $fetch(`/api/visits/${visitId}/preconsult`)
    preconsult.value = res.data ?? res
    if (selectedVisit.value?.id === visitId && preconsult.value?.status) {
      const alert =
        preconsult.value.aiUrgency === 'red' || preconsult.value.answers?.urgency === 'high'
          ? 'urgent'
          : selectedVisit.value.preconsultAlert
      selectedVisit.value = {
        ...selectedVisit.value,
        preconsultStatus: preconsult.value.status,
        preconsultAlert: alert || undefined,
      }
      const patch = (list: CalendarVisit[]) => {
        const idx = list.findIndex((v) => v.id === visitId)
        if (idx < 0) return list
        const next = list.slice()
        next[idx] = {
          ...next[idx],
          preconsultStatus: preconsult.value?.status,
          preconsultAlert: alert || undefined,
        }
        return next
      }
      visits.value = patch(visits.value)
      pending.value = patch(pending.value)
    }
  } catch {
    preconsult.value = null
  }
}

async function saveVisitAddress() {
  if (!selectedVisit.value) return
  addressBusy.value = true
  addressMsg.value = ''
  try {
    const nextAddress = visitAddress.value.trim()
    const prevAddress = (selectedVisit.value.addressText || '').trim()
    const clearCoords = nextAddress !== prevAddress &&
      (selectedVisit.value.lat != null || selectedVisit.value.lng != null)
    const res: any = await $fetch(`/api/visits/${selectedVisit.value.id}/location`, {
      method: 'PATCH',
      body: {
        addressText: nextAddress,
        ...(clearCoords ? { clearCoords: true } : {}),
      },
    })
    const data = res.data ?? res
    selectedVisit.value = {
      ...selectedVisit.value,
      addressText: data.addressText || nextAddress,
      lat: clearCoords ? (data.lat ?? null) : (data.lat ?? selectedVisit.value.lat),
      lng: clearCoords ? (data.lng ?? null) : (data.lng ?? selectedVisit.value.lng),
    }
    addressMsg.value = t('calendar.addressSaved')
    await load()
  } catch (e: any) {
    addressMsg.value = mapError(e)
  } finally {
    addressBusy.value = false
  }
}

function openReschedule(v: CalendarVisit) {
  rescheduleVisitId.value = v.id
  rescheduleAt.value = ''
  rescheduleError.value = ''
  rescheduleMode.value = 'propose'
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
    const action = rescheduleMode.value === 'direct' ? 'reschedule_direct' : 'propose_reschedule'
    await $fetch(`/api/visits/${rescheduleVisitId.value}`, {
      method: 'PATCH',
      body: { action, proposedScheduledAt: iso },
    })
    rescheduleOpen.value = false
    rescheduleError.value = ''
    await load()
    actionSuccess.value = action === 'reschedule_direct'
      ? t('calendar.rescheduleDirectDone')
      : t('calendar.proposeSent')
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
  calendarReady.value = true
  await load()
  if (typeof route.query.visit === 'string' && route.query.visit) {
    await focusVisit(route.query.visit)
  }
  if (import.meta.client) {
    window.addEventListener('pf-site-changed', onSiteChanged as EventListener)
  }
})

onUnmounted(() => {
  if (import.meta.client) {
    window.removeEventListener('pf-site-changed', onSiteChanged as EventListener)
  }
})

function onSiteChanged() {
  void load()
}

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
.calendar-today-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.25rem;
  margin-left: 0.35rem;
  padding: 0 0.35rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 600;
  background: color-mix(in srgb, var(--pf-vet-accent) 18%, transparent);
  color: var(--pf-vet-accent);
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
.visit-resources {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.75rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius);
  background: var(--pf-vet-bg);
}
.visit-type-dot {
  display: inline-block;
  width: 0.65rem;
  height: 0.65rem;
  border-radius: 2px;
  margin: 0 0.35rem 0 0.15rem;
  vertical-align: middle;
}
.preconsult-answers {
  margin: 0.75rem 0 1rem;
  padding: 0.75rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius-md, 8px);
  background: var(--pf-vet-surface);
}
.preconsult-dl {
  margin: 0.5rem 0 0;
  display: grid;
  gap: 0.5rem;
}
.preconsult-dl dt {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--pf-vet-muted, #64748b);
}
.preconsult-dl dd {
  margin: 0.15rem 0 0;
}
.preconsult-ai {
  margin-top: 1rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--pf-vet-border);
}
.pro-inline-feedback--error {
  background: color-mix(in srgb, var(--pf-vet-alert) 10%, var(--pf-vet-surface));
  border-color: color-mix(in srgb, var(--pf-vet-alert) 35%, transparent);
  color: var(--pf-vet-alert);
}
</style>
