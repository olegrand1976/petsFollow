<template>
  <div data-testid="calendar-grid" class="cal-week" :class="{ 'cal-week--stack': stackMobile }">
    <div class="cal-week__header">
      <div
        v-for="cell in dayCells"
        :key="cell.key"
        class="cal-week__head"
        :class="{
          'cal-week__head--today': cell.key === todayKey,
          'cal-week__head--vacation': !!cell.vacation,
        }"
      >
        <span class="cal-week__weekday">{{ cell.weekday }}</span>
        <span class="cal-week__date">{{ cell.day.getDate() }}</span>
        <span v-if="cell.vacation" class="cal-week__vac-label">
          {{ cell.vacation.label || $t('calendar.vacation') }}
        </span>
      </div>
    </div>
    <div class="cal-week__body">
      <div
        v-for="cell in dayCells"
        :key="`body-${cell.key}`"
        class="cal-week__col"
        :class="{
          'cal-week__col--today': cell.key === todayKey,
          'cal-week__col--vacation': !!cell.vacation,
        }"
      >
        <p v-if="stackMobile" class="cal-week__mobile-day">
          {{ cell.weekday }} {{ cell.day.getDate() }}
        </p>
        <button
          v-for="v in cell.visits"
          :id="`calendar-chip-${v.id}`"
          :key="v.id"
          type="button"
          class="cal-chip cal-chip--column"
          :class="[
            `cal-chip--${statusVariant(v.status)}`,
            {
              'cal-chip--focus': focusVisitId === v.id,
              'cal-chip--walkin': !!v.consultationSession,
              'cal-chip--typed': !!v.visitTypeColor,
              'cal-chip--urgent': v.preconsultAlert === 'urgent',
              'cal-chip--waiting': !!v.waitingRoomAt,
            },
          ]"
          :style="v.visitTypeColor ? { '--cal-type-color': v.visitTypeColor } : undefined"
          :title="chipTitle(v) || undefined"
          :data-testid="`calendar-chip-${v.id}`"
          @click="emit('select-visit', v)"
        >
          <span class="cal-chip__time">{{ chipTime(v) }}</span>
          <span v-if="v.visitTypeName" class="cal-chip__type">{{ v.visitTypeName }}</span>
          <span
            v-if="showSiteLabel && v.siteName"
            class="cal-chip__site"
            data-testid="calendar-chip-site"
          >{{ v.siteName }}</span>
          <span v-if="v.consultationSession" class="cal-chip__walkin">{{ $t('calendar.walkInShort') }}</span>
          <span
            v-if="v.waitingRoomAt"
            class="cal-chip__waiting"
            data-testid="calendar-chip-waiting-room"
            :title="$t('calendar.waitingRoomTooltip')"
          >{{ $t('calendar.waitingRoomTag') }}</span>
          <span
            v-if="v.preconsultAlert === 'urgent'"
            class="cal-chip__urgent"
            data-testid="calendar-chip-preconsult-urgent"
            :title="$t('calendar.preconsultUrgentTooltip')"
          >{{ $t('calendar.preconsultUrgentShort') }}</span>
          <span
            v-else-if="v.preconsultStatus === 'submitted'"
            class="cal-chip__preconsult cal-chip__preconsult--answered"
            data-testid="calendar-chip-preconsult-answered"
            :title="$t('calendar.preconsultAnsweredTooltip')"
          >{{ $t('calendar.preconsultAnsweredTag') }}</span>
          <span
            v-else-if="v.preconsultStatus === 'pending'"
            class="cal-chip__preconsult cal-chip__preconsult--sent"
            data-testid="calendar-chip-preconsult-sent"
            :title="$t('calendar.preconsultSentTooltip')"
          >{{ $t('calendar.preconsultSentTag') }}</span>
          <span class="cal-chip__title">
            {{ v.petName || '—' }} · {{ v.clientName || '—' }}
            <span v-if="v.isWalkinPlaceholder" class="cal-chip__walkin" data-testid="calendar-chip-walkin">!</span>
          </span>
          <span
            v-if="v.callbackPhone"
            class="cal-chip__phone"
            data-testid="calendar-chip-phone"
            role="link"
            tabindex="0"
            :title="$t('clients.walkin.callbackPhone')"
            @click.stop="dialCallbackPhone(v.callbackPhone!)"
            @keydown.enter.stop="dialCallbackPhone(v.callbackPhone!)"
          >{{ v.callbackPhone }}</span>
          <span v-if="v.addressText" class="cal-chip__place">{{ v.addressText }}</span>
        </button>
        <p v-if="!cell.visits.length" class="cal-week__empty">—</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { calendarChipTooltip, dialCallbackPhone, type CalendarVacation, type CalendarVisit } from '~/composables/useCalendarGrid'

