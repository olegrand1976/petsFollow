<template>
  <div
    data-testid="calendar-day-grid"
    class="cal-day"
    :style="{ '--cal-day-cols': String(columns.length || 1) }"
  >
    <div class="cal-day__header">
      <div class="cal-day__gutter" aria-hidden="true" />
      <div
        v-for="col in columns"
        :key="col.id || '__empty'"
        class="cal-day__head"
        :data-testid="`calendar-day-col-${col.id || 'empty'}`"
      >
        {{ col.label }}
      </div>
    </div>
    <div class="cal-day__body">
      <div
        v-for="hour in hours"
        :key="hour"
        class="cal-day__row"
      >
        <div class="cal-day__gutter">
          <span class="cal-day__hour">{{ formatHour(hour) }}</span>
        </div>
        <div
          v-for="col in columns"
          :key="`${hour}-${col.id || '__empty'}`"
          class="cal-day__cell"
        >
          <button
            v-for="v in visitsFor(col.id, hour)"
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
              v-if="resourceMode === 'rooms' && v.assigneeName"
              class="cal-chip__site"
            >{{ v.assigneeName }}</span>
            <span
              v-else-if="resourceMode === 'people' && v.roomName"
              class="cal-chip__site"
            >{{ v.roomName }}</span>
            <span v-if="v.consultationSession" class="cal-chip__walkin">{{ $t('calendar.walkInShort') }}</span>
            <span
              v-if="v.waitingRoomAt"
              class="cal-chip__waiting"
              data-testid="calendar-chip-waiting-room"
            >{{ $t('calendar.waitingRoomTag') }}</span>
            <span
              v-if="v.preconsultAlert === 'urgent'"
              class="cal-chip__urgent"
              data-testid="calendar-chip-preconsult-urgent"
            >{{ $t('calendar.preconsultUrgentShort') }}</span>
            <span
              v-else-if="v.preconsultStatus === 'submitted'"
              class="cal-chip__preconsult cal-chip__preconsult--answered"
              data-testid="calendar-chip-preconsult-answered"
            >{{ $t('calendar.preconsultAnsweredTag') }}</span>
            <span
              v-else-if="v.preconsultStatus === 'pending'"
              class="cal-chip__preconsult cal-chip__preconsult--sent"
              data-testid="calendar-chip-preconsult-sent"
            >{{ $t('calendar.preconsultSentTag') }}</span>
            <span class="cal-chip__title">{{ v.petName || '—' }} · {{ v.clientName || '—' }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { calendarChipTooltip, type CalendarVisit } from '~/composables/useCalendarGrid'

export type CalendarDayColumn = { id: string; label: string }
export type CalendarResourceMode = 'people' | 'rooms'

const props = withDefaults(defineProps<{
  day: Date
  visits: CalendarVisit[]
  columns: CalendarDayColumn[]
  resourceMode?: CalendarResourceMode | null
  focusVisitId?: string
  /** First hour shown (inclusive). */
  hourStart?: number
  /** Last hour shown (inclusive). */
  hourEnd?: number
}>(), {
  resourceMode: null,
  hourStart: 7,
  hourEnd: 20,
})

const emit = defineEmits<{ 'select-visit': [visit: CalendarVisit] }>()

const { t } = useI18n()
const {
  dayKey,
  visitDisplayAt,
  isUnscheduled,
  statusVariant,
  startOfDay,
} = useCalendarGrid()

const targetKey = computed(() => dayKey(startOfDay(props.day)))

const hours = computed(() => {
  const start = Math.max(0, Math.min(23, props.hourStart ?? 7))
  const end = Math.max(start, Math.min(23, props.hourEnd ?? 20))
  return Array.from({ length: end - start + 1 }, (_, i) => start + i)
})

const dayVisits = computed(() =>
  props.visits.filter((v) => {
    const at = visitDisplayAt(v)
    if (!at) return false
    return dayKey(at) === targetKey.value
  }),
)

function resourceId(v: CalendarVisit): string {
  if (props.resourceMode === 'rooms') return v.roomId || ''
  if (props.resourceMode === 'people') return v.assigneeUserId || ''
  return ''
}

function visitHour(v: CalendarVisit): number {
  const at = visitDisplayAt(v)
  if (!at) return props.hourStart ?? 7
  const h = at.getHours()
  const start = props.hourStart ?? 7
  const end = props.hourEnd ?? 20
  if (h < start) return start
  if (h > end) return end
  return h
}

function visitsFor(colId: string, hour: number) {
  return dayVisits.value.filter((v) => resourceId(v) === colId && visitHour(v) === hour)
}

function formatHour(h: number) {
  return `${String(h).padStart(2, '0')}:00`
}

function chipTitle(v: CalendarVisit) {
  return calendarChipTooltip(v, t)
}

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
.cal-day {
  overflow-x: auto;
}
.cal-day__header,
.cal-day__row {
  display: grid;
  grid-template-columns: 3.25rem repeat(var(--cal-day-cols, 1), minmax(7rem, 1fr));
  gap: 0.35rem;
}
.cal-day__header {
  margin-bottom: 0.35rem;
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--pf-vet-bg, #fff);
}
.cal-day__gutter {
  display: flex;
  align-items: flex-start;
  justify-content: flex-end;
  padding: 0.35rem 0.25rem 0 0;
}
.cal-day__hour {
  font-size: 0.7rem;
  color: var(--pf-vet-text-muted);
  font-variant-numeric: tabular-nums;
}
.cal-day__head {
  text-align: center;
  padding: 0.4rem 0.25rem;
  border-radius: var(--pf-vet-radius);
  border: 1px solid var(--pf-vet-border);
  background: var(--pf-vet-surface);
  font-size: 0.8rem;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.cal-day__body {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}
.cal-day__cell {
  min-height: 3.25rem;
  padding: 0.25rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius);
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  background: var(--pf-vet-surface);
}
</style>
