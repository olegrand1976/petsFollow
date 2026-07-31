<template>
  <ProModal :open="open" :title="$t('calendar.newAppointment')" @update:open="onOpenChange">
    <form class="new-appt" data-testid="new-appointment-modal" @submit.prevent="submit(false)">
      <div class="pro-field">
        <label class="pro-label" for="new-appt-client">{{ $t('calendar.columnClient') }}</label>
        <select
          id="new-appt-client"
          v-model="clientId"
          class="pro-select"
          required
          data-testid="new-appt-client"
        >
          <option value="">{{ $t('calendar.selectClient') }}</option>
          <option v-for="c in clients" :key="c.userId" :value="c.userId">
            {{ c.displayName || c.email || c.userId }}
          </option>
        </select>
      </div>

      <div class="pro-field">
        <label class="pro-label" for="new-appt-pet">{{ $t('calendar.columnPet') }}</label>
        <select
          id="new-appt-pet"
          v-model="petId"
          class="pro-select"
          required
          :disabled="!clientId || petsLoading"
          data-testid="new-appt-pet"
        >
          <option value="">{{ $t('calendar.selectPet') }}</option>
          <option v-for="p in pets" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
      </div>

      <div class="pro-field">
        <label class="pro-label" for="new-appt-type">{{ $t('calendar.visitType') }}</label>
        <select
          id="new-appt-type"
          v-model="visitTypeId"
          class="pro-select"
          data-testid="new-appt-type"
        >
          <option value="">{{ $t('calendar.visitTypeNone') }}</option>
          <option v-for="vt in visitTypes" :key="vt.id" :value="vt.id">
            {{ vt.name }} ({{ vt.durationMinutes }} min)
          </option>
        </select>
      </div>

      <div class="pro-field">
        <label class="pro-label" for="new-appt-day">{{ $t('calendar.newAppointmentDay') }}</label>
        <input
          id="new-appt-day"
          v-model="day"
          type="date"
          class="pro-input"
          required
          data-testid="new-appt-day"
        >
      </div>

      <div v-if="day" class="new-appt__day" data-testid="new-appt-day-agenda">
        <h3 class="pro-section-title">{{ $t('calendar.dayAgenda') }}</h3>
        <p v-if="!dayVisits.length" class="text-muted">{{ $t('calendar.dayAgendaEmpty') }}</p>
        <ul v-else class="new-appt__day-list">
          <li
            v-for="item in dayAgenda"
            :key="item.key"
            class="new-appt__day-item"
            :class="{ 'new-appt__day-item--free': item.kind === 'free' }"
          >
            <span
              v-if="item.kind === 'visit' && item.color"
              class="new-appt__swatch"
              :style="{ background: item.color }"
            />
            <span class="new-appt__when">{{ item.label }}</span>
            <span v-if="item.title" class="new-appt__title">{{ item.title }}</span>
          </li>
        </ul>
      </div>

      <div class="new-appt__row">
        <div class="pro-field">
          <label class="pro-label" for="new-appt-time">{{ $t('calendar.newAppointmentTime') }}</label>
          <input
            id="new-appt-time"
            v-model="time"
            type="time"
            class="pro-input"
            required
            data-testid="new-appt-time"
          >
        </div>
        <div class="pro-field">
          <label class="pro-label" for="new-appt-duration">{{ $t('calendar.durationMinutes') }}</label>
          <input
            id="new-appt-duration"
            v-model.number="durationMinutes"
            type="number"
            class="pro-input"
            min="5"
            max="480"
            step="5"
            :disabled="!!visitTypeId"
            required
            data-testid="new-appt-duration"
          >
        </div>
      </div>
      <p v-if="visitTypeId" class="pro-settings-hint">{{ $t('calendar.durationFromType') }}</p>

      <div class="pro-field">
        <label class="pro-label" for="new-appt-notes">{{ $t('clients.pet.visitNotes') }}</label>
        <input id="new-appt-notes" v-model="notes" type="text" class="pro-input" data-testid="new-appt-notes">
      </div>

      <label class="pro-checkbox-row">
        <input v-model="requestPreconsult" type="checkbox" data-testid="new-appt-preconsult">
        <span>{{ $t('clients.pet.visitRequestPreconsult') }}</span>
      </label>

      <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>

      <div class="create-client-actions">
        <ProButton type="button" variant="ghost" :disabled="busy" @click="onOpenChange(false)">
          {{ $t('common.cancel') }}
        </ProButton>
        <ProButton type="submit" variant="secondary" :disabled="busy || !canSubmit" data-testid="new-appt-propose">
          {{ $t('clients.pet.visitPropose') }}
        </ProButton>
        <ProButton type="button" :loading="busy" :disabled="!canSubmit" data-testid="new-appt-confirm" @click="submit(true)">
          {{ $t('clients.pet.visitConfirmDirect') }}
        </ProButton>
      </div>
    </form>
  </ProModal>
