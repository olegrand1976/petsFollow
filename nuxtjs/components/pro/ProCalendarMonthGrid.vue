<template>
  <div data-testid="calendar-grid" class="cal-month">
    <div class="cal-month__weekdays">
      <span v-for="d in weekdayHeaders" :key="d" class="cal-month__weekday">{{ d }}</span>
    </div>
    <div class="cal-month__grid">
      <div
        v-for="cell in dayCells"
        :key="cell.key"
        class="cal-month__cell"
        :class="{
          'cal-month__cell--outside': cell.outside,
          'cal-month__cell--today': cell.key === todayKey,
          'cal-month__cell--vacation': !!cell.vacation,
        }"
      >
        <div class="cal-month__cell-head">
          <span class="cal-month__daynum">{{ cell.day.getDate() }}</span>
          <span
            v-if="cell.vacation"
            class="cal-month__vac"
            :title="cell.vacation.label || $t('calendar.vacation')"
          >
            {{ $t('calendar.vacation') }}
          </span>
        </div>
        <ProCalendarChip
          v-for="v in cell.visible"
          :key="v.id"
          :visit="v"
          variant="row"
          title-mode="pet"
          address-mode="badge"
          :focused="focusVisitId === v.id"
          :busy="busyId === v.id"
          :show-site-label="showSiteLabel"
          @select="emit('select-visit', v)"
          @action="(action) => emit('visit-action', { visit: v, action })"
        />
        <button
          v-if="cell.overflow > 0"
          type="button"
          class="cal-month__more"
          @click="emit('select-day', cell.day)"
        >
          {{ $t('calendar.moreCount', { n: cell.overflow }) }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { CalendarDeskAction, CalendarVacation, CalendarVisit } from '~/composables/useCalendarGrid'

const props = withDefaults(
  defineProps<{
    monthStart: Date
    visits: CalendarVisit[]
    vacations: CalendarVacation[]
    focusVisitId?: string
    busyId?: string
    maxPerDay?: number
    showSiteLabel?: boolean
  }>(),
  { maxPerDay: 3, showSiteLabel: false, busyId: '' },
)

const emit = defineEmits<{
  'select-visit': [visit: CalendarVisit]
  'select-day': [day: Date]
  'visit-action': [payload: { visit: CalendarVisit; action: CalendarDeskAction }]
}>()

const { t } = useI18n()
const { dateLocale } = useFormatters()
const {
  dayKey,
  monthGridRange,
  visitsByDay,
  vacationOnDay,
  startOfDay,
} = useCalendarGrid()

const byDay = computed(() => visitsByDay(props.visits))
const todayKey = dayKey(startOfDay(new Date()))

const weekdayHeaders = computed(() => {
  const loc = dateLocale()
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(2024, 0, 1 + i)
    return d.toLocaleDateString(loc, { weekday: 'short' })
  })
})

const dayCells = computed(() =>
  monthGridRange(props.monthStart).days.map((day) => {
    const key = dayKey(day)
    const all = byDay.value.get(key) || []
    return {
      day,
      key,
      outside: day.getMonth() !== props.monthStart.getMonth(),
      vacation: vacationOnDay(props.vacations, day),
      visible: all.slice(0, props.maxPerDay),
      overflow: Math.max(0, all.length - props.maxPerDay),
    }
  }),
)
</script>

<style scoped>
.cal-month__weekdays,
.cal-month__grid {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 0.25rem;
}
.cal-month__weekday {
  text-align: center;
  font-size: 0.7rem;
  text-transform: capitalize;
  color: var(--pf-vet-text-muted);
  padding: 0.25rem;
}
.cal-month__cell {
  min-height: 6.5rem;
  padding: 0.3rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius);
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  background: var(--pf-vet-surface);
}
.cal-month__cell--outside {
  opacity: 0.45;
}
.cal-month__cell--today {
  box-shadow: inset 0 0 0 1px var(--pf-vet-accent);
}
.cal-month__cell--vacation {
  background: color-mix(in srgb, var(--pf-vet-border) 45%, transparent);
}
.cal-month__cell-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.25rem;
}
.cal-month__daynum {
  font-weight: 600;
  font-size: 0.8rem;
}
.cal-month__vac {
  font-size: 0.6rem;
  color: var(--pf-vet-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 55%;
}
.cal-month__more {
  border: none;
  background: transparent;
  color: var(--pf-vet-accent);
  font-size: 0.7rem;
  cursor: pointer;
  padding: 0.1rem 0;
  text-align: left;
}

@media (max-width: 768px) {
  .cal-month__cell {
    min-height: 4.5rem;
  }
  :deep(.cal-chip__title) {
    display: none;
  }
}
</style>