const props = defineProps<{
  weekStart: Date
  visits: CalendarVisit[]
  vacations: CalendarVacation[]
  focusVisitId?: string
  /** Show site name on chips (multi-site practices). */
  showSiteLabel?: boolean
}>()

const emit = defineEmits<{ 'select-visit': [visit: CalendarVisit] }>()

const { t } = useI18n()
const {
  dayKey,
  weekDays,
  visitsByDay,
  vacationOnDay,
  visitDisplayAt,
  isUnscheduled,
  statusVariant,
  startOfDay,
} = useCalendarGrid()

function chipTitle(v: CalendarVisit) {
  return calendarChipTooltip(v, t)
}

const byDay = computed(() => visitsByDay(props.visits))
const todayKey = dayKey(startOfDay(new Date()))

const dayCells = computed(() =>
  weekDays(props.weekStart).map((day) => {
    const key = dayKey(day)
    return {
      day,
      key,
      weekday: t(`settings.calendar.weekday.${day.getDay()}` as any),
      vacation: vacationOnDay(props.vacations, day),
      visits: byDay.value.get(key) || [],
    }
  }),
)

const stackMobile = ref(false)
let mql: MediaQueryList | null = null

function onMql() {
  stackMobile.value = !!mql?.matches
}

onMounted(() => {
  mql = window.matchMedia('(max-width: 768px)')
  onMql()
  mql.addEventListener('change', onMql)
})

onBeforeUnmount(() => {
  mql?.removeEventListener('change', onMql)
})

function chipTime(v: CalendarVisit) {
  if (isUnscheduled(v)) return t('calendar.unscheduled')
  const at = visitDisplayAt(v)
  if (!at) return '—'
  const pad = (n: number) => String(n).padStart(2, '0')
  const time = `${pad(at.getHours())}:${pad(at.getMinutes())}`
  if (v.proposedScheduledAt) return `${time} (${t('calendar.proposed')})`
  return time
}
</script>

<style scoped>
.cal-week__header,
.cal-week__body {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 0.35rem;
}
.cal-week__head {
  text-align: center;
  padding: 0.4rem 0.25rem;
  border-radius: var(--pf-vet-radius);
  border: 1px solid var(--pf-vet-border);
  background: var(--pf-vet-surface);
  font-size: 0.8rem;
}
.cal-week__head--today {
  border-color: var(--pf-vet-accent);
}
.cal-week__head--vacation,
.cal-week__col--vacation {
  background: color-mix(in srgb, var(--pf-vet-border) 45%, transparent);
}
.cal-week__weekday {
  display: block;
  color: var(--pf-vet-text-muted);
  font-size: 0.7rem;
  text-transform: uppercase;
}
.cal-week__date {
  font-weight: 600;
}
.cal-week__vac-label {
  display: block;
  font-size: 0.65rem;
  color: var(--pf-vet-text-muted);
  margin-top: 0.15rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.cal-week__col {
  min-height: 8rem;
  padding: 0.35rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius);
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}
.cal-week__col--today {
  box-shadow: inset 0 0 0 1px var(--pf-vet-accent);
}
.cal-week__empty {
  margin: 0;
  text-align: center;
  color: var(--pf-vet-text-muted);
  font-size: 0.85rem;
  padding: 0.5rem 0;
}
.cal-week__mobile-day {
  display: none;
  margin: 0 0 0.25rem;
  font-weight: 600;
  font-size: 0.85rem;
}
.cal-chip__place {
  display: block;
  font-size: 0.65rem;
  color: var(--pf-vet-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 768px) {
  .cal-week--stack .cal-week__header {
    display: none;
  }
  .cal-week--stack .cal-week__body {
    grid-template-columns: 1fr;
  }
  .cal-week--stack .cal-week__mobile-day {
    display: block;
  }
  .cal-week--stack .cal-week__col {
    min-height: 0;
  }
}
.cal-chip__walkin {
  display: inline-block;
  margin-left: 0.2rem;
  color: var(--pf-vet-alert);
  font-weight: 700;
}
</style>