</template>

<script setup lang="ts">
import type { CalendarVisit } from '~/composables/useCalendarGrid'

type ClientRow = { userId: string; displayName?: string; email?: string }
type PetRow = { id: string; name: string }
type VisitTypeRow = { id: string; name: string; durationMinutes: number; color: string }

const props = defineProps<{
  open: boolean
  visits: CalendarVisit[]
  defaultDay?: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  created: []
}>()

const { t } = useI18n()
const { mapError } = useApiError()
const { dayKey, visitDisplayAt, visitsByDay, startOfDay } = useCalendarGrid()

const clients = ref<ClientRow[]>([])
const pets = ref<PetRow[]>([])
const visitTypes = ref<VisitTypeRow[]>([])
const petsLoading = ref(false)
const busy = ref(false)
const error = ref('')

const clientId = ref('')
const petId = ref('')
const visitTypeId = ref('')
const day = ref('')
const time = ref('09:00')
const durationMinutes = ref(30)
const notes = ref('')
const requestPreconsult = ref(false)
const defaultDuration = ref(30)

const canSubmit = computed(() => !!clientId.value && !!petId.value && !!day.value && !!time.value)

const dayVisits = computed(() => {
  if (!day.value) return [] as CalendarVisit[]
  const map = visitsByDay(props.visits)
  return map.get(day.value) || []
})

type AgendaItem = {
  key: string
  kind: 'visit' | 'free'
  label: string
  title?: string
  color?: string
}

const dayAgenda = computed((): AgendaItem[] => {
  const items: AgendaItem[] = []
  const list = dayVisits.value
  for (let i = 0; i < list.length; i++) {
    const v = list[i]
    const start = visitDisplayAt(v)
    if (!start) continue
    const dur = v.durationMinutes || defaultDuration.value
    const end = new Date(start.getTime() + dur * 60_000)
    const pad = (n: number) => String(n).padStart(2, '0')
    const fmt = (d: Date) => `${pad(d.getHours())}:${pad(d.getMinutes())}`
    items.push({
      key: `v-${v.id}`,
      kind: 'visit',
      label: `${fmt(start)} – ${fmt(end)}`,
      title: `${v.petName || '—'} · ${v.clientName || '—'}${v.visitTypeName ? ` (${v.visitTypeName})` : ''}`,
      color: v.visitTypeColor || undefined,
    })
    const next = list[i + 1]
    const nextStart = next ? visitDisplayAt(next) : null
    if (nextStart && nextStart.getTime() > end.getTime()) {
      const gapMin = Math.round((nextStart.getTime() - end.getTime()) / 60_000)
      items.push({
        key: `f-${v.id}`,
        kind: 'free',
        label: `${fmt(end)} – ${fmt(nextStart)}`,
        title: t('calendar.freeSlot', { minutes: gapMin }),
      })
    }
  }
  return items
})

