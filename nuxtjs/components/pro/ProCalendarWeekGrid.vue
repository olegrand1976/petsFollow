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
            { 'cal-chip--focus': focusVisitId === v.id },
          ]"
          :data-testid="`calendar-chip-${v.id}`"
          @click="emit('select-visit', v)"
        >
          <span class="cal-chip__time">{{ chipTime(v) }}</span>
          <span class="cal-chip__title">{{ v.petName || '—' }} · {{ v.clientName || '—' }}</span>
        </button>
        <p v-if="!cell.visits.length" class="cal-week__empty">—</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { CalendarVacation, CalendarVisit } from '~/composables/useCalendarGrid'

const props = defineProps<{
  weekStart: Date
  visits: CalendarVisit[]
  vacations: CalendarVacation[]
  focusVisitId?: string
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
</style>
