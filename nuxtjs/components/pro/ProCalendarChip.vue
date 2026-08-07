<template>
  <div
    class="cal-chip-wrap"
    :class="{ 'cal-chip-wrap--actions': showActions }"
  >
    <button
      :id="`calendar-chip-${visit.id}`"
      type="button"
      class="cal-chip"
      :class="[
        variant === 'row' ? 'cal-chip--row' : 'cal-chip--column',
        `cal-chip--${statusVariant(visit.status)}`,
        {
          'cal-chip--focus': focused,
          'cal-chip--walkin': !!visit.consultationSession,
          'cal-chip--typed': !!visit.visitTypeColor,
          'cal-chip--urgent': visit.preconsultAlert === 'urgent',
          'cal-chip--waiting': !!visit.waitingRoomAt,
        },
      ]"
      :style="visit.visitTypeColor ? { '--cal-type-color': visit.visitTypeColor } : undefined"
      :title="tooltip || undefined"
      :data-testid="`calendar-chip-${visit.id}`"
      @click="emit('select')"
    >
      <span class="cal-chip__time">{{ chipTime }}</span>
      <span v-if="visit.visitTypeName" class="cal-chip__type">{{ visit.visitTypeName }}</span>
      <span
        v-if="resourceMode === 'rooms' && visit.assigneeName"
        class="cal-chip__site"
      >{{ visit.assigneeName }}</span>
      <span
        v-else-if="resourceMode === 'people' && visit.roomName"
        class="cal-chip__site"
      >{{ visit.roomName }}</span>
      <span
        v-else-if="showSiteLabel && visit.siteName"
        class="cal-chip__site"
        data-testid="calendar-chip-site"
      >{{ visit.siteName }}</span>
      <span v-if="visit.consultationSession" class="cal-chip__walkin">{{ $t('calendar.walkInShort') }}</span>
      <span
        v-if="visit.waitingRoomAt"
        class="cal-chip__waiting"
        data-testid="calendar-chip-waiting-room"
        :title="$t('calendar.waitingRoomTooltip')"
      >{{ $t('calendar.waitingRoomTag') }}</span>
      <span
        v-if="visit.preconsultAlert === 'urgent'"
        class="cal-chip__urgent"
        data-testid="calendar-chip-preconsult-urgent"
        :title="$t('calendar.preconsultUrgentTooltip')"
      >{{ $t('calendar.preconsultUrgentShort') }}</span>
      <span
        v-else-if="visit.preconsultStatus === 'submitted'"
        class="cal-chip__preconsult"
        :class="{ 'cal-chip__preconsult--answered': variant === 'column' }"
        data-testid="calendar-chip-preconsult-answered"
        :title="$t('calendar.preconsultAnsweredTooltip')"
      >{{ $t('calendar.preconsultAnsweredTag') }}</span>
      <span
        v-else-if="visit.preconsultStatus === 'pending'"
        class="cal-chip__preconsult"
        :class="{ 'cal-chip__preconsult--sent': variant === 'column' }"
        data-testid="calendar-chip-preconsult-sent"
        :title="$t('calendar.preconsultSentTooltip')"
      >{{ $t('calendar.preconsultSentTag') }}</span>
      <span class="cal-chip__title">
        <template v-if="titleMode === 'pet'">{{ visit.petName || '—' }}</template>
        <template v-else>
          {{ visit.petName || '—' }} · {{ visit.clientName || '—' }}
          <span
            v-if="visit.isWalkinPlaceholder"
            class="cal-chip__walkin"
            data-testid="calendar-chip-walkin"
          >!</span>
        </template>
      </span>
      <span
        v-if="visit.callbackPhone"
        class="cal-chip__phone"
        data-testid="calendar-chip-phone"
        role="link"
        tabindex="0"
        :title="phoneTitle"
        @click.stop="onDial"
        @keydown.enter.stop="onDial"
      >{{ visit.callbackPhone }}</span>
      <span
        v-if="addressMode === 'full' && visit.addressText"
        class="cal-chip__place"
      >{{ visit.addressText }}</span>
      <span
        v-else-if="addressMode === 'badge' && visit.addressText"
        class="cal-chip__place"
        :title="visit.addressText"
      >{{ $t('calendar.placeBadge') }}</span>
    </button>
    <ProCalendarChipActions
      :visit="visit"
      :disabled="busy"
      @action="(action) => emit('action', action)"
    />
  </div>
</template>

<script setup lang="ts">
import {
  calendarChipTooltip,
  canShowCalendarChipDeskActions,
  dialCallbackPhone,
  type CalendarDeskAction,
  type CalendarVisit,
} from '~/composables/useCalendarGrid'

const props = withDefaults(
  defineProps<{
    visit: CalendarVisit
    variant?: 'column' | 'row'
    titleMode?: 'pet-client' | 'pet'
    addressMode?: 'none' | 'full' | 'badge'
    focused?: boolean
    busy?: boolean
    showSiteLabel?: boolean
    resourceMode?: 'people' | 'rooms' | null
  }>(),
  {
    variant: 'column',
    titleMode: 'pet-client',
    addressMode: 'none',
    focused: false,
    busy: false,
    showSiteLabel: false,
    resourceMode: null,
  },
)

const emit = defineEmits<{
  select: []
  action: [CalendarDeskAction]
}>()

const { t } = useI18n()
const { statusVariant, visitDisplayAt, isUnscheduled } = useCalendarGrid()

const showActions = computed(() => canShowCalendarChipDeskActions(props.visit))

const tooltip = computed(() => calendarChipTooltip(props.visit, t))

const phoneTitle = computed(() =>
  props.titleMode === 'pet'
    ? (props.visit.callbackPhone || '')
    : t('clients.walkin.callbackPhone'),
)

const chipTime = computed(() => {
  if (isUnscheduled(props.visit)) return t('calendar.unscheduled')
  const at = visitDisplayAt(props.visit)
  if (!at) return '—'
  const pad = (n: number) => String(n).padStart(2, '0')
  const time = `${pad(at.getHours())}:${pad(at.getMinutes())}`
  if (props.visit.proposedScheduledAt) return `${time} (${t('calendar.proposed')})`
  return time
})

function onDial() {
  const phone = props.visit.callbackPhone?.trim()
  if (phone) dialCallbackPhone(phone)
}
</script>