watch(visitTypeId, (id) => {
  if (!id) return
  const vt = visitTypes.value.find((x) => x.id === id)
  if (vt) durationMinutes.value = vt.durationMinutes
})

watch(clientId, async (id) => {
  petId.value = ''
  pets.value = []
  if (!id) return
  petsLoading.value = true
  try {
    const res: any = await $fetch(`/api/clients/${id}/pets`)
    const list = res.data ?? res ?? []
    pets.value = (Array.isArray(list) ? list : []).map((p: any) => ({
      id: p.id,
      name: p.name || p.id,
    }))
  } catch {
    pets.value = []
  } finally {
    petsLoading.value = false
  }
})

watch(
  () => props.open,
  async (isOpen) => {
    if (!isOpen) return
    error.value = ''
    notes.value = ''
    requestPreconsult.value = false
    visitTypeId.value = ''
    clientId.value = ''
    petId.value = ''
    day.value = props.defaultDay || dayKey(startOfDay(new Date()))
    time.value = '09:00'
    await loadMeta()
    durationMinutes.value = defaultDuration.value
  },
)

async function loadMeta() {
  try {
    const [clientsRes, typesRes, schedRes]: any[] = await Promise.all([
      $fetch('/api/clients'),
      $fetch('/api/vet/visit-types?active=1'),
      $fetch('/api/vet/schedule'),
    ])
    const cl = clientsRes.data ?? clientsRes ?? []
    clients.value = (Array.isArray(cl) ? cl : [])
      .filter((c: any) => c?.userId)
      .map((c: any) => ({
        userId: c.userId,
        displayName: c.displayName || c.fullName || '',
        email: c.email || '',
      }))
    const types = typesRes.data ?? typesRes ?? []
    visitTypes.value = (Array.isArray(types) ? types : [])
      .filter((vt: any) => vt?.id && vt.isActive !== false)
      .map((vt: any) => ({
        id: vt.id,
        name: vt.name,
        durationMinutes: vt.durationMinutes || 30,
        color: vt.color || '#2A9D8F',
      }))
    const sched = schedRes.data ?? schedRes
    defaultDuration.value = sched?.slotDurationMinutes || 30
  } catch (e: any) {
    error.value = mapError(e) || t('calendar.newAppointmentLoadFailed')
  }
}

function onOpenChange(value: boolean) {
  emit('update:open', value)
}

async function submit(confirmDirect: boolean) {
  if (!canSubmit.value || busy.value) return
  busy.value = true
  error.value = ''
  try {
    const scheduledAt = new Date(`${day.value}T${time.value}:00`)
    const body: Record<string, unknown> = {
      notes: notes.value,
      confirmDirect,
      requestPreconsult: requestPreconsult.value,
      scheduledAt: scheduledAt.toISOString(),
    }
    if (visitTypeId.value) {
      body.visitTypeId = visitTypeId.value
    } else {
      body.durationMinutes = Number(durationMinutes.value) || defaultDuration.value
    }
    await $fetch(`/api/pets/${petId.value}/visits`, { method: 'POST', body })
    emit('created')
    emit('update:open', false)
  } catch (e: any) {
    error.value = mapError(e) || t('calendar.newAppointmentFailed')
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
.new-appt {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.new-appt__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}
.new-appt__day {
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius-md, 8px);
  padding: 0.75rem;
  background: var(--pf-vet-bg);
  max-height: 14rem;
  overflow: auto;
}
.new-appt__day-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}
.new-appt__day-item {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  align-items: baseline;
  font-size: 0.85rem;
}
.new-appt__day-item--free {
  color: var(--pf-vet-muted, #64748b);
  font-style: italic;
}
.new-appt__swatch {
  width: 0.65rem;
  height: 0.65rem;
  border-radius: 2px;
  flex-shrink: 0;
}
.new-appt__when {
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
.create-client-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 0.5rem;
}
@media (max-width: 560px) {
  .new-appt__row {
    grid-template-columns: 1fr;
  }
}
</style>
